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

func TestUpdateMcqQuestion(t *testing.T) {
	maxScore := int32(10)

	tests := []UpdateQuestionCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update MCQ question",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Paper, *models.Question) {
				paper := createTestPaper(t, userID)
				err := db.DB.Model(&paper).Update("question_counts", models.QuestionCount{MCQ: 1}).Error
				require.NoError(t, err)

				var category models.QuestionCategory
				require.NoError(t, db.DB.Where("paper_id = ?", paper.ID).First(&category).Error)

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Old MCQ","options":["A","B"]}`),
						Tags:       json.RawMessage(`["old"]`),
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

				verifyQuestionCounts(t, paper.ID, &models.QuestionCount{MCQ: 1})

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
	}

	runUpdateQuestionTests(t, tests)
}

func TestUpdateSubjectiveQuestion(t *testing.T) {
	maxScore := int32(10)

	tests := []UpdateQuestionCase{
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
	}

	runUpdateQuestionTests(t, tests)
}

func TestUpdateCodingQuestion(t *testing.T) {
	maxScore := int32(10)

	tests := []UpdateQuestionCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Update coding question",
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
						Type:       constants.QUESTION_TYPE_CODING,
						Question: json.RawMessage(`{
							"title": "Old Title",
							"statement": "Old statement",
							"input_definitions": [
								{ "variable_name": "a", "type": 1 },
								{ "variable_name": "b", "type": 1 }
							],
							"output_definition": {
								"variable_name": "sum",
								"type": 1
							},
							"test_cases": [
								{
									"inputs": ["1", "1"],
									"output": "2"
								}
							]
						}`),
						Tags: json.RawMessage(`["old"]`),
					},
				})
				return &paper, &questions[0]
			},
			request: &proto.UpdateQuestionRequest{
				MaxScore: &maxScore,
				Tags:     []string{"updated", "coding"},
				RawQuestion: []byte(`{
					"title": "Updated Coding Question",
					"statement": "Write an optimized solution",
					"input_definitions": [
						{
							"variable_name": "arr",
							"type": 4,
							"items": [{ "type": 1 }]
						}
					],
					"output_definition": {
						"variable_name": "sum",
						"type": 1
					},
					"test_cases": [
						{
							"inputs": ["[1, 2, 3]"],
							"output": "6",
							"explanation": "1 + 2 + 3 = 6"
						}
					]
				}`),
			},
			validate: func(t *testing.T, paper *models.Paper, question *models.Question) {
				var updated models.Question
				require.NoError(t, db.DB.First(&updated, question.ID).Error)

				// Verify question content
				var coding structs.CodingQuestion
				require.NoError(t, json.Unmarshal(updated.Question, &coding))
				assert.Equal(t, "Updated Coding Question", coding.Title)
				assert.Equal(t, "Write an optimized solution", coding.Statement)
				assert.Len(t, coding.TestCases, 1)
				assert.Equal(t, "[1, 2, 3]", coding.TestCases[0].Inputs[0])
				assert.Equal(t, "6", coding.TestCases[0].Output)

				// Verify other fields
				assert.EqualValues(t, 10, updated.MaxScore)
				var tags []string
				require.NoError(t, json.Unmarshal(updated.Tags, &tags))
				assert.ElementsMatch(t, []string{"updated", "coding"}, tags)
			},
		},
	}

	runUpdateQuestionTests(t, tests)
}

