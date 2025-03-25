package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
)

func (s *PaperServer) GetPaperQuestions(ctx context.Context, req *proto.PaperRequest) (*proto.QuestionList, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Verify paper access
	var paper models.Paper
	err = db.DB.Preload("PaperOwnership", "user_id = ?", userID).
		Take(&paper, req.PaperId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "paper not found")
		}
		return nil, status.Error(codes.Internal, "failed to find paper")
	}

	if paper.PaperOwnership.ID == 0 {
		return nil, status.Error(codes.NotFound, "paper not found")
	}

	var questions []models.Question
	if err := db.DB.Where("paper_id = ?", req.PaperId).Find(&questions).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve questions")
	}

	response := &proto.QuestionList{
		Questions: make([]*proto.QuestionMinimal, len(questions)),
	}

	for i, question := range questions {
		response.Questions[i] = &proto.QuestionMinimal{
			Id:         int32(question.ID),
			CategoryId: int32(question.CategoryID),
			PaperId:    int32(question.PaperID),
			Order:      int32(question.Order),
			Question:   nil,
		}

		switch question.Type {
		case constants.QUESTION_TYPE_MCQ:
			var mcq models.MCQQuestion
			if err := json.Unmarshal(question.Question, &mcq); err != nil {
				return nil, status.Error(codes.Internal, "invalid question data")
			}
			response.Questions[i].Question = &proto.QuestionMinimal_Mcq{
				Mcq: &proto.McqQuestion{
					Statement: mcq.Statement,
					Options:   mcq.Options,
				},
			}
		default:
			var general models.GeneralQuestion
			if err := json.Unmarshal(question.Question, &general); err != nil {
				return nil, status.Error(codes.Internal, "invalid question data")
			}
			response.Questions[i].Question = &proto.QuestionMinimal_General{
				General: &proto.GeneralQuestion{
					Statement: general.Statement,
				},
			}
		}
	}

	return response, nil
}

func (s *PaperServer) GetQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.QuestionResponse, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	var question models.Question
	err = db.DB.Preload("Category").
		Preload("Paper").
		Preload("Paper.PaperOwnership", "user_id = ?", userID).
		Take(&question, req.QuestionId).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "question not found")
		}
		return nil, status.Error(codes.Internal, "failed to find question")
	}

	// Verify access
	if question.Paper.PaperOwnership.ID == 0 {
		return nil, status.Error(codes.NotFound, "question not found")
	}

	return questionToProto(question, true)
}

func (s *PaperServer) CreateQuestion(ctx context.Context, req *proto.CreateQuestionRequest) (*proto.QuestionResponse, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Verify paper access
	var paper models.Paper
	err = db.DB.Preload("PaperOwnership", "user_id = ?", userID).
		Take(&paper, req.PaperId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "paper not found")
		}
		return nil, status.Error(codes.Internal, "failed to find paper")
	}

	if paper.PaperOwnership.ID == 0 || paper.PaperOwnership.Type != constants.PAPER_OWNERSHIP_TYPE_OWNER {
		return nil, status.Error(codes.PermissionDenied, "only owner can create questions")
	}

	switch q := req.Question.(type) {
	case *proto.CreateQuestionRequest_Mcq:
		if err := validateQuestionData(req.Type, q.Mcq); err != nil {
			return nil, err
		}
	case *proto.CreateQuestionRequest_General:
		if err := validateQuestionData(req.Type, q.General); err != nil {
			return nil, err
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid question type")
	}

	var questionData json.RawMessage
	switch q := req.Question.(type) {
	case *proto.CreateQuestionRequest_Mcq:
		questionData, _ = json.Marshal(q.Mcq)
	case *proto.CreateQuestionRequest_General:
		questionData, _ = json.Marshal(q.General)
	}

	tags, _ := json.Marshal(req.Tags)

	var question models.Question
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Get max order for this category
		var maxOrder struct{ MaxOrder int }
		err := tx.Model(&models.Question{}).
			Where("category_id = ?", req.CategoryId).
			Select("COALESCE(MAX(\"order\"), 0) as max_order").
			Scan(&maxOrder).Error
		if err != nil {
			return err
		}

		question = models.Question{
			PaperID:    int(req.PaperId),
			CategoryID: int(req.CategoryId),
			Order:      maxOrder.MaxOrder + 1,
			Question:   questionData,
			Type:       req.Type,
			Tags:       tags,
			MaxScore:   int(req.MaxScore),
			CorrectAnswer: sql.NullString{
				String: req.GetCorrectAnswer(),
				Valid:  req.CorrectAnswer != nil,
			},
		}

		if err := tx.Create(&question).Error; err != nil {
			return err
		}

		// Update paper max score
		if err := tx.Model(&paper).
			UpdateColumn("max_score", gorm.Expr("max_score + ?", req.MaxScore)).
			Error; err != nil {
			return err
		}

		if err := tx.Preload("Category").Take(&question, question.ID).Error; err != nil {
			return err
		}

		// Update question counts
		return updateQuestionCounts(tx, paper.ID, req.Type, 1)
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create question")
	}

	return questionToProto(question, true)
}

