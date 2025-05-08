package handlers

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/paper/internal/config/db"
)

func TestGetPaperQuestions(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.Paper, []models.Question)
		userID       int64
		expectedCode codes.Code
		validate     func(t *testing.T, questions []models.Question, resp *proto.QuestionList)
	}{
		{
			name: "Success - Get questions for paper",
			setup: func(t *testing.T) (*models.Paper, []models.Question) {
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
				return &paper, questions
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, questions []models.Question, resp *proto.QuestionList) {
				assert.Equal(t, len(questions), len(resp.Questions))
				assert.Equal(t, "MCQ Question", resp.Questions[0].GetMcq().Statement)
				assert.Equal(t, "Short Question", resp.Questions[1].GetGeneral().Statement)
			},
		},
		{
			name: "Paper not found",
			setup: func(t *testing.T) (*models.Paper, []models.Question) {
				return &models.Paper{ID: 999}, nil
			},
			userID:       userID,
			expectedCode: codes.NotFound,
		},
		{
			name: "No access to paper",
			setup: func(t *testing.T) (*models.Paper, []models.Question) {
				paper := createTestPaper(t, 2) // Create paper owned by different user
				return &paper, nil
			},
			userID:       userID,
			expectedCode: codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, questions := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.GetPaperQuestions(ctx, &proto.PaperRequest{
				PaperId: paper.ID,
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, questions, resp)
		})
	}
}

func TestCreateQuestion(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.Paper, *models.QuestionCategory)
		userID       int64
		request      *proto.CreateQuestionRequest
		expectedCode codes.Code
		validate     func(t *testing.T, paper *models.Paper, resp *proto.QuestionResponse)
	}{
		{
			name: "Success - Create MCQ question",
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			userID: userID,
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{"statement":"Test MCQ","options":["A","B","C"]}`),
				Type:        constants.QUESTION_TYPE_MCQ,
				MaxScore:    5,
				CategoryId:  1,
			},
			expectedCode: codes.OK,
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
				assert.EqualValues(t, 0, counts.Short)
				assert.EqualValues(t, 0, counts.Long)
			},
		},
		{
			name: "Success - Create SHORT question",
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			userID: userID,
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{"statement":"Test Short Answer"}`),
				Type:        constants.QUESTION_TYPE_SHORT,
				MaxScore:    10,
				CategoryId:  1,
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper, resp *proto.QuestionResponse) {
				// Validate question response
				assert.Equal(t, "Test Short Answer", resp.GetGeneral().Statement)
				assert.EqualValues(t, 10, resp.MaxScore)
				assert.Equal(t, constants.QUESTION_TYPE_SHORT, resp.Type)

				// Validate paper's question counts
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.EqualValues(t, 0, counts.MCQ)
				assert.EqualValues(t, 1, counts.Short)
				assert.EqualValues(t, 0, counts.Long)
			},
		},
		{
			name: "Success - Create LONG question",
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			userID: userID,
			request: &proto.CreateQuestionRequest{
				RawQuestion: []byte(`{"statement":"Test Long Answer"}`),
				Type:        constants.QUESTION_TYPE_LONG,
				MaxScore:    15,
				CategoryId:  1,
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper, resp *proto.QuestionResponse) {
				// Validate question response
				assert.Equal(t, "Test Long Answer", resp.GetGeneral().Statement)
				assert.EqualValues(t, 15, resp.MaxScore)
				assert.Equal(t, constants.QUESTION_TYPE_LONG, resp.Type)

				// Validate paper's question counts
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.EqualValues(t, 0, counts.MCQ)
				assert.EqualValues(t, 0, counts.Short)
				assert.EqualValues(t, 1, counts.Long)
			},
		},
		{
			name: "Error - Max score too high",
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
			userID:       userID,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error - Negative max score",
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
			userID:       userID,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, category := tt.setup(t)
			tt.request.PaperId = paper.ID
			tt.request.CategoryId = category.ID

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.CreateQuestion(ctx, tt.request)

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, paper, resp)
		})
	}
}

