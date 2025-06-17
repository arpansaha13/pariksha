package services

import (
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
	"pariksha/paper/internal/repositories"
)

type Category struct {
	categoryRepo *repositories.Category
	paperRepo    *repositories.Paper
	questionRepo *repositories.Question
}

// NewCategoryService creates a new category service instance
func NewCategory(
	categoryRepo *repositories.Category,
	paperRepo *repositories.Paper,
	questionRepo *repositories.Question,
) *Category {
	return &Category{
		categoryRepo: categoryRepo,
		paperRepo:    paperRepo,
		questionRepo: questionRepo,
	}
}

// GetPaperCategories handles fetching all categories for a paper
func (s *Category) GetPaperCategories(paperHash string) (*proto.CategoryList, error) {
	categories, err := s.categoryRepo.GetAllByPaperHash(paperHash)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch paper categories")
	}

	return categoriesToProto(categories), nil
}

// ReorderCategories handles the business logic for reordering categories
func (s *Category) ReorderCategories(paperHash string, categoryIDs []int64) error {
	return utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		paper, err := s.paperRepo.GetByHash(tx, paperHash)
		if err != nil {
			return utils.HandleDBError(err, "paper not found")
		}

		count, err := s.categoryRepo.ValidatePaperCategories(tx, paper.ID, categoryIDs)
		if err != nil {
			return err
		}

		if int(count) != len(categoryIDs) {
			return status.Error(codes.InvalidArgument, "invalid category IDs")
		}

		for i, categoryID := range categoryIDs {
			if err := s.categoryRepo.UpdateOrder(tx, categoryID, int16(i+1)); err != nil {
				return err
			}
		}

		return nil
	})
}

// CreateCategory handles creating a new category
func (s *Category) CreateCategory(paperHash string) (*proto.CategoryResponse, error) {
	var category models.QuestionCategory

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		paper, err := s.paperRepo.GetByHash(tx, paperHash)
		if err != nil {
			return utils.HandleDBError(err, "paper not found")
		}

		maxOrder, err := s.categoryRepo.GetMaxOrder(tx, paper.ID)
		if err != nil {
			return err
		}

		category = models.QuestionCategory{
			PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
			Name:    fmt.Sprintf("Category %d", maxOrder+1),
			Order:   maxOrder + 1,
		}

		return s.categoryRepo.Create(tx, &category)
	})

	if err != nil {
		return nil, err
	}

	return categoryToProto(category), nil
}

// UpdateCategory handles updating a category's name
func (s *Category) UpdateCategory(categoryID int64, name string) error {
	return utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Get existing category
		category, err := s.categoryRepo.GetByID(tx, categoryID)
		if err != nil {
			return utils.HandleDBError(err, "category not found")
		}

		// If category is locked, create a new row with updates
		if category.Locked {
			// Create new category with updates
			newCategory := models.QuestionCategory{
				PaperID: category.PaperID,
				Name:    name,
				Order:   category.Order,
				Locked:  false,
			}

			// Create new category
			if err := s.categoryRepo.Create(tx, &newCategory); err != nil {
				return status.Error(codes.Internal, constants.ErrInternalServer)
			}

			// Update questions to point to new category
			if err := s.questionRepo.UpdateCategory(tx, category.ID, newCategory.ID); err != nil {
				return status.Error(codes.Internal, constants.ErrInternalServer)
			}

			// Unlink old category
			return s.categoryRepo.UnlinkFromPaper(tx, category.ID)
		}

		// Update existing category
		return s.categoryRepo.UpdateName(tx, categoryID, name)
	})
}

// DeleteCategory handles deleting a category and its contents
func (s *Category) DeleteCategory(categoryID int64) error {
	return utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		category, err := s.categoryRepo.GetByID(tx, categoryID)
		if err != nil {
			return utils.HandleDBError(err, "category not found")
		}

		// Check if there is only one category
		categoryCount, err := s.categoryRepo.GetCountByPaperId(tx, category.PaperID)
		if err != nil {
			return grpcerror.Internal(err, "failed to get count by paper id")
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
		questions, err := s.questionRepo.GetAllInCategory(tx, category.ID)
		if err != nil {
			return grpcerror.Internal(err, "failed to get questions in category")
		}

		// Accumulate the diff in max_score and question_counts
		var totalScoreDiff int32
		questionCounts := paper.QuestionCounts

		for _, q := range questions {
			totalScoreDiff -= int32(q.MaxScore)
			if questionCounts, err = updateQuestionCounts(questionCounts, q.Type, -1); err != nil {
				return err
			}
		}

		if err := updatePaperStats(tx, *paper, totalScoreDiff, questionCounts); err != nil {
			return err
		}

		// Handle questions based on lock status
		if err := s.questionRepo.DeleteNonLocked(tx, category.ID); err != nil {
			return grpcerror.Internal(err, "failed to delete non-locked questions")
		}

		if err := s.questionRepo.UnlinkLockedInCategoryFromPaper(tx, category.ID); err != nil {
			return grpcerror.Internal(err, "failed to unlink questions")
		}

		if category.Locked {
			return s.categoryRepo.UnlinkFromPaper(tx, category.ID)
		}

		return s.categoryRepo.Delete(tx, category.ID)
	})
}

// GetCategoriesByIds retrieves multiple categories by their IDs
func (s *Category) GetCategoriesByIds(categoryIDs []int64) (*proto.CategoryBatchResponse, error) {
	categories, err := s.categoryRepo.GetByIds(nil, categoryIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	response := &proto.CategoryBatchResponse{
		Categories: make([]*proto.CategoryBatchItem, len(categories)),
	}

	for i, category := range categories {
		response.Categories[i] = &proto.CategoryBatchItem{
			CategoryId: int64(category.ID),
			Name:       category.Name,
		}
	}

	return response, nil
}
