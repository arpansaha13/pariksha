package controllers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/service"
)

type Category struct {
	categorySvc *service.Category
}

func NewCategory(s *service.Category) *Category {
	return &Category{
		categorySvc: s,
	}
}

func (c *Category) HandleGetPaperCategories(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoryList, error) {
	return c.categorySvc.GetPaperCategories(ctx, req.PaperHash)
}

func (c *Category) HandleCreateCategory(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoryResponse, error) {
	return c.categorySvc.CreateCategory(ctx, req.PaperHash)
}

func (c *Category) HandleUpdateCategory(ctx context.Context, req *proto.UpdatePaperCategoryRequest) (*emptypb.Empty, error) {
	if err := c.categorySvc.UpdateCategory(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (c *Category) HandleReorderCategories(ctx context.Context, req *proto.ReorderPaperCategoriesRequest) (*emptypb.Empty, error) {
	if err := c.categorySvc.ReorderCategories(ctx, req.PaperHash, req.CategoryIds); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (c *Category) HandleDeleteCategory(ctx context.Context, req *proto.PaperCategoryRequest) (*emptypb.Empty, error) {
	if err := c.categorySvc.DeleteCategory(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (c *Category) HandleGetPaperCategoriesMeta(ctx context.Context, req *proto.PaperRequest) (*proto.PaperCategoriesMeta, error) {
	return c.categorySvc.GetPaperCategoriesMeta(ctx, req)
}
