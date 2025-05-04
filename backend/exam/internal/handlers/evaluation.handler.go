package handlers

import (
	"context"
	"database/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/db"
)

func (s *ExamServer) UpdateAnswerForEvaluation(ctx context.Context, req *proto.UpdateAnswerRequest) (*proto.Empty, error) {
	// Get max score from exam_questions using joins
	var maxScore int16
	if err := db.DB.Table("answers").
		Joins("INNER JOIN exam_participants ON exam_participants.id = answers.exam_participant_id").
		Joins("INNER JOIN exam_questions ON exam_questions.exam_id = exam_participants.exam_id AND exam_questions.question_id = answers.question_id").
		Where("answers.id = ?", req.AnswerId).
		Select("exam_questions.max_score").
		Take(&maxScore).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "answer not found")
		}
		return nil, status.Error(codes.Internal, "failed to get max score")
	}

	if req.NewScore != nil && req.GetNewScore() > int32(maxScore) {
		return nil, status.Error(codes.InvalidArgument, "new score exceeds max score for the question")
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var answer models.Answer
		if err := tx.Take(&answer, req.AnswerId).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "answer not found")
			}
			return status.Error(codes.Internal, "database error")
		}

		isUpdated := false

		if req.NewScore != nil {
			var examParticipant models.ExamParticipant
			if err := tx.Take(&examParticipant, answer.ExamParticipantID).Error; err != nil {
				return status.Error(codes.NotFound, "exam participant not found")
			}

			examParticipant.ScoreAwarded = examParticipant.ScoreAwarded - answer.ScoreAwarded + int(*req.NewScore)
			answer.ScoreAwarded = int(*req.NewScore)

			if err := tx.Save(&examParticipant).Error; err != nil {
				return status.Error(codes.Internal, "failed to update exam participant")
			}

			isUpdated = true
		}

		if req.Evaluated != nil {
			answer.Evaluated = *req.Evaluated
			isUpdated = true
		}

		if req.Comments != nil {
			answer.Comments = sql.NullString{String: *req.Comments, Valid: true}
			isUpdated = true
		}

		if isUpdated {
			if err := tx.Save(&answer).Error; err != nil {
				return status.Error(codes.Internal, "failed to update answer")
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

func (s *ExamServer) MarkParticipantAsEvaluated(ctx context.Context, req *proto.ParticipantRequest) (*proto.EvaluationStatusResponse, error) {
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var examParticipant models.ExamParticipant
		if err := tx.Take(&examParticipant, req.ParticipantId).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "exam participant not found")
			}
			return status.Error(codes.Internal, "database error")
		}

		if examParticipant.Status != constants.PARTICIPANT_STATUS_ENDED {
			return status.Error(codes.FailedPrecondition, "evaluation can only start if the exam has ended")
		}

		var unevaluatedCount int64
		if err := tx.Model(&models.Answer{}).Where("exam_participant_id = ? AND evaluated = ?", req.ParticipantId, false).Count(&unevaluatedCount).Error; err != nil {
			return status.Error(codes.Internal, "failed to count unevaluated answers")
		}

		if unevaluatedCount == 0 {
			examParticipant.Status = constants.PARTICIPANT_STATUS_EVALUATED
			if err := tx.Save(&examParticipant).Error; err != nil {
				return status.Error(codes.Internal, "failed to update exam participant status")
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Get final unevaluated count
	var unevaluatedCount int64
	if err := db.DB.Model(&models.Answer{}).Where("exam_participant_id = ? AND evaluated = ?", req.ParticipantId, false).Count(&unevaluatedCount).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to count unevaluated answers")
	}

	return &proto.EvaluationStatusResponse{
		UnevaluatedCount: int32(unevaluatedCount),
	}, nil
}
