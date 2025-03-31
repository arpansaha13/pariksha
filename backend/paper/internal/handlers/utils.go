package handlers

import (
	"database/sql"
	"encoding/json"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/models"
)

// Helper function to convert Paper model to proto response
func paperToProto(paper models.Paper) *proto.PaperResponse {
	var questionCounts proto.QuestionCount
	json.Unmarshal(paper.QuestionCounts, &questionCounts)

	return &proto.PaperResponse{
		Id:              int32(paper.ID),
		Title:           paper.Title,
		MaxScore:        int32(paper.MaxScore),
		DurationMinutes: int32(paper.DurationMinutes),
		QuestionCounts:  &questionCounts,
		Ownership: &proto.PaperOwnership{
			Id:   int32(paper.PaperOwnership.ID),
			Path: paper.PaperOwnership.Path,
			Type: paper.PaperOwnership.Type,
		},
	}
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
		PaperId:       int32(question.PaperID.Int64),
		MaxScore:      int32(question.MaxScore),
		CorrectAnswer: &question.CorrectAnswer.String,
	}

	switch questionType {
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
	case constants.QUESTION_TYPE_SHORT, constants.QUESTION_TYPE_LONG:
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

// Helper function to verify paper access
func verifyPaperAccess(tx *gorm.DB, paperID interface{}, userID int32, ownershipType string) error {
	if tx == nil {
		tx = db.DB
	}

	var actualPaperID int
	switch v := paperID.(type) {
	case sql.NullInt64:
		if !v.Valid {
			return status.Error(codes.InvalidArgument, "invalid paper id")
		}
		actualPaperID = int(v.Int64)
	case int:
		actualPaperID = v
	default:
		return status.Error(codes.InvalidArgument, "invalid paper id type")
	}

	var condition string
	var args []any

	args = append(args, actualPaperID, userID)
	condition = "po.paper_id = ? AND po.user_id = ?"

	if ownershipType != "" {
		condition += " AND po.type = ?"
		args = append(args, ownershipType)
	}

	var exists bool
	err := tx.Raw(`SELECT EXISTS (
			SELECT 1 FROM paper_ownerships po
			WHERE `+condition+`)`, args...).
		Scan(&exists).Error

	if err != nil {
		return status.Error(codes.Internal, "failed to check paper access")
	}

	if !exists {
		return status.Error(codes.PermissionDenied, "no permission to perform this action")
	}

	return nil
}

// Helper function to update question counts
func updateQuestionCounts(rawCounts json.RawMessage, questionType string, delta int) (json.RawMessage, error) {
	var counts models.QuestionCount
	if err := json.Unmarshal(rawCounts, &counts); err != nil {
		return nil, err
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
		return nil, err
	}

	return newCounts, nil
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

// Helper function to apply updates to a question
func applyQuestionUpdates(question models.Question, req *proto.UpdateQuestionRequest) (models.Question, error) {
	if req.Type != nil {
		question.Type = req.GetType()
	}

	if req.Question != nil {
		questionType := question.Type
		switch q := req.Question.(type) {
		case *proto.UpdateQuestionRequest_Mcq:
			if err := validateQuestionData(questionType, q.Mcq); err != nil {
				return question, err
			}
			questionData, _ := json.Marshal(q.Mcq)
			question.Question = questionData
		case *proto.UpdateQuestionRequest_General:
			if err := validateQuestionData(questionType, q.General); err != nil {
				return question, err
			}
			questionData, _ := json.Marshal(q.General)
			question.Question = questionData
		}
	}

	if req.CategoryId != nil {
		question.CategoryID = int(req.GetCategoryId())
	}

	if req.MaxScore != nil {
		question.MaxScore = int(req.GetMaxScore())
	}

	if len(req.Tags) > 0 {
		tags, _ := json.Marshal(req.Tags)
		question.Tags = tags
	}

	if req.CorrectAnswer != nil {
		question.CorrectAnswer = sql.NullString{
			String: req.GetCorrectAnswer(),
			Valid:  true,
		}
	}

	return question, nil
}

// Helper function to update paper stats when question type or max score changes
func updatePaperStats(tx *gorm.DB, paper models.Paper, oldType, newType string, oldScore, newScore int) error {
	if oldScore != newScore {
		scoreDiff := newScore - oldScore
		if err := tx.Model(&paper).
			UpdateColumn("max_score", gorm.Expr("max_score + ?", scoreDiff)).
			Error; err != nil {
			return err
		}
	}

	if oldType != newType {
		newCounts, err := updateQuestionCounts(paper.QuestionCounts, oldType, -1)
		if err != nil {
			return err
		}

		newCounts, err = updateQuestionCounts(newCounts, newType, 1)
		if err != nil {
			return err
		}

		if err := tx.Model(&paper).Update("question_counts", newCounts).Error; err != nil {
			return err
		}
	}

	return nil
}
