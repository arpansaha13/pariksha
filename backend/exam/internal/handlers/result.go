package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/interceptors"
)

// GetExamResults retrieves all answers for a participant in an exam
func (s *ExamServer) GetExamResults(ctx context.Context, req *proto.ExamRequest) (*proto.ExamResultsResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	examID, ok := interceptors.GetExamIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "exam ID not found in context")
	}

	// Get participant ID for this user in this exam
	var participant models.ExamParticipant
	if err := db.DB.Where("exam_id = ? AND user_id = ?", examID, userID).
		Take(&participant).Error; err != nil {
		return nil, utils.HandleDBError(err, "participant not found")
	}

	// Get all answers for this participant
	var answers []models.Answer
	if err := db.DB.Where("exam_participant_id = ?", participant.ID).
		Find(&answers).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch answers")
	}

	// Build response items from answers
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
