package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/common/pkg/utils/generate"
	"pariksha/common/pkg/utils/grpcerror"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/interceptors"
	"pariksha/paper/internal/utils/validate"
)

func (s *PaperServer) GetPaperQuestions(ctx context.Context, req *proto.PaperRequest) (*proto.QuestionList, error) {
	var questions []models.Question
	if err := db.DB.Select("questions.id, question, category_id, paper_id, \"order\", questions.hash").
		Joins("INNER JOIN papers ON papers.id = questions.paper_id").
		Where("papers.hash = ?", req.PaperHash).
		Order("\"order\" ASC").
		Find(&questions).Error; err != nil {
		return nil, grpcerror.Internal(err, "failed to fetch paper questions")
	}

	response := &proto.QuestionList{
		Questions: make([]*proto.QuestionMinimal, len(questions)),
	}

	for i, question := range questions {
		var err error
		question.Paper.Hash = req.PaperHash
		response.Questions[i], err = questionToMinimalProto(question)
		if err != nil {
			return nil, err
		}
	}

	return response, nil
}

func (s *PaperServer) GetPaperQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.QuestionResponse, error) {
	question, err := interceptors.GetQuestionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get associated paper
	if err := db.DB.
		Select("papers.id, papers.hash").
		Where("id = ?", question.PaperID).
		Take(&question.Paper).Error; err != nil {
		return nil, utils.HandleDBError(err, "failed to fetch paper details")
	}

	var testCases []models.TestCase
	if question.Type == proto.QuestionType_CODING {
		// Fetch test cases for coding questions
		if err := db.DB.Where("question_id = ?", question.ID).Find(&testCases).Error; err != nil {
			return nil, status.Error(codes.Internal, constants.ErrInternalServer)
		}
	}

	return questionToProto(*question, testCases)
}

func (s *PaperServer) CreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
	if err := validate.MaxScore(req.MaxScore); err != nil {
		return nil, err
	}

	// Validate question data based on type
	var coding structs.CodingQuestion
	switch req.Type {
	case proto.QuestionType_MCQ:
		var mcq structs.MCQQuestion
		if err := utils.StrictUnmarshal(req.RawQuestion, &mcq); err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid MCQ question format: %s", err.Error()))
		}
		if err := validate.McqQuestionData(&mcq); err != nil {
			return nil, err
		}
	case proto.QuestionType_SUBJECTIVE:
		var subjective structs.SubjectiveQuestion
		if err := utils.StrictUnmarshal(req.RawQuestion, &subjective); err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid subjective question format: %s", err.Error()))
		}
		if err := validate.SubjectiveQuestionData(&subjective); err != nil {
			return nil, err
		}
	case proto.QuestionType_CODING:
		if err := utils.StrictUnmarshal(req.RawQuestion, &coding); err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid coding question format: %s", err.Error()))
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
		// Get paper by hash
		var paper models.Paper
		if err := tx.Where("hash = ?", req.PaperHash).Take(&paper).Error; err != nil {
			return utils.HandleDBError(err, "paper not found")
		}

		// Get max order for this category
		var maxOrder struct{ MaxOrder int16 }
		if err := tx.Model(&models.Question{}).
			Where("category_id = ?", req.CategoryId).
			Select("COALESCE(MAX(\"order\"), 0) as max_order").
			Scan(&maxOrder).Error; err != nil {
			return grpcerror.Internal(err, "failed to get max order forcategory")
		}

		question = models.Question{
			PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
			CategoryID: types.CategoryID(req.CategoryId),
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

		// Generate and store hash
		question.Hash = generate.HMACHash(int64(question.ID))
		if err := tx.Model(&question).Update("hash", question.Hash).Error; err != nil {
			return status.Error(codes.Internal, "failed to store question hash")
		}

		// Create boilerplates for coding questions
		if req.Type == proto.QuestionType_CODING {
			if err := upsertBoilerplates(tx, question.ID, coding.InputDefinitions, coding.OutputDefinition); err != nil {
				return status.Error(codes.Internal, "failed to create boilerplates")
			}
		}

		newCounts, err := updateQuestionCounts(paper.QuestionCounts, req.Type, 1)
		if err != nil {
			return err
		}

		return updatePaperStats(tx, paper, int32(req.MaxScore), newCounts)
	})

	if err != nil {
		return nil, err
	}

	return &proto.CreateQuestionResponse{
		QuestionHash: question.Hash,
	}, nil
}

func (s *PaperServer) DeleteQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.Empty, error) {
	// Get question by hash
	var question models.Question
	if err := db.DB.Where("hash = ?", req.QuestionHash).Take(&question).Error; err != nil {
		return nil, utils.HandleDBError(err, "question not found")
	}

	// Preload paper with required fields
	if err := db.DB.
		Select("id, question_counts").
		Where("id = ?", question.PaperID).
		Take(&question.Paper).Error; err != nil {
		return nil, utils.HandleDBError(err, "failed to fetch question details")
	}

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Update paper max score and question counts
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
			// Delete the question
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
			Where("category_id = ? AND hash IN ?", req.CategoryId, req.QuestionHashes).
			Count(&categoryQuestionsCount).Error
		if err != nil {
			return err
		}

		if int(categoryQuestionsCount) != len(req.QuestionHashes) {
			return status.Error(codes.InvalidArgument, "invalid question hashes")
		}

		// Update orders
		for i, questionHash := range req.QuestionHashes {
			if err := tx.Model(&models.Question{}).
				Where("hash = ?", questionHash).
				Update("order", i+1).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, grpcerror.Internal(err, "failed to reorder questions")
	}

	return &proto.Empty{}, nil
}
