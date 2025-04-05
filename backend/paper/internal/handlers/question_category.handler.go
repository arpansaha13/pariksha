package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/paper/internal/config/db"
)

func (s *PaperServer) GetPaperCategories(ctx context.Context, req *proto.PaperRequest) (*proto.CategoryList, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	if err := verifyPaperAccess(nil, int(req.PaperId), userID, ""); err != nil {
		return nil, err
	}

	var categories []models.QuestionCategory
	err = db.DB.Where("paper_id = ?", req.PaperId).
		Order("\"order\"").
		Find(&categories).Error

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch categories")
	}

	response := &proto.CategoryList{
		Categories: make([]*proto.CategoryResponse, len(categories)),
	}

	for i, category := range categories {
		response.Categories[i] = &proto.CategoryResponse{
			Id:    int32(category.ID),
			Name:  category.Name,
			Order: int32(category.Order),
		}
	}

	return response, nil
}

func (s *PaperServer) CreateCategory(ctx context.Context, req *proto.CreateCategoryRequest) (*proto.CategoryResponse, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var transactionErr error
	var category models.QuestionCategory
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var paper models.Paper
		err := tx.Preload("PaperOwnership", "user_id = ?", userID).
			Take(&paper, req.PaperId).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				transactionErr = status.Error(codes.NotFound, "paper not found")
				return transactionErr
			}
			return err
		}

		if paper.PaperOwnership.ID == 0 || paper.PaperOwnership.Type != constants.PAPER_OWNERSHIP_TYPE_OWNER {
			transactionErr = status.Error(codes.PermissionDenied, "only owner can create categories")
			return transactionErr
		}

		var count int64
		err = tx.Model(&models.QuestionCategory{}).
			Where("paper_id = ?", req.PaperId).
			Count(&count).Error
		if err != nil {
			return err
		}

		var maxOrder struct{ MaxOrder int }
		err = tx.Model(&models.QuestionCategory{}).
			Where("paper_id = ?", req.PaperId).
			Select("COALESCE(MAX(\"order\"), 0) as max_order").
			Scan(&maxOrder).Error
		if err != nil {
			return err
		}

		category = models.QuestionCategory{
			PaperID: sql.NullInt64{Int64: int64(req.PaperId), Valid: true},
			Name:    fmt.Sprintf("Category %d", count+1),
			Order:   maxOrder.MaxOrder + 1,
		}

		return tx.Create(&category).Error
	})

	if err != nil {
		if transactionErr != nil {
			return nil, transactionErr
		}
		return nil, status.Error(codes.Internal, "failed to create category")
	}

	return &proto.CategoryResponse{
		Id:    int32(category.ID),
		Name:  category.Name,
		Order: int32(category.Order),
	}, nil
}

