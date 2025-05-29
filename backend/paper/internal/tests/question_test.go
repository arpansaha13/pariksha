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
	"pariksha/common/pkg/utils/testrunner"
	"pariksha/paper/internal/config/db"
)

func TestGetPaperQuestions(t *testing.T) {
	tests := []QuestionListCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get questions for paper",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, []models.Question) {
				paper, category := setupTestCategory(t, userID, false)
				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"MCQ Question","options":["A","B","C"]}`),
					},
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_SUBJECTIVE,
						Question:   json.RawMessage(`{"statement":"Subjective Question"}`),
					},
				})
				return paper, questions
			},
			validate: func(t *testing.T, resp *proto.QuestionList) {
				assert.Equal(t, 2, len(resp.Questions))

				// Validate first question (MCQ)
				mcqJson, _ := json.Marshal(structs.MCQQuestion{
					Statement: "MCQ Question",
					Options:   []string{"A", "B", "C"},
				})
				assert.Equal(t, mcqJson, resp.Questions[0].RawQuestion)

				// Validate second question (Subjective)
				subjJson, _ := json.Marshal(structs.SubjectiveQuestion{
					Statement: "Subjective Question",
				})
				assert.Equal(t, subjJson, resp.Questions[1].RawQuestion)
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
				&proto.PaperRequest{PaperId: paper.ID},
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
				return &paper, &questions[0]
			},
			validate: func(t *testing.T, paper *models.Paper, questionID int64) {
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

				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
					CategoryID: category.ID,
					Order:      1,
					Type:       constants.QUESTION_TYPE_MCQ,
					Question:   json.RawMessage(`{"statement":"Test MCQ","options":["A","B"]}`),

					Locked: true,
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			validate: func(t *testing.T, paper *models.Paper, questionID int64) {
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
				&proto.QuestionRequest{QuestionId: question.ID},
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
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Q1"}`),
					},
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
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

			tt.request.CategoryId = category.ID
			tt.request.QuestionIds = []int64{questions[1].ID, questions[0].ID} // Reverse order

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
