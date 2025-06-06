package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
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

// GetAnswerForEvaluation retrieves an answer for evaluation purposes
func (s *ExamServer) GetAnswerForEvaluation(ctx context.Context, req *proto.ParticipantQuestionRequest) (*proto.AnswerMinimalResponse, error) {
	var answer struct {
		ID         int64
		QuestionID int64
		Answer     *json.RawMessage
	}

	err := db.DB.Model(&models.Answer{}).
		Select("id", "question_id", "answer").
		Where("exam_participant_id = ? AND question_id = ?",
			req.ParticipantId,
			req.QuestionId,
		).Take(&answer).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &proto.AnswerMinimalResponse{
				QuestionId: req.QuestionId,
			}, nil
		}
		return nil, status.Error(codes.Internal, "failed to fetch answer")
	}

	var answerBytes []byte
	if answer.Answer != nil {
		answerBytes = *answer.Answer
	}

	return &proto.AnswerMinimalResponse{
		Id:         answer.ID,
		Answer:     answerBytes,
		QuestionId: answer.QuestionID,
	}, nil
}

func (s *ExamServer) GetAnswerEvaluationData(ctx context.Context, req *proto.ParticipantQuestionRequest) (*proto.GetAnswerEvaluationDataResponse, error) {
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

	var answer struct {
		ID           int64
		QuestionID   int64
		ScoreAwarded int16
		Comments     sql.NullString
	}

	if err := db.DB.Model(&models.Answer{}).
		Select("id", "question_id", "score_awarded", "comments").
		Where("exam_participant_id = ? AND question_id = ? AND answer IS NOT NULL",
			req.ParticipantId, req.QuestionId).Take(&answer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &proto.GetAnswerEvaluationDataResponse{
				QuestionId: req.QuestionId,
			}, nil
		}
		return nil, utils.HandleDBError(err, "answer not found")
	}

	return &proto.GetAnswerEvaluationDataResponse{
		Id:           answer.ID,
		QuestionId:   answer.QuestionID,
		ScoreAwarded: int32(answer.ScoreAwarded),
		Comments:     answer.Comments.String,
	}, nil
}

func (s *ExamServer) UpdateAnswerForEvaluation(ctx context.Context, req *proto.UpdateAnswerRequest) (*proto.GetAnswerEvaluationDataResponse, error) {
	// Get max score and participant_status using joins
	type QueryResult struct {
		MaxScore          int16 `gorm:"column:max_score"`
		ParticipantStatus int16 `gorm:"column:status"`
	}

	queryResult := QueryResult{}
	if err := db.DB.Table("answers").
		Joins("INNER JOIN exam_participants ON exam_participants.id = answers.exam_participant_id").
		Joins("INNER JOIN exam_questions ON exam_questions.exam_id = exam_participants.exam_id AND exam_questions.question_id = answers.question_id").
		Where("answers.id = ?", req.AnswerId).
		Select("exam_questions.max_score", "exam_participants.status").
		Take(&queryResult).Error; err != nil {
		return nil, utils.HandleDBError(err, "answer not found")
	}

	if queryResult.ParticipantStatus != constants.PARTICIPANT_STATUS_ENDED {
		return nil, status.Error(codes.FailedPrecondition, "participant cannot be evaluated")
	}

	if req.NewScore != nil && req.GetNewScore() > int32(queryResult.MaxScore) {
		return nil, status.Error(codes.InvalidArgument, "new score exceeds max score for the question")
	}

	var answer *models.Answer
	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		var err error
		answer, err = utils.FindRecord[models.Answer](tx, req.AnswerId, "answer not found")
		if err != nil {
			return err
		}

		if req.NewScore != nil {
			// Use atomic update query to prevent race conditions
			if err := tx.Exec(
				`UPDATE exam_participants 
				SET score_awarded = score_awarded - ? + ?
				WHERE id = ?`,
				answer.ScoreAwarded, req.GetNewScore(), answer.ExamParticipantID,
			).Error; err != nil {
				return status.Error(codes.Internal, "failed to update exam participant")
			}

			answer.ScoreAwarded = int16(req.GetNewScore())
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

	return &proto.GetAnswerEvaluationDataResponse{
		Id:           int64(answer.ID),
		QuestionId:   int64(answer.QuestionID),
		ScoreAwarded: int32(answer.ScoreAwarded),
		Comments:     answer.Comments.String,
	}, nil
}

func (s *ExamServer) MarkParticipantAsEvaluated(ctx context.Context, req *proto.ParticipantRequest) (*proto.EvaluationStatusResponse, error) {
	var unevaluatedCount int64

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Find and validate participant exists
		participant, err := utils.FindRecord[models.ExamParticipant](tx, req.ParticipantId, "exam participant not found")
		if err != nil {
			return err
		}

		// Ensure participant has finished the exam before evaluation
		if participant.Status != constants.PARTICIPANT_STATUS_ENDED {
			return status.Error(codes.FailedPrecondition, "evaluation can only start if the exam has ended")
		}

		// Count answers that are still pending evaluation
		if err := tx.Model(&models.Answer{}).Where("exam_participant_id = ? AND evaluated = ? AND answer IS NOT NULL", req.ParticipantId, false).Count(&unevaluatedCount).Error; err != nil {
			return status.Error(codes.Internal, "failed to count unevaluated answers")
		}

		// If all answers are evaluated, mark the participant's exam as fully evaluated
		if unevaluatedCount == 0 {
			participant.Status = constants.PARTICIPANT_STATUS_EVALUATED
			if err := tx.Save(participant).Error; err != nil {
				return status.Error(codes.Internal, "failed to update exam participant status")
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.EvaluationStatusResponse{
		UnevaluatedCount: int32(unevaluatedCount),
	}, nil
}
