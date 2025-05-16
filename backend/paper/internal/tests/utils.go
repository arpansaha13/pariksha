package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/structs"
	"pariksha/paper/internal/config/db"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

const defaultPaperCategoryName string = "Category 1"

func createContextWithUserID(userID int64) context.Context {
	md := metadata.New(map[string]string{
		"user_id": strconv.FormatInt(userID, 10),
	})
	return metadata.NewOutgoingContext(context.Background(), md)
}

func createTestPaper(t *testing.T, userID int64) models.Paper {
	paper := models.Paper{
		Title:           "Test Paper",
		MaxScore:        0,
		DurationMinutes: 60,
		CreatedBy:       userID,
	}
	err := db.DB.Create(&paper).Error
	require.NoError(t, err)

	// Create permissions entry with write access
	permissions := models.PaperPermissions{
		UserID:  userID,
		PaperID: paper.ID,
	}
	permissions.SetWrite()
	err = db.DB.Create(&permissions).Error
	require.NoError(t, err)

	category := models.QuestionCategory{
		PaperID: sql.NullInt64{Int64: paper.ID, Valid: true},
		Name:    defaultPaperCategoryName,
		Order:   1,
	}
	err = db.DB.Create(&category).Error
	require.NoError(t, err)

	return paper
}

// QuestionBuilder holds configuration for creating test questions
type QuestionBuilder struct {
	PaperID    int64
	CategoryID int64
	Order      int16
	Type       string
	Statement  string
	Options    []string
	MaxScore   int16
	Locked     bool
}

// createMCQQuestion creates a test MCQ question
func createMCQQuestion(t *testing.T, builder QuestionBuilder) models.Question {
	mcq := structs.MCQQuestion{
		Statement: builder.Statement,
		Options:   builder.Options,
	}
	rawQuestion, err := json.Marshal(mcq)
	require.NoError(t, err)

	question := models.Question{
		PaperID:    sql.NullInt64{Int64: builder.PaperID, Valid: true},
		CategoryID: builder.CategoryID,
		Order:      builder.Order,
		Type:       constants.QUESTION_TYPE_MCQ,
		Question:   rawQuestion,
		MaxScore:   builder.MaxScore,
		Locked:     builder.Locked,
	}
	require.NoError(t, db.DB.Create(&question).Error)
	return question
}

// createSubjectiveQuestion creates a test Subjective question
func createSubjectiveQuestion(t *testing.T, builder QuestionBuilder) models.Question {
	subjective := structs.SubjectiveQuestion{
		Statement: builder.Statement,
	}
	rawQuestion, err := json.Marshal(subjective)
	require.NoError(t, err)

	question := models.Question{
		PaperID:    sql.NullInt64{Int64: builder.PaperID, Valid: true},
		CategoryID: builder.CategoryID,
		Order:      builder.Order,
		Type:       builder.Type,
		Question:   rawQuestion,
		MaxScore:   builder.MaxScore,
		Locked:     builder.Locked,
	}
	require.NoError(t, db.DB.Create(&question).Error)
	return question
}

// verifyQuestionCounts validates the question counts on a paper
func verifyQuestionCounts(t *testing.T, paperID int64, expected models.QuestionCount) {
	var paper models.Paper
	require.NoError(t, db.DB.First(&paper, paperID).Error)
	counts, err := paper.GetQuestionCounts()
	require.NoError(t, err)
	assert.Equal(t, expected.MCQ, counts.MCQ, "MCQ count mismatch")
	assert.Equal(t, expected.Subjective, counts.Subjective, "Subjective count mismatch")
}

// verifyMCQContent validates the content of an MCQ question
func verifyMCQContent(t *testing.T, question models.Question, expectedStatement string, expectedOptions []string) {
	var mcq structs.MCQQuestion
	require.NoError(t, json.Unmarshal(question.Question, &mcq))
	assert.Equal(t, expectedStatement, mcq.Statement)
	assert.Equal(t, expectedOptions, mcq.Options)
}

