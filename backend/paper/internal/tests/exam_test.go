package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/utils/generate"
	"pariksha/common/pkg/utils/testrunner"
	"pariksha/paper/internal/config/env"
)

func TestGetQuestionsByIds(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) []models.Question
		request      *proto.GetQuestionsByIdsRequest
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.GetQuestionsByIdsResponse, questions []models.Question)
	}{
		{
			name: "Success - Get multiple questions",
			setup: func(t *testing.T) []models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				mcqContent, _ := json.Marshal(structs.MCQQuestion{
					Statement: "MCQ Question",
					Options:   []string{"A", "B", "C"},
				})
				subjectiveContent, _ := json.Marshal(structs.SubjectiveQuestion{
					Statement: "Subjective Question",
				})

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Order:      1,
						Type:       proto.QuestionType_MCQ,
						Question:   mcqContent,
						MaxScore:   5,
					},
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Order:      2,
						Type:       proto.QuestionType_SUBJECTIVE,
						Question:   subjectiveContent,
						MaxScore:   10,
					},
				})
				return questions
			},
			request: &proto.GetQuestionsByIdsRequest{
				QuestionIds: []int64{1, 2}, // IDs will be filled in test
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.GetQuestionsByIdsResponse, questions []models.Question) {
				require.Len(t, resp.Questions, 2)

				// Validate MCQ question
				mcqResp := resp.Questions[0]
				assert.EqualValues(t, questions[0].Hash, mcqResp.QuestionHash)
				assert.EqualValues(t, questions[0].Type, mcqResp.Type)
				assert.EqualValues(t, questions[0].MaxScore, mcqResp.MaxScore)
				assert.True(t, compareJSONByteArrays(questions[0].Question, mcqResp.RawQuestion))

				// Validate Subjective question
				subjectiveResp := resp.Questions[1]
				assert.EqualValues(t, questions[1].Hash, subjectiveResp.QuestionHash)
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
			validate: func(t *testing.T, resp *proto.GetQuestionsByIdsResponse, questions []models.Question) {
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

			// Create context with exam API token
			ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
				constants.X_EXAM_API_TOKEN: env.EXAM_API_TOKEN,
			}))

			testrunner.Runner(t, ctx, tt.expectedCode,
				tt.request,
				client.GetQuestionsByIds,
				func(t *testing.T, resp *proto.GetQuestionsByIdsResponse) {
					if tt.validate != nil {
						tt.validate(t, resp, questions)
					}
				},
			)
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
				categories := createDefaultTestCategories(t, paper.ID, 3)
				return categories
			},
			request: &proto.GetCategoriesByIdsRequest{
				CategoryIds: []int64{1, 2, 3},
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.CategoryBatchResponse, categories []models.QuestionCategory) {
				require.Len(t, resp.Categories, len(categories))
				for i, category := range resp.Categories {
					assert.EqualValues(t, categories[i].ID, category.CategoryId)
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
		validate     func(t *testing.T, resp *proto.ExamQuestionResponse)
	}{
		{
			name: "Success - Get MCQ question",
			setup: func(t *testing.T) *models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_MCQ,
						Question:   json.RawMessage(`{"statement":"MCQ Question","options":["A","B","C"]}`),
					},
				})
				return &questions[0]
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.ExamQuestionResponse) {
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
			validate: func(t *testing.T, resp *proto.ExamQuestionResponse) {
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
				return &models.Question{
					ID:   999,
					Hash: generate.HMACHash(999),
				}
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			question := tt.setup(t)

			testrunner.Runner(t, context.Background(), tt.expectedCode,
				&proto.QuestionRequest{QuestionHash: question.Hash},
				client.GetExamQuestionByHash,
				func(t *testing.T, resp *proto.ExamQuestionResponse) {
					if tt.validate != nil {
						tt.validate(t, resp)
					}
				})
		})
	}
}

