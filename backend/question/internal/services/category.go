package services

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils"
	"pariksha/question/internal/config/db"
	"pariksha/question/internal/models"
	"pariksha/question/internal/repositories"
)

type Category struct {
	categoryRepo *repositories.Category
}

func NewCategory(repo *repositories.Category) *Category {
	return &Category{categoryRepo: repo}
}

// CreateCategory creates a new question category
func (s *Category) CreateCategory(req *proto.CreateCategoryRequest) (*proto.CategoryResponse, error) {
	var category models.Category

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Create new category
		category = models.Category{
			Name:          req.Name,
			PaperIndegree: 0,
			ExamIndegree:  0,
		}

		if err := s.categoryRepo.Create(tx, &category); err != nil {
			return status.Error(codes.Internal, "failed to create category")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.CategoryResponse{
		Id:   int64(category.ID),
		Name: category.Name,
	}, nil
}

// GetCategoriesByIds fetches category details by IDs
func (s *Category) GetCategoriesByIds(ids []int64) (*proto.GetCategoriesResponse, error) {
	categories, err := s.categoryRepo.GetByIDs(nil, ids)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch categories")
	}

	response := &proto.GetCategoriesResponse{
		Categories: make([]*proto.CategoryResponse, len(categories)),
	}

	for i, cat := range categories {
		response.Categories[i] = &proto.CategoryResponse{
			Id:   int64(cat.ID),
			Name: cat.Name,
		}
	}

	return response, nil
}

// UpdateCategoryName updates the name of a question category
func (s *Category) UpdateCategoryName(req *proto.UpdateCategoryRequest) (*proto.UpdateCategoryResponse, error) {
	if err := s.categoryRepo.UpdateName(nil, req.Id, req.Name); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "could not find category")
		}
		return nil, status.Error(codes.Internal, "failed to update category name")
	}

	return &proto.UpdateCategoryResponse{
		Id: req.Id,
	}, nil
}
