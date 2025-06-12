package handlers

import (
	"context"
	"encoding/json"
	"slices"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/interceptors"
)

func (s *ExamServer) GetParticipantAnswers(ctx context.Context, req *proto.ParticipantRequest) (*proto.AnswerList, error) {
	exam, ok := interceptors.GetExamFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam not found in context")
	}

	type QueryResult struct {
		ID                int64            `gorm:"primaryKey;type:bigint"`
		ExamParticipantID int64            `gorm:"type:bigint"`
		Answer            *json.RawMessage `gorm:"type:json"`

		QuestionID int64 `gorm:"type:bigint"`
		Order      int16 `gorm:"column:order"`
		CategoryID int64 `gorm:"column:category_id"`
		Type       int32 `gorm:"column:type"`
		MaxScore   int16 `gorm:"column:max_score"`
	}

	var results []QueryResult
	err := db.DB.Table("exam_questions").
		Select("exam_questions.question_id, exam_questions.order, exam_questions.category_id, exam_questions.type, exam_questions.max_score", "answers.id, answers.answer, answers.exam_participant_id").
		Joins("LEFT JOIN answers ON exam_questions.question_id = answers.question_id AND answers.exam_participant_id = ?", req.ParticipantId).
		Where("exam_questions.exam_id = ?", exam.ID).
		Find(&results).Error

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch answers")
	}

	if len(results) == 0 {
		return nil, status.Error(codes.NotFound, "no questions found")
	}

	response := &proto.AnswerList{
		Answers: make([]*proto.AnswerResponse, len(results)),
	}

	for i, result := range results {
		response.Answers[i] = &proto.AnswerResponse{
			AnswerId:          result.ID, // Will be 0 if question is unanswered
			ExamParticipantId: result.ExamParticipantID,
			QuestionId:        result.QuestionID,
			Order:             int32(result.Order),
			CategoryId:        result.CategoryID,
			QuestionType:      result.Type,
			MaxScore:          int32(result.MaxScore),
		}
		if result.Answer != nil {
			response.Answers[i].Answer = *result.Answer
		}
	}

	return response, nil
}

// GetAnswerForExam finds an answer using participant ID and question ID and returns minimal info
func (s *ExamServer) GetAnswerForExam(ctx context.Context, req *proto.GetAnswerRequest) (*proto.AnswerMinimalResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var participant models.ExamParticipant
	if err := db.DB.Where("exam_id = ? AND user_id = ?", req.ExamId, userID).Take(&participant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "participant not found")
		}
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	var answer models.Answer
	err = db.DB.Where("exam_participant_id = ? AND question_id = ? AND answer IS NOT NULL",
		participant.ID, req.QuestionId).Take(&answer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &proto.AnswerMinimalResponse{
				QuestionId: req.QuestionId,
			}, nil
		}
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	return &proto.AnswerMinimalResponse{
		AnswerId:   int64(answer.ID),
		Answer:     *answer.Answer,
		QuestionId: int64(answer.QuestionID),
	}, nil
}

func (s *ExamServer) UpsertAnswer(ctx context.Context, req *proto.UpsertAnswersRequest) (*proto.UpsertAnswersResponse, error) {
	// Should be added to context by EndExamInterceptor
	participant, ok := interceptors.GetParticipantFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "participant not found in context")
	}

	if participant.Status != constants.PARTICIPANT_STATUS_STARTED {
		return nil, status.Error(codes.FailedPrecondition, "participant has not started the exam")
	}

	if participant.Status == constants.PARTICIPANT_STATUS_ENDED {
		return nil, status.Error(codes.FailedPrecondition, "participant has ended the exam")
	}

	// Check participant status
	if !slices.Contains([]int16{constants.PARTICIPANT_STATUS_STARTED}, participant.Status) {
		return nil, status.Error(codes.FailedPrecondition, "participant must be in STARTED state")
	}

	typedQuestionId := types.QuestionID(req.Answer.QuestionId)

	var examQuestion models.ExamQuestion
	if err := db.DB.Model(&models.ExamQuestion{}).
		Select("type").
		Where("exam_id = ? AND question_id = ?", req.ExamId, typedQuestionId).
		Find(&examQuestion).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch question")
	}
	if err := validateAnswerJSON(req.Answer.Answer, examQuestion.Type); err != nil {
		return nil, err
	}

	// Convert answer bytes to *json.RawMessage
	// go-staticcheck: Should omit nil check; len() for []byte is defined as zero (S1009)
	var answerContent *json.RawMessage
	if len(req.Answer.Answer) > 0 {
		raw := json.RawMessage(req.Answer.Answer)
		answerContent = &raw
	}

	var answer models.Answer
	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		err := tx.Where("exam_participant_id = ? AND question_id = ?",
			participant.ID, typedQuestionId).Take(&answer).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new answer
				answer = models.Answer{
					ExamParticipantID: participant.ID,
					QuestionID:        typedQuestionId,
					Answer:            answerContent,
				}
				return tx.Create(&answer).Error
			}
			return err
		}

		// Update existing answer
		answer.Answer = answerContent
		return tx.Save(&answer).Error
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to upsert answer")
	}

	response := &proto.UpsertAnswersResponse{
		AnswerId:   int64(answer.ID),
		QuestionId: req.Answer.QuestionId,
		Answer:     nil,
	}

	if answer.Answer != nil {
		response.Answer = *answer.Answer
	}

	return response, nil
}
