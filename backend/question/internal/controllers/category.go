package controllers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/proto"
	"pariksha/question/internal/services"
)

type Category struct {
	categorySvc *services.Category
}

func NewCategory(s *services.Category) *Category {
	return &Category{
		categorySvc: s,
	}
}

func (c *Category) HandleCreateCategory(ctx context.Context, req *proto.CreateCategoryRequest) (*proto.CategoryResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "category name is required")
	}

	return c.categorySvc.CreateCategory(req)
}

func (c *Category) HandleGetCategoriesByIds(ctx context.Context, req *proto.CategoryIdsRequest) (*proto.GetCategoriesResponse, error) {
	return c.categorySvc.GetCategoriesByIds(req.Ids)
}

func (c *Category) HandleUpdateCategoryName(ctx context.Context, req *proto.UpdateCategoryRequest) (*proto.UpdateCategoryResponse, error) {
	return c.categorySvc.UpdateCategoryName(req)
}

// func (c *Category) HandleIncCategoryPaperIndegree(ctx context.Context, req *proto.CategoryIdsRequest) (*emptypb.Empty, error) {
// 	return &emptypb.Empty{}, c.categorySvc.IncPaperIndegree(req.Ids)
// }

// func (c *Category) HandleDecCategoryPaperIndegree(ctx context.Context, req *proto.CategoryIdsRequest) (*emptypb.Empty, error) {
// 	return &emptypb.Empty{}, c.categorySvc.DecPaperIndegree(req.Ids)
// }

// func (c *Category) HandleIncCategoryExamIndegree(ctx context.Context, req *proto.CategoryIdsRequest) (*emptypb.Empty, error) {
// 	return &emptypb.Empty{}, c.categorySvc.IncExamIndegree(req.Ids)
// }

// func (c *Category) HandleDecCategoryExamIndegree(ctx context.Context, req *proto.CategoryIdsRequest) (*emptypb.Empty, error) {
// 	return &emptypb.Empty{}, c.categorySvc.DecExamIndegree(req.Ids)
// }
