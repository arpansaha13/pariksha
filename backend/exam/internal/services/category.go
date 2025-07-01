package services

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/exam/internal/interservice/paper"
	"pariksha/exam/internal/repositories"
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

	categoryIDs := make([]types.CategoryID, len(examCategories))
	for i, ec := range examCategories {
		categoryIDs[i] = ec.CategoryID
	}

	paperCategories, err := paper.FetchCategoriesByIds(categoryIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch categories from paper service")
	}

	categories := make([]*proto.ExamCategory, len(examCategories))
	for i, ec := range examCategories {
		categories[i] = &proto.ExamCategory{
			CategoryId: int64(ec.CategoryID),
			Name:       paperCategories[i].Name,
			Order:      int32(ec.Order),
		}
	}

	return &proto.ExamCategoriesResponse{
		Categories: categories,
	}, nil
}
