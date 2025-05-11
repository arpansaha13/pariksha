package handlers

import (
	"context"
	"encoding/json"
	"slices"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/interceptors"
)

func (s *ExamServer) GetParticipantAnswers(ctx context.Context, req *proto.ParticipantRequest) (*proto.AnswerList, error) {
	var answers []models.Answer
	if err := db.DB.Where("exam_participant_id = ?", req.ParticipantId).Find(&answers).Error; err != nil {
		return nil, status.Error(codes.NotFound, "answers not found")
	}

	if len(answers) == 0 {
		return nil, status.Error(codes.NotFound, "answers not found")
	}

	response := &proto.AnswerList{
		Answers: make([]*proto.AnswerResponse, len(answers)),
	}

	for i, answer := range answers {
		response.Answers[i] = answerToProto(answer)
	}

	return response, nil
}

// GetAnswerForExam finds an answer using participant ID and question ID and returns minimal info
func (s *ExamServer) GetAnswerForExam(ctx context.Context, req *proto.GetAnswerRequest) (*proto.AnswerMinimalResponse, error) {
	var participant models.ExamParticipant
	if err := db.DB.Where("exam_id = ?", req.ExamId).Take(&participant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "participant not found")
		}
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	var answer models.Answer
	err := db.DB.Where("exam_participant_id = ? AND question_id = ? AND answer IS NOT NULL",
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
		Id:         answer.ID,
		Answer:     *answer.Answer,
		QuestionId: answer.QuestionID,
	}, nil
}

func (s *ExamServer) UpsertAnswer(ctx context.Context, req *proto.UpsertAnswersRequest) (*proto.UpsertAnswersResponse, error) {
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

	// Get question type from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "missing metadata")
	}

	questionTypes := md.Get("question_type")
	if len(questionTypes) == 0 {
		return nil, status.Error(codes.Internal, "missing question type in metadata")
	}

	// Check participant status
	if !slices.Contains([]int16{constants.PARTICIPANT_STATUS_STARTED}, participant.Status) {
		return nil, status.Error(codes.FailedPrecondition, "participant must be in STARTED state")
	}

	if err := validateAnswerJSON(req.Answer.Answer, questionTypes[0]); err != nil {
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
			participant.ID, req.Answer.QuestionId).Take(&answer).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new answer
				answer = models.Answer{
					ExamParticipantID: participant.ID,
					QuestionID:        req.Answer.QuestionId,
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

	return &proto.UpsertAnswersResponse{
		AnswerId: answer.ID,
	}, nil
}
