package controllers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

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

func (c *Category) HandleGetPaperCategories(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoryList, error) {
	return c.categorySvc.GetPaperCategories(req.PaperHash)
}

func (c *Category) HandleCreateCategory(ctx context.Context, req *proto.CreatePaperCategoryRequest) (*proto.PaperCategoryResponse, error) {
	return c.categorySvc.CreateCategory(req.PaperHash)
}

func (c *Category) HandleUpdateCategory(ctx context.Context, req *proto.UpdatePaperCategoryRequest) (*emptypb.Empty, error) {
	if err := c.categorySvc.UpdateCategory(req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (c *Category) HandleReorderCategories(ctx context.Context, req *proto.ReorderPaperCategoriesRequest) (*emptypb.Empty, error) {
	if err := c.categorySvc.ReorderCategories(req.PaperHash, req.CategoryIds); err != nil {
		return nil, grpcerror.Internal(err, "failed to reorder categories")
	}
	return &emptypb.Empty{}, nil
}

func (c *Category) HandleDeleteCategory(ctx context.Context, req *proto.PaperCategoryRequest) (*emptypb.Empty, error) {
	if err := c.categorySvc.DeleteCategory(req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (c *Category) HandleGetPaperCategoriesMeta(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoriesMeta, error) {
	return c.categorySvc.GetPaperCategoriesMeta(req)
}
