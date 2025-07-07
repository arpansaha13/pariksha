package services

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/exam/internal/interservice"
	"pariksha/exam/internal/repositories"
)

type Question struct {
	questionRepo *repositories.Question
}

func NewQuestion(questionRepo *repositories.Question) *Question {
	return &Question{questionRepo: questionRepo}
}

// GetExamQuestions retrieves all questions associated with an exam
func (s *Question) GetExamQuestions(req *proto.ExamRequest) (*proto.ExamQuestionsResponse, error) {
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
	questionHashes, err := interservice.GetQuestionHashesByIds(questionIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch question hashes")
	}

	questionTypes, err := interservice.GetQuestionTypesByIds(questionIDs)
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
func (s *Question) GetExamQuestion(req *proto.ExamQuestionRequest) (*proto.ExamQuestionResponse, error) {
	// Fetch question hashes from paper service
	question, err := interservice.GetQuestionByHash(req.QuestionHash)
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
