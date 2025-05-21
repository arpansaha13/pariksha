package handlers

import (
	"context"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PaperServer) GetQuestionsByIds(ctx context.Context, req *proto.GetQuestionsByIdsRequest) (*proto.QuestionBatchResponse, error) {
	var questions []models.Question
	if err := db.DB.Where("id IN ?", req.QuestionIds).Find(&questions).Error; err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	response := &proto.QuestionBatchResponse{
		Questions: make([]*proto.QuestionBatchItem, len(questions)),
	}

	for i, question := range questions {
		response.Questions[i] = &proto.QuestionBatchItem{
			Id:          question.ID,
			MaxScore:    int32(question.MaxScore),
			Type:        question.Type,
			RawQuestion: question.Question,
		}
	}

	return response, nil
}

// GetCategoriesByIds retrieves multiple categories by their IDs in a single request
func (s *PaperServer) GetCategoriesByIds(ctx context.Context, req *proto.GetCategoriesByIdsRequest) (*proto.CategoryBatchResponse, error) {
	var categories []models.QuestionCategory
	if err := db.DB.Where("id IN ?", req.CategoryIds).Find(&categories).Error; err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	response := &proto.CategoryBatchResponse{
		Categories: make([]*proto.CategoryBatchItem, len(categories)),
	}

	for i, category := range categories {
		response.Categories[i] = &proto.CategoryBatchItem{
			Id:   category.ID,
			Name: category.Name,
		}
	}

	return response, nil
}

// GetExamQuestion retrieves minimal question data needed for exam taking
func (s *PaperServer) GetExamQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.QuestionResponse, error) {
	var question models.Question
	if err := db.DB.Select("id, question, type").
		Take(&question, req.QuestionId).Error; err != nil {
		return nil, status.Error(codes.NotFound, constants.ErrNotFound)
	}

	return &proto.QuestionResponse{
		Id:          question.ID,
		Type:        question.Type,
		RawQuestion: question.Question,
	}, nil
}