func TestUpdateLockedQuestion(t *testing.T) {
	maxScore := int32(10)

	tests := []UpdateQuestionCase{
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

				questions := createTestQuestions(t, []models.Question{
					{
						PaperID:    sql.NullInt64{Int64: paper.ID, Valid: true},
						CategoryID: category.ID,
						Type:       constants.QUESTION_TYPE_MCQ,
						Question:   json.RawMessage(`{"statement":"Test MCQ","options":["A","B"]}`),
						MaxScore:   5,
						Locked:     true,
					},
				})
				return &paper, &questions[0]
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
				assert.EqualValues(t, 10, newQuestion.MaxScore)
				assert.False(t, newQuestion.Locked)

				var mcq structs.MCQQuestion
				require.NoError(t, json.Unmarshal(newQuestion.Question, &mcq))
				assert.Equal(t, "Updated MCQ", mcq.Statement)
				assert.Equal(t, []string{"X", "Y", "Z"}, mcq.Options)
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
				func(t *testing.T, resp *proto.UpdateQuestionResponse) {
					if tt.validate != nil {
						// For locked questions, we expect a new ID
						assert.NotEqual(t, question.ID, resp.QuestionId, "A new ID should be created")
						tt.validate(t, paper, question)
					}
				})
		})
	}
}

func TestUpdateQuestionTypeChange(t *testing.T) {
	maxScore := int32(10)
	questionTypeSubjective := constants.QUESTION_TYPE_SUBJECTIVE

	tests := []UpdateQuestionCase{
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
				verifyQuestionCounts(t, paper.ID, &models.QuestionCount{Subjective: 1})
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
	}

	runUpdateQuestionTests(t, tests)
}

func TestUpdateCodingQuestionBoilerplates(t *testing.T) {
	tests := []struct {
		name            string
		initialQuestion string
		updatedQuestion string
		wantOldCode     string
		wantNewCode     string
	}{
		{
			name: "Update from two params to one array param",
			initialQuestion: `{
				"title": "Add Numbers",
				"statement": "Add two numbers",
				"input_definitions": [
					{"variable_name": "a", "type": 1},
					{"variable_name": "b", "type": 1}
				],
				"output_definition": {"type": 1}
			}`,
			updatedQuestion: `{
				"title": "Sum Array",
				"statement": "Sum array elements",
				"input_definitions": [
					{
						"variable_name": "arr",
						"type": 4,
						"items": [{"type": 1}]
					}
				],
				"output_definition": {"type": 1}
			}`,
			wantOldCode: "function solve(a, b) {}",
			wantNewCode: "function solve(arr) {}",
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

			// Create initial coding question
			ctx := createContextWithUserID(userID)
			createReq := &proto.CreateQuestionRequest{
				PaperId:     paper.ID,
				CategoryId:  category.ID,
				Type:        constants.QUESTION_TYPE_CODING,
				RawQuestion: []byte(tt.initialQuestion),
				MaxScore:    10,
			}

			var createResp *proto.CreateQuestionResponse
			testrunner.Runner(t, ctx, codes.OK, createReq, client.CreateQuestion, func(t *testing.T, r *proto.CreateQuestionResponse) {
				createResp = r
			})

			// Verify initial boilerplate
			var initialBoilerplate models.Boilerplate
			require.NoError(t, db.DB.First(&initialBoilerplate, "question_id = ?", createResp.Id).Error)
			assert.Equal(t, tt.wantOldCode, initialBoilerplate.Code)

			// Update question
			updateReq := &proto.UpdateQuestionRequest{
				QuestionId:  createResp.Id,
				RawQuestion: []byte(tt.updatedQuestion),
			}

			testrunner.Runner(t, ctx, codes.OK, updateReq, client.UpdateQuestion, nil)

			// Verify updated boilerplate
			var updatedBoilerplate models.Boilerplate
			require.NoError(t, db.DB.First(&updatedBoilerplate, "question_id = ?", createResp.Id).Error)
			assert.Equal(t, tt.wantNewCode, updatedBoilerplate.Code)
		})
	}
}

// Helper function to run update question tests
func runUpdateQuestionTests(t *testing.T, tests []UpdateQuestionCase) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper, question := tt.setup(t)
			tt.request.QuestionId = question.ID

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				tt.request,
				client.UpdateQuestion,
				func(t *testing.T, resp *proto.UpdateQuestionResponse) {
					if tt.validate != nil {
						// Question ID should remain same for non-locked questions
						assert.Equal(t, question.ID, resp.QuestionId, "Question ID should remain same")
						tt.validate(t, paper, question)
					}
				})
		})
	}
}
