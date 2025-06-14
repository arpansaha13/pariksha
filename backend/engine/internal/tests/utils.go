package tests

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/generate"
	"pariksha/engine/internal/config/db"
)

const (
	defaultMaxScore int16            = 10
	userID          int64            = 1
	typedUserID     types.UserID     = 1
	paperID         int64            = 1
	typedPaperID    types.PaperID    = 1
	typedQuestionID types.QuestionID = 1
	typedCategoryID types.CategoryID = 1
)

// createTestPaper creates a test paper entry in the database
func createTestPaper(t *testing.T) types.PaperID {
	paper := models.Paper{
		ID:              typedPaperID,
		Title:           "Test Paper",
		MaxScore:        int32(defaultMaxScore),
		DurationMinutes: 60,
		CreatedBy:       typedUserID,
		QuestionCounts:  json.RawMessage(`{"mcq":0,"subjective":0,"coding":1}`),
	}

	err := db.Papers.Create(&paper).Error
	require.NoError(t, err)

	return paper.ID
}

// createTestCategory creates a test category entry in the database
func createTestCategory(t *testing.T, paperID types.PaperID) types.CategoryID {
	category := models.QuestionCategory{
		ID:      typedCategoryID,
		PaperID: sql.NullInt64{Int64: int64(paperID), Valid: true},
		Name:    "Test Category",
		Order:   1,
	}

	err := db.Papers.Create(&category).Error
	require.NoError(t, err)

	return category.ID
}

// createCodingQuestion creates a test coding question in the database
func createCodingQuestion(t *testing.T, inputDefs []structs.InputDefinition, outputDef structs.OutputDefinition) string {
	// Create paper and category to satisfy foreign key constraints
	paperID := createTestPaper(t)
	categoryID := createTestCategory(t, paperID)

	codingQ := structs.CodingQuestion{
		Title:            "Test Question",
		Statement:        "Write a function that solves the problem",
		InputDefinitions: inputDefs,
		OutputDefinition: outputDef,
	}

	rawQuestion, err := json.Marshal(codingQ)
	require.NoError(t, err)

	question := models.Question{
		ID:         typedQuestionID,
		CategoryID: categoryID,
		PaperID:    sql.NullInt64{Int64: int64(paperID), Valid: true},
		Type:       proto.QuestionType_CODING,
		Question:   rawQuestion,
		MaxScore:   defaultMaxScore,
		Order:      1,
	}

	err = db.Papers.Create(&question).Error
	require.NoError(t, err)

	// Create question hashes
	questionHash := models.QuestionHash{
		ID:   question.ID,
		Hash: generate.HMACHash(int64(question.ID)),
	}
	err = db.Papers.Create(&questionHash).Error
	require.NoError(t, err)

	return question.QuestionHash.Hash
}

// getAbsPath returns the absolute path by joining the given path with test tmp directory
func getAbsPath(path string) string {
	return filepath.Join(".", "tmp", path)
}
