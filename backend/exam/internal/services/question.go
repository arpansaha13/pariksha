package services

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/exam/internal/interservice"
	"pariksha/exam/internal/repositories"
)

type Question struct {
	questionRepo   *repositories.Question
	questionIntSvc *interservice.Question
}

func NewQuestion(
	questionRepo *repositories.Question,
	questionIntSvc *interservice.Question,
) *Question {
	return &Question{
		questionRepo:   questionRepo,
		questionIntSvc: questionIntSvc,
	}
}

// GetExamQuestions retrieves all questions associated with an exam
func (s *Question) GetExamQuestions(ctx context.Context, req *proto.ExamRequest) (*proto.ExamQuestionsResponse, error) {
	examQuestions, err := s.questionRepo.GetExamQuestions(nil, req.ExamHash)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch questions")
	}

	// Get question IDs for hash lookup
	questionIDs := make([]types.QuestionID, len(examQuestions))
	for i, eq := range examQuestions {
		questionIDs[i] = types.QuestionID(eq.QuestionID)
	}

	// Fetch question hashes from paper service
	questionHashes, err := s.questionIntSvc.GetQuestionHashesByIds(ctx, questionIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch question hashes")
	}

	questionTypes, err := s.questionIntSvc.GetQuestionTypesByIds(ctx, questionIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch question")
	}

	questions := make([]*proto.ExamQuestion, len(examQuestions))
	for i, eq := range examQuestions {
		questions[i] = &proto.ExamQuestion{
			QuestionHash: questionHashes[i],
			CategoryId:   int64(eq.CategoryID),
			Order:        int32(eq.Order),
			MaxScore:     int32(eq.MaxScore),
			Type:         questionTypes[i],
		}
	}

	return &proto.ExamQuestionsResponse{
		Questions: questions,
	}, nil
}

// GetExamQuestion retrieves the question specified by the hash
func (s *Question) GetExamQuestion(ctx context.Context, req *proto.ExamQuestionRequest) (*proto.ExamQuestionResponse, error) {
	// Fetch question hashes from paper service
	question, err := s.questionIntSvc.GetQuestionByHash(ctx, req.QuestionHash)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch question hashes")
	}

	return &proto.ExamQuestionResponse{
		QuestionHash: question.Hash,
		Type:         question.Type,
		RawQuestion:  question.RawQuestion,
		TestCases:    question.TestCases,
	}, nil
}
