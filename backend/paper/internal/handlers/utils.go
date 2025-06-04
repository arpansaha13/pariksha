package handlers

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/paper/internal/interceptors"
	"pariksha/paper/internal/utils/boilerplate"
)

// validateEntityIDs checks if all provided IDs exist in the given table
func validateEntityIDs(tx *gorm.DB, tableName string, ids []int64) error {
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
	Permissions *models.PaperPermission
}

// NewPaperContext creates a new paper context from the given gRPC context
func NewPaperContext(ctx context.Context) (*PaperContext, error) {
	pc := &PaperContext{}

	if perm, ok := ctx.Value(interceptors.PermissionsCtxKey{}).(models.PaperPermission); ok {
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

func questionToProto(question models.Question, testCases []models.TestCase) (*proto.QuestionResponse, error) {
	var tags []string
	if question.Tags != nil {
		if err := json.Unmarshal(*question.Tags, &tags); err != nil {
			return nil, status.Error(codes.Internal, "invalid tags data")
		}
	}

	response := &proto.QuestionResponse{
		Id:            question.ID,
		CategoryId:    question.CategoryID,
		Type:          question.Type,
		Tags:          tags,
		PaperId:       question.PaperID.Int64,
		MaxScore:      int32(question.MaxScore),
		CorrectAnswer: &question.CorrectAnswer.String,
		RawQuestion:   question.Question,
	}

	// Add test cases if this is a coding question
	if question.Type == constants.QUESTION_TYPE_CODING && len(testCases) > 0 {
		protoTestCases := make([]*proto.PaperTestCase, 0, len(testCases))
		for _, tc := range testCases {
			var content models.TestCaseContent
			if err := json.Unmarshal(tc.Content, &content); err != nil {
				return nil, status.Error(codes.Internal, "invalid test case format")
			}
			protoTestCases = append(protoTestCases, &proto.PaperTestCase{
				Inputs:      content.Inputs,
				Output:      content.Output,
				Explanation: content.Explanation,
				Hidden:      tc.Hidden,
				Order:       int32(tc.Order),
			})
		}
		response.TestCases = protoTestCases
	}

	return response, nil
}

func questionToMinimalProto(question models.Question) (*proto.QuestionMinimal, error) {
	response := &proto.QuestionMinimal{
		Id:          question.ID,
		CategoryId:  question.CategoryID,
		PaperId:     question.PaperID.Int64,
		Order:       int32(question.Order),
		RawQuestion: question.Question,
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
	case constants.QUESTION_TYPE_CODING:
		counts.Coding += delta
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid question type")
	}

	newCounts, err := json.Marshal(counts)
	if err != nil {
		return nil, err
	}

	return newCounts, nil
}

// Helper function to update paper stats (max score and question counts)
func updatePaperStats(tx *gorm.DB, paper models.Paper, scoreDiff int32, newQuestionCounts json.RawMessage) error {
	return tx.Model(&paper).
		Updates(map[string]any{
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

	if scoreDiff != 0 || oldType != newType {
		return updatePaperStats(tx, paper, scoreDiff, newCounts)
	}
	return nil
}

// upsertBoilerplates creates or updates boilerplate code for all languages
func upsertBoilerplates(tx *gorm.DB, questionID int64, inputs []structs.InputDefinition, output structs.OutputDefinition) error {
	var languages []models.Language
	if err := tx.Where("is_enabled = ?", true).Find(&languages).Error; err != nil {
		return err
	}

	for _, lang := range languages {
		code := boilerplate.Generate(&lang, inputs, output)
		if code == "" {
			continue // Skip if language is not supported
		}

		// Upsert using ON CONFLICT
		if err := tx.Exec(`
					INSERT INTO boilerplates (question_id, language_id, code)
					VALUES (?, ?, ?)
					ON CONFLICT (question_id, language_id) DO UPDATE
					SET code = EXCLUDED.code
			`, questionID, lang.ID, code).Error; err != nil {
			return err
		}
	}
	return nil
}
