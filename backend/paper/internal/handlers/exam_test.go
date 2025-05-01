package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
	"testing"

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

				questions := []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Order:      1,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"MCQ Question","options":["A","B","C"]}`),
						MaxScore:   5,
					},
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Order:      2,
						Type:       constants.QUESTION_TYPE_SHORT,
						Question:   json.RawMessage(`{"statement":"Short Question"}`),
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
				assert.Equal(t, questions[0].ID, mcqResp.Id)
				assert.Equal(t, questions[0].Type, mcqResp.Type)
				assert.Equal(t, int32(questions[0].MaxScore), mcqResp.MaxScore)
				assert.Equal(t, "MCQ Question", mcqResp.GetMcq().Statement)
				assert.Equal(t, []string{"A", "B", "C"}, mcqResp.GetMcq().Options)

				// Validate Short question
				shortResp := resp.Questions[1]
				assert.Equal(t, questions[1].ID, shortResp.Id)
				assert.Equal(t, questions[1].Type, shortResp.Type)
				assert.Equal(t, int32(questions[1].MaxScore), shortResp.MaxScore)
				assert.Equal(t, "Short Question", shortResp.GetGeneral().Statement)
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
					tt.request.QuestionIds[i] = q.ID
				}
			}

			resp, err := client.GetQuestionsByIds(context.Background(), tt.request)

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp, questions)
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
						PaperID: sql.NullInt64{Int64: paper.ID, Valid: true},
						Name:    "Category 2",
						Order:   2,
					},
					{
						PaperID: sql.NullInt64{Int64: paper.ID, Valid: true},
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
					assert.Equal(t, categories[i].ID, category.Id)
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
					tt.request.CategoryIds[i] = c.ID
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
