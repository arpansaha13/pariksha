package controllers

import (
	"context"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils/grpcerror"
	"pariksha/paper/internal/services"
)

type Category struct {
	categorySvc *services.Category
}

func NewCategory(s *services.Category) *Category {
	return &Category{
		categorySvc: s,
	}
}

// HandleGetPaperCategories handles getting all categories for a paper
func (c *Category) HandleGetPaperCategories(ctx context.Context, req *proto.PaperRequest) (*proto.CategoryList, error) {
	return c.categorySvc.GetPaperCategories(req.PaperHash)
}

// HandleCreateCategory handles creating a new category
func (c *Category) HandleCreateCategory(ctx context.Context, req *proto.CreateCategoryRequest) (*proto.CategoryResponse, error) {
	return c.categorySvc.CreateCategory(req.PaperHash)
}

// HandleUpdateCategory handles updating a category
func (c *Category) HandleUpdateCategory(ctx context.Context, req *proto.UpdateCategoryRequest) (*proto.Empty, error) {
	if err := c.categorySvc.UpdateCategory(req.CategoryId, req.Name); err != nil {
		return nil, err
	}
	return &proto.Empty{}, nil
}

// HandleReorderCategories handles reordering categories in a paper
func (c *Category) HandleReorderCategories(ctx context.Context, req *proto.ReorderCategoriesRequest) (*proto.Empty, error) {
	if err := c.categorySvc.ReorderCategories(req.PaperHash, req.CategoryIds); err != nil {
		return nil, grpcerror.Internal(err, "failed to reorder categories")
	}
	return &proto.Empty{}, nil
}

// HandleDeleteCategory handles deleting a category
func (c *Category) HandleDeleteCategory(ctx context.Context, req *proto.CategoryRequest) (*proto.Empty, error) {
	if err := c.categorySvc.DeleteCategory(req.CategoryId); err != nil {
		return nil, err
	}
	return &proto.Empty{}, nil
}

// GetCategoriesByIds retrieves multiple categories by their IDs
func (s *Category) HandleGetCategoriesByIds(ctx context.Context, req *proto.GetCategoriesByIdsRequest) (*proto.CategoryBatchResponse, error) {
	categories, err := s.categorySvc.GetCategoriesByIds(req.CategoryIds)
	if err != nil {
		return nil, grpcerror.Internal(err, "failed to reorder categories")
	}
	return categories, nil
}
