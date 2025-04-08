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
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, questions []models.Question, resp *proto.QuestionList)
	}{
		{
			name: "Success - Get questions for paper",
			setup: func(t *testing.T) (*models.Paper, []models.Question) {
				paper := createTestPaper(t, int(userID))
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Order:      1,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"MCQ Question","options":["A","B","C"]}`),
						MaxScore:   5,
					},
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
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
				PaperId: int32(paper.ID),
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
		userID       int32
		request      *proto.CreateQuestionRequest
		expectedCode codes.Code
		validate     func(t *testing.T, paper *models.Paper, resp *proto.QuestionResponse)
	}{
		{
			name: "Success - Create MCQ question",
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, int(userID))
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			userID: userID,
			request: &proto.CreateQuestionRequest{
				Question: &proto.CreateQuestionRequest_Mcq{
					Mcq: &proto.McqQuestion{
						Statement: "Test MCQ",
						Options:   []string{"A", "B", "C"},
					},
				},
				Type:       constants.QUESTION_TYPE_MCQ,
				MaxScore:   5,
				CategoryId: 1,
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper, resp *proto.QuestionResponse) {
				// Validate question response
				assert.Equal(t, "Test MCQ", resp.GetMcq().Statement)
				assert.Equal(t, []string{"A", "B", "C"}, resp.GetMcq().Options)
				assert.Equal(t, int32(5), resp.MaxScore)

				// Validate paper's question counts were updated
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.Equal(t, 1, counts.MCQ)
				assert.Equal(t, 0, counts.Short)
				assert.Equal(t, 0, counts.Long)
			},
		},
		{
			name: "Success - Create SHORT question",
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, int(userID))
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			userID: userID,
			request: &proto.CreateQuestionRequest{
				Question: &proto.CreateQuestionRequest_General{
					General: &proto.GeneralQuestion{
						Statement: "Test Short Answer",
					},
				},
				Type:       constants.QUESTION_TYPE_SHORT,
				MaxScore:   10,
				CategoryId: 1,
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper, resp *proto.QuestionResponse) {
				// Validate question response
				assert.Equal(t, "Test Short Answer", resp.GetGeneral().Statement)
				assert.Equal(t, int32(10), resp.MaxScore)
				assert.Equal(t, constants.QUESTION_TYPE_SHORT, resp.Type)

				// Validate paper's question counts
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.Equal(t, 0, counts.MCQ)
				assert.Equal(t, 1, counts.Short)
				assert.Equal(t, 0, counts.Long)
			},
		},
		{
			name: "Success - Create LONG question",
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory) {
				paper := createTestPaper(t, int(userID))
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)
				return &paper, &category
			},
			userID: userID,
			request: &proto.CreateQuestionRequest{
				Question: &proto.CreateQuestionRequest_General{
					General: &proto.GeneralQuestion{
						Statement: "Test Long Answer",
					},
				},
				Type:       constants.QUESTION_TYPE_LONG,
				MaxScore:   15,
				CategoryId: 1,
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper, resp *proto.QuestionResponse) {
				// Validate question response
				assert.Equal(t, "Test Long Answer", resp.GetGeneral().Statement)
				assert.Equal(t, int32(15), resp.MaxScore)
				assert.Equal(t, constants.QUESTION_TYPE_LONG, resp.Type)

				// Validate paper's question counts
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.Equal(t, 0, counts.MCQ)
				assert.Equal(t, 0, counts.Short)
				assert.Equal(t, 1, counts.Long)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, category := tt.setup(t)
			tt.request.PaperId = int32(paper.ID)
			tt.request.CategoryId = int32(category.ID)

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
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, paper *models.Paper, question *models.Question)
	}{
		{
			name: "Success - Update question without type change",
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, int(userID))
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 1, "short": 0, "long": 0}`).Error
				require.NoError(t, err)

				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
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
				Question: &proto.UpdateQuestionRequest_Mcq{
					Mcq: &proto.McqQuestion{
						Statement: "Updated MCQ",
						Options:   []string{"X", "Y", "Z"},
					},
				},
				MaxScore: &maxScore,
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
				assert.Equal(t, 10, updated.MaxScore)

				// Verify question counts didn't change
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.Equal(t, 1, counts.MCQ)
				assert.Equal(t, 0, counts.Short)
				assert.Equal(t, 0, counts.Long)
			},
		},
		{
			name: "Success - Update question with type change",
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, int(userID))
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 1, "short": 0, "long": 0}`).Error
				require.NoError(t, err)

				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
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
				Type: &questionTypeShort,
				Question: &proto.UpdateQuestionRequest_General{
					General: &proto.GeneralQuestion{
						Statement: "Updated to Short",
					},
				},
				MaxScore: &maxScore,
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
				assert.Equal(t, 0, counts.MCQ, "MCQ count should decrease")
				assert.Equal(t, 1, counts.Short, "Short count should increase")
				assert.Equal(t, 0, counts.Long)
			},
		},
		{
			name: "Success - Update locked question creates new copy",
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, int(userID))
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
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
				Question: &proto.UpdateQuestionRequest_Mcq{
					Mcq: &proto.McqQuestion{
						Statement: "Updated MCQ",
						Options:   []string{"X", "Y", "Z"},
					},
				},
				MaxScore: &maxScore,
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
				assert.NotEqual(t, question.ID, newQuestion.ID)
				var mcq structs.MCQQuestion
				require.NoError(t, json.Unmarshal(newQuestion.Question, &mcq))
				assert.Equal(t, "Updated MCQ", mcq.Statement)
				assert.Equal(t, []string{"X", "Y", "Z"}, mcq.Options)
				assert.Equal(t, 10, newQuestion.MaxScore)
				assert.False(t, newQuestion.Locked)
			},
		},
		{
			name: "Error - Type change without question content",
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, int(userID))
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
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
				paper := createTestPaper(t, int(userID))
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
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
				Type: &questionTypeShort, // Changing to SHORT type
				Question: &proto.UpdateQuestionRequest_Mcq{ // But providing MCQ content
					Mcq: &proto.McqQuestion{
						Statement: "Wrong content type",
						Options:   []string{"X", "Y", "Z"},
					},
				},
			},
			userID:       userID,
			expectedCode: codes.InvalidArgument,
		},
		// Add more test cases for other scenarios
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, question := tt.setup(t)
			tt.request.QuestionId = int32(question.ID)

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
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, paper *models.Paper, questionID int)
	}{
		{
			name: "Success - Delete MCQ question",
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, int(userID))
				// Set initial question counts
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 1, "short": 1, "long": 1}`).Error
				require.NoError(t, err)

				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
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
			validate: func(t *testing.T, paper *models.Paper, questionID int) {
				// Verify question was deleted
				var question models.Question
				err := db.DB.First(&question, questionID).Error
				assert.Error(t, err) // Question should not exist

				// Verify question counts were updated
				var updatedPaper models.Paper
				require.NoError(t, db.DB.First(&updatedPaper, paper.ID).Error)
				counts, err := updatedPaper.GetQuestionCounts()
				require.NoError(t, err)
				assert.Equal(t, 0, counts.MCQ, "MCQ count should decrease")
				assert.Equal(t, 1, counts.Short)
				assert.Equal(t, 1, counts.Long)
			},
		},
		{
			name: "Success - Delete locked question unlinks it",
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, int(userID))
				// Set initial question counts
				err := db.DB.Model(&paper).Update("question_counts", `{"mcq": 2, "short": 1, "long": 0}`).Error
				require.NoError(t, err)

				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				question := models.Question{
					PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
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
			validate: func(t *testing.T, paper *models.Paper, questionID int) {
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
				assert.Equal(t, 1, counts.MCQ, "MCQ count should decrease")
				assert.Equal(t, 1, counts.Short)
				assert.Equal(t, 0, counts.Long)
			},
		},
		// Add more test cases for other scenarios
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, question := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.DeleteQuestion(ctx, &proto.QuestionRequest{QuestionId: int32(question.ID)})

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
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, questions []models.Question)
	}{
		{
			name: "Success - Reorder questions",
			setup: func(t *testing.T) (*models.Paper, *models.QuestionCategory, []models.Question) {
				paper := createTestPaper(t, int(userID))
				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
						CategoryID: category.ID,
						Order:      1,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Q1"}`),
					},
					{
						PaperID:    sql.NullInt64{Int64: int64(paper.ID), Valid: true},
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
				CategoryId:  int32(category.ID),
				QuestionIds: []int32{int32(questions[1].ID), int32(questions[0].ID)}, // Reverse order
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
