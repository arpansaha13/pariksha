package services

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
	"pariksha/exam/internal/interceptors"
	"pariksha/exam/internal/interservice"
	"pariksha/exam/internal/repositories"
)

type Answer struct {
	answerRepo      *repositories.Answer
	questionRepo    *repositories.Question
	participantRepo *repositories.Participant
}

func NewAnswer(answerRepo *repositories.Answer, questionRepo *repositories.Question, participantRepo *repositories.Participant) *Answer {
	return &Answer{
		answerRepo:      answerRepo,
		questionRepo:    questionRepo,
		participantRepo: participantRepo,
	}
}

// GetParticipantAnswers retrieves all answers for a participant in an exam.
func (s *Answer) GetParticipantAnswers(ctx context.Context, req *proto.ParticipantRequest) (*proto.AnswerList, error) {
	exam, ok := interceptors.GetExamFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam not found in context")
	}

	// Fetch answers and question info
	results, err := s.answerRepo.GetParticipantAnswers(nil, types.ExamID(exam.ID), types.ParticipantID(req.ParticipantId))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch answers")
	}
	if len(results) == 0 {
		return nil, status.Error(codes.NotFound, "no questions found")
	}

	// Collect all question IDs
	questionIDs := make([]types.QuestionID, len(results))
	for i, result := range results {
		questionIDs[i] = result.QuestionID
	}

	// Fetch question content from paper service
	questions, err := interservice.GetQuestionsByIDs(questionIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch questions")
	}

	response := &proto.AnswerList{
		Answers: make([]*proto.AnswerResponse, len(results)),
	}

	for i, result := range results {
		response.Answers[i] = &proto.AnswerResponse{
			AnswerId:          int64(result.ID),
			ExamParticipantId: int64(result.ExamParticipantID),
			Order:             int32(result.Order),
			CategoryId:        int64(result.CategoryID),
			QuestionType:      questions[i].Type,
			MaxScore:          int32(result.MaxScore),
			QuestionHash:      questions[i].Hash,
			Question:          questions[i].RawQuestion,
		}
		if result.Answer != nil {
			response.Answers[i].Answer = *result.Answer
		}
	}

	return response, nil
}

// GetAnswerForExam finds an answer using participant ID and question ID and returns minimal info.
func (s *Answer) GetAnswerForExam(ctx context.Context, req *proto.GetAnswerRequest) (*proto.AnswerMinimalResponse, error) {
	questionID, ok := interceptors.GetQuestionIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "question ID not found in context")
	}

	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// Use repository to fetch participant
	participant, err := s.participantRepo.GetByExamHashAndUserID(nil, req.ExamHash, types.UserID(userID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "participant not found")
		}
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	// Use repository to fetch answer
	answer, err := s.answerRepo.GetAnswerByParticipantAndQuestion(nil, participant.ID, questionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &proto.AnswerMinimalResponse{
				QuestionHash: req.QuestionHash,
			}, nil
		}
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	return &proto.AnswerMinimalResponse{
		AnswerId:     int64(answer.ID),
		Answer:       *answer.Answer,
		QuestionHash: req.QuestionHash,
	}, nil
}

// UpsertAnswer creates or updates an answer for a participant and question.
func (s *Answer) UpsertAnswer(ctx context.Context, req *proto.UpsertAnswersRequest) (*proto.UpsertAnswersResponse, error) {
	questionID, ok := interceptors.GetQuestionIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "question ID not found in context")
	}

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

	questionTypes, err := interservice.GetQuestionTypesByIds([]types.QuestionID{questionID})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch question")
	}

	if err := validateAnswerJSON(req.Answer.Answer, questionTypes[0]); err != nil {
		return nil, err
	}

	// Convert answer bytes to *json.RawMessage
	var answerContent *json.RawMessage
	if len(req.Answer.Answer) > 0 {
		raw := json.RawMessage(req.Answer.Answer)
		answerContent = &raw
	}

	var answer *models.Answer
	// Use repository transaction and upsert method
	err = s.answerRepo.Transaction(func(tx *gorm.DB) error {
		var txAnswer *models.Answer
		var err error
		txAnswer, err = s.answerRepo.UpsertAnswer(tx, participant.ID, questionID, answerContent)
		if err != nil {
			return err
		}
		answer = txAnswer
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to upsert answer")
	}

	// Convert question ID to hash
	questionHashes, err := interservice.GetQuestionHashesByIds([]types.QuestionID{answer.QuestionID})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch question hash")
	}

	response := &proto.UpsertAnswersResponse{
		AnswerId:     int64(answer.ID),
		QuestionHash: questionHashes[0],
		Answer:       nil,
	}

	if answer.Answer != nil {
		response.Answer = *answer.Answer
	}

	return response, nil
}
