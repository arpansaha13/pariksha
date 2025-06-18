package services

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/interservice/paper"
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
	questionIDs := make([]int64, len(examQuestions))
	for i, eq := range examQuestions {
		questionIDs[i] = int64(eq.QuestionID)
	}

	// Fetch question hashes from paper service
	questionHashes, err := paper.FetchQuestionHashesForIds(questionIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch question hashes")
	}

	questions := make([]*proto.ExamQuestion, len(examQuestions))
	for i, eq := range examQuestions {
		questions[i] = &proto.ExamQuestion{
			QuestionHash: questionHashes[i],
			CategoryId:   int64(eq.CategoryID),
			Order:        int32(eq.Order),
			MaxScore:     int32(eq.MaxScore),
			Type:         eq.Type,
		}
	}

	return &proto.ExamQuestionsResponse{
		Questions: questions,
	}, nil
}
