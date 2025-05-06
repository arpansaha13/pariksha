package handlers

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
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
		response.Answers[i] = &proto.AnswerResponse{
			Id:                answer.ID,
			ExamParticipantId: answer.ExamParticipantID,
			QuestionId:        answer.QuestionID,
			Answer:            *answer.Answer,
			Comments:          answer.Comments.String,
			ScoreAwarded:      int32(answer.ScoreAwarded),
		}
	}

	return response, nil
}

// GetAnswerForExam finds an answer using participant ID and question ID and returns minimal info
func (s *ExamServer) GetAnswerForExam(ctx context.Context, req *proto.GetAnswerRequest) (*proto.GetAnswerResponse, error) {
	participant, ok := interceptors.GetParticipantFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "participant not found in context")
	}

	var answer models.Answer
	if err := db.DB.Where("exam_participant_id = ? AND question_id = ? AND answer IS NOT NULL",
		participant.ID, req.QuestionId).Take(&answer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &proto.GetAnswerResponse{
				QuestionId: req.QuestionId,
			}, nil
		}
		return nil, status.Error(codes.Internal, "database error")
	}

	return &proto.GetAnswerResponse{
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
	questionType := questionTypes[0]

	// Convert answer bytes to *json.RawMessage
	var answerContent *json.RawMessage
	if req.Answer.Answer != nil && len(req.Answer.Answer) > 0 {
		// Validate answer JSON based on question type
		if err := validateAnswerJSON(req.Answer.Answer, questionType); err != nil {
			return nil, err
		}
		raw := json.RawMessage(req.Answer.Answer)
		answerContent = &raw
	}

	var answer models.Answer
	if err := db.DB.Where("exam_participant_id = ? AND question_id = ?", participant.ID, req.Answer.QuestionId).Take(&answer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new answer
			answer = models.Answer{
				ExamParticipantID: participant.ID,
				QuestionID:        req.Answer.QuestionId,
				Answer:            answerContent,
			}
			if err := db.DB.Create(&answer).Error; err != nil {
				return nil, status.Error(codes.Internal, "failed to create answer")
			}
		} else {
			return nil, status.Error(codes.Internal, "database error")
		}
	} else {
		// Update existing answer
		answer.Answer = answerContent
		if err := db.DB.Save(&answer).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to update answer")
		}
	}

	return &proto.UpsertAnswersResponse{
		AnswerId: answer.ID,
	}, nil
}
