package handlers

import (
	"context"
	"database/sql"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/utils"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/models"
)

func (s *PaperServer) GetPaperQuestions(ctx context.Context, req *proto.PaperRequest) (*proto.QuestionList, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	exists, err := paperExists(req.PaperId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, status.Error(codes.NotFound, "paper not found")
	}

	if err := verifyPaperAccess(nil, int(req.PaperId), userID, ""); err != nil {
		return nil, err
	}

	var questions []models.Question
	if err := db.DB.Where("paper_id = ?", req.PaperId).Find(&questions).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve questions")
	}

	response := &proto.QuestionList{
		Questions: make([]*proto.QuestionMinimal, len(questions)),
	}

	for i, question := range questions {
		response.Questions[i] = &proto.QuestionMinimal{
			Id:         int32(question.ID),
			CategoryId: int32(question.CategoryID),
			PaperId:    int32(question.PaperID.Int64),
			Order:      int32(question.Order),
			Question:   nil,
		}

		switch question.Type {
		case constants.QUESTION_TYPE_MCQ:
			var mcq structs.MCQQuestion
			if err := json.Unmarshal(question.Question, &mcq); err != nil {
				return nil, status.Error(codes.Internal, "invalid question data")
			}
			response.Questions[i].Question = &proto.QuestionMinimal_Mcq{
				Mcq: &proto.McqQuestion{
					Statement: mcq.Statement,
					Options:   mcq.Options,
				},
			}
		default:
			var general structs.GeneralQuestion
			if err := json.Unmarshal(question.Question, &general); err != nil {
				return nil, status.Error(codes.Internal, "invalid question data")
			}
			response.Questions[i].Question = &proto.QuestionMinimal_General{
				General: &proto.GeneralQuestion{
					Statement: general.Statement,
				},
			}
		}
	}

	return response, nil
}

func (s *PaperServer) GetQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.QuestionResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var question models.Question
	err = db.DB.Preload("Category").Take(&question, req.QuestionId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "question not found")
		}
		return nil, status.Error(codes.Internal, "failed to find question")
	}

	if err := verifyPaperAccess(nil, question.PaperID, userID, ""); err != nil {
		return nil, err
	}

	return questionToProto(question, true)
}

func (s *PaperServer) CreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest) (*proto.QuestionResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// Verify paper access
	var paper models.Paper
	err = db.DB.Preload("PaperOwnership", "user_id = ?", userID).
		Take(&paper, req.PaperId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "paper not found")
		}
		return nil, status.Error(codes.Internal, "failed to find paper")
	}

	if paper.PaperOwnership.ID == 0 || paper.PaperOwnership.Type != constants.PAPER_OWNERSHIP_TYPE_OWNER {
		return nil, status.Error(codes.PermissionDenied, "only owner can create questions")
	}

	switch q := req.Question.(type) {
	case *proto.CreateQuestionRequest_Mcq:
		if err := validateQuestionData(req.Type, q.Mcq); err != nil {
			return nil, err
		}
	case *proto.CreateQuestionRequest_General:
		if err := validateQuestionData(req.Type, q.General); err != nil {
			return nil, err
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid question type")
	}

	var questionData json.RawMessage
	switch q := req.Question.(type) {
	case *proto.CreateQuestionRequest_Mcq:
		questionData, _ = json.Marshal(q.Mcq)
	case *proto.CreateQuestionRequest_General:
		questionData, _ = json.Marshal(q.General)
	}

	tags, _ := json.Marshal(req.Tags)

	var question models.Question
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Get max order for this category
		var maxOrder struct{ MaxOrder int }
		err := tx.Model(&models.Question{}).
			Where("category_id = ?", req.CategoryId).
			Select("COALESCE(MAX(\"order\"), 0) as max_order").
			Scan(&maxOrder).Error
		if err != nil {
			return err
		}

		question = models.Question{
			PaperID:    sql.NullInt64{Int64: int64(req.PaperId), Valid: true},
			CategoryID: int(req.CategoryId),
			Order:      maxOrder.MaxOrder + 1,
			Question:   questionData,
			Type:       req.Type,
			Tags:       tags,
			MaxScore:   int(req.MaxScore),
			CorrectAnswer: sql.NullString{
				String: req.GetCorrectAnswer(),
				Valid:  req.CorrectAnswer != nil,
			},
		}

		if err := tx.Create(&question).Error; err != nil {
			return err
		}

		// Update paper max score
		if err := tx.Model(&paper).
			UpdateColumn("max_score", gorm.Expr("max_score + ?", req.MaxScore)).
			Error; err != nil {
			return err
		}

		if err := tx.Preload("Category").Take(&question, question.ID).Error; err != nil {
			return err
		}

		newCounts, err := updateQuestionCounts(paper.QuestionCounts, req.Type, 1)
		if err != nil {
			return err
		}

		return tx.Model(&paper).Update("question_counts", newCounts).Error
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create question")
	}

	return questionToProto(question, true)
}

