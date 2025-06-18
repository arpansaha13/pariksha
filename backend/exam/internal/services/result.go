package services

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/repositories"
)

type Result struct {
	participantRepo *repositories.Participant
	answerRepo      *repositories.Answer
}

func NewResult(participantRepo *repositories.Participant, answerRepo *repositories.Answer) *Result {
	return &Result{
		participantRepo: participantRepo,
		answerRepo:      answerRepo,
	}
}

// GetExamResults retrieves all answers for a participant in an exam.
func (s *Result) GetExamResults(ctx context.Context, req *proto.ExamRequest) (*proto.ExamResultsResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	participant, err := s.participantRepo.GetByExamHashAndUserID(nil, req.ExamHash, types.UserID(userID))
	if err != nil {
		return nil, utils.HandleDBError(err, "participant not found")
	}

	answers, err := s.answerRepo.GetAllByParticipantID(nil, participant.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch answers")
	}

	results := make([]*proto.ExamResultItem, len(answers))
	for i, answer := range answers {
		results[i] = &proto.ExamResultItem{
			AnswerId:     int64(answer.ID),
			ScoreAwarded: int32(answer.ScoreAwarded),
			Comments:     answer.Comments.String,
		}
	}

	return &proto.ExamResultsResponse{
		Results: results,
	}, nil
}
