package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/utils"
	"pariksha/paper/internal/interceptors"
)

// ValidateEntityIDs checks if all provided IDs exist in the given table
func ValidateEntityIDs(tx *gorm.DB, tableName string, ids []int64) error {
	var count int64
	err := tx.Table(tableName).Where("id IN ?", ids).Count(&count).Error
	if err != nil {
		return status.Error(codes.Internal, constants.ErrInternalServer)
	}
	if int(count) != len(ids) {
		return status.Error(codes.InvalidArgument, "invalid ids provided")
	}
	return nil
}

// PaperContext represents the context of paper-related operations
type PaperContext struct {
	Paper       models.Paper
	Question    *models.Question
	Category    *models.QuestionCategory
	Permissions *models.PaperPermissions
}

// NewPaperContext creates a new paper context from the given gRPC context
func NewPaperContext(ctx context.Context) (*PaperContext, error) {
	pc := &PaperContext{}

	if perm, ok := ctx.Value(interceptors.PermissionsCtxKey{}).(models.PaperPermissions); ok {
		pc.Permissions = &perm
	}

	if q, ok := ctx.Value(interceptors.QuestionCtxKey{}).(models.Question); ok {
		pc.Question = &q
	}

	if cat, ok := ctx.Value(interceptors.CategoryCtxKey{}).(models.QuestionCategory); ok {
		pc.Category = &cat
	}

	return pc, nil
}

// Helper function to convert Paper model to proto response
func paperToProto(paper models.Paper) *proto.PaperResponse {
	var questionCounts proto.QuestionCount
	json.Unmarshal(paper.QuestionCounts, &questionCounts)

	return &proto.PaperResponse{
		Id:              paper.ID,
		Title:           paper.Title,
		MaxScore:        int32(paper.MaxScore),
		DurationMinutes: int32(paper.DurationMinutes),
		QuestionCounts:  &questionCounts,
		CreatedBy:       paper.CreatedBy,
	}
}

// Helper function to unmarshal and convert question data based on type
func unmarshalQuestionData(questionType string, rawQuestion json.RawMessage) (any, error) {
	switch questionType {
	case constants.QUESTION_TYPE_MCQ:
		var mcq structs.MCQQuestion
		if err := json.Unmarshal(rawQuestion, &mcq); err != nil {
			return nil, status.Error(codes.Internal, "invalid question data")
		}
		return &proto.McqQuestion{
			Statement: mcq.Statement,
			Options:   mcq.Options,
		}, nil
	default:
		var subjective structs.SubjectiveQuestion
		if err := json.Unmarshal(rawQuestion, &subjective); err != nil {
			return nil, status.Error(codes.Internal, "invalid question data")
		}
		return &proto.SubjectiveQuestion{
			Statement: subjective.Statement,
		}, nil
	}
}

func questionToProto(question models.Question) (*proto.QuestionResponse, error) {
	var tags []string
	if err := json.Unmarshal(question.Tags, &tags); err != nil {
		return nil, status.Error(codes.Internal, "invalid tags data")
	}

	questionData, err := unmarshalQuestionData(question.Type, question.Question)
	if err != nil {
		return nil, err
	}

	response := &proto.QuestionResponse{
		Id:            question.ID,
		CategoryId:    question.CategoryID,
		Type:          question.Type,
		Tags:          tags,
		PaperId:       question.PaperID.Int64,
		MaxScore:      int32(question.MaxScore),
		CorrectAnswer: &question.CorrectAnswer.String,
	}

	switch question.Type {
	case constants.QUESTION_TYPE_MCQ:
		response.Question = &proto.QuestionResponse_Mcq{Mcq: questionData.(*proto.McqQuestion)}
	default:
		response.Question = &proto.QuestionResponse_Subjective{Subjective: questionData.(*proto.SubjectiveQuestion)}
	}

	return response, nil
}

func questionToMinimalProto(question models.Question) (*proto.QuestionMinimal, error) {
	questionData, err := unmarshalQuestionData(question.Type, question.Question)
	if err != nil {
		return nil, err
	}

	response := &proto.QuestionMinimal{
		Id:         question.ID,
		CategoryId: question.CategoryID,
		PaperId:    question.PaperID.Int64,
		Order:      int32(question.Order),
	}

	switch question.Type {
	case constants.QUESTION_TYPE_MCQ:
		response.Question = &proto.QuestionMinimal_Mcq{Mcq: questionData.(*proto.McqQuestion)}
	default:
		response.Question = &proto.QuestionMinimal_Subjective{Subjective: questionData.(*proto.SubjectiveQuestion)}
	}

	return response, nil
}

