package tests

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/utils/testrunner"
	"pariksha/paper/internal/config/db"
)

func TestCreateMcqQuestion(t *testing.T) {
	tests := []CreateQuestionCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Create MCQ question",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{"statement":"Test MCQ","options":["A","B","C"]}`),
				Type:        int32(constants.QUESTION_TYPE_MCQ),
				MaxScore:    5,
				CategoryId:  1,
			},
			validate: func(t *testing.T, paper *models.Paper, resp *proto.CreateQuestionResponse) {
				// Fetch the created question from database
				var question models.Question
				require.NoError(t, db.DB.First(&question, resp.Id).Error)

				// Validate question data
				assert.True(t, compareJSONByteArrays([]byte(`{"statement":"Test MCQ","options":["A","B","C"]}`), question.Question))
				assert.Equal(t, constants.QUESTION_TYPE_MCQ, question.Type)
				assert.EqualValues(t, 5, question.MaxScore)

				// Validate paper's question counts
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.EqualValues(t, 1, counts.MCQ)
				assert.EqualValues(t, 0, counts.Subjective)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, category := tt.setup(t)
			tt.request.PaperId = int64(paper.ID)
			tt.request.CategoryId = int64(category.ID)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				tt.request,
				client.CreateQuestion,
				func(t *testing.T, resp *proto.CreateQuestionResponse) {
					if tt.validate != nil {
						tt.validate(t, paper, resp)
					}
				})
		})
	}
}

func TestCreateSubjectiveQuestion(t *testing.T) {
	tests := []CreateQuestionCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Create SUBJECTIVE question",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{"statement":"Test Subjective Answer"}`),
				Type:        int32(constants.QUESTION_TYPE_SUBJECTIVE),
				MaxScore:    10,
				CategoryId:  1,
			},
			validate: func(t *testing.T, paper *models.Paper, resp *proto.CreateQuestionResponse) {
				// Fetch the created question from database
				var question models.Question
				require.NoError(t, db.DB.First(&question, resp.Id).Error)

				// Validate question data
				assert.True(t, compareJSONByteArrays([]byte(`{"statement":"Test Subjective Answer"}`), question.Question))
				assert.Equal(t, constants.QUESTION_TYPE_SUBJECTIVE, question.Type)
				assert.EqualValues(t, 10, question.MaxScore)

				// Validate paper's question counts
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.EqualValues(t, 0, counts.MCQ)
				assert.EqualValues(t, 1, counts.Subjective)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, category := tt.setup(t)
			tt.request.PaperId = int64(paper.ID)
			tt.request.CategoryId = int64(category.ID)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				tt.request,
				client.CreateQuestion,
				func(t *testing.T, resp *proto.CreateQuestionResponse) {
					if tt.validate != nil {
						tt.validate(t, paper, resp)
					}
				})
		})
	}
}

func TestCreateCodingQuestion(t *testing.T) {
	tests := []CreateQuestionCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Create coding question with primitive inputs",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{
					"title": "Sum of Numbers",
					"statement": "Write a program to add two numbers",
					"input_definitions": [
						{ "variable_name": "a", "type": 1 },
						{ "variable_name": "b",  "type": 1 }
					],
					"output_definition": {
						"variable_name": "sum",
						"type": 1
					}
				}`),
				Type:     int32(constants.QUESTION_TYPE_CODING),
				MaxScore: 15,
			},
			validate: func(t *testing.T, paper *models.Paper, resp *proto.CreateQuestionResponse) {
				// Fetch the created question from database
				var question models.Question
				require.NoError(t, db.DB.First(&question, resp.Id).Error)

				// Validate question data
				assert.Equal(t, constants.QUESTION_TYPE_CODING, question.Type)
				assert.EqualValues(t, 15, question.MaxScore)

				var codingQ structs.CodingQuestion
				err := json.Unmarshal(question.Question, &codingQ)
				require.NoError(t, err)
				assert.Equal(t, "Sum of Numbers", codingQ.Title)
				assert.Equal(t, "Write a program to add two numbers", codingQ.Statement)

				// Validate paper's question counts were updated
				verifyQuestionCounts(t, paper.ID, &models.QuestionCount{Coding: 1})
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Create coding question with array output",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{
					"title": "Generate Sequence",
					"statement": "Generate a sequence of numbers",
					"input_definitions": [
						{ "variable_name": "n", "type": 1 }
					],
					"output_definition": {
						"variable_name": "sequence",
						"type": 4,
						"items": [{ "type": 1 }]
					}
				}`),
				Type:     int32(constants.QUESTION_TYPE_CODING),
				MaxScore: 15,
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Missing output definition",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{
					"title": "Sum Numbers",
					"statement": "Add numbers",
					"input_definitions": [
						{ "variable_name": "arr", "type": 4, "items": [{ "type": 1 }] }
					]
				}`),
				Type:     int32(constants.QUESTION_TYPE_CODING),
				MaxScore: 15,
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Missing variable name",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper, category := setupTestCategory(t, userID, false)
				return paper, category
			},
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{
					"title": "Array Sum",
					"statement": "Write a program to sum an array",
					"input_definitions": [
						{
							"type": 4,
							"items": [{ "type": 1 }]
						}
					]
				}`),
				Type:     int32(constants.QUESTION_TYPE_CODING),
				MaxScore: 15,
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Array input with property_name",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper, category := setupTestCategory(t, userID, false)
				return paper, category
			},
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{
					"title": "Invalid Array",
					"statement": "Test",
					"input_definitions": [
						{
							"variable_name": "arr",
							"type": 4,
							"items": [{ "property_name": "invalid", "type": 1 }]
						}
					]
				}`),
				Type:     int32(constants.QUESTION_TYPE_CODING),
				MaxScore: 15,
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Array without items",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper, category := setupTestCategory(t, userID, false)
				return paper, category
			},
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{
					"title": "Invalid Array",
					"statement": "Test",
					"input_definitions": [{ "variable_name": "arr", "type": 4 }]
				}`),
				Type:     int32(constants.QUESTION_TYPE_CODING),
				MaxScore: 15,
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Array with non-primitive item type",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper, category := setupTestCategory(t, userID, false)
				return paper, category
			},
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{
					"title": "Invalid Array",
					"statement": "Test",
					"input_definitions": [
						{
							"variable_name": "arr",
							"type": 4,
							"items": [{ "type": 4 }]
						}
					]
				}`),
				Type:     int32(constants.QUESTION_TYPE_CODING),
				MaxScore: 15,
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Missing title",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{
					"title": "",
					"statement": "Write a program",
					"input_definitions": [{ "variable_name": "str", "type": 2 }]
				}`),
				Type: int32(constants.QUESTION_TYPE_CODING),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, category := tt.setup(t)
			tt.request.PaperId = int64(paper.ID)
			tt.request.CategoryId = int64(category.ID)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				tt.request,
				client.CreateQuestion,
				func(t *testing.T, resp *proto.CreateQuestionResponse) {
					if tt.validate != nil {
						tt.validate(t, paper, resp)
					}
				})
		})
	}
}

