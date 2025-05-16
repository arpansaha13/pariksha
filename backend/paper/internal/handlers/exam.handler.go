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
		questionData, err := unmarshalQuestionData(question.Type, question.Question)
		if err != nil {
			return nil, err
		}

		batchItem := &proto.QuestionBatchItem{
			Id:       question.ID,
			MaxScore: int32(question.MaxScore),
			Type:     question.Type,
		}

		switch question.Type {
		case constants.QUESTION_TYPE_MCQ:
			batchItem.Question = &proto.QuestionBatchItem_Mcq{
				Mcq: questionData.(*proto.McqQuestion),
			}
		default:
			batchItem.Question = &proto.QuestionBatchItem_General{
				General: questionData.(*proto.GeneralQuestion),
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

	questionData, err := unmarshalQuestionData(question.Type, question.Question)
	if err != nil {
		return nil, err
	}

	response := &proto.QuestionResponse{
		Id: question.ID,
	}

	switch question.Type {
	case constants.QUESTION_TYPE_MCQ:
		response.Question = &proto.QuestionResponse_Mcq{
			Mcq: questionData.(*proto.McqQuestion),
		}
	default:
		response.Question = &proto.QuestionResponse_General{
			General: questionData.(*proto.GeneralQuestion),
		}
	}

	return response, nil
}
