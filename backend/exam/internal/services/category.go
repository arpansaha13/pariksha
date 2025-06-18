package services

import (
	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/repositories"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Category struct {
	categoryRepo *repositories.Category
}

func NewCategory(categoryRepo *repositories.Category) *Category {
	return &Category{categoryRepo: categoryRepo}
}

// GetExamCategories retrieves all category IDs associated with an exam
func (s *Category) GetExamCategories(req *proto.ExamRequest) (*proto.ExamCategoriesResponse, error) {
	examCategories, err := s.categoryRepo.GetExamCategories(nil, req.ExamHash)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch categories")
	}

	categories := make([]*proto.ExamCategory, len(examCategories))
	for i, ec := range examCategories {
		categories[i] = &proto.ExamCategory{
			CategoryId: int64(ec.CategoryID),
			Order:      int32(ec.Order),
		}
	}

	return &proto.ExamCategoriesResponse{
		Categories: categories,
	}, nil
}
