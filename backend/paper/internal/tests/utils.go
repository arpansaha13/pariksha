package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/structs"
	"pariksha/paper/internal/config/db"
	paperUtils "pariksha/paper/internal/utils"
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
	permissions := models.PaperPermission{
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

// verifyQuestionCounts validates the question counts on a paper
func verifyQuestionCounts(t *testing.T, paperID int64, expected *models.QuestionCount) {
	var paper models.Paper
	require.NoError(t, db.DB.Take(&paper, paperID).Error)
	counts, err := paper.GetQuestionCounts()
	require.NoError(t, err)
	assert.Equal(t, expected.MCQ, counts.MCQ, "MCQ count mismatch")
	assert.Equal(t, expected.Subjective, counts.Subjective, "Subjective count mismatch")
	assert.Equal(t, expected.Coding, counts.Coding, "Coding count mismatch")
}

// verifyMCQContent validates the content of an MCQ question
func verifyMCQContent(t *testing.T, question models.Question, expectedStatement string, expectedOptions []string) {
	var mcq structs.MCQQuestion
	require.NoError(t, json.Unmarshal(question.Question, &mcq))
	assert.Equal(t, expectedStatement, mcq.Statement)
	assert.Equal(t, expectedOptions, mcq.Options)
}

// verifyPaperPermissions validates the permissions of a user for a paper
func verifyPaperPermissions(t *testing.T, paperID, userID int64, expectedRead, expectedWrite bool) {
	var permissions models.PaperPermission
	err := db.DB.Where("paper_id = ? AND user_id = ?", paperID, userID).Take(&permissions).Error
	require.NoError(t, err)
	assert.Equal(t, expectedRead, permissions.CanRead(), "Read permission mismatch")
	assert.Equal(t, expectedWrite, permissions.CanWrite(), "Write permission mismatch")
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

// createTestQuestions creates test questions in the database with default values and auto-generated order
func createTestQuestions(t *testing.T, questions []models.Question) []models.Question {
	// Map to track order counter per category
	categoryOrders := make(map[int64]int16)

	for i := range questions {
		// Set default MaxScore if not provided
		if questions[i].MaxScore == 0 {
			questions[i].MaxScore = 5
		}

		// Auto-generate Order based on category
		currentOrder := categoryOrders[questions[i].CategoryID] + 1
		questions[i].Order = currentOrder
		categoryOrders[questions[i].CategoryID] = currentOrder
	}

	err := db.DB.Create(&questions).Error
	require.NoError(t, err)
	return questions
}

// compareJSONByteArrays checks if two JSON byte arrays contain equivalent data, ignoring key order
func compareJSONByteArrays(a, b []byte) bool {
	var obj1, obj2 map[string]interface{}

	// Unmarshal JSON into maps
	if err := json.Unmarshal(a, &obj1); err != nil {
		fmt.Println("Error unmarshalling first JSON:", err)
		return false
	}
	if err := json.Unmarshal(b, &obj2); err != nil {
		fmt.Println("Error unmarshalling second JSON:", err)
		return false
	}

	// Compare the two maps
	return reflect.DeepEqual(obj1, obj2)
}

// createTestCases creates test cases in the database for a coding question
func createTestCases(t *testing.T, testCases []models.TestCase) []models.TestCase {
	for i := range testCases {
		if testCases[i].Content == nil {
			// Set default content if not provided
			content := models.TestCaseContent{
				Inputs: []string{"1", "2"},
				Output: "3",
			}
			contentBytes, err := json.Marshal(content)
			require.NoError(t, err)
			testCases[i].Content = contentBytes
		}

		var unmarshaled models.TestCaseContent
		json.Unmarshal(testCases[i].Content, &unmarshaled)

		testCases[i].Order = int16(i + 1)
		testCases[i].DataHash = paperUtils.GenerateDataHash(unmarshaled, testCases[i].Hidden)
	}

	err := db.DB.Create(&testCases).Error
	require.NoError(t, err)
	return testCases
}
