package services

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/exam/internal/interservice"
	"pariksha/exam/internal/repositories"
)

type Category struct {
	categoryRepo   *repositories.Category
	questionIntSvc *interservice.Question
}

func NewCategory(
	categoryRepo *repositories.Category,
	questionIntSvc *interservice.Question,
) *Category {
	return &Category{
		categoryRepo:   categoryRepo,
		questionIntSvc: questionIntSvc,
	}
}

// GetExamCategories retrieves all category IDs associated with an exam
func (s *Category) GetExamCategories(ctx context.Context, req *proto.ExamRequest) (*proto.ExamCategoriesResponse, error) {
	examCategories, err := s.categoryRepo.GetExamCategories(nil, req.ExamHash)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch categories")
	}

	categoryIDs := make([]types.CategoryID, len(examCategories))
	for i, ec := range examCategories {
		categoryIDs[i] = ec.CategoryID
	}

	categories, err := s.questionIntSvc.GetCategoriesByIDs(ctx, categoryIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch categories from paper service")
	}

	response := proto.ExamCategoriesResponse{
		Categories: make([]*proto.ExamCategory, len(examCategories)),
	}

	for i, ec := range examCategories {
		response.Categories[i] = &proto.ExamCategory{
			CategoryId: int64(ec.CategoryID),
			Name:       categories[i].Name,
			Order:      int32(ec.Order),
		}
	}

	return &response, nil
}