func (s *PaperServer) UpdateQuestion(ctx context.Context, req *proto.UpdateQuestionRequest) (*proto.Empty, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var question models.Question
		err := tx.Preload("Paper.PaperOwnership", "user_id = ?", userID).
			Take(&question, req.QuestionId).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "question not found")
			}
			return err
		}

		if question.Paper.PaperOwnership.ID == 0 ||
			question.Paper.PaperOwnership.Type != constants.PAPER_OWNERSHIP_TYPE_OWNER {
			return status.Error(codes.PermissionDenied, "only owner can update questions")
		}

		isUpdated := false
		oldType := question.Type
		oldMaxScore := question.MaxScore

		if req.Type != nil && req.GetType() != question.Type {
			question.Type = req.GetType()
			isUpdated = true
		}

		if req.Question != nil {
			questionType := question.Type
			if req.Type != nil {
				questionType = req.GetType()
			}

			switch q := req.Question.(type) {
			case *proto.UpdateQuestionRequest_Mcq:
				if err := validateQuestionData(questionType, q.Mcq); err != nil {
					return err
				}
			case *proto.UpdateQuestionRequest_General:
				if err := validateQuestionData(questionType, q.General); err != nil {
					return err
				}
			default:
				return status.Error(codes.InvalidArgument, "invalid question type")
			}

			var questionData json.RawMessage
			switch q := req.Question.(type) {
			case *proto.UpdateQuestionRequest_Mcq:
				questionData, _ = json.Marshal(q.Mcq)
			case *proto.UpdateQuestionRequest_General:
				questionData, _ = json.Marshal(q.General)
			}
			question.Question = questionData
			isUpdated = true
		}

		if req.CategoryId != nil {
			question.CategoryID = int(req.GetCategoryId())
			isUpdated = true
		}

		if req.MaxScore != nil {
			question.MaxScore = int(req.GetMaxScore())
			isUpdated = true
		}

		if len(req.Tags) > 0 {
			tags, _ := json.Marshal(req.Tags)
			question.Tags = tags
			isUpdated = true
		}

		if req.CorrectAnswer != nil {
			question.CorrectAnswer = sql.NullString{
				String: req.GetCorrectAnswer(),
				Valid:  true,
			}
			isUpdated = true
		}

		if isUpdated {
			if err := tx.Save(&question).Error; err != nil {
				return err
			}

			// Update paper max score if changed
			if req.MaxScore != nil {
				scoreDiff := question.MaxScore - oldMaxScore
				if err := tx.Model(&question.Paper).
					UpdateColumn("max_score", gorm.Expr("max_score + ?", scoreDiff)).
					Error; err != nil {
					return err
				}
			}

			// Update question counts if type changed
			if req.Type != nil && req.GetType() != oldType {
				if err := updateQuestionCounts(tx, question.PaperID, oldType, -1); err != nil {
					return err
				}
				if err := updateQuestionCounts(tx, question.PaperID, question.Type, 1); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update question")
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) DeleteQuestion(ctx context.Context, req *proto.QuestionRequest) (*proto.Empty, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var question models.Question
		err := tx.Preload("Paper.PaperOwnership", "user_id = ?", userID).
			Take(&question, req.QuestionId).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "question not found")
			}
			return err
		}

		if question.Paper.PaperOwnership.ID == 0 ||
			question.Paper.PaperOwnership.Type != constants.PAPER_OWNERSHIP_TYPE_OWNER {
			return status.Error(codes.PermissionDenied, "only owner can delete questions")
		}

		// Update paper max score
		if err := tx.Model(&question.Paper).
			UpdateColumn("max_score", gorm.Expr("max_score - ?", question.MaxScore)).
			Error; err != nil {
			return err
		}

		// Update question counts
		if err := updateQuestionCounts(tx, question.PaperID, question.Type, -1); err != nil {
			return err
		}

		// Delete question
		if err := tx.Delete(&question).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete question")
	}

	return &proto.Empty{}, nil
}

