package handlers

import (
	"context"
	"encoding/json"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/paper/internal/config/db"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PaperServer) GetQuestionsByIds(ctx context.Context, req *proto.GetQuestionsByIdsRequest) (*proto.QuestionBatchResponse, error) {
	var questions []models.Question
	if err := db.DB.Where("id IN ?", req.QuestionIds).Find(&questions).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve questions")
	}

	response := &proto.QuestionBatchResponse{
		Questions: make([]*proto.QuestionBatchItem, len(questions)),
	}

	for i, question := range questions {
		batchItem := &proto.QuestionBatchItem{
			Id:       question.ID,
			MaxScore: int32(question.MaxScore),
			Type:     question.Type,
		}

		switch question.Type {
		case constants.QUESTION_TYPE_MCQ:
			var mcq structs.MCQQuestion
			if err := json.Unmarshal(question.Question, &mcq); err != nil {
				return nil, status.Error(codes.Internal, "invalid question data")
			}
			batchItem.Question = &proto.QuestionBatchItem_Mcq{
				Mcq: &proto.McqQuestion{
					Statement: mcq.Statement,
					Options:   mcq.Options,
				},
			}
		default:
			var general structs.GeneralQuestion
			if err := json.Unmarshal(question.Question, &general); err != nil {
				return nil, status.Error(codes.Internal, "invalid question data")
			}
			batchItem.Question = &proto.QuestionBatchItem_General{
				General: &proto.GeneralQuestion{
					Statement: general.Statement,
				},
			}
		}

		response.Questions[i] = batchItem
	}

	return response, nil
}

// GetCategoriesByIds retrieves multiple categories by their IDs in a single request
func (s *PaperServer) GetCategoriesByIds(ctx context.Context, req *proto.GetCategoriesByIdsRequest) (*proto.CategoryBatchResponse, error) {
	var categories []models.QuestionCategory
	if err := db.DB.Where("id IN ?", req.CategoryIds).Find(&categories).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve categories")
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
	if err := db.DB.Select("id, question, type, max_score, category_id").
		Take(&question, req.QuestionId).Error; err != nil {
		return nil, status.Error(codes.NotFound, "question not found")
	}

	response := &proto.QuestionResponse{
		Id:       question.ID,
		Type:     question.Type,
		MaxScore: int32(question.MaxScore),

		CategoryId: question.CategoryID,
	}

	switch question.Type {
	case constants.QUESTION_TYPE_MCQ:
		var mcq structs.MCQQuestion
		if err := json.Unmarshal(question.Question, &mcq); err != nil {
			return nil, status.Error(codes.Internal, "invalid question data")
		}
		response.Question = &proto.QuestionResponse_Mcq{
			Mcq: &proto.McqQuestion{
				Statement: mcq.Statement,
				Options:   mcq.Options,
			},
		}
	default:
		var general structs.GeneralQuestion
		if err := json.Unmarshal(question.Question, &general); err != nil {
			return nil, status.Error(codes.Internal, "invalid question data")
		}
		response.Question = &proto.QuestionResponse_General{
			General: &proto.GeneralQuestion{
				Statement: general.Statement,
			},
		}
	}

	return response, nil
}
