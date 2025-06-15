package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/interceptors"
)

func (s *PaperServer) GetQuestionsByIds(ctx context.Context, req *proto.GetQuestionsByIdsRequest) (*proto.GetQuestionsByIdsResponse, error) {
	var questions []models.Question
	if err := db.DB.Joins("QuestionHash").Where("questions.id IN ?", req.QuestionIds).Find(&questions).Error; err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	response := &proto.GetQuestionsByIdsResponse{
		Questions: make([]*proto.QuestionBatchItem, len(questions)),
	}

	for i, question := range questions {
		response.Questions[i] = &proto.QuestionBatchItem{
			QuestionId:   int64(question.ID),
			QuestionHash: question.QuestionHash.Hash,
			MaxScore:     int32(question.MaxScore),
			Type:         question.Type,
			RawQuestion:  question.Question,
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
			CategoryId: int64(category.ID),
			Name:       category.Name,
		}
	}

	return response, nil
}

// GetExamQuestion retrieves minimal question data needed for exam taking
func (s *PaperServer) GetExamQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.QuestionResponse, error) {
	questionID, ok := interceptors.GetQuestionIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "question ID not found in context")
	}

	var question models.Question
	if err := db.DB.Preload("QuestionHash").Select("id, question, type").
		Take(&question, questionID).Error; err != nil {
		return nil, status.Error(codes.NotFound, constants.ErrNotFound)
	}

	var testCases []models.TestCase
	if question.Type == proto.QuestionType_CODING {
		// Fetch only non-hidden test cases for coding questions
		if err := db.DB.Where("question_id = ? AND hidden = ?", question.ID, false).Find(&testCases).Error; err != nil {
			return nil, status.Error(codes.Internal, constants.ErrInternalServer)
		}
	}

	return questionToProto(question, testCases)
}
