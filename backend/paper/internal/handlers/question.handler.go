package handlers

import (
	"context"
	"database/sql"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/utils"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/interceptors"
	"pariksha/paper/internal/utils/validate"
)

func (s *PaperServer) GetPaperQuestions(ctx context.Context, req *proto.PaperRequest) (*proto.QuestionList, error) {
	var questions []models.Question
	if err := db.DB.Where("paper_id = ?", req.PaperId).Find(&questions).Error; err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	response := &proto.QuestionList{
		Questions: make([]*proto.QuestionMinimal, len(questions)),
	}

	for i, question := range questions {
		var err error
		response.Questions[i], err = questionToMinimalProto(question)
		if err != nil {
			return nil, err
		}
	}

	return response, nil
}

func (s *PaperServer) GetPaperQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.QuestionResponse, error) {
	pc, err := NewPaperContext(ctx)
	if err != nil || pc.Question == nil {
		return nil, status.Error(codes.Internal, "question data not found in context")
	}
	return questionToProto(*pc.Question)
}

func (s *PaperServer) CreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
	if err := validate.MaxScore(req.MaxScore); err != nil {
		return nil, err
	}

	// Validate question data based on type
	var coding structs.CodingQuestion
	switch req.Type {
	case constants.QUESTION_TYPE_MCQ:
		var mcq structs.MCQQuestion
		if err := utils.StrictUnmarshal(req.RawQuestion, &mcq); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid MCQ question format")
		}
		if err := validate.McqQuestionData(&mcq); err != nil {
			return nil, err
		}
	case constants.QUESTION_TYPE_SUBJECTIVE:
		var subjective structs.SubjectiveQuestion
		if err := utils.StrictUnmarshal(req.RawQuestion, &subjective); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid subjective question format")
		}
		if err := validate.SubjectiveQuestionData(&subjective); err != nil {
			return nil, err
		}
	case constants.QUESTION_TYPE_CODING:
		if err := utils.StrictUnmarshal(req.RawQuestion, &coding); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid coding question format")
		}
		if err := validate.CodingQuestionData(&coding); err != nil {
			return nil, err
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid question type")
	}

	tags, _ := json.Marshal(req.Tags)
	var question models.Question

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Get max order for this category
		var maxOrder struct{ MaxOrder int16 }
		if err := tx.Model(&models.Question{}).
			Where("category_id = ?", req.CategoryId).
			Select("COALESCE(MAX(\"order\"), 0) as max_order").
			Scan(&maxOrder).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		paper, err := utils.FindRecord[models.Paper](tx, req.PaperId, "paper not found")
		if err != nil {
			return err
		}

		question = models.Question{
			PaperID:    sql.NullInt64{Int64: req.PaperId, Valid: true},
			CategoryID: req.CategoryId,
			Order:      maxOrder.MaxOrder + 1,
			Question:   json.RawMessage(req.RawQuestion),
			Type:       req.Type,
			Tags:       ptr.JsonRawMessage(tags),
			MaxScore:   int16(req.MaxScore),
			CorrectAnswer: sql.NullString{
				String: req.GetCorrectAnswer(),
				Valid:  req.CorrectAnswer != nil,
			},
		}

		if err := tx.Create(&question).Error; err != nil {
			return err
		}

		// Create boilerplates for coding questions
		if req.Type == constants.QUESTION_TYPE_CODING {
			if err := upsertBoilerplates(tx, question.ID, coding.InputDefinitions, coding.OutputDefinition); err != nil {
				return status.Error(codes.Internal, "failed to create boilerplates")
			}
		}

		newCounts, err := updateQuestionCounts(paper.QuestionCounts, req.Type, 1)
		if err != nil {
			return err
		}

		return updatePaperStats(tx, *paper, int32(req.MaxScore), newCounts)
	})

	if err != nil {
		return nil, err
	}

	return &proto.CreateQuestionResponse{
		Id: question.ID,
	}, nil
}

func (s *PaperServer) DeleteQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.Empty, error) {
	question, ok := ctx.Value(interceptors.QuestionCtxKey{}).(models.Question)
	if !ok {
		return nil, status.Error(codes.Internal, "question data not found in context")
	}

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Update paper max score
		if err := tx.Model(&question.Paper).
			UpdateColumn("max_score", gorm.Expr("max_score - ?", question.MaxScore)).
			Error; err != nil {
			return err
		}

		newCounts, err := updateQuestionCounts(question.Paper.QuestionCounts, question.Type, -1)
		if err != nil {
			return err
		}

		if err := tx.Model(&question.Paper).Update("question_counts", newCounts).Error; err != nil {
			return err
		}

		// If question is locked, just unlink it from the paper
		if question.Locked {
			if err := tx.Model(models.Question{}).
				Where("id = ?", question.ID).
				Update("paper_id", sql.NullInt64{}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Delete(&question).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) ReorderQuestions(ctx context.Context, req *proto.ReorderQuestionsRequest) (*proto.Empty, error) {
	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Verify all questions belong to the category
		var categoryQuestionsCount int64
		err := tx.Model(&models.Question{}).
			Where("category_id = ? AND id IN ?", req.CategoryId, req.QuestionIds).
			Count(&categoryQuestionsCount).Error
		if err != nil {
			return err
		}

		if int(categoryQuestionsCount) != len(req.QuestionIds) {
			return status.Error(codes.InvalidArgument, "invalid question ids")
		}

		// Update orders
		for i, questionID := range req.QuestionIds {
			if err := tx.Model(&models.Question{}).
				Where("id = ?", questionID).
				Update("order", i+1).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to reorder questions")
	}

	return &proto.Empty{}, nil
}