func TestGetQuestionHashes(t *testing.T) {
	mcqContent, _ := json.Marshal(structs.MCQQuestion{
		Statement: "MCQ Question",
		Options:   []string{"A", "B", "C"},
	})
	subjectiveContent, _ := json.Marshal(structs.SubjectiveQuestion{
		Statement: "Subjective Question",
	})

	tests := []struct {
		name         string
		setup        func(t *testing.T) []models.Question
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.GetQuestionHashesResponse, questions []models.Question)
	}{
		{
			name: "Success - Get multiple question hashes in sequence",
			setup: func(t *testing.T) []models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)
				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_MCQ,
						Question:   mcqContent,
						MaxScore:   5,
					},
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_SUBJECTIVE,
						Question:   subjectiveContent,
						MaxScore:   5,
					},
				})
				return questions
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.GetQuestionHashesResponse, questions []models.Question) {
				require.Len(t, resp.QuestionHashes, len(questions))
				// Verify hashes are in same order as requested IDs
				for i, question := range questions {
					assert.Equal(t, question.Hash, resp.QuestionHashes[i])
				}
			},
		},
		{
			name: "Success - Request includes non-existent questions",
			setup: func(t *testing.T) []models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)
				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_MCQ,
						Question:   mcqContent,
						MaxScore:   5,
					},
				})
				questions = append(
					questions,
					models.Question{ID: 999}, // Non-existent question
				)
				return questions
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.GetQuestionHashesResponse, questions []models.Question) {
				require.Len(t, resp.QuestionHashes, 2)
				assert.Equal(t, questions[0].Hash, resp.QuestionHashes[0])
				assert.Empty(t, resp.QuestionHashes[1]) // Non-existent question should return empty hash
			},
		},
		{
			name: "Success - Empty request",
			setup: func(t *testing.T) []models.Question {
				return nil
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.GetQuestionHashesResponse, questions []models.Question) {
				assert.Empty(t, resp.QuestionHashes)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			questions := tt.setup(t)

			request := &proto.GetQuestionHashesRequest{}
			if len(questions) > 0 {
				request.QuestionIds = make([]int64, 0, len(questions))
				for _, q := range questions {
					request.QuestionIds = append(request.QuestionIds, int64(q.ID))
				}
			}

			// Create context with exam API token
			ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
				constants.X_EXAM_API_TOKEN: env.EXAM_API_TOKEN,
			}))

			testrunner.Runner(t, ctx, tt.expectedCode,
				request,
				client.GetQuestionHashes,
				func(t *testing.T, resp *proto.GetQuestionHashesResponse) {
					if tt.validate != nil {
						tt.validate(t, resp, questions)
					}
				})
		})
	}
}

func TestGetQuestionIds(t *testing.T) {
	mcqContent, _ := json.Marshal(structs.MCQQuestion{
		Statement: "MCQ Question",
		Options:   []string{"A", "B", "C"},
	})
	subjectiveContent, _ := json.Marshal(structs.SubjectiveQuestion{
		Statement: "Subjective Question",
	})

	tests := []struct {
		name         string
		setup        func(t *testing.T) []models.Question
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.GetQuestionIdsResponse, questions []models.Question)
	}{
		{
			name: "Success - Get multiple question IDs in sequence",
			setup: func(t *testing.T) []models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)
				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_MCQ,
						Question:   mcqContent,
						MaxScore:   5,
					},
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_SUBJECTIVE,
						Question:   subjectiveContent,
						MaxScore:   5,
					},
				})
				return questions
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.GetQuestionIdsResponse, questions []models.Question) {
				require.Len(t, resp.QuestionIds, len(questions))
				// Verify IDs are in same order as requested hashes
				for i, question := range questions {
					assert.Equal(t, int64(question.ID), resp.QuestionIds[i])
				}
			},
		},
		{
			name: "Success - Request includes non-existent questions",
			setup: func(t *testing.T) []models.Question {
				paper := createTestPaper(t, userID)
				category := createDefaultTestCategory(t, paper.ID)
				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Type:       proto.QuestionType_MCQ,
						Question:   mcqContent,
						MaxScore:   5,
					},
				})
				return questions
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.GetQuestionIdsResponse, questions []models.Question) {
				require.Len(t, resp.QuestionIds, 1)
				assert.Equal(t, int64(questions[0].ID), resp.QuestionIds[0])
			},
		},
		{
			name: "Success - Empty request",
			setup: func(t *testing.T) []models.Question {
				return nil
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.GetQuestionIdsResponse, questions []models.Question) {
				assert.Empty(t, resp.QuestionIds)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			questions := tt.setup(t)

			request := &proto.GetQuestionIdsRequest{}
			if len(questions) > 0 {
				request.QuestionHashes = make([]string, 0, len(questions))
				for _, q := range questions {
					request.QuestionHashes = append(request.QuestionHashes, q.Hash)
				}
			}

			// Create context with exam API token
			ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
				constants.X_EXAM_API_TOKEN: env.EXAM_API_TOKEN,
			}))

			testrunner.Runner(t, ctx, tt.expectedCode,
				request,
				client.GetQuestionIds,
				func(t *testing.T, resp *proto.GetQuestionIdsResponse) {
					if tt.validate != nil {
						tt.validate(t, resp, questions)
					}
				})
		})
	}
}
