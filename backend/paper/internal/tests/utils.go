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

	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/generate"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/models"
	paperUtils "pariksha/paper/internal/utils"
)

func createContextWithUserID(userID types.UserID) context.Context {
	md := metadata.New(map[string]string{
		"user_id": strconv.FormatInt(int64(userID), 10),
	})
	return metadata.NewOutgoingContext(context.Background(), md)
}

func createTestPaper(t *testing.T, userID types.UserID) models.Paper {
	paper := models.Paper{
		Title:           "Test Paper",
		MaxScore:        0,
		DurationMinutes: 60,
		CreatedBy:       userID,
	}
	err := db.DB.Create(&paper).Error
	require.NoError(t, err)

	// Generate and update paper hash
	paper.Hash = generate.HMACHash(int64(paper.ID))
	err = db.DB.Model(&paper).Update("hash", paper.Hash).Error
	require.NoError(t, err)

	// Create permissions entry with write access
	permissions := models.PaperPermission{
		UserID:  userID,
		PaperID: paper.ID,
	}
	permissions.SetWrite()
	err = db.DB.Create(&permissions).Error
	require.NoError(t, err)

	return paper
}

// verifyQuestionCounts validates the question counts on a paper
func verifyQuestionCounts(t *testing.T, paperID types.PaperID, expected *models.QuestionCount) {
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
func verifyPaperPermissions(t *testing.T, paperHash string, userID types.UserID, expectedRead, expectedWrite bool) {
	var paper models.Paper
	err := db.DB.Where("hash = ?", paperHash).Take(&paper).Error
	require.NoError(t, err)

	var permissions models.PaperPermission
	err = db.DB.Where("paper_id = ? AND user_id = ?", paper.ID, userID).Take(&permissions).Error
	require.NoError(t, err)
	assert.Equal(t, expectedRead, permissions.CanRead(), "Read permission mismatch")
	assert.Equal(t, expectedWrite, permissions.CanWrite(), "Write permission mismatch")
}

// updatePaperCounts updates a paper's question counts
func updatePaperCounts(t *testing.T, paper *models.Paper, counts string) {
	err := db.DB.Model(paper).Update("question_counts", counts).Error
	require.NoError(t, err)
}

// createTestCategories creates test categories in the database with default values and auto-generated order
func createTestCategories(t *testing.T, categories []models.QuestionCategory) []models.QuestionCategory {
	for i := range categories {
		// PaperID must be provided, validate it
		require.True(t, categories[i].PaperID.Valid, "PaperID is required")

		// Set default name if not provided
		if categories[i].Name == "" {
			categories[i].Name = fmt.Sprintf("Category %d", i+1)
		}

		// Auto-generate Order if not provided
		if categories[i].Order == 0 {
			categories[i].Order = int16(i + 1)
		}

		// Default Locked status is false, no need to set explicitly
	}

	err := db.DB.Create(&categories).Error
	require.NoError(t, err)
	return categories
}

func createDefaultTestCategory(t *testing.T, paperID types.PaperID) models.QuestionCategory {
	categories := createTestCategories(t, []models.QuestionCategory{
		{
			PaperID: sql.NullInt64{Int64: int64(paperID), Valid: true},
		},
	})
	return categories[0]
}

func createDefaultTestCategories(t *testing.T, paperID types.PaperID, count int8) []models.QuestionCategory {
	require.Greater(t, count, int8(0), "invalid count argument to createDefaultTestCategories")

	categories := make([]models.QuestionCategory, count)
	for i := range categories {
		categories[i] = models.QuestionCategory{
			PaperID: sql.NullInt64{Int64: int64(paperID), Valid: true},
		}
	}

	categories = createTestCategories(t, categories)
	return categories
}

// createTestQuestions creates test questions in the database with default values and auto-generated order
func createTestQuestions(t *testing.T, questions []models.Question) []models.Question {
	// Map to track order counter per category
	categoryOrders := make(map[types.CategoryID]int16)

	// Create each question separately to handle hash generation properly
	result := make([]models.Question, len(questions))
	for i := range questions {
		// Set default MaxScore if not provided
		if questions[i].MaxScore == 0 {
			questions[i].MaxScore = 5
		}

		// Auto-generate Order based on category
		currentOrder := categoryOrders[questions[i].CategoryID] + 1
		questions[i].Order = currentOrder
		categoryOrders[questions[i].CategoryID] = currentOrder

		// Create question individually
		err := db.DB.Create(&questions[i]).Error
		require.NoError(t, err)

		// Generate and store hash immediately after creation
		questions[i].Hash = generate.HMACHash(int64(questions[i].ID))
		err = db.DB.Model(&questions[i]).Update("hash", questions[i].Hash).Error
		require.NoError(t, err)

		result[i] = questions[i]
	}

	return result
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