func (s *PaperServer) UpdateQuestion(ctx context.Context, req *proto.UpdateQuestionRequest) (*proto.Empty, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var question models.Question
		err := tx.Preload("Paper").Take(&question, req.QuestionId).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "question not found")
			}
			return err
		}

		if err := verifyPaperAccess(tx, question.PaperID, userID, constants.PAPER_OWNERSHIP_TYPE_OWNER); err != nil {
			return err
		}

		oldType := question.Type
		oldMaxScore := question.MaxScore

		if question.Locked {
			newQuestion := question // Copy original values
			newQuestion.ID = 0      // Clear ID for new record
			newQuestion.Locked = false

			updatedQuestion, err := applyQuestionUpdates(newQuestion, req)
			if err != nil {
				return err
			}

			// Create new question with updated values
			if err := tx.Create(&updatedQuestion).Error; err != nil {
				return err
			}

			// Unlink old question from paper
			if err := tx.Model(models.Question{}).
				Where("id = ?", question.ID).
				Update("paper_id", sql.NullInt64{}).Error; err != nil {
				return err
			}

			return updatePaperStats(tx, question.Paper, oldType, updatedQuestion.Type, oldMaxScore, updatedQuestion.MaxScore)
		}

		// Apply updates to existing question
		updatedQuestion, err := applyQuestionUpdates(question, req)
		if err != nil {
			return err
		}

		if err := tx.Save(&updatedQuestion).Error; err != nil {
			return err
		}

		return updatePaperStats(tx, updatedQuestion.Paper, oldType, updatedQuestion.Type, oldMaxScore, updatedQuestion.MaxScore)
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update question")
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) DeleteQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.Empty, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var question models.Question
		err := tx.Preload("Paper").Take(&question, req.QuestionId).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "question not found")
			}
			return err
		}

		if err := verifyPaperAccess(tx, question.PaperID, userID, constants.PAPER_OWNERSHIP_TYPE_OWNER); err != nil {
			return err
		}

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
		return nil, status.Error(codes.Internal, "failed to delete question")
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) ReorderQuestions(ctx context.Context, req *proto.ReorderQuestionsRequest) (*proto.Empty, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Verify all questions belong to the category and user has access
		var questions []models.Question
		err := tx.Preload("Paper.PaperOwnership", "user_id = ?", userID).
			Where("category_id = ? AND id IN ?", req.CategoryId, req.QuestionIds).
			Find(&questions).Error
		if err != nil {
			return err
		}

		if len(questions) != len(req.QuestionIds) {
			return status.Error(codes.InvalidArgument, "invalid question ids")
		}

		// Verify ownership
		if len(questions) > 0 && (questions[0].Paper.PaperOwnership.ID == 0 ||
			questions[0].Paper.PaperOwnership.Type != constants.PAPER_OWNERSHIP_TYPE_OWNER) {
			return status.Error(codes.PermissionDenied, "only owner can reorder questions")
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
