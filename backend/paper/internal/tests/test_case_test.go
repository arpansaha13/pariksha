package tests

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/common/pkg/utils/testrunner"
	"pariksha/paper/internal/config/db"
)

type UpsertTestCasesCase struct {
	BaseTestCase
	setup    func(t *testing.T) *models.Question
	request  *proto.UpsertTestCasesRequest
	validate func(t *testing.T, questionID int64)
}

func TestUpsertPaperTestCases(t *testing.T) {
	tests := []UpsertTestCasesCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Create new test cases",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_CODING,
						Question: json.RawMessage(`{
							"title": "Sum Numbers",
							"statement": "Add two numbers",
							"input_definitions": [
								{ "variable_name": "a", "type": 1 },
								{ "variable_name": "b", "type": 1 }
							],
							"output_definition": {"type": 1}
						}`),
					},
				})
				return &questions[0]
			},
			request: &proto.UpsertTestCasesRequest{
				TestCases: []*proto.UpsertTestCase{
					{
						Inputs:      []string{"1", "2"},
						Output:      "3",
						Explanation: ptr.String("First test case"),
						Hidden:      false,
					},
					{
						Inputs: []string{"4", "5"},
						Output: "9",
						Hidden: true,
					},
				},
			},
			validate: func(t *testing.T, questionID int64) {
				var testCases []models.TestCase
				require.NoError(t, db.DB.Where("question_id = ?", questionID).Order("hidden").Find(&testCases).Error)
				require.Len(t, testCases, 2)

				// Verify first test case (non-hidden)
				var content1 models.TestCaseContent
				require.NoError(t, json.Unmarshal(testCases[0].Content, &content1))
				assert.Equal(t, []string{"1", "2"}, content1.Inputs)
				assert.Equal(t, "3", content1.Output)
				assert.Equal(t, "First test case", *content1.Explanation)
				assert.False(t, testCases[0].Hidden)

				// Verify second test case (hidden)
				var content2 models.TestCaseContent
				require.NoError(t, json.Unmarshal(testCases[1].Content, &content2))
				assert.Equal(t, []string{"4", "5"}, content2.Inputs)
				assert.Equal(t, "9", content2.Output)
				assert.Nil(t, content2.Explanation)
				assert.True(t, testCases[1].Hidden)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update existing test cases",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_CODING,
						Question:   json.RawMessage(`{"title":"Test","statement":"Test","input_definitions":[{"variable_name":"x","type":1}],"output_definition":{"type":1}}`),
					},
				})

				createTestCases(t, []models.TestCase{
					{
						QuestionID: questions[0].ID,
						Content:    json.RawMessage(`{"inputs": ["1"], "output": "1"}`),
						Hidden:     false,
					},
				})

				return &questions[0]
			},
			request: &proto.UpsertTestCasesRequest{
				TestCases: []*proto.UpsertTestCase{
					{
						Id:          ptr.Int64(1), // Existing test case
						Inputs:      []string{"2"},
						Output:      "4",
						Explanation: ptr.String("Updated"),
						Hidden:      true,
					},
					{
						Inputs: []string{"3"},
						Output: "9",
						Hidden: false,
					},
				},
			},
			validate: func(t *testing.T, questionID int64) {
				var testCases []models.TestCase
				require.NoError(t, db.DB.Where("question_id = ?", questionID).Order("id").Find(&testCases).Error)
				require.Len(t, testCases, 2)

				// Verify updated test case
				var content1 models.TestCaseContent
				require.NoError(t, json.Unmarshal(testCases[0].Content, &content1))
				assert.Equal(t, []string{"2"}, content1.Inputs)
				assert.Equal(t, "4", content1.Output)
				assert.Equal(t, "Updated", *content1.Explanation)
				assert.True(t, testCases[0].Hidden)

				// Verify new test case
				var content2 models.TestCaseContent
				require.NoError(t, json.Unmarshal(testCases[1].Content, &content2))
				assert.Equal(t, []string{"3"}, content2.Inputs)
				assert.Equal(t, "9", content2.Output)
				assert.False(t, testCases[1].Hidden)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Duplicate test case IDs",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_CODING,
						Question:   json.RawMessage(`{"title":"Test","statement":"Test","input_definitions":[{"variable_name":"x","type":1}],"output_definition":{"type":1}}`),
					},
				})
				return &questions[0]
			},
			request: &proto.UpsertTestCasesRequest{
				TestCases: []*proto.UpsertTestCase{
					{
						Id:     ptr.Int64(1),
						Inputs: []string{"1"},
						Output: "1",
					},
					{
						Id:     ptr.Int64(1), // Duplicate ID
						Inputs: []string{"2"},
						Output: "4",
					},
				},
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Non-coding question",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Test MCQ","options":["A","B"]}`),
					},
				})
				return &questions[0]
			},
			request: &proto.UpsertTestCasesRequest{
				TestCases: []*proto.UpsertTestCase{
					{
						Inputs: []string{"1"},
						Output: "1",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			question := tt.setup(t)
			tt.request.QuestionId = question.ID

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				tt.request,
				client.UpsertPaperTestCases,
				func(t *testing.T, _ *proto.Empty) {
					if tt.validate != nil {
						tt.validate(t, question.ID)
					}
				})
		})
	}
}
