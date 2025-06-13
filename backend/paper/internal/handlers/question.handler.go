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
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/common/pkg/utils/generate"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/interceptors"
	"pariksha/paper/internal/utils/validate"
)

func (s *PaperServer) GetPaperQuestions(ctx context.Context, req *proto.PaperRequest) (*proto.QuestionList, error) {
	paperID, ok := interceptors.GetPaperIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "paper ID not found in context")
	}

	var questions []models.Question
	if err := db.DB.Select("id, question, category_id, paper_id, \"order\"").
		Preload("QuestionHash").
		Preload("Paper.PaperHash").
		Where("paper_id = ?", paperID).
		Order("\"order\" ASC").
		Find(&questions).Error; err != nil {
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
	question, err := interceptors.GetQuestionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Define struct to hold the query result
	type QueryResult struct {
		QuestionHash string
		PaperHash    string
	}

	var result QueryResult
	// Join with both hash tables to get question and paper hashes
	err = db.DB.
		Model(&models.Question{}).
		Select("qh.hash as question_hash, ph.hash as paper_hash").
		Joins("INNER JOIN question_hashes qh ON qh.id = questions.id").
		Joins("INNER JOIN paper_hashes ph ON ph.id = questions.paper_id").
		Where("questions.id = ?", question.ID).
		Scan(&result).Error

	if err != nil {
		return nil, utils.HandleDBError(err, "failed to fetch question details")
	}

	// Update question with fetched hashes
	question.QuestionHash.Hash = result.QuestionHash
	question.Paper.PaperHash.Hash = result.PaperHash

	var testCases []models.TestCase
	if question.Type == constants.QUESTION_TYPE_CODING {
		// Fetch test cases for coding questions
		if err := db.DB.Where("question_id = ?", question.ID).Find(&testCases).Error; err != nil {
			return nil, status.Error(codes.Internal, constants.ErrInternalServer)
		}
	}

	return questionToProto(*question, testCases)
}

func (s *PaperServer) CreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest) (*proto.CreateQuestionResponse, error) {
	paperID, ok := interceptors.GetPaperIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "paper ID not found in context")
	}

	if err := validate.MaxScore(req.MaxScore); err != nil {
		return nil, err
	}

	// Validate question data based on type
	var coding structs.CodingQuestion
	switch int16(req.Type) {
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
	var questionHash models.QuestionHash
	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {

		var paper models.Paper
		if err := db.DB.Take(&paper, paperID).Error; err != nil {
			return utils.HandleDBError(err, "paper not found")
		}

		// Get max order for this category
		var maxOrder struct{ MaxOrder int16 }
		if err := tx.Model(&models.Question{}).
			Where("category_id = ?", req.CategoryId).
			Select("COALESCE(MAX(\"order\"), 0) as max_order").
			Scan(&maxOrder).Error; err != nil {
			return status.Error(codes.Internal, constants.ErrInternalServer)
		}

		question = models.Question{
			PaperID:    sql.NullInt64{Int64: int64(paperID), Valid: true},
			CategoryID: types.CategoryID(req.CategoryId),
			Order:      maxOrder.MaxOrder + 1,
			Question:   json.RawMessage(req.RawQuestion),
			Type:       int16(req.Type),
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

		// Create HMAC hash for the new question
		questionHash = models.QuestionHash{
			ID:   question.ID,
			Hash: generate.HMACHash(int64(question.ID)),
		}
		if err := tx.Create(&questionHash).Error; err != nil {
			return status.Error(codes.Internal, "failed to create question hash")
		}

		// Create boilerplates for coding questions
		if int16(req.Type) == constants.QUESTION_TYPE_CODING {
			if err := upsertBoilerplates(tx, question.ID, coding.InputDefinitions, coding.OutputDefinition); err != nil {
				return status.Error(codes.Internal, "failed to create boilerplates")
			}
		}

		newCounts, err := updateQuestionCounts(paper.QuestionCounts, int16(req.Type), 1)
		if err != nil {
			return err
		}

		return updatePaperStats(tx, paper, int32(req.MaxScore), newCounts)
	})

	if err != nil {
		return nil, err
	}

	return &proto.CreateQuestionResponse{
		QuestionHash: questionHash.Hash,
	}, nil
}

func (s *PaperServer) DeleteQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.Empty, error) {
	question, err := interceptors.GetQuestionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
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
	questionIDs, ok := interceptors.GetQuestionIDsFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "question ID list not found in context")
	}

	err := utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Verify all questions belong to the category
		var categoryQuestionsCount int64
		err := tx.Model(&models.Question{}).
			Where("category_id = ? AND id IN ?", req.CategoryId, questionIDs).
			Count(&categoryQuestionsCount).Error
		if err != nil {
			return err
		}

		if int(categoryQuestionsCount) != len(questionIDs) {
			return status.Error(codes.InvalidArgument, "invalid question ids")
		}

		// Update orders
		for i, questionID := range questionIDs {
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