func TestUpdateQuestion(t *testing.T) {
	maxScore := int32(10)
	questionTypeShort := constants.QUESTION_TYPE_SHORT

	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.Paper, *models.Question)
		request      *proto.UpdateQuestionRequest
		userID       int64
		expectedCode codes.Code
		validate     func(t *testing.T, paper *models.Paper, question *models.Question)
	}{
		{
			name: "Success - Update question without type change",
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 1, "short": 0, "long": 0}`).Error
				require.NoError(t, err)

				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
					CategoryID: category.ID,
					Order:      1,
					Type:       constants.QUESTION_TYPE_MCQ,
					Question:   json.RawMessage(`{"statement":"Old MCQ","options":["A","B"]}`),
					MaxScore:   5,
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			request: &proto.UpdateQuestionRequest{
				RawQuestion: []byte(`{"statement":"Updated MCQ","options":["X","Y","Z"]}`),
				MaxScore:    &maxScore,
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper, question *models.Question) {
				// Verify question was updated
				var updated models.Question
				require.NoError(t, db.DB.First(&updated, question.ID).Error)
				var mcq structs.MCQQuestion
				require.NoError(t, json.Unmarshal(updated.Question, &mcq))
				assert.Equal(t, "Updated MCQ", mcq.Statement)
				assert.Equal(t, []string{"X", "Y", "Z"}, mcq.Options)
				assert.EqualValues(t, 10, updated.MaxScore)

				// Verify question counts didn't change
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.EqualValues(t, 1, counts.MCQ)
				assert.EqualValues(t, 0, counts.Short)
				assert.EqualValues(t, 0, counts.Long)
			},
		},
		{
			name: "Success - Update question with type change",
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 1, "short": 0, "long": 0}`).Error
				require.NoError(t, err)

				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
					CategoryID: category.ID,
					Order:      1,
					Type:       constants.QUESTION_TYPE_MCQ,
					Question:   json.RawMessage(`{"statement":"Old MCQ","options":["A","B"]}`),
					MaxScore:   5,
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			request: &proto.UpdateQuestionRequest{
				Type:        &questionTypeShort,
				RawQuestion: []byte(`{"statement":"Updated to Short"}`),
				MaxScore:    &maxScore,
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper, question *models.Question) {
				// Verify question was updated
				var updated models.Question
				require.NoError(t, db.DB.First(&updated, question.ID).Error)
				assert.Equal(t, constants.QUESTION_TYPE_SHORT, updated.Type)

				// Verify question counts were updated
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.EqualValues(t, 0, counts.MCQ, "MCQ count should decrease")
				assert.EqualValues(t, 1, counts.Short, "Short count should increase")
				assert.EqualValues(t, 0, counts.Long)
			},
		},
		{
			name: "Success - Update locked question creates new copy",
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
			userID:       userID,
			expectedCode: codes.OK,
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
			name: "Error - Type change without question content",
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
					MaxScore:   5,
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			request: &proto.UpdateQuestionRequest{
				Type: &questionTypeShort, // Trying to change type without providing question
			},
			userID:       userID,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error - Type change with mismatched question content",
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
					MaxScore:   5,
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			request: &proto.UpdateQuestionRequest{
				Type:        &questionTypeShort,                                                   // Changing to SHORT type
				RawQuestion: []byte(`{"statement":"Wrong content type","options":["X","Y","Z"]}`), // But providing MCQ content
			},
			userID:       userID,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Success - Update MCQ question",
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
					MaxScore:   5,
					Tags:       json.RawMessage(`["old"]`),
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			request: &proto.UpdateQuestionRequest{
				MaxScore:    &maxScore,
				Tags:        []string{"updated", "mcq"},
				RawQuestion: []byte(`{"statement":"Updated MCQ Question","options":["X","Y","Z","W"]}`),
			},
			userID:       userID,
			expectedCode: codes.OK,
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
			name: "Success - Update Short question",
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
					CategoryID: category.ID,
					Order:      1,
					Type:       constants.QUESTION_TYPE_SHORT,
					Question:   json.RawMessage(`{"statement":"Old Short Question"}`),
					MaxScore:   5,
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			request: &proto.UpdateQuestionRequest{
				MaxScore:      &maxScore,
				RawQuestion:   []byte(`{"statement":"Updated Short Question"}`),
				CorrectAnswer: &[]string{"Expected answer"}[0],
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper, question *models.Question) {
				var updated models.Question
				require.NoError(t, db.DB.First(&updated, question.ID).Error)

				// Verify question content
				var general structs.GeneralQuestion
				require.NoError(t, json.Unmarshal(updated.Question, &general))
				assert.Equal(t, "Updated Short Question", general.Statement)

				// Verify other fields
				assert.EqualValues(t, 10, updated.MaxScore)
				assert.Equal(t, "Expected answer", updated.CorrectAnswer.String)
				assert.True(t, updated.CorrectAnswer.Valid)
			},
		},
		{
			name: "Error - Max score too high",
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
					MaxScore:   5,
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			request: &proto.UpdateQuestionRequest{
				MaxScore: &[]int32{1001}[0], // Exceeds maximum
			},
			userID:       userID,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error - Negative max score",
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
					MaxScore:   5,
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			request: &proto.UpdateQuestionRequest{
				MaxScore: &[]int32{-1}[0], // Negative score
			},
			userID:       userID,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, question := tt.setup(t)
			tt.request.QuestionId = question.ID

			ctx := createContextWithUserID(tt.userID)
			_, err := client.UpdateQuestion(ctx, tt.request)

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			tt.validate(t, paper, question)
		})
	}
}