func (s *PaperServer) UpdateCategory(ctx context.Context, req *proto.UpdateCategoryRequest) (*proto.Empty, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var category models.QuestionCategory
	err = db.DB.Take(&category, req.CategoryId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "category not found")
		}
		return nil, status.Error(codes.Internal, "failed to find category")
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err := verifyPaperAccess(tx, category.PaperID, userID, constants.PAPER_OWNERSHIP_TYPE_OWNER); err != nil {
			return err
		}

		// If category is locked, create a new row with updates
		if category.Locked {
			newCategory := models.QuestionCategory{
				PaperID: category.PaperID,
				Name:    req.Name, // Use the updated name directly
				Order:   category.Order,
				Locked:  false,
			}

			if err := tx.Create(&newCategory).Error; err != nil {
				return err
			}

			// Unlink old category from paper
			if err := tx.Model(models.QuestionCategory{}).
				Where("id = ?", category.ID).
				Update("paper_id", sql.NullInt64{}).Error; err != nil {
				return err
			}

			return nil
		}

		// Apply updates to existing category
		category.Name = req.Name
		if err := tx.Save(&category).Error; err != nil {
			return status.Error(codes.Internal, "failed to update category")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) DeleteCategory(ctx context.Context, req *proto.CategoryRequest) (*proto.Empty, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var transactionErr error
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var category models.QuestionCategory
		err = tx.Take(&category, req.CategoryId).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "category not found")
			}
			return status.Error(codes.Internal, "failed to find category")
		}

		if err := verifyPaperAccess(tx, category.PaperID, userID, constants.PAPER_OWNERSHIP_TYPE_OWNER); err != nil {
			transactionErr = err
			return err
		}

		// Check if there is only one category
		var categoryCount int64
		if err := tx.Model(&models.QuestionCategory{}).
			Where("paper_id = ?", category.PaperID).
			Count(&categoryCount).Error; err != nil {
			return err
		}

		if categoryCount <= 1 {
			transactionErr = status.Error(codes.FailedPrecondition, "cannot delete the last category")
			return transactionErr
		}

		// Get paper to update question counts
		var paper models.Paper
		if err := tx.First(&paper, category.PaperID).Error; err != nil {
			return err
		}

		// Get both locked and non-locked questions in this category
		var questions []models.Question
		if err := tx.Where("category_id = ?", req.CategoryId).Find(&questions).Error; err != nil {
			return err
		}

		// Update paper's question counts and max score
		var questionCounts models.QuestionCount
		if err := json.Unmarshal(paper.QuestionCounts, &questionCounts); err != nil {
			return err
		}

		for _, q := range questions {
			switch q.Type {
			case constants.QUESTION_TYPE_MCQ:
				questionCounts.MCQ--
			case constants.QUESTION_TYPE_SHORT:
				questionCounts.Short--
			case constants.QUESTION_TYPE_LONG:
				questionCounts.Long--
			}
			paper.MaxScore -= q.MaxScore
		}

		newCounts, err := json.Marshal(questionCounts)
		if err != nil {
			return err
		}
		paper.QuestionCounts = newCounts

		if err := tx.Save(&paper).Error; err != nil {
			return err
		}

		// Handle locked and non-locked questions differently
		// Delete non-locked questions
		if err := tx.Where("category_id = ? AND locked = false", req.CategoryId).
			Delete(&models.Question{}).Error; err != nil {
			return err
		}

		// Unlink locked questions from the paper
		if err := tx.Model(&models.Question{}).
			Where("category_id = ? AND locked = true", req.CategoryId).
			Update("paper_id", nil).Error; err != nil {
			return err
		}

		// If category is locked, just unlink it from the paper
		if category.Locked {
			if err := tx.Model(models.QuestionCategory{}).
				Where("id = ?", category.ID).
				Update("paper_id", sql.NullInt64{}).Error; err != nil {
				return err
			}
		} else {
			// Delete the category if not locked
			if err := tx.Delete(&category).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if transactionErr != nil {
			return nil, transactionErr
		}
		return nil, status.Error(codes.Internal, "failed to delete category")
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) ReorderCategories(ctx context.Context, req *proto.ReorderCategoriesRequest) (*proto.Empty, error) {
	userID, err := utils.GetUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	var transactionErr error
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var paper models.Paper
		err := tx.Preload("PaperOwnership", "user_id = ?", userID).
			Take(&paper, req.PaperId).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				transactionErr = status.Error(codes.NotFound, "paper not found")
				return transactionErr
			}
			return err
		}

		if paper.PaperOwnership.ID == 0 || paper.PaperOwnership.Type != constants.PAPER_OWNERSHIP_TYPE_OWNER {
			transactionErr = status.Error(codes.PermissionDenied, "only owner can reorder categories")
			return transactionErr
		}

		var count int64
		err = tx.Model(&models.QuestionCategory{}).
			Where("paper_id = ? AND id IN ?", req.PaperId, req.CategoryIds).
			Count(&count).Error
		if err != nil {
			return err
		}

		if int(count) != len(req.CategoryIds) {
			transactionErr = status.Error(codes.InvalidArgument, "invalid category ids")
			return transactionErr
		}

		for i, categoryID := range req.CategoryIds {
			if err := tx.Model(&models.QuestionCategory{}).
				Where("id = ?", categoryID).
				Update("order", i+1).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if transactionErr != nil {
			return nil, transactionErr
		}
		return nil, status.Error(codes.Internal, "failed to reorder categories")
	}

	return &proto.Empty{}, nil
}
