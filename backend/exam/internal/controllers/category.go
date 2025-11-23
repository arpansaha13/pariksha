package controllers

import (
	"context"

	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/services"
)

type Category struct {
	categorySvc *services.Category
}

func NewCategory(s *services.Category) *Category {
	return &Category{
		categorySvc: s,
	}
}

// HandleGetExamCategories handles retrieving exam categories
func (c *Category) HandleGetExamCategories(ctx context.Context, req *proto.ExamRequest) (*proto.ExamCategoriesResponse, error) {
	return c.categorySvc.GetExamCategories(ctx, req)
}
