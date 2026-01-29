package service

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/common/pkg/utils/grpcerror"
	"pariksha/paper/internal/domain"
	"pariksha/paper/internal/interservice"
	"pariksha/paper/internal/repository"
)

type Category struct {
	paperRepo      *repository.Paper
	paperCatRepo   *repository.PaperCategory
	paperQuestRepo *repository.PaperQuestion
	questionIntSvc *interservice.Question
}

// NewCategoryService creates a new category service instance
func NewCategory(
	paperRepo *repository.Paper,
	paperCatRepo *repository.PaperCategory,
	paperQuestRepo *repository.PaperQuestion,
	questionIntSvc *interservice.Question,
) *Category {
	return &Category{
		paperRepo:      paperRepo,
		paperCatRepo:   paperCatRepo,
		paperQuestRepo: paperQuestRepo,
		questionIntSvc: questionIntSvc,
	}
}

// GetPaperCategories handles fetching all categories for a paper
func (s *Category) GetPaperCategories(ctx context.Context, paperHash string) (*proto.PaperCategoryList, error) {
	paperCategories, err := s.paperCatRepo.GetAllByPaperHash(nil, paperHash)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch paper categories")
	}

	categoryIDs := make([]types.CategoryID, len(paperCategories))
	for i, pc := range paperCategories {
		categoryIDs[i] = pc.CategoryID
	}

	categories, err := s.questionIntSvc.GetCategoriesByIDs(ctx, categoryIDs)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch question meta")
	}

	response := &proto.PaperCategoryList{
		Categories: make([]*proto.PaperCategoryResponse, len(categories)),
	}

	for i, c := range categories {
		pc := paperCategories[i]
		response.Categories[i] = &proto.PaperCategoryResponse{
			Id:    int64(pc.CategoryID),
			Name:  c.Name,
			Order: int32(pc.Order),
		}
	}

	return response, nil
}

func (s *Category) GetPaperCategoriesMeta(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoriesMeta, error) {
	paperCategories, err := s.paperCatRepo.GetAllByPaperHash(nil, req.PaperHash)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch paper categories")
	}

	if len(paperCategories) == 0 {
		return nil, status.Error(codes.NotFound, "could not find paper categories")
	}

	response := &proto.PaperCategoriesMeta{
		Categories: make([]*proto.PaperCategoryMeta, len(paperCategories)),
	}

	for i, pc := range paperCategories {
		response.Categories[i] = &proto.PaperCategoryMeta{
			Id:    int64(pc.CategoryID),
			Order: int32(pc.Order),
		}
	}

	return response, nil
}

// ReorderCategories handles the business logic for reordering categories
func (s *Category) ReorderCategories(ctx context.Context, paperHash string, categoryIDs []int64) error {
	return s.paperRepo.Transaction(func(tx *gorm.DB) error {
		count, err := s.paperCatRepo.GetCountByPaperHash(tx, paperHash)
		if err != nil {
			return err
		}

		if int(count) != len(categoryIDs) {
			return status.Error(codes.InvalidArgument, "invalid category IDs")
		}

		for i, categoryID := range categoryIDs {
			if err := s.paperCatRepo.UpdateOrder(tx, categoryID, int16(i+1)); err != nil {
				if err == gorm.ErrRecordNotFound {
					return status.Error(codes.NotFound, "could not find category")
				}
				return grpcerror.Internal(err, "failed to update category order")
			}
		}

		return nil
	})
}

// CreateCategory handles creating a new category
func (s *Category) CreateCategory(ctx context.Context, paperHash string) (*proto.PaperCategoryResponse, error) {
	var paperCategory domain.PaperCategory
	var category *proto.CategoryResponse
	err := s.paperRepo.Transaction(func(tx *gorm.DB) error {
		paper, err := s.paperRepo.GetByHash(tx, paperHash)
		if err != nil {
			return utils.HandleDBError(err, "paper not found")
		}

		maxOrder, err := s.paperCatRepo.GetMaxOrder(tx, paper.ID)
		if err != nil {
			return grpcerror.Internal(err, "failed to get max category order")
		}

		category, err = s.questionIntSvc.CreateCategory(ctx, fmt.Sprintf("Category %d", maxOrder+1))
		if err != nil {
			return grpcerror.Internal(err, "failed to create category")
		}

		paperCategory = domain.PaperCategory{
			PaperID:    paper.ID,
			CategoryID: types.CategoryID(category.Id),
			Order:      maxOrder + 1,
		}

		return s.paperCatRepo.Create(tx, &paperCategory)
	})

	if err != nil {
		return nil, err
	}

	return &proto.PaperCategoryResponse{
		Id:    int64(category.Id),
		Name:  category.Name,
		Order: int32(paperCategory.Order),
	}, nil
}

// UpdateCategory handles updating a category's name
func (s *Category) UpdateCategory(ctx context.Context, req *proto.UpdatePaperCategoryRequest) error {
	return s.paperRepo.Transaction(func(tx *gorm.DB) error {
		// Get existing category
		category, err := s.paperCatRepo.GetByPaperHashAndCategoryID(tx, req.PaperHash, types.CategoryID(req.CategoryId))
		if err != nil {
			return utils.HandleDBError(err, "category not found")
		}

		_, err = s.questionIntSvc.UpdateCategoryName(ctx, &proto.UpdateCategoryRequest{
			Id:   int64(category.CategoryID),
			Name: req.Name,
		})
		if err != nil {
			return grpcerror.Internal(err, "failed to update category")
		}

		return nil
	})
}

// DeleteCategory handles deleting a category and its contents
func (s *Category) DeleteCategory(ctx context.Context, req *proto.PaperCategoryRequest) error {
	return s.paperRepo.Transaction(func(tx *gorm.DB) error {
		paperCategory, err := s.paperCatRepo.GetByPaperHashAndCategoryID(tx, req.PaperHash, types.CategoryID(req.CategoryId))
		if err != nil {
			return utils.HandleDBError(err, "category not found")
		}

		// Check if there is only one category
		categoryCount, err := s.paperCatRepo.GetCountByPaperId(tx, paperCategory.PaperID)
		if err != nil {
			return grpcerror.Internal(err, "failed to get count by paper id")
		}

		if categoryCount <= 1 {
			return status.Error(codes.FailedPrecondition, "cannot delete the last category")
		}

		paperQuests, err := s.paperQuestRepo.GetAllByCategoryID(tx, paperCategory.CategoryID)
		if err != nil {
			return grpcerror.Internal(err, "failed to get questions in category")
		}

		questionIDs := make([]types.QuestionID, len(paperQuests))
		for i, pq := range paperQuests {
			questionIDs[i] = pq.QuestionID
		}

		// Decrement paper indegree for deleted paper questions
		if err := s.questionIntSvc.DecQuestionPaperIndegreeByIds(ctx, questionIDs); err != nil {
			return grpcerror.Internal(err, "failed to decrement question paper indegrees")
		}

		// Delete paper questions in this category
		if err := s.paperQuestRepo.BulkDeleteByPaperIDAndCategoryID(tx, paperCategory.PaperID, paperCategory.CategoryID); err != nil {
			return grpcerror.Internal(err, "failed to delete questions in category")
		}

		return s.paperCatRepo.DeleteByID(tx, paperCategory.PaperID, paperCategory.CategoryID)
	})
}
