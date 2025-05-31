package handlers

import (
	"context"
	"database/sql"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/utils"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/utils/validate"
)

// UpdateQuestion handles question updates with proper locking to prevent race conditions
func (s *PaperServer) UpdateQuestion(ctx context.Context, req *proto.UpdateQuestionRequest) (*proto.UpdateQuestionResponse, error) {
	pc, err := NewPaperContext(ctx)
	if err != nil || pc.Question == nil {
		return nil, status.Error(codes.Internal, "question data not found in context")
	}

	if req.MaxScore != nil {
		if err := validate.MaxScore(*req.MaxScore); err != nil {
			return nil, err
		}
	}

	var updatedQuestionID *int64
	err = utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Row-level lock for both question and paper rows
		var question models.Question
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			First(&question, pc.Question.ID).Error; err != nil {
			return status.Error(codes.Internal, "failed to lock question")
		}

		var paper models.Paper
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Take(&paper, question.PaperID.Int64).Error; err != nil {
			return status.Error(codes.Internal, "failed to lock paper")
		}

		oldType := question.Type
		oldMaxScore := question.MaxScore

		if question.Locked {
			updatedQuestionID, err = handleLockedQuestionUpdate(tx, question, paper, req, oldType, oldMaxScore)
			return err
		}

		// In case of unlocked question, the the questionId will remain same
		updatedQuestionID = &question.ID
		return handleUnlockedQuestionUpdate(tx, question, paper, req, oldType, oldMaxScore)
	})

	if err != nil {
		return nil, err
	}

	return &proto.UpdateQuestionResponse{QuestionId: *updatedQuestionID}, nil
}

// handleLockedQuestionUpdate handles updates to a locked (column) question by creating a new one
func handleLockedQuestionUpdate(tx *gorm.DB, question models.Question, paper models.Paper, req *proto.UpdateQuestionRequest, oldType string, oldMaxScore int16) (*int64, error) {
	newQuestion := question
	newQuestion.ID = 0
	newQuestion.Locked = false

	updatedQuestion, err := applyQuestionUpdates(newQuestion, req)
	if err != nil {
		return nil, err
	}

	if err := tx.Create(&updatedQuestion).Error; err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	// Atomic update to unlink old question
	if err := tx.Model(&models.Question{}).
		Where("id = ?", question.ID).
		Update("paper_id", sql.NullInt64{}).Error; err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	err = updateQuestionStats(tx, paper, oldType, updatedQuestion.Type, oldMaxScore, updatedQuestion.MaxScore)
	if err != nil {
		return nil, err
	}

	return &updatedQuestion.ID, nil
}

// handleUnlockedQuestionUpdate handles updates to an unlocked (column) question
func handleUnlockedQuestionUpdate(tx *gorm.DB, question models.Question, paper models.Paper, req *proto.UpdateQuestionRequest, oldType string, oldMaxScore int16) error {
	updatedQuestion, err := applyQuestionUpdates(question, req)
	if err != nil {
		return err
	}

	if err := tx.Save(&updatedQuestion).Error; err != nil {
		return status.Error(codes.Internal, constants.ErrInternalServer)
	}

	return updateQuestionStats(tx, paper, oldType, updatedQuestion.Type, oldMaxScore, updatedQuestion.MaxScore)
}

// Helper function to apply updates to a question
func applyQuestionUpdates(question models.Question, req *proto.UpdateQuestionRequest) (models.Question, error) {
	if req.Type != nil {
		if req.RawQuestion == nil {
			return question, status.Error(codes.InvalidArgument, "question content must be provided when changing question type")
		}
		question.Type = req.GetType()
	}

	if req.RawQuestion != nil {
		switch question.Type {
		case constants.QUESTION_TYPE_MCQ:
			var mcq structs.MCQQuestion
			if err := utils.StrictUnmarshal(req.RawQuestion, &mcq); err != nil {
				return question, status.Error(codes.InvalidArgument, "invalid MCQ question format")
			}
			if err := validate.McqQuestionData(&mcq); err != nil {
				return question, err
			}
		case constants.QUESTION_TYPE_SUBJECTIVE:
			var subjective structs.SubjectiveQuestion
			if err := utils.StrictUnmarshal(req.RawQuestion, &subjective); err != nil {
				return question, status.Error(codes.InvalidArgument, "invalid subjective question format")
			}
			if err := validate.SubjectiveQuestionData(&subjective); err != nil {
				return question, err
			}
		case constants.QUESTION_TYPE_CODING:
			var coding structs.CodingQuestion
			if err := utils.StrictUnmarshal(req.RawQuestion, &coding); err != nil {
				return question, status.Error(codes.InvalidArgument, "invalid coding question format")
			}
			if err := validate.CodingQuestionData(&coding); err != nil {
				return question, err
			}
		default:
			return question, status.Error(codes.InvalidArgument, "invalid question type")
		}
		question.Question = json.RawMessage(req.RawQuestion)
	}

	// TODO: Disable category updating for now
	// if req.CategoryId != nil {
	// 	question.CategoryID = req.GetCategoryId()
	// }

	if req.MaxScore != nil {
		question.MaxScore = int16(req.GetMaxScore())
	}

	if len(req.Tags) > 0 {
		tags, _ := json.Marshal(req.Tags)
		question.Tags = tags
	}

	if req.CorrectAnswer != nil {
		question.CorrectAnswer = sql.NullString{
			String: req.GetCorrectAnswer(),
			Valid:  true,
		}
	}

	return question, nil
}
