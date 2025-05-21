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
				assert.Equal(t, "MCQ Question", resp.Questions[0].GetMcq().Statement)
				assert.Equal(t, "Subjective Question", resp.Questions[1].GetSubjective().Statement)
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

func TestCreateQuestion(t *testing.T) {
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
				Type:        constants.QUESTION_TYPE_MCQ,
				MaxScore:    5,
				CategoryId:  1,
			},
			validate: func(t *testing.T, paper *models.Paper, resp *proto.QuestionResponse) {
				// Validate question response
				assert.Equal(t, "Test MCQ", resp.GetMcq().Statement)
				assert.Equal(t, []string{"A", "B", "C"}, resp.GetMcq().Options)
				assert.EqualValues(t, 5, resp.MaxScore)

				// Validate paper's question counts were updated
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.EqualValues(t, 1, counts.MCQ)
				assert.EqualValues(t, 0, counts.Subjective)
			},
		},
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
				Type:        constants.QUESTION_TYPE_SUBJECTIVE,
				MaxScore:    10,
				CategoryId:  1,
			},
			validate: func(t *testing.T, paper *models.Paper, resp *proto.QuestionResponse) {
				// Validate question response
				assert.Equal(t, "Test Subjective Answer", resp.GetSubjective().Statement)
				assert.EqualValues(t, 10, resp.MaxScore)
				assert.Equal(t, constants.QUESTION_TYPE_SUBJECTIVE, resp.Type)

				// Validate paper's question counts
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.EqualValues(t, 0, counts.MCQ)
				assert.EqualValues(t, 1, counts.Subjective)
			},
		},
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
				Type:        constants.QUESTION_TYPE_MCQ,
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
				Type:        constants.QUESTION_TYPE_MCQ,
				MaxScore:    -1, // Negative score
				CategoryId:  1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, category := tt.setup(t)
			tt.request.PaperId = paper.ID
			tt.request.CategoryId = category.ID

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				tt.request,
				client.CreateQuestion,
				func(t *testing.T, resp *proto.QuestionResponse) {
					if tt.validate != nil {
						tt.validate(t, paper, resp)
					}
				})
		})
	}
}