// setupTestQuestion creates a test question with the given configuration
func setupTestQuestion(t *testing.T, userID int64, qType string, initialCounts string) (*models.Paper, *models.Question) {
	paper := createTestPaper(t, userID)
	err := db.DB.Model(&paper).Update("question_counts", initialCounts).Error
	require.NoError(t, err)

	var category models.QuestionCategory
	require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

	builder := QuestionBuilder{
		PaperID:    paper.ID,
		CategoryID: category.ID,
		Order:      1,
		Type:       qType,
		Statement:  "Test Question",
		Options:    []string{"A", "B", "C"},
		MaxScore:   5,
	}

	var question models.Question
	if qType == constants.QUESTION_TYPE_MCQ {
		question = createMCQQuestion(t, builder)
	} else {
		question = createSubjectiveQuestion(t, builder)
	}

	return &paper, &question
}

// setupLockedPaper creates a test paper with specified question counts
func setupLockedPaper(t *testing.T, builderID int64, counts string) (*models.Paper, *models.Question) {
	paper := createTestPaper(t, builderID)
	err := db.DB.Model(&paper).Update("question_counts", counts).Error
	require.NoError(t, err)

	var category models.QuestionCategory
	require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

	question := models.Question{
		PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
		CategoryID: category.ID,
		Order:      1,
		Type:       constants.QUESTION_TYPE_MCQ,
		Question:   json.RawMessage(`{"statement":"Test MCQ","options":["A","B"]}`),
		MaxScore:   5,
		Locked:     true,
	}
	require.NoError(t, db.DB.Create(&question).Error)
	return &paper, &question
}

// verifyPaperPermissions validates the permissions of a user for a paper
func verifyPaperPermissions(t *testing.T, paperID, userID int64, expectedRead, expectedWrite bool) {
	var permissions models.PaperPermissions
	err := db.DB.Where("paper_id = ? AND user_id = ?", paperID, userID).Take(&permissions).Error
	require.NoError(t, err)
	assert.Equal(t, expectedRead, permissions.CanRead(), "Read permission mismatch")
	assert.Equal(t, expectedWrite, permissions.CanWrite(), "Write permission mismatch")
}

// setupPaperWithSharedAccess creates a paper owned by one user and shared with another
func setupPaperWithSharedAccess(t *testing.T, ownerID, sharedWithID int64, readOnly bool) *models.Paper {
	paper := createTestPaper(t, ownerID)

	permissions := models.PaperPermissions{
		UserID:  sharedWithID,
		PaperID: paper.ID,
	}
	if readOnly {
		permissions.SetRead()
	} else {
		permissions.SetWrite()
	}
	err := db.DB.Create(&permissions).Error
	require.NoError(t, err)

	return &paper
}

// updatePaperCounts updates a paper's question counts
func updatePaperCounts(t *testing.T, paper *models.Paper, counts string) {
	err := db.DB.Model(paper).Update("question_counts", counts).Error
	require.NoError(t, err)
}

// setupTestCategory creates a test category with the given configuration
func setupTestCategory(t *testing.T, userID int64, isLocked bool) (*models.Paper, *models.QuestionCategory) {
	paper := createTestPaper(t, userID)
	category := models.QuestionCategory{
		PaperID: sql.NullInt64{Int64: paper.ID, Valid: true},
		Name:    "Test Category",
		Order:   2,
		Locked:  isLocked,
	}
	require.NoError(t, db.DB.Create(&category).Error)
	return &paper, &category
}

// setupCategoryWithQuestions creates a category and adds questions to it
func setupCategoryWithQuestions(t *testing.T, userID int64, isLocked bool, questionCount int32) (*models.Paper, *models.QuestionCategory) {
	paper, category := setupTestCategory(t, userID, isLocked)

	for i := int32(0); i < questionCount; i++ {
		builder := QuestionBuilder{
			PaperID:    paper.ID,
			CategoryID: category.ID,
			Order:      int16(i + 1),
			Type:       constants.QUESTION_TYPE_MCQ,
			Statement:  "Test Question",
			Options:    []string{"A", "B", "C"},
			MaxScore:   5,
			Locked:     isLocked,
		}
		createMCQQuestion(t, builder)
	}

	return paper, category
}
