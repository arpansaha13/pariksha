package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/config/db"
)

// GetExamResults retrieves all questions and corresponding answers for a participant in an exam
func (s *ExamServer) GetExamResults(ctx context.Context, req *proto.ExamRequest) (*proto.ExamResultsResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// Get participant ID for this user in this exam
	var participant models.ExamParticipant
	if err := db.DB.Where("exam_id = ? AND user_id = ?", req.ExamId, userID).
		Take(&participant).Error; err != nil {
		return nil, utils.HandleDBError(err, "participant not found")
	}

	// Return empty array if participant is not evaluated
	if participant.Status != constants.PARTICIPANT_STATUS_EVALUATED {
		return &proto.ExamResultsResponse{
			Results: []*proto.ExamResultItem{},
		}, nil
	}

	// Get all questions for this exam
	var examQuestions []models.ExamQuestion
	if err := db.DB.Where("exam_id = ?", req.ExamId).
		Order("\"order\"").
		Find(&examQuestions).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch exam questions")
	}

	// Get all answers for this participant
	var answers []models.Answer
	if err := db.DB.Where("exam_participant_id = ?", participant.ID).
		Find(&answers).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch answers")
	}

	// Create a map of questionID -> answer for efficient lookup
	answerMap := make(map[int64]models.Answer)
	for _, answer := range answers {
		answerMap[answer.QuestionID] = answer
	}

	// Build response items by combining questions with answers
	results := make([]*proto.ExamResultItem, len(examQuestions))
	for i, question := range examQuestions {
		var answerBytes []byte
		var scoreAwarded int32
		var comments string

		// If answer exists for this question, include it
		if answer, exists := answerMap[question.QuestionID]; exists {
			if answer.Answer != nil {
				answerBytes = *answer.Answer
			}
			scoreAwarded = int32(answer.ScoreAwarded)
			comments = answer.Comments.String
		}

		results[i] = &proto.ExamResultItem{
			QuestionId:   question.QuestionID,
			Order:        int32(question.Order),
			CategoryId:   question.CategoryID,
			Answer:       answerBytes,
			ScoreAwarded: scoreAwarded,
			Comments:     comments,
			MaxScore:     int32(question.MaxScore),
		}
	}

	return &proto.ExamResultsResponse{
		Results: results,
	}, nil
}