func TestDeleteQuestion(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.Paper, *models.Question)
		userID       int64
		expectedCode codes.Code
		validate     func(t *testing.T, paper *models.Paper, questionID int64)
	}{
		{
			name: "Success - Delete MCQ question",
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				// Set initial question counts
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 1, "short": 1, "long": 1}`).Error
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
				}
				require.NoError(t, db.DB.Create(&question).Error)
				return &paper, &question
			},
			userID:       userID,
			expectedCode: codes.OK,
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
				assert.EqualValues(t, 1, counts.Short)
				assert.EqualValues(t, 1, counts.Long)
			},
		},
		{
			name: "Success - Delete locked question unlinks it",
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				// Set initial question counts
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 2, "short": 1, "long": 0}`).Error
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
			userID:       userID,
			expectedCode: codes.OK,
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
				assert.EqualValues(t, 1, counts.Short)
				assert.EqualValues(t, 0, counts.Long)
			},
		},
		// Add more test cases for other scenarios
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, question := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.DeleteQuestion(ctx, &proto.QuestionRequest{QuestionId: question.ID})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			tt.validate(t, paper, question.ID)
		})
	}
}

func TestReorderQuestions(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.Paper, *models.QuestionCategory, []models.Question)
		request      *proto.ReorderQuestionsRequest
		userID       int64
		expectedCode codes.Code
		validate     func(t *testing.T, questions []models.Question)
	}{
		{
			name: "Success - Reorder questions",
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory, []models.Question) {
				paper := createTestPaper(t, userID)
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Order:      1,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Q1"}`),
					},
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Order:      2,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Q2"}`),
					},
				}
				require.NoError(t, db.DB.Create(&questions).Error)
				return &paper, &category, questions
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, questions []models.Question) {
				var updated []models.Question
				require.NoError(t, db.DB.Order("\"order\"").Find(&updated).Error)
				assert.Equal(t, questions[1].ID, updated[0].ID) // Q2 should be first
				assert.Equal(t, questions[0].ID, updated[1].ID) // Q1 should be second
			},
		},
		// Add more test cases for other scenarios
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			_, category, questions := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.ReorderQuestions(ctx, &proto.ReorderQuestionsRequest{
				CategoryId:  category.ID,
				QuestionIds: []int64{questions[1].ID, questions[0].ID}, // Reverse order
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			tt.validate(t, questions)
		})
	}
}
