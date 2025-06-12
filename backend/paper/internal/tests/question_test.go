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
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/testrunner"
	"pariksha/paper/internal/config/db"
)

func TestGetPaperQuestions(t *testing.T) {
	tests := []QuestionListCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get questions for paper with all test cases",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, []models.Question) {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"MCQ Question","options":["A","B","C"]}`),
					},
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_SUBJECTIVE,
						Question:   json.RawMessage(`{"statement":"Subjective Question"}`),
					},
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
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

				// Create test cases for coding question
				createTestCases(t, []models.TestCase{
					{
						QuestionID: questions[2].ID,
						Content:    json.RawMessage(`{"inputs": ["1", "2"], "output": "3"}`),
						Hidden:     false,
					},
					{
						QuestionID: questions[2].ID,
						Content:    json.RawMessage(`{"inputs": ["4", "5"], "output": "9"}`),
						Hidden:     true,
					},
				})

				return &paper, questions
			},
			validate: func(t *testing.T, resp *proto.QuestionList) {
				assert.Equal(t, 3, len(resp.Questions))

				// Validate MCQ question
				mcqJson, _ := json.Marshal(structs.MCQQuestion{
					Statement: "MCQ Question",
					Options:   []string{"A", "B", "C"},
				})
				assert.True(t, compareJSONByteArrays(mcqJson, resp.Questions[0].RawQuestion))

				// Validate subjective question
				subjJson, _ := json.Marshal(structs.SubjectiveQuestion{
					Statement: "Subjective Question",
				})
				assert.True(t, compareJSONByteArrays(subjJson, resp.Questions[1].RawQuestion))

				// Validate coding question
				var coding structs.CodingQuestion
				require.NoError(t, json.Unmarshal(resp.Questions[2].RawQuestion, &coding))
				assert.Equal(t, "Sum Numbers", coding.Title)
				assert.Equal(t, "Add two numbers", coding.Statement)
				assert.Len(t, coding.InputDefinitions, 2)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Paper not found",
				userID:       userID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) (*models.Paper, []models.Question) {
				return &models.Paper{ID: 999}, nil
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "No access to paper",
				userID:       userID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) (*models.Paper, []models.Question) {
				paper := createTestPaper(t, 2) // Create paper owned by different user
				return &paper, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, _ := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.PaperRequest{PaperId: int64(paper.ID)},
				client.GetPaperQuestions,
				tt.validate,
			)
		})
	}
}

func TestDeleteQuestion(t *testing.T) {
	tests := []DeleteQuestionCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Delete MCQ question",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				// Set initial question counts
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 1, "subjective": 1}`).Error
				require.NoError(t, err)

				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Test MCQ","options":["A","B"]}`),
					},
				})
				return &paper, &questions[0]
			},
			validate: func(t *testing.T, paper *models.Paper, questionID types.QuestionID) {
				// Verify question was deleted
				var question models.Question
				err := db.DB.First(&question, questionID).Error
				assert.Error(t, err) // Question should not exist

				// Verify question counts were updated
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.EqualValues(t, 0, counts.MCQ, "MCQ count should decrease")
				assert.EqualValues(t, 1, counts.Subjective)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Delete locked question unlinks it",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				// Set initial question counts
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 2, "subjective": 1}`).Error
				require.NoError(t, err)

				category := createDefaultTestCategory(t, paper.ID)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
					CategoryID: category.ID,
					Order:      1,
					Type:       constants.QUESTION_TYPE_MCQ,
					Question:   json.RawMessage(`{"statement":"Test MCQ","options":["A","B"]}`),

					Locked: true,
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			validate: func(t *testing.T, paper *models.Paper, questionID types.QuestionID) {
				// Verify question is unlinked but exists
				var question models.Question
				require.NoError(t, db.DB.First(&question, questionID).Error)
				assert.True(t, question.Locked)
				assert.Zero(t, question.PaperID)

				// Verify question counts were updated
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.EqualValues(t, 1, counts.MCQ, "MCQ count should decrease")
				assert.EqualValues(t, 1, counts.Subjective)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, question := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.QuestionRequest{QuestionId: int64(question.ID)},
				client.DeleteQuestion,
				func(t *testing.T, resp *proto.Empty) {
					if tt.validate != nil {
						tt.validate(t, paper, question.ID)
					}
				})
		})
	}
}

func TestReorderQuestions(t *testing.T) {
	tests := []ReorderQuestionsCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Reorder questions",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory, []models.Question) {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Q1"}`),
					},
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Q2"}`),
					},
				})
				return &paper, &category, questions
			},
			request: &proto.ReorderQuestionsRequest{},
			validate: func(t *testing.T, questions []models.Question) {
				var updated []models.Question
				require.NoError(t, db.DB.Order("\"order\"").Find(&updated).Error)
				assert.Equal(t, questions[1].ID, updated[0].ID) // Q2 should be first
				assert.Equal(t, questions[0].ID, updated[1].ID) // Q1 should be second
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			_, category, questions := tt.setup(t)

			tt.request.CategoryId = int64(category.ID)
			tt.request.QuestionIds = []int64{int64(questions[1].ID), int64(questions[0].ID)} // Reverse order

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				tt.request,
				client.ReorderQuestions,
				func(t *testing.T, resp *proto.Empty) {
					if tt.validate != nil {
						tt.validate(t, questions)
					}
				})
		})
	}
}

