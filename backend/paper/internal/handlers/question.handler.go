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
	"pariksha/paper/internal/interceptors"
)

func (s *PaperServer) GetPaperQuestions(ctx context.Context, req *proto.PaperRequest) (*proto.QuestionList, error) {
	var questions []models.Question
	if err := db.DB.Where("paper_id = ?", req.PaperId).Find(&questions).Error; err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	response := &proto.QuestionList{
		Questions: make([]*proto.QuestionMinimal, len(questions)),
	}

	for i, question := range questions {
		var err error
		response.Questions[i], err = questionToMinimalProto(question)
		if err != nil {
			return nil, err
		}
	}

	return response, nil
}

func (s *PaperServer) GetPaperQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.QuestionResponse, error) {
	pc, err := NewPaperContext(ctx)
	if err != nil || pc.Question == nil {
		return nil, status.Error(codes.Internal, "question data not found in context")
	}
	return questionToProto(*pc.Question)
}

func (s *PaperServer) CreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
	if err := validateMaxScore(req.MaxScore); err != nil {
		return nil, err
	}

	// Validate question data based on type
	switch req.Type {
	case constants.QUESTION_TYPE_MCQ:
		var mcq structs.MCQQuestion
		if err := utils.StrictUnmarshal(req.RawQuestion, &mcq); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid MCQ question format")
		}
		if err := validateMcqQuestionData(&mcq); err != nil {
			return nil, err
		}
	case constants.QUESTION_TYPE_SUBJECTIVE:
		var subjective structs.SubjectiveQuestion
		if err := utils.StrictUnmarshal(req.RawQuestion, &subjective); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid subjective question format")
		}
		if err := validateSubjectiveQuestionData(&subjective); err != nil {
			return nil, err
		}
	case constants.QUESTION_TYPE_CODING:
		var coding structs.CodingQuestion
		if err := utils.StrictUnmarshal(req.RawQuestion, &coding); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid coding question format")
		}
		if err := validateCodingQuestionData(&coding); err != nil {
			return nil, err
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid question type")
	}

	tags, _ := json.Marshal(req.Tags)
	var question models.Question

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Get max order for this category
		var maxOrder struct{ MaxOrder int16 }
		if err := tx.Model(&models.Question{}).
			Where("category_id = ?", req.CategoryId).
			Select("COALESCE(MAX(\"order\"), 0) as max_order").
			Scan(&maxOrder).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		paper, err := utils.FindRecord[models.Paper](tx, req.PaperId, "paper not found")
		if err != nil {
			return err
		}

		question = models.Question{
			PaperID:    sql.NullInt64{Int64: req.PaperId, Valid: true},
			CategoryID: req.CategoryId,
			Order:      maxOrder.MaxOrder + 1,
			Question:   json.RawMessage(req.RawQuestion),
			Type:       req.Type,
			Tags:       tags,
			MaxScore:   int16(req.MaxScore),
			CorrectAnswer: sql.NullString{
				String: req.GetCorrectAnswer(),
				Valid:  req.CorrectAnswer != nil,
			},
		}

		if err := tx.Create(&question).Error; err != nil {
			return err
		}

		newCounts, err := updateQuestionCounts(paper.QuestionCounts, req.Type, 1)
		if err != nil {
			return err
		}

		return updatePaperStats(tx, *paper, int32(req.MaxScore), newCounts)
	})

	if err != nil {
		return nil, err
	}

	return &proto.CreateQuestionResponse{
		Id: question.ID,
	}, nil
}

// UpdateQuestion handles question updates with proper locking to prevent race conditions
func (s *PaperServer) UpdateQuestion(ctx context.Context, req *proto.UpdateQuestionRequest) (*proto.UpdateQuestionResponse, error) {
	pc, err := NewPaperContext(ctx)
	if err != nil || pc.Question == nil {
		return nil, status.Error(codes.Internal, "question data not found in context")
	}

	if req.MaxScore != nil {
		if err := validateMaxScore(*req.MaxScore); err != nil {
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

func (s *PaperServer) DeleteQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.Empty, error) {
	question, ok := ctx.Value(interceptors.QuestionCtxKey{}).(models.Question)
	if !ok {
		return nil, status.Error(codes.Internal, "question data not found in context")
	}

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Update paper max score
		if err := tx.Model(&question.Paper).
			UpdateColumn("max_score", gorm.Expr("max_score - ?", question.MaxScore)).
			Error; err != nil {
			return err
		}

		newCounts, err := updateQuestionCounts(question.Paper.QuestionCounts, question.Type, -1)
		if err != nil {
			return err
		}

		if err := tx.Model(&question.Paper).Update("question_counts", newCounts).Error; err != nil {
			return err
		}

		// If question is locked, just unlink it from the paper
		if question.Locked {
			if err := tx.Model(models.Question{}).
				Where("id = ?", question.ID).
				Update("paper_id", sql.NullInt64{}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Delete(&question).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) ReorderQuestions(ctx context.Context, req *proto.ReorderQuestionsRequest) (*proto.Empty, error) {
	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Verify all questions belong to the category
		var categoryQuestionsCount int64
		err := tx.Model(&models.Question{}).
			Where("category_id = ? AND id IN ?", req.CategoryId, req.QuestionIds).
			Count(&categoryQuestionsCount).Error
		if err != nil {
			return err
		}

		if int(categoryQuestionsCount) != len(req.QuestionIds) {
			return status.Error(codes.InvalidArgument, "invalid question ids")
		}

		// Update orders
		for i, questionID := range req.QuestionIds {
			if err := tx.Model(&models.Question{}).
				Where("id = ?", questionID).
				Update("order", i+1).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to reorder questions")
	}

	return &proto.Empty{}, nil
}
