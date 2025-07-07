package tests

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/common/pkg/utils/testrunner"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/models"
)

type UpsertTestCasesCase struct {
	BaseTestCase
	setup    func(t *testing.T) *models.Question
	request  *proto.UpsertTestCasesRequest
	validate func(t *testing.T, questionID types.QuestionID)
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
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_CODING,
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
			validate: func(t *testing.T, questionID types.QuestionID) {
				var testCases []models.TestCase
				require.NoError(t, db.DB.Where("question_id = ?", questionID).Order("\"order\"").Find(&testCases).Error)
				require.Len(t, testCases, 2)

				// Verify first test case (non-hidden)
				assert.EqualValues(t, 1, testCases[0].Order)
				var content1 models.TestCaseContent
				require.NoError(t, json.Unmarshal(testCases[0].Content, &content1))
				assert.Equal(t, []string{"1", "2"}, content1.Inputs)
				assert.NotEmpty(t, testCases[0].DataHash)

				// Verify second test case (hidden)
				assert.EqualValues(t, 2, testCases[1].Order)
				var content2 models.TestCaseContent
				require.NoError(t, json.Unmarshal(testCases[1].Content, &content2))
				assert.Equal(t, []string{"4", "5"}, content2.Inputs)
				assert.NotEmpty(t, testCases[1].DataHash)
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
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_CODING,
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
			validate: func(t *testing.T, questionID types.QuestionID) {
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
				name:         "Error - Non-coding question",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_MCQ,
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
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Partial update with unchanged content",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_CODING,
						Question:   json.RawMessage(`{"title":"Test","statement":"Test","input_definitions":[{"variable_name":"x","type":1}],"output_definition":{"type":1}}`),
					},
				})

				createTestCases(t, []models.TestCase{
					{
						QuestionID: questions[0].ID,
						Content:    json.RawMessage(`{"inputs": ["1"], "output": "1"}`),
						Hidden:     false,
					},
					{
						QuestionID: questions[0].ID,
						Content:    json.RawMessage(`{"inputs": ["2"], "output": "4"}`),
						Hidden:     true,
					},
				})

				return &questions[0]
			},
			request: &proto.UpsertTestCasesRequest{
				TestCases: []*proto.UpsertTestCase{
					{
						Inputs: []string{"2"},
						Output: "4",
					},
					{
						Inputs: []string{"3"},
						Output: "9",
					},
				},
			},
			validate: func(t *testing.T, questionID types.QuestionID) {
				var testCases []models.TestCase
				require.NoError(t, db.DB.Where("question_id = ?", questionID).Order("id").Find(&testCases).Error)
				require.Len(t, testCases, 2)

				// Verify modified test case
				var content1 models.TestCaseContent
				require.NoError(t, json.Unmarshal(testCases[0].Content, &content1))
				assert.Equal(t, []string{"2"}, content1.Inputs)
				assert.Equal(t, "4", content1.Output)

				// Verify unchanged test case
				var content2 models.TestCaseContent
				require.NoError(t, json.Unmarshal(testCases[1].Content, &content2))
				assert.Equal(t, []string{"3"}, content2.Inputs)
				assert.Equal(t, "9", content2.Output)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Revive soft-deleted test case",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_CODING,
						Question:   json.RawMessage(`{"title":"Test","statement":"Test","input_definitions":[{"variable_name":"x","type":1}],"output_definition":{"type":1}}`),
					},
				})

				// Create a test case and soft delete it
				testCases := createTestCases(t, []models.TestCase{
					{
						QuestionID: questions[0].ID,
						Order:      1,
						Content:    json.RawMessage(`{"inputs": ["1"], "output": "1"}`),
						Hidden:     false,
					},
				})
				require.NoError(t, db.DB.Delete(&testCases[0]).Error)

				return &questions[0]
			},
			request: &proto.UpsertTestCasesRequest{
				TestCases: []*proto.UpsertTestCase{
					{
						Inputs: []string{"1"},
						Output: "1",
						Hidden: true,
					},
				},
			},
			validate: func(t *testing.T, questionID types.QuestionID) {
				var testCases []models.TestCase
				// Check both active and soft-deleted records
				require.NoError(t, db.DB.Unscoped().Where("question_id = ?", questionID).Find(&testCases).Error)
				require.Len(t, testCases, 1)

				// Verify the test case was revived
				assert.False(t, testCases[0].DeletedAt.Valid)
				assert.True(t, testCases[0].Hidden)

				var content models.TestCaseContent
				require.NoError(t, json.Unmarshal(testCases[0].Content, &content))
				assert.Equal(t, []string{"1"}, content.Inputs)
				assert.Equal(t, "1", content.Output)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Handle unique constraint with soft-deleted records",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_CODING,
						Question:   json.RawMessage(`{"title":"Test","statement":"Test","input_definitions":[{"variable_name":"x","type":1}],"output_definition":{"type":1}}`),
					},
				})

				// Create and soft-delete multiple test cases
				testCases := createTestCases(t, []models.TestCase{
					{
						QuestionID: questions[0].ID,
						Order:      1,
						Content:    json.RawMessage(`{"inputs": ["1"], "output": "1"}`),
					},
					{
						QuestionID: questions[0].ID,
						Order:      2,
						Content:    json.RawMessage(`{"inputs": ["2"], "output": "4"}`),
					},
				})
				require.NoError(t, db.DB.Delete(&testCases).Error)

				return &questions[0]
			},
			request: &proto.UpsertTestCasesRequest{
				TestCases: []*proto.UpsertTestCase{
					{
						Inputs: []string{"3"},
						Output: "9",
					},
					{
						Inputs: []string{"4"},
						Output: "16",
					},
				},
			},
			validate: func(t *testing.T, questionID types.QuestionID) {
				var testCases []models.TestCase
				require.NoError(t, db.DB.Unscoped().Where("question_id = ?", questionID).Order("\"order\"").Find(&testCases).Error)
				require.Len(t, testCases, 2)

				// Verify both test cases were revived
				for _, tc := range testCases {
					assert.False(t, tc.DeletedAt.Valid)
				}

				// Verify first test case content
				var content1 models.TestCaseContent
				require.NoError(t, json.Unmarshal(testCases[0].Content, &content1))
				assert.Equal(t, []string{"3"}, content1.Inputs)
				assert.Equal(t, "9", content1.Output)

				// Verify second test case content
				var content2 models.TestCaseContent
				require.NoError(t, json.Unmarshal(testCases[1].Content, &content2))
				assert.Equal(t, []string{"4"}, content2.Inputs)
				assert.Equal(t, "16", content2.Output)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Input count mismatch",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_CODING,
						Question: json.RawMessage(`{
							"title": "Test",
							"statement": "Test",
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
						Inputs: []string{"1"}, // Only one input when two are required
						Output: "1",
					},
				},
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Empty input value",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_CODING,
						Question: json.RawMessage(`{
							"title": "Test",
							"statement": "Test",
							"input_definitions": [{"variable_name": "x", "type": 1}],
							"output_definition": {"type": 1}
						}`),
					},
				})
				return &questions[0]
			},
			request: &proto.UpsertTestCasesRequest{
				TestCases: []*proto.UpsertTestCase{
					{
						Inputs: []string{"   "}, // Empty input (whitespace only)
						Output: "1",
					},
				},
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Empty output value",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_CODING,
						Question: json.RawMessage(`{
							"title": "Test",
							"statement": "Test",
							"input_definitions": [{"variable_name": "x", "type": 1}],
							"output_definition": {"type": 1}
						}`),
					},
				})
				return &questions[0]
			},
			request: &proto.UpsertTestCasesRequest{
				TestCases: []*proto.UpsertTestCase{
					{
						Inputs: []string{"1"},
						Output: "  ", // Empty output (whitespace only)
					},
				},
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Empty explanation is converted to nil",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_CODING,
						Question: json.RawMessage(`{
							"title": "Test",
							"statement": "Test",
							"input_definitions": [{"variable_name": "x", "type": 1}],
							"output_definition": {"type": 1}
						}`),
					},
				})
				return &questions[0]
			},
			request: &proto.UpsertTestCasesRequest{
				TestCases: []*proto.UpsertTestCase{
					{
						Inputs:      []string{"1"},
						Output:      "1",
						Explanation: ptr.String("   "), // Empty explanation (whitespace only)
					},
				},
			},
			validate: func(t *testing.T, questionID types.QuestionID) {
				var testCases []models.TestCase
				require.NoError(t, db.DB.Where("question_id = ?", questionID).Find(&testCases).Error)
				require.Len(t, testCases, 1)

				var content models.TestCaseContent
				require.NoError(t, json.Unmarshal(testCases[0].Content, &content))
				assert.Nil(t, content.Explanation, "Empty explanation should be converted to nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			question := tt.setup(t)
			tt.request.QuestionHash = question.Hash

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				tt.request,
				client.UpsertPaperTestCases,
				func(t *testing.T, _ *emptypb.Empty) {
					if tt.validate != nil {
						tt.validate(t, question.ID)
					}
				})
		})
	}
}