// Helper function to convert QuestionCategory model to proto response
func categoryToProto(category models.QuestionCategory) *proto.CategoryResponse {
	return &proto.CategoryResponse{
		Id:    category.ID,
		Name:  category.Name,
		Order: int32(category.Order),
	}
}

// Helper function to convert slice of models to proto responses
func categoriesToProto(categories []models.QuestionCategory) *proto.CategoryList {
	response := &proto.CategoryList{
		Categories: make([]*proto.CategoryResponse, len(categories)),
	}

	for i, category := range categories {
		response.Categories[i] = categoryToProto(category)
	}

	return response
}

// Helper function to update question counts
func updateQuestionCounts(rawCounts json.RawMessage, questionType string, delta int16) (json.RawMessage, error) {
	var counts models.QuestionCount
	if err := json.Unmarshal(rawCounts, &counts); err != nil {
		return nil, err
	}

	switch questionType {
	case constants.QUESTION_TYPE_MCQ:
		counts.MCQ += delta
	case constants.QUESTION_TYPE_SUBJECTIVE:
		counts.Subjective += delta
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
		mcq, ok := question.(*structs.MCQQuestion)
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

	case constants.QUESTION_TYPE_SUBJECTIVE:
		subjective, ok := question.(*structs.SubjectiveQuestion)
		log.Println(subjective)
		if !ok {
			return status.Error(codes.InvalidArgument, "invalid subjective question format")
		}

		if strings.TrimSpace(subjective.Statement) == "" {
			return status.Error(codes.InvalidArgument, "question statement cannot be empty")
		}
	}

	return nil
}

// Helper function to apply updates to a question
func applyQuestionUpdates(question models.Question, req *proto.UpdateQuestionRequest) (models.Question, error) {
	if req.Type != nil {
		if req.RawQuestion == nil {
			return question, status.Error(codes.InvalidArgument, "question content must be provided when changing question type")
		}
		question.Type = req.GetType()
	}

	if req.RawQuestion != nil {
		switch question.Type {
		case constants.QUESTION_TYPE_MCQ:
			var mcq structs.MCQQuestion
			if err := utils.StrictUnmarshal(req.RawQuestion, &mcq); err != nil {
				return question, status.Error(codes.InvalidArgument, "invalid MCQ question format")
			}
			if err := validateQuestionData(question.Type, &mcq); err != nil {
				return question, err
			}
		default:
			var subjective structs.SubjectiveQuestion
			if err := utils.StrictUnmarshal(req.RawQuestion, &subjective); err != nil {
				return question, status.Error(codes.InvalidArgument, "invalid subjective question format")
			}
			if err := validateQuestionData(question.Type, &subjective); err != nil {
				return question, err
			}
		}
		question.Question = json.RawMessage(req.RawQuestion)
	}

	// TODO: Disable category updating for now
	// if req.CategoryId != nil {
	// 	question.CategoryID = req.GetCategoryId()
	// }

	if req.MaxScore != nil {
		question.MaxScore = int16(req.GetMaxScore())
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

// Helper function to update paper stats (max score and question counts)
func updatePaperStats(tx *gorm.DB, paper models.Paper, scoreDiff int32, newQuestionCounts json.RawMessage) error {
	return tx.Model(&paper).
		Updates(map[string]interface{}{
			"max_score":       gorm.Expr("max_score + ?", scoreDiff),
			"question_counts": newQuestionCounts,
		}).Error
}

// updateQuestionStats updates the paper's statistics after a question update
func updateQuestionStats(tx *gorm.DB, paper models.Paper, oldType, newType string, oldScore, newScore int16) error {
	scoreDiff := int32(newScore - oldScore)
	newCounts := paper.QuestionCounts
	var err error

	if oldType != newType {
		if newCounts, err = updateQuestionCounts(newCounts, oldType, -1); err != nil {
			return err
		}
		if newCounts, err = updateQuestionCounts(newCounts, newType, 1); err != nil {
			return err
		}
	}

	return updatePaperStats(tx, paper, scoreDiff, newCounts)
}

// validateMaxScore checks if the given score is within valid range (0 to MAX_SCORE_PER_QUESTION)
func validateMaxScore(score int32) error {
	if score < 0 || score > constants.MAX_SCORE_PER_QUESTION {
		return status.Errorf(codes.InvalidArgument, "max score must be between 0 and %d", constants.MAX_SCORE_PER_QUESTION)
	}
	return nil
}

// validateDuration checks if the given duration in minutes is within valid range
func validateDuration(durationMinutes int32) error {
	if durationMinutes < 0 {
		return status.Error(codes.InvalidArgument, "duration must be positive")
	}
	if durationMinutes > int32(constants.MAX_EXAM_DURATION_MINUTES) {
		return status.Error(codes.InvalidArgument, "duration cannot exceed 24 hours")
	}
	return nil
}
