package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/paper/internal/config/db"
	"testing"

	"pariksha/common/pkg/utils/testrunner"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetQuestionsByIds(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) []models.Question
		request      *proto.GetQuestionsByIdsRequest
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.QuestionBatchResponse, questions []models.Question)
	}{
		{
			name: "Success - Get multiple questions",
			setup: func(t *testing.T) []models.Question {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				mcqContent, _ := json.Marshal(structs.MCQQuestion{
					Statement: "MCQ Question",
					Options:   []string{"A", "B", "C"},
				})
				subjectiveContent, _ := json.Marshal(structs.SubjectiveQuestion{
					Statement: "Subjective Question",
				})

				questions := []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Order:      1,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   mcqContent,
						MaxScore:   5,
					},
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Order:      2,
						Type:       constants.QUESTION_TYPE_SUBJECTIVE,
						Question:   subjectiveContent,
						MaxScore:   10,
					},
				}
				require.NoError(t, db.DB.Create(&questions).Error)
				return questions
			},
			request: &proto.GetQuestionsByIdsRequest{
				QuestionIds: []int64{1, 2}, // IDs will be filled in test
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.QuestionBatchResponse, questions []models.Question) {
				require.Len(t, resp.Questions, 2)

				// Validate MCQ question
				mcqResp := resp.Questions[0]
				assert.EqualValues(t, questions[0].ID, mcqResp.Id)
				assert.EqualValues(t, questions[0].Type, mcqResp.Type)
				assert.EqualValues(t, questions[0].MaxScore, mcqResp.MaxScore)
				assert.True(t, compareJSONByteArrays(questions[0].Question, mcqResp.RawQuestion))

				// Validate Subjective question
				subjectiveResp := resp.Questions[1]
				assert.EqualValues(t, questions[1].ID, subjectiveResp.Id)
				assert.EqualValues(t, questions[1].Type, subjectiveResp.Type)
				assert.EqualValues(t, questions[1].MaxScore, subjectiveResp.MaxScore)
				assert.True(t, compareJSONByteArrays(questions[1].Question, subjectiveResp.RawQuestion))
			},
		},
		{
			name: "Success - Empty result for non-existent IDs",
			setup: func(t *testing.T) []models.Question {
				return []models.Question{}
			},
			request: &proto.GetQuestionsByIdsRequest{
				QuestionIds: []int64{999, 1000},
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.QuestionBatchResponse, questions []models.Question) {
				assert.Empty(t, resp.Questions)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			questions := tt.setup(t)
			if len(questions) > 0 {
				tt.request.QuestionIds = make([]int64, len(questions))
				for i, q := range questions {
					tt.request.QuestionIds[i] = int64(q.ID)
				}
			}

			testrunner.Runner(t, context.Background(), tt.expectedCode,
				tt.request,
				client.GetQuestionsByIds,
				func(t *testing.T, resp *proto.QuestionBatchResponse) {
					if tt.validate != nil {
						tt.validate(t, resp, questions)
					}
				})
		})
	}
}

func TestGetCategoriesByIds(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) []models.QuestionCategory
		request      *proto.GetCategoriesByIdsRequest
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.CategoryBatchResponse, categories []models.QuestionCategory)
	}{
		{
			name: "Success - Get multiple categories",
			setup: func(t *testing.T) []models.QuestionCategory {
				paper := createTestPaper(t, userID)
				var defaultCategory models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&defaultCategory).Error)

				categories := []models.QuestionCategory{
					defaultCategory,
					{
						PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						Name:    "Category 2",
						Order:   2,
					},
					{
						PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						Name:    "Category 3",
						Order:   3,
					},
				}
				// Skip first category as it's already created
				require.NoError(t, db.DB.Create(categories[1:]).Error)
				return categories
			},
			request: &proto.GetCategoriesByIdsRequest{
				CategoryIds: []int64{1, 2, 3},
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.CategoryBatchResponse, categories []models.QuestionCategory) {
				require.Len(t, resp.Categories, len(categories))
				for i, category := range resp.Categories {
					assert.EqualValues(t, categories[i].ID, category.Id)
					assert.Equal(t, categories[i].Name, category.Name)
				}
			},
		},
		{
			name: "Success - Empty result for non-existent IDs",
			setup: func(t *testing.T) []models.QuestionCategory {
				return []models.QuestionCategory{}
			},
			request: &proto.GetCategoriesByIdsRequest{
				CategoryIds: []int64{999, 1000},
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.CategoryBatchResponse, categories []models.QuestionCategory) {
				assert.Empty(t, resp.Categories)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			categories := tt.setup(t)
			if len(categories) > 0 {
				tt.request.CategoryIds = make([]int64, len(categories))
				for i, c := range categories {
					tt.request.CategoryIds[i] = int64(c.ID)
				}
			}

			resp, err := client.GetCategoriesByIds(context.Background(), tt.request)

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp, categories)
		})
	}
}

func TestGetExamQuestion(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Question
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.QuestionResponse)
	}{
		{
			name: "Success - Get MCQ question",
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"MCQ Question","options":["A","B","C"]}`),
					},
				})
				return &questions[0]
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.QuestionResponse) {
				var mcq structs.MCQQuestion
				require.NoError(t, json.Unmarshal(resp.RawQuestion, &mcq))
				assert.Equal(t, "MCQ Question", mcq.Statement)
				assert.Equal(t, []string{"A", "B", "C"}, mcq.Options)
				assert.Empty(t, resp.TestCases)
			},
		},
		{
			name: "Success - Get coding question with only non-hidden test cases",
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

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
					},
				})

				// Create test cases with both hidden and non-hidden
				createTestCases(t, []models.TestCase{
					{
						QuestionID: questions[0].ID,
						Content:    json.RawMessage(`{"inputs": ["1", "2"], "output": "3"}`),
						Hidden:     false,
					},
					{
						QuestionID: questions[0].ID,
						Content:    json.RawMessage(`{"inputs": ["4", "5"], "output": "9", "explanation": "hidden case"}`),
						Hidden:     true, // This should not appear in response
					},
				})

				return &questions[0]
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.QuestionResponse) {
				// Verify coding question content
				var coding structs.CodingQuestion
				require.NoError(t, json.Unmarshal(resp.RawQuestion, &coding))
				assert.Equal(t, "Sum Numbers", coding.Title)
				assert.Equal(t, "Add two numbers", coding.Statement)

				// Verify only non-hidden test cases are included
				require.Len(t, resp.TestCases, 1, "Should only include non-hidden test cases")
				assert.Equal(t, []string{"1", "2"}, resp.TestCases[0].Inputs)
				assert.Equal(t, "3", resp.TestCases[0].Output)
				assert.False(t, resp.TestCases[0].Hidden)
			},
		},
		{
			name: "Question not found",
			setup: func(t *testing.T) *models.Question {
				return &models.Question{ID: 999}
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			question := tt.setup(t)

			testrunner.Runner(t, context.Background(), tt.expectedCode,
				&proto.QuestionRequest{QuestionId: int64(question.ID)},
				client.GetExamQuestion,
				func(t *testing.T, resp *proto.QuestionResponse) {
					if tt.validate != nil {
						tt.validate(t, resp)
					}
				})
		})
	}
}