func TestUpdateQuestion(t *testing.T) {
	maxScore := int32(10)
	questionTypeSubjective := constants.QUESTION_TYPE_SUBJECTIVE

	tests := []UpdateQuestionCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update question without type change",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 1, "subjective": 0}`).Error
				require.NoError(t, err)

				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Test Question","options":["A","B","C"]}`),
					},
				})
				return &paper, &questions[0]
			},
			request: &proto.UpdateQuestionRequest{
				RawQuestion: []byte(`{"statement":"Updated MCQ","options":["X","Y","Z"]}`),
				MaxScore:    &maxScore,
			},
			validate: func(t *testing.T, paper *models.Paper, question *models.Question) {
				var updated models.Question
				require.NoError(t, db.DB.First(&updated, question.ID).Error)
				verifyMCQContent(t, updated, "Updated MCQ", []string{"X", "Y", "Z"})
				assert.EqualValues(t, 10, updated.MaxScore)

				verifyQuestionCounts(t, paper.ID, models.QuestionCount{
					MCQ:        1,
					Subjective: 0,
				})
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update question with type change",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 1, "subjective": 0}`).Error
				require.NoError(t, err)

				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Old MCQ","options":["A","B"]}`),
					},
				})
				return &paper, &questions[0]
			},
			request: &proto.UpdateQuestionRequest{
				Type:        &questionTypeSubjective,
				RawQuestion: []byte(`{"statement":"Updated to Subjective"}`),
				MaxScore:    &maxScore,
			},
			validate: func(t *testing.T, paper *models.Paper, question *models.Question) {
				// Verify question was updated
				var updated models.Question
				require.NoError(t, db.DB.First(&updated, question.ID).Error)
				assert.Equal(t, constants.QUESTION_TYPE_SUBJECTIVE, updated.Type)

				// Verify question counts were updated
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.EqualValues(t, 0, counts.MCQ, "MCQ count should decrease")
				assert.EqualValues(t, 1, counts.Subjective, "Subjective count should increase")
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update locked question creates new copy",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 1, "subjective": 0}`).Error
				require.NoError(t, err)

				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
					CategoryID: category.ID,
					Order:      1,
					Type:       constants.QUESTION_TYPE_MCQ,
					Question:   json.RawMessage(`{"statement":"Test MCQ","options":["A","B"]}`),
					MaxScore:   5,
					Locked:     true,
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			request: &proto.UpdateQuestionRequest{
				RawQuestion: []byte(`{"statement":"Updated MCQ","options":["X","Y","Z"]}`),
				MaxScore:    &maxScore,
			},
			validate: func(t *testing.T, paper *models.Paper, question *models.Question) {
				// Original question should be unlinked
				var original models.Question
				require.NoError(t, db.DB.First(&original, question.ID).Error)
				assert.True(t, original.Locked)
				assert.Zero(t, original.PaperID) // Unlinked

				// New question should be created
				var newQuestion models.Question
				require.NoError(t, db.DB.Where("paper_id = ?", question.PaperID).First(&newQuestion).Error)
				assert.NotEqualValues(t, question.ID, newQuestion.ID)
				var mcq structs.MCQQuestion
				require.NoError(t, json.Unmarshal(newQuestion.Question, &mcq))
				assert.Equal(t, "Updated MCQ", mcq.Statement)
				assert.Equal(t, []string{"X", "Y", "Z"}, mcq.Options)
				assert.EqualValues(t, 10, newQuestion.MaxScore)
				assert.False(t, newQuestion.Locked)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Type change without question content",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
					CategoryID: category.ID,
					Order:      1,
					Type:       constants.QUESTION_TYPE_MCQ,
					Question:   json.RawMessage(`{"statement":"Old MCQ","options":["A","B"]}`),
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			request: &proto.UpdateQuestionRequest{
				Type: &questionTypeSubjective, // Trying to change type without providing question
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Type change with mismatched question content",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
					CategoryID: category.ID,
					Order:      1,
					Type:       constants.QUESTION_TYPE_MCQ,
					Question:   json.RawMessage(`{"statement":"Old MCQ","options":["A","B"]}`),
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			request: &proto.UpdateQuestionRequest{
				// Changing to SUBJECTIVE type
				Type: &questionTypeSubjective,
				// But providing MCQ content
				RawQuestion: []byte(`{"statement":"Wrong content type","options":["X","Y","Z"]}`),
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update MCQ question",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Old MCQ","options":["A","B"]}`),

						Tags: json.RawMessage(`["old"]`),
					},
				})
				return &paper, &questions[0]
			},
			request: &proto.UpdateQuestionRequest{
				MaxScore:    &maxScore,
				Tags:        []string{"updated", "mcq"},
				RawQuestion: []byte(`{"statement":"Updated MCQ Question","options":["X","Y","Z","W"]}`),
			},
			validate: func(t *testing.T, paper *models.Paper, question *models.Question) {
				var updated models.Question
				require.NoError(t, db.DB.First(&updated, question.ID).Error)

				// Verify question content
				var mcq structs.MCQQuestion
				require.NoError(t, json.Unmarshal(updated.Question, &mcq))
				assert.Equal(t, "Updated MCQ Question", mcq.Statement)
				assert.Equal(t, []string{"X", "Y", "Z", "W"}, mcq.Options)

				// Verify other fields
				assert.EqualValues(t, 10, updated.MaxScore)
				var tags []string
				require.NoError(t, json.Unmarshal(updated.Tags, &tags))
				assert.ElementsMatch(t, []string{"updated", "mcq"}, tags)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update Subjective question",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_SUBJECTIVE,
						Question:   json.RawMessage(`{"statement":"Old Subjective Question"}`),
					},
				})
				return &paper, &questions[0]
			},
			request: &proto.UpdateQuestionRequest{
				MaxScore:      &maxScore,
				RawQuestion:   []byte(`{"statement":"Updated Subjective Question"}`),
				CorrectAnswer: &[]string{"Expected answer"}[0],
			},
			validate: func(t *testing.T, paper *models.Paper, question *models.Question) {
				var updated models.Question
				require.NoError(t, db.DB.First(&updated, question.ID).Error)

				// Verify question content
				var subjective structs.SubjectiveQuestion
				require.NoError(t, json.Unmarshal(updated.Question, &subjective))
				assert.Equal(t, "Updated Subjective Question", subjective.Statement)

				// Verify other fields
				assert.EqualValues(t, 10, updated.MaxScore)
				assert.Equal(t, "Expected answer", updated.CorrectAnswer.String)
				assert.True(t, updated.CorrectAnswer.Valid)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Max score too high",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Old MCQ","options":["A","B"]}`),
					},
				})
				return &paper, &questions[0]
			},
			request: &proto.UpdateQuestionRequest{
				MaxScore: &[]int32{1001}[0], // Exceeds maximum
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Error - Negative max score",
				userID:       userID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Old MCQ","options":["A","B"]}`),
					},
				})
				return &paper, &questions[0]
			},
			request: &proto.UpdateQuestionRequest{
				MaxScore: &[]int32{-1}[0], // Negative score
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, question := tt.setup(t)
			tt.request.QuestionId = question.ID

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				tt.request,
				client.UpdateQuestion,
				func(t *testing.T, resp *proto.Empty) {
					if tt.validate != nil {
						tt.validate(t, paper, question)
					}
				})
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
