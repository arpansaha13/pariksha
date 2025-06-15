package handlers

import (
	"context"
	"database/sql"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/common/pkg/utils/grpcerror"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/interceptors"
)

func (s *PaperServer) GetPaperCategories(ctx context.Context, req *proto.PaperRequest) (*proto.CategoryList, error) {
	var categories []models.QuestionCategory
	err := db.DB.Select("categories.id, categories.name, categories.order").
		Joins("INNER JOIN papers ON papers.id = categories.paper_id").
		Where("papers.hash = ?", req.PaperHash).
		Order("\"order\" ASC").
		Find(&categories).Error

	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch categories")
	}

	return categoriesToProto(categories), nil
}

func (s *PaperServer) CreateCategory(ctx context.Context, req *proto.CreateCategoryRequest) (*proto.CategoryResponse, error) {
	var category models.QuestionCategory
	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// First get paper using hash
		var paper models.Paper
		if err := tx.Where("hash = ?", req.PaperHash).Take(&paper).Error; err != nil {
			return status.Error(codes.NotFound, "paper not found")
		}

		// Count existing categories
		var count int64
		if err := tx.Model(&models.QuestionCategory{}).
			Where("paper_id = ?", paper.ID).
			Count(&count).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		// Get max order
		var maxOrder struct{ MaxOrder int16 }
		if err := tx.Model(&models.QuestionCategory{}).
			Where("paper_id = ?", paper.ID).
			Select("COALESCE(MAX(\"order\"), 0) as max_order").
			Scan(&maxOrder).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		// Create new category
		category = models.QuestionCategory{
			PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
			Name:    fmt.Sprintf("Category %d", count+1),
			Order:   maxOrder.MaxOrder + 1,
		}

		return tx.Create(&category).Error
	})

	if err != nil {
		return nil, err
	}

	return categoryToProto(category), nil
}

func (s *PaperServer) UpdateCategory(ctx context.Context, req *proto.UpdateCategoryRequest) (*proto.Empty, error) {
	category, err := interceptors.GetCategoryFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// If category is locked, create a new row with updates
		if category.Locked {
			newCategory := models.QuestionCategory{
				PaperID: category.PaperID,
				Name:    req.Name,
				Order:   category.Order,
				Locked:  false,
			}

			if err := tx.Create(&newCategory).Error; err != nil {
				return status.Error(codes.Internal, constants.ErrInternalServer)
			}

			// Update all questions to point to the new category
			if err := tx.Model(&models.Question{}).
				Where("category_id = ?", category.ID).
				Update("category_id", newCategory.ID).Error; err != nil {
				return status.Error(codes.Internal, constants.ErrInternalServer)
			}

			// Unlink old category from paper
			return tx.Model(models.QuestionCategory{}).
				Where("id = ?", category.ID).
				Update("paper_id", sql.NullInt64{}).Error
		}

		// Apply updates to existing category
		category.Name = req.Name
		return tx.Save(category).Error
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) DeleteCategory(ctx context.Context, req *proto.CategoryRequest) (*proto.Empty, error) {
	category, err := interceptors.GetCategoryFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Check if there is only one category
		var categoryCount int64
		if err := tx.Model(&models.QuestionCategory{}).
			Where("paper_id = ?", category.PaperID).
			Count(&categoryCount).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		if categoryCount <= 1 {
			return status.Error(codes.FailedPrecondition, "cannot delete the last category")
		}

		// Get paper to update question counts
		paper, err := utils.FindRecord[models.Paper](tx, category.PaperID.Int64, "paper not found")
		if err != nil {
			return err
		}

		// Get both locked and non-locked questions in this category
		var questions []models.Question
		if err := tx.Where("category_id = ?", req.CategoryId).Find(&questions).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		// Accumulate the diff in max_score and question_counts
		var totalScoreDiff int32
		questionCounts := paper.QuestionCounts

		for _, q := range questions {
			totalScoreDiff -= int32(q.MaxScore)
			var err error
			if questionCounts, err = updateQuestionCounts(questionCounts, q.Type, -1); err != nil {
				return err
			}
		}

		// Update paper stats
		if err := updatePaperStats(tx, *paper, totalScoreDiff, questionCounts); err != nil {
			return err
		}

		// Handle locked and non-locked questions
		if err := handleCategoryDeletion(tx, *category, req.CategoryId); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

// handleCategoryDeletion handles the deletion of questions and category based on lock status
func handleCategoryDeletion(tx *gorm.DB, category models.QuestionCategory, categoryID int64) error {
	// Delete non-locked questions
	if err := tx.Where("category_id = ? AND locked = false", categoryID).
		Delete(&models.Question{}).Error; err != nil {
		return status.Error(codes.Internal, constants.ErrInternalServer)
	}

	// Unlink locked questions from the paper
	if err := tx.Model(&models.Question{}).
		Where("category_id = ? AND locked = true", categoryID).
		Update("paper_id", nil).Error; err != nil {
		return status.Error(codes.Internal, constants.ErrInternalServer)
	}

	// Handle category based on lock status
	if category.Locked {
		return tx.Model(models.QuestionCategory{}).
			Where("id = ?", category.ID).
			Update("paper_id", sql.NullInt64{}).Error
	}

	return tx.Delete(&category).Error
}

func (s *PaperServer) ReorderCategories(ctx context.Context, req *proto.ReorderCategoriesRequest) (*proto.Empty, error) {
	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// First get paper using hash
		var paper models.Paper
		if err := tx.Where("hash = ?", req.PaperHash).Take(&paper).Error; err != nil {
			return utils.HandleDBError(err, "paper not found")
		}

		if err := validateEntityIDs(tx, constants.TABLE_CATEGORIES, req.CategoryIds); err != nil {
			return err
		}

		// Additional validation to ensure categories belong to the paper
		var count int64
		if err := tx.Model(&models.QuestionCategory{}).
			Where("paper_id = ? AND id IN ?", paper.ID, req.CategoryIds).
			Count(&count).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		if int(count) != len(req.CategoryIds) {
			return status.Error(codes.InvalidArgument, "some categories do not belong to this paper")
		}

		// Update orders
		for i, categoryID := range req.CategoryIds {
			if err := tx.Model(&models.QuestionCategory{}).
				Where("id = ?", categoryID).
				Update("order", i+1).Error; err != nil {
				return status.Error(codes.Internal, constants.ErrInternalServer)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}
