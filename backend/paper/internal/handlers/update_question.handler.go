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
	"pariksha/common/pkg/utils/generate"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/interceptors"
	"pariksha/paper/internal/utils/validate"
)

// UpdateQuestion handles question updates with proper locking to prevent race conditions
func (s *PaperServer) UpdateQuestion(ctx context.Context, req *proto.UpdateQuestionRequest) (*proto.UpdateQuestionResponse, error) {
	question, err := interceptors.GetQuestionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.MaxScore != nil {
		if err := validate.MaxScore(*req.MaxScore); err != nil {
			return nil, err
		}
	}

	var updatedQuestionHash *string
	err = utils.TransactionHandler(db.DB, func(tx *gorm.DB) error {
		// Row-level lock for both question and paper rows
		var questionForUpdate models.Question
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Take(&questionForUpdate, question.ID).Error; err != nil {
			return status.Error(codes.Internal, "failed to lock question")
		}

		var paper models.Paper
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Take(&paper, question.PaperID.Int64).Error; err != nil {
			return status.Error(codes.Internal, "failed to lock paper")
		}

		oldType := question.Type
		oldMaxScore := question.MaxScore

		if question.Locked {
			updatedQuestionHash, err = handleLockedQuestionUpdate(tx, *question, paper, req, oldType, oldMaxScore)
			return err
		}

		// In case of unlocked question, the the questionId will remain same
		updatedQuestionHash = &req.QuestionHash
		return handleUnlockedQuestionUpdate(tx, *question, paper, req, oldType, oldMaxScore)
	})

	if err != nil {
		return nil, err
	}

	return &proto.UpdateQuestionResponse{QuestionHash: *updatedQuestionHash}, nil
}

// handleLockedQuestionUpdate handles updates to a locked (column) question by creating a new one
func handleLockedQuestionUpdate(tx *gorm.DB, question models.Question, paper models.Paper, req *proto.UpdateQuestionRequest, oldType proto.QuestionType, oldMaxScore int16) (*string, error) {
	newQuestion := question
	newQuestion.ID = 0
	newQuestion.Locked = false

	updatedQuestion, err := applyQuestionUpdates(newQuestion, req)
	if err != nil {
		return nil, err
	}

	if err := tx.Create(&updatedQuestion).Error; err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	// Create HMAC hash for the new question
	questionHash := models.QuestionHash{
		ID:   updatedQuestion.ID,
		Hash: generate.HMACHash(int64(updatedQuestion.ID)),
	}
	if err := tx.Create(&questionHash).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to create question hash")
	}

	// Create new boilerplate entry if it's a coding question
	if updatedQuestion.Type == proto.QuestionType_CODING && req.RawQuestion != nil {
		var coding structs.CodingQuestion
		if err := json.Unmarshal(req.RawQuestion, &coding); err != nil {
			return nil, status.Error(codes.Internal, "invalid coding question format")
		}

		if err := upsertBoilerplates(tx, updatedQuestion.ID, coding.InputDefinitions, coding.OutputDefinition); err != nil {
			return nil, status.Error(codes.Internal, "failed to create boilerplates")
		}
	}

	// Atomic update to unlink old question
	if err := tx.Model(&models.Question{}).
		Where("id = ?", question.ID).
		Update("paper_id", sql.NullInt64{}).Error; err != nil {
		return nil, status.Error(codes.Internal, constants.ErrInternalServer)
	}

	err = updateQuestionStats(tx, paper, oldType, updatedQuestion.Type, oldMaxScore, updatedQuestion.MaxScore)
	if err != nil {
		return nil, err
	}

	return &questionHash.Hash, nil
}

// handleUnlockedQuestionUpdate handles updates to an unlocked (column) question
func handleUnlockedQuestionUpdate(tx *gorm.DB, question models.Question, paper models.Paper, req *proto.UpdateQuestionRequest, oldType proto.QuestionType, oldMaxScore int16) error {
	updatedQuestion, err := applyQuestionUpdates(question, req)
	if err != nil {
		return err
	}

	// Check if this is a coding question and content has changed
	if updatedQuestion.Type == proto.QuestionType_CODING && req.RawQuestion != nil {
		var newCoding, oldCoding structs.CodingQuestion
		if err := json.Unmarshal(req.RawQuestion, &newCoding); err != nil {
			return status.Error(codes.Internal, "invalid coding question format")
		}
		if err := json.Unmarshal(question.Question, &oldCoding); err != nil {
			return status.Error(codes.Internal, "invalid existing question format")
		}

		// Check if input/output definitions have changed
		if shouldUpdateBoilerplates(oldCoding, newCoding) {
			// Delete all test cases for this question as they may no longer be valid
			if err := tx.Delete(&models.TestCase{}, "question_id = ?", question.ID).Error; err != nil {
				return status.Error(codes.Internal, "failed to delete test cases")
			}

			// Update boilerplates
			if err := upsertBoilerplates(tx, updatedQuestion.ID, newCoding.InputDefinitions, newCoding.OutputDefinition); err != nil {
				return status.Error(codes.Internal, "failed to update boilerplates")
			}
		}
	}

	if err := tx.Save(&updatedQuestion).Error; err != nil {
		return status.Error(codes.Internal, constants.ErrInternalServer)
	}

	return updateQuestionStats(tx, paper, oldType, updatedQuestion.Type, oldMaxScore, updatedQuestion.MaxScore)
}