func TestGetPaperQuestion(t *testing.T) {
	tests := []GetPaperQuestionCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get MCQ question",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"MCQ Question","options":["A","B","C"]}`),
						MaxScore:   5,
					},
				})
				return &paper, &questions[0]
			},
			validate: func(t *testing.T, resp *proto.QuestionResponse) {
				var mcq structs.MCQQuestion
				require.NoError(t, json.Unmarshal(resp.RawQuestion, &mcq))
				assert.Equal(t, "MCQ Question", mcq.Statement)
				assert.Equal(t, []string{"A", "B", "C"}, mcq.Options)
				assert.Equal(t, int32(5), resp.MaxScore)
				assert.Empty(t, resp.TestCases)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get coding question with test cases",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
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
						MaxScore: 10,
					},
				})

				createTestCases(t, []models.TestCase{
					{
						QuestionID: questions[0].ID,
						Content:    json.RawMessage(`{"inputs": ["1", "2"], "output": "3"}`),
						Hidden:     false,
					},
					{
						QuestionID: questions[0].ID,
						Content:    json.RawMessage(`{"inputs": ["4", "5"], "output": "9", "explanation": "hidden case"}`),
						Hidden:     true,
					},
				})

				return &paper, &questions[0]
			},
			validate: func(t *testing.T, resp *proto.QuestionResponse) {
				// Verify coding question content
				var coding structs.CodingQuestion
				require.NoError(t, json.Unmarshal(resp.RawQuestion, &coding))
				assert.Equal(t, "Sum Numbers", coding.Title)
				assert.Equal(t, "Add two numbers", coding.Statement)
				assert.Len(t, coding.InputDefinitions, 2)
				assert.Equal(t, int32(10), resp.MaxScore)

				// Verify test cases
				require.Len(t, resp.TestCases, 2)
				// First test case (non-hidden)
				assert.Equal(t, []string{"1", "2"}, resp.TestCases[0].Inputs)
				assert.Equal(t, "3", resp.TestCases[0].Output)
				assert.False(t, resp.TestCases[0].Hidden)
				assert.Nil(t, resp.TestCases[0].Explanation)

				// Second test case (hidden)
				assert.Equal(t, []string{"4", "5"}, resp.TestCases[1].Inputs)
				assert.Equal(t, "9", resp.TestCases[1].Output)
				assert.True(t, resp.TestCases[1].Hidden)
				assert.NotNil(t, resp.TestCases[1].Explanation)
				assert.Equal(t, "hidden case", *resp.TestCases[1].Explanation)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Question not found",
				userID:       userID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				return nil, &models.Question{ID: 999}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "No access to paper",
				userID:       userID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, 2) // different user
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_SUBJECTIVE,
						Question:   json.RawMessage(`{"statement":"Q1"}`),
					},
				})
				return &paper, &questions[0]
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			_, question := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.QuestionRequest{QuestionId: int64(question.ID)},
				client.GetPaperQuestion,
				tt.validate,
			)
		})
	}
}