func TestCreateCodingQuestionBoilerplates(t *testing.T) {
	tests := []struct {
		name     string
		question string
		want     string
	}{
		{
			name: "Two number inputs",
			question: `{
				"title": "Add Numbers",
				"statement": "Add two numbers",
				"input_definitions": [
					{"variable_name": "a", "type": 1},
					{"variable_name": "b", "type": 1}
				],
				"output_definition": {"type": 1}
			}`,
			want: "function solve(a, b) {\n\n}",
		},
		{
			name: "Single array input",
			question: `{
				"title": "Sum Array",
				"statement": "Sum array elements",
				"input_definitions": [
					{"variable_name": "arr", "type": 4, "items": [{"type": 1}]}
				],
				"output_definition": {"type": 1}
			}`,
			want: "function solve(arr) {\n\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)

			// Create test language
			lang := models.Language{
				ID:        1,
				Slug:      constants.LangNode,
				Name:      "Node.js",
				Extension: "js",
				Version:   "16.x",
				IsEnabled: "true",
			}
			require.NoError(t, db.DB.Create(&lang).Error)

			// Create test paper and category
			paper, category := setupTestCategory(t, userID, false)

			// Create coding question
			ctx := createContextWithUserID(userID)
			req := &proto.CreateQuestionRequest{
				PaperId:     int64(paper.ID),
				CategoryId:  int64(category.ID),
				Type:        int32(constants.QUESTION_TYPE_CODING),
				RawQuestion: []byte(tt.question),
				MaxScore:    10,
			}

			var resp *proto.CreateQuestionResponse
			testrunner.Runner(t, ctx, codes.OK, req, client.CreateQuestion, func(t *testing.T, r *proto.CreateQuestionResponse) {
				resp = r
			})

			// Verify boilerplate was created
			var boilerplate models.Boilerplate
			require.NoError(t, db.DB.First(&boilerplate, "question_id = ?", resp.Id).Error)
			assert.Equal(t, tt.want, boilerplate.Code)
		})
	}
}

func TestGeneralCreateQuestion(t *testing.T) {
	tests := []CreateQuestionCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Max score too high",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{"statement":"Test MCQ","options":["A","B","C"]}`),
				Type:        int32(constants.QUESTION_TYPE_MCQ),
				MaxScore:    1001, // Exceeds maximum
				CategoryId:  1,
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Negative max score",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{"statement":"Test MCQ","options":["A","B","C"]}`),
				Type:        int32(constants.QUESTION_TYPE_MCQ),
				MaxScore:    -1, // Negative score
				CategoryId:  1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, category := tt.setup(t)
			tt.request.PaperId = int64(paper.ID)
			tt.request.CategoryId = int64(category.ID)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				tt.request,
				client.CreateQuestion,
				func(t *testing.T, resp *proto.CreateQuestionResponse) {
					if tt.validate != nil {
						tt.validate(t, paper, resp)
					}
				})
		})
	}
}