// shouldUpdateBoilerplates checks if input/output definitions have changed
func shouldUpdateBoilerplates(old, new structs.CodingQuestion) bool {
	if len(old.InputDefinitions) != len(new.InputDefinitions) {
		return true
	}

	// Compare input definitions
	for i := range old.InputDefinitions {
		if old.InputDefinitions[i].Type != new.InputDefinitions[i].Type ||
			old.InputDefinitions[i].VariableName != new.InputDefinitions[i].VariableName ||
			!itemsEqual(old.InputDefinitions[i].Items, new.InputDefinitions[i].Items) {
			return true
		}
	}

	// Compare output definition
	return old.OutputDefinition.Type != new.OutputDefinition.Type ||
		!itemsEqual(old.OutputDefinition.Items, new.OutputDefinition.Items)
}

// itemsEqual safely compares two optional item slices
func itemsEqual(a, b *[]structs.ParameterItem) bool {
	if a == nil || b == nil {
		return a == b // true if both nil
	}
	if len(*a) != len(*b) {
		return false
	}
	for i := range *a {
		if (*a)[i].Type != (*b)[i].Type {
			return false
		}
		// Compare PropertyName pointers safely
		if ((*a)[i].PropertyName == nil) != ((*b)[i].PropertyName == nil) {
			return false
		}
		if (*a)[i].PropertyName != nil && (*b)[i].PropertyName != nil &&
			*(*a)[i].PropertyName != *(*b)[i].PropertyName {
			return false
		}
	}
	return true
}

// applyQuestionUpdates handles common question updates and delegates type-specific updates
func applyQuestionUpdates(question models.Question, req *proto.UpdateQuestionRequest) (models.Question, error) {
	if req.Type != nil {
		if req.RawQuestion == nil {
			return question, status.Error(codes.InvalidArgument, "question content must be provided when changing question type")
		}
		question.Type = req.GetType()
	}

	if req.RawQuestion != nil {
		var err error
		switch question.Type {
		case proto.QuestionType_MCQ:
			question, err = applyMcqQuestionUpdates(question, req.RawQuestion)
		case proto.QuestionType_SUBJECTIVE:
			question, err = applySubjectiveQuestionUpdates(question, req.RawQuestion)
		case proto.QuestionType_CODING:
			question, err = applyCodingQuestionUpdates(question, req.RawQuestion)
		default:
			return question, status.Error(codes.InvalidArgument, "invalid question type")
		}
		if err != nil {
			return question, err
		}
	}

	if req.MaxScore != nil {
		question.MaxScore = int16(req.GetMaxScore())
	}

	if len(req.Tags) > 0 {
		tags, _ := json.Marshal(req.Tags)
		question.Tags = ptr.JsonRawMessage(tags)
	}

	if req.CorrectAnswer != nil {
		question.CorrectAnswer = sql.NullString{
			String: req.GetCorrectAnswer(),
			Valid:  true,
		}
	}

	return question, nil
}

// applyMcqQuestionUpdates handles MCQ-specific question updates
func applyMcqQuestionUpdates(question models.Question, rawQuestion []byte) (models.Question, error) {
	var mcq structs.MCQQuestion
	if err := utils.StrictUnmarshal(rawQuestion, &mcq); err != nil {
		return question, status.Error(codes.InvalidArgument, "invalid MCQ question format")
	}
	if err := validate.McqQuestionData(&mcq); err != nil {
		return question, err
	}
	question.Question = json.RawMessage(rawQuestion)
	return question, nil
}

// applySubjectiveQuestionUpdates handles Subjective-specific question updates
func applySubjectiveQuestionUpdates(question models.Question, rawQuestion []byte) (models.Question, error) {
	var subjective structs.SubjectiveQuestion
	if err := utils.StrictUnmarshal(rawQuestion, &subjective); err != nil {
		return question, status.Error(codes.InvalidArgument, "invalid subjective question format")
	}
	if err := validate.SubjectiveQuestionData(&subjective); err != nil {
		return question, err
	}
	question.Question = json.RawMessage(rawQuestion)
	return question, nil
}

// applyCodingQuestionUpdates handles Coding-specific question updates
func applyCodingQuestionUpdates(question models.Question, rawQuestion []byte) (models.Question, error) {
	var coding structs.CodingQuestion
	if err := utils.StrictUnmarshal(rawQuestion, &coding); err != nil {
		return question, status.Error(codes.InvalidArgument, "invalid coding question format")
	}

	if err := validate.CodingQuestionData(&coding); err != nil {
		return question, err
	}

	question.Question = json.RawMessage(rawQuestion)
	return question, nil
}