func (s *PaperServer) ReorderQuestions(ctx context.Context, req *proto.ReorderQuestionsRequest) (*proto.Empty, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Verify all questions belong to the category and user has access
		var questions []models.Question
		err := tx.Preload("Paper.PaperOwnership", "user_id = ?", userID).
			Where("category_id = ? AND id IN ?", req.CategoryId, req.QuestionIds).
			Find(&questions).Error
		if err != nil {
			return err
		}

		if len(questions) != len(req.QuestionIds) {
			return status.Error(codes.InvalidArgument, "invalid question ids")
		}

		// Verify ownership
		if len(questions) > 0 && (questions[0].Paper.PaperOwnership.ID == 0 ||
			questions[0].Paper.PaperOwnership.Type != constants.PAPER_OWNERSHIP_TYPE_OWNER) {
			return status.Error(codes.PermissionDenied, "only owner can reorder questions")
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

func questionToProto(question models.Question, includeCategory bool) (*proto.QuestionResponse, error) {
	var category *proto.CategoryResponse
	if includeCategory {
		category = &proto.CategoryResponse{
			Id:    int32(question.Category.ID),
			Name:  question.Category.Name,
			Order: int32(question.Category.Order),
		}
	}

	questionType := question.Type

	var tags []string
	if err := json.Unmarshal(question.Tags, &tags); err != nil {
		return nil, status.Error(codes.Internal, "invalid tags data")
	}

	response := &proto.QuestionResponse{
		Id:            int32(question.ID),
		Question:      nil,
		Category:      category,
		Type:          question.Type,
		Tags:          tags,
		PaperId:       int32(question.PaperID),
		MaxScore:      int32(question.MaxScore),
		CorrectAnswer: &question.CorrectAnswer.String,
	}

	switch questionType {
	case constants.QUESTION_TYPE_MCQ:
		var mcq models.MCQQuestion
		if err := json.Unmarshal(question.Question, &mcq); err != nil {
			return nil, status.Error(codes.Internal, "invalid question data")
		}
		response.Question = &proto.QuestionResponse_Mcq{
			Mcq: &proto.McqQuestion{
				Statement: mcq.Statement,
				Options:   mcq.Options,
			},
		}
	case constants.QUESTION_TYPE_SHORT, constants.QUESTION_TYPE_LONG:
		var general models.GeneralQuestion
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

// Helper function to update question counts
func updateQuestionCounts(tx *gorm.DB, paperID int, questionType string, delta int) error {
	var paper models.Paper
	if err := tx.Take(&paper, paperID).Error; err != nil {
		return err
	}

	var counts struct {
		MCQ   int `json:"mcq"`
		Short int `json:"short"`
		Long  int `json:"long"`
	}
	if err := json.Unmarshal(paper.QuestionCounts, &counts); err != nil {
		return err
	}

	switch questionType {
	case constants.QUESTION_TYPE_MCQ:
		counts.MCQ += delta
	case constants.QUESTION_TYPE_SHORT:
		counts.Short += delta
	case constants.QUESTION_TYPE_LONG:
		counts.Long += delta
	}

	newCounts, err := json.Marshal(counts)
	if err != nil {
		return err
	}

	return tx.Model(&paper).Update("question_counts", newCounts).Error
}

func validateQuestionData(questionType string, question interface{}) error {
	switch questionType {
	case constants.QUESTION_TYPE_MCQ:
		mcq, ok := question.(*proto.McqQuestion)
		if !ok {
			return status.Error(codes.InvalidArgument, "invalid MCQ question format")
		}

		if strings.TrimSpace(mcq.Statement) == "" {
			return status.Error(codes.InvalidArgument, "question statement cannot be empty")
		}
		if len(mcq.Options) < 2 {
			return status.Error(codes.InvalidArgument, "MCQ questions must have at least 2 options")
		}
		if len(mcq.Options) > 5 {
			return status.Error(codes.InvalidArgument, "MCQ questions cannot have more than 5 options")
		}

	case constants.QUESTION_TYPE_SHORT, constants.QUESTION_TYPE_LONG:
		general, ok := question.(*proto.GeneralQuestion)
		if !ok {
			return status.Error(codes.InvalidArgument, "invalid general question format")
		}

		if strings.TrimSpace(general.Statement) == "" {
			return status.Error(codes.InvalidArgument, "question statement cannot be empty")
		}
	}

	return nil
}
