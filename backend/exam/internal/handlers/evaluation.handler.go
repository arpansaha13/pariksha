package handlers

import (
	"context"
	"database/sql"
	"slices"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
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
		return nil, utils.HandleDBError(err, "answer not found")
	}

	if req.NewScore != nil && req.GetNewScore() > int32(maxScore) {
		return nil, status.Error(codes.InvalidArgument, "new score exceeds max score for the question")
	}

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		answer, err := utils.FindRecord[models.Answer](tx, req.AnswerId, "answer not found")
		if err != nil {
			return err
		}

		if req.NewScore != nil {
			participant, err := utils.FindRecord[models.ExamParticipant](tx, answer.ExamParticipantID, "exam participant not found")
			if err != nil {
				return err
			}

			participant.ScoreAwarded = participant.ScoreAwarded - int32(answer.ScoreAwarded) + req.GetNewScore()
			answer.ScoreAwarded = int16(req.GetNewScore())

			if err := tx.Save(participant).Error; err != nil {
				return status.Error(codes.Internal, "failed to update exam participant")
			}
		}

		if req.Evaluated != nil {
			answer.Evaluated = *req.Evaluated
		}
		if req.Comments != nil {
			answer.Comments = sql.NullString{String: *req.Comments, Valid: true}
		}

		return tx.Save(answer).Error
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

func (s *ExamServer) GetAnswerForEvaluation(ctx context.Context, req *proto.GetAnswerForEvaluationRequest) (*proto.GetAnswerForEvaluationResponse, error) {
	// First check if participant exists and has ended the exam
	participant, err := utils.FindRecord[models.ExamParticipant](db.DB, req.ParticipantId, "participant not found")
	if err != nil {
		return nil, err
	}

	if !slices.Contains([]int16{
		constants.PARTICIPANT_STATUS_ENDED,
		constants.PARTICIPANT_STATUS_EVALUATED,
	}, participant.Status) {
		return nil, status.Error(codes.FailedPrecondition, "cannot evaluate answers before exam completion")
	}

	var answer models.Answer
	if err := db.DB.Where("exam_participant_id = ? AND question_id = ? AND answer IS NOT NULL",
		req.ParticipantId, req.QuestionId).Take(&answer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &proto.GetAnswerForEvaluationResponse{
				QuestionId: req.QuestionId,
			}, nil
		}
		return nil, utils.HandleDBError(err, "answer not found")
	}

	return &proto.GetAnswerForEvaluationResponse{
		Id:           answer.ID,
		QuestionId:   answer.QuestionID,
		ScoreAwarded: int32(answer.ScoreAwarded),
		Comments:     answer.Comments.String,
	}, nil
}
