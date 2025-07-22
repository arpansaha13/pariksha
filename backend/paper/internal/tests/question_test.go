package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
	"pariksha/common/pkg/types"
	"pariksha/paper/internal/config/env"
	"pariksha/paper/internal/models"
)

func TestGetPaperQuestions(t *testing.T) {
	type SetupReturn struct {
		Paper      models.Paper
		Questions  []models.PaperQuestion
		Categories []models.PaperCategory
	}

	testCases := []test.TestCase[*proto.PaperRequest, *proto.QuestionList, *SetupReturn]{
		{
			Name:     "Get multiple questions from paper",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
					{
						PaperID:    papers[0].ID,
						CategoryID: 2,
					},
				})

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 1,
						CategoryID: categories[0].CategoryID,
						MaxScore:   10,
					},
					{
						PaperID:    papers[0].ID,
						QuestionID: 2,
						CategoryID: categories[1].CategoryID,
						MaxScore:   15,
					},
				})

				return &SetupReturn{
					Paper:      papers[0],
					Questions:  questions,
					Categories: categories,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: setupData.Paper.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.QuestionList, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Questions), len(resp.Questions))
				for i, q := range resp.Questions {
					assert.NotEmpty(t, q.QuestionHash)
					assert.Equal(t, int64(setupData.Questions[i].CategoryID), q.CategoryId)
					assert.Equal(t, setupData.Paper.Hash, q.PaperHash)
					assert.Equal(t, int32(setupData.Questions[i].Order), q.Order)
					assert.NotEmpty(t, q.RawQuestion)
				}
			},
		},
		{
			Name:     "Paper with no questions",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				return &SetupReturn{
					Paper:      papers[0],
					Categories: categories,
					Questions:  []models.PaperQuestion{},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: setupData.Paper.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.QuestionList, setupData *SetupReturn) {
				assert.Empty(t, resp.Questions)
			},
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: "nonexistent",
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetPaperQuestions)
		})
	}
}

func TestGetPaperQuestion(t *testing.T) {
	type SetupReturn struct {
		Paper    models.Paper
		Question models.PaperQuestion
		Category models.PaperCategory
	}

	testCases := []test.TestCase[*proto.PaperQuestionRequest, *proto.PaperQuestionResponse, *SetupReturn]{
		{
			Name:     "Get existing paper question",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 1,
						CategoryID: categories[0].CategoryID,
						MaxScore:   10,
					},
				})

				return &SetupReturn{
					Paper:    papers[0],
					Question: questions[0],
					Category: categories[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperQuestionRequest {
				return &proto.PaperQuestionRequest{
					PaperHash:    setupData.Paper.Hash,
					QuestionHash: "q_hash_1", // Static question hash in mock
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.PaperQuestionResponse, setupData *SetupReturn) {
				assert.NotEmpty(t, resp.QuestionHash)
				assert.NotEmpty(t, resp.RawQuestion)
				assert.Equal(t, int64(setupData.Question.CategoryID), resp.CategoryId)
				assert.Equal(t, setupData.Paper.Hash, resp.PaperHash)
				assert.Equal(t, int32(setupData.Question.MaxScore), resp.MaxScore)
				assert.Nil(t, resp.TestCases)
			},
		},
		{
			Name:     "Get existing coding question with test cases",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 3,
						CategoryID: categories[0].CategoryID,
						MaxScore:   10,
					},
				})

				return &SetupReturn{
					Paper:    papers[0],
					Question: questions[0],
					Category: categories[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperQuestionRequest {
				return &proto.PaperQuestionRequest{
					PaperHash:    setupData.Paper.Hash,
					QuestionHash: "q_hash_3", // Static question hash in mock
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.PaperQuestionResponse, setupData *SetupReturn) {
				assert.NotEmpty(t, resp.QuestionHash)
				assert.NotEmpty(t, resp.RawQuestion)
				assert.Equal(t, int64(setupData.Question.CategoryID), resp.CategoryId)
				assert.Equal(t, setupData.Paper.Hash, resp.PaperHash)
				assert.Equal(t, int32(setupData.Question.MaxScore), resp.MaxScore)
				assert.Equal(t, resp.Type, proto.QuestionType_CODING)
				assert.NotNil(t, resp.TestCases)
			},
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.PaperQuestionRequest {
				return &proto.PaperQuestionRequest{
					PaperHash:    "nonexistent",
					QuestionHash: "question-hash-1",
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Non-existent question hash",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)
				return &SetupReturn{Paper: papers[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperQuestionRequest {
				return &proto.PaperQuestionRequest{
					PaperHash:    setupData.Paper.Hash,
					QuestionHash: "nonexistent",
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetPaperQuestion)
		})
	}
}

func TestCreatePaperQuestion(t *testing.T) {
	type SetupReturn struct {
		Paper    models.Paper
		Category models.PaperCategory
	}

	testCases := []test.TestCase[*proto.CreatePaperQuestionRequest, *proto.CreatePaperQuestionResponse, *SetupReturn]{
		{
			Name:     "Create MCQ question in paper",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				return &SetupReturn{
					Paper:    papers[0],
					Category: categories[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.CreatePaperQuestionRequest {
				return &proto.CreatePaperQuestionRequest{
					PaperHash:   setupData.Paper.Hash,
					CategoryId:  int64(setupData.Category.CategoryID),
					MaxScore:    10,
					Type:        proto.QuestionType_MCQ,
					RawQuestion: []byte(`{"statement": "What is 2+2?", "options": ["3", "4", "5", "6"]}`),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.CreatePaperQuestionResponse, setupData *SetupReturn) {
				assert.NotEmpty(t, resp.QuestionHash)

				// Verify paper question was created
				var paperQuestion models.PaperQuestion
				err := dbInst.Where("paper_id = ? AND category_id = ?", setupData.Paper.ID, setupData.Category.CategoryID).
					Take(&paperQuestion).Error
				require.NoError(t, err)
				assert.Equal(t, int16(10), paperQuestion.MaxScore)
				assert.Equal(t, int16(1), paperQuestion.Order)
			},
		},
		{
			Name:     "Invalid max score",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				return &SetupReturn{
					Paper:    papers[0],
					Category: categories[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.CreatePaperQuestionRequest {
				return &proto.CreatePaperQuestionRequest{
					PaperHash:   setupData.Paper.Hash,
					CategoryId:  int64(setupData.Category.CategoryID),
					MaxScore:    1001, // Max allowed is 1000
					Type:        proto.QuestionType_MCQ,
					RawQuestion: []byte(`{"statement": "What is 2+2?", "options": ["3", "4", "5", "6"]}`),
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.CreatePaperQuestionRequest {
				return &proto.CreatePaperQuestionRequest{
					PaperHash:   "nonexistent",
					CategoryId:  1,
					MaxScore:    10,
					Type:        proto.QuestionType_MCQ,
					RawQuestion: []byte(`{"statement": "What is 2+2?", "options": ["3", "4", "5", "6"]}`),
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.CreatePaperQuestion)
		})
	}
}

func TestUpdatePaperQuestion(t *testing.T) {
	type SetupReturn struct {
		Paper    models.Paper
		Question models.PaperQuestion
		Category models.PaperCategory
	}

	testCases := []test.TestCase[*proto.UpdatePaperQuestionRequest, *proto.UpdatePaperQuestionResponse, *SetupReturn]{
		{
			Name:     "Update MCQ question",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 1,
						CategoryID: categories[0].CategoryID,
						MaxScore:   10,
					},
				})

				return &SetupReturn{
					Paper:    papers[0],
					Question: questions[0],
					Category: categories[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdatePaperQuestionRequest {
				maxScore := int32(15)
				return &proto.UpdatePaperQuestionRequest{
					PaperHash:    setupData.Paper.Hash,
					QuestionHash: "q_hash_1", // Static hash from mock
					MaxScore:     &maxScore,
					Type:         proto.QuestionType_MCQ.Enum(),
					RawQuestion:  []byte(`{"statement": "Updated question", "options": ["1", "2", "3", "4"]}`),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.UpdatePaperQuestionResponse, setupData *SetupReturn) {
				assert.NotEmpty(t, resp.QuestionHash)

				// Verify paper question was updated
				var paperQuestion models.PaperQuestion
				err := dbInst.Where("paper_id = ? AND question_id = ?", setupData.Paper.ID, setupData.Question.QuestionID).
					Take(&paperQuestion).Error
				require.NoError(t, err)
				assert.Equal(t, int16(15), paperQuestion.MaxScore)
			},
		},
		{
			Name:     "Invalid max score",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)
				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})
				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 1,
						CategoryID: categories[0].CategoryID,
						MaxScore:   10,
					},
				})
				return &SetupReturn{
					Paper:    papers[0],
					Question: questions[0],
					Category: categories[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdatePaperQuestionRequest {
				maxScore := int32(1001) // Max allowed is 1000
				return &proto.UpdatePaperQuestionRequest{
					PaperHash:    setupData.Paper.Hash,
					QuestionHash: "q_hash_1",
					MaxScore:     &maxScore,
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.UpdatePaperQuestionRequest {
				maxScore := int32(15)
				return &proto.UpdatePaperQuestionRequest{
					PaperHash:    "nonexistent",
					QuestionHash: "q_hash_1",
					MaxScore:     &maxScore,
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Non-existent question hash",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)
				return &SetupReturn{Paper: papers[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdatePaperQuestionRequest {
				maxScore := int32(15)
				return &proto.UpdatePaperQuestionRequest{
					PaperHash:    setupData.Paper.Hash,
					QuestionHash: "nonexistent",
					MaxScore:     &maxScore,
				}
			},
			ExpectedCode: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.UpdatePaperQuestion)
		})
	}
}

func TestDeletePaperQuestion(t *testing.T) {
	type SetupReturn struct {
		Paper    models.Paper
		Question models.PaperQuestion
		Category models.PaperCategory
	}

	testCases := []test.TestCase[*proto.PaperQuestionRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name:     "Delete existing question",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 1,
						CategoryID: categories[0].CategoryID,
						MaxScore:   10,
					},
				})

				return &SetupReturn{
					Paper:    papers[0],
					Category: categories[0],
					Question: questions[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperQuestionRequest {
				return &proto.PaperQuestionRequest{
					PaperHash:    setupData.Paper.Hash,
					QuestionHash: "q_hash_1", // Mock client will handle this
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				// Verify question was deleted
				var count int64
				err := dbInst.Model(&models.PaperQuestion{}).
					Where("paper_id = ? AND question_id = ?", setupData.Paper.ID, setupData.Question.QuestionID).
					Count(&count).Error
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.PaperQuestionRequest {
				return &proto.PaperQuestionRequest{
					PaperHash:    "nonexistent",
					QuestionHash: "q_hash_1",
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Non-existent question hash",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)
				return &SetupReturn{Paper: papers[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperQuestionRequest {
				return &proto.PaperQuestionRequest{
					PaperHash:    setupData.Paper.Hash,
					QuestionHash: "nonexistent",
				}
			},
			ExpectedCode: codes.OK, // Ignore
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.DeletePaperQuestion)
		})
	}
}

func TestReorderQuestions(t *testing.T) {
	type SetupReturn struct {
		Paper     models.Paper
		Category  models.PaperCategory
		Questions []models.PaperQuestion
	}

	testCases := []test.TestCase[*proto.ReorderPaperQuestionsRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name:     "Reorder multiple questions",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 1,
						CategoryID: categories[0].CategoryID,
						MaxScore:   10,
					},
					{
						PaperID:    papers[0].ID,
						QuestionID: 2,
						CategoryID: categories[0].CategoryID,
						MaxScore:   15,
					},
					{
						PaperID:    papers[0].ID,
						QuestionID: 3,
						CategoryID: categories[0].CategoryID,
						MaxScore:   20,
					},
				})

				return &SetupReturn{
					Paper:     papers[0],
					Category:  categories[0],
					Questions: questions,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ReorderPaperQuestionsRequest {
				// Reverse the order of questions
				questionHashes := []string{"q_hash_3", "q_hash_2", "q_hash_1"}
				return &proto.ReorderPaperQuestionsRequest{
					CategoryId:     int64(setupData.Category.CategoryID),
					QuestionHashes: questionHashes,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				var questions []models.PaperQuestion
				err := dbInst.Where("category_id = ?", setupData.Category.CategoryID).
					Order("\"order\" asc").
					Find(&questions).Error
				require.NoError(t, err)
				require.Len(t, questions, 3)

				// Verify reversed order
				assert.Equal(t, types.QuestionID(3), questions[0].QuestionID)
				assert.Equal(t, types.QuestionID(2), questions[1].QuestionID)
				assert.Equal(t, types.QuestionID(1), questions[2].QuestionID)

				// Verify order numbers are sequential
				assert.Equal(t, int16(1), questions[0].Order)
				assert.Equal(t, int16(2), questions[1].Order)
				assert.Equal(t, int16(3), questions[2].Order)
			},
		},
		{
			Name:     "Invalid question hashes",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)
				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
				})
				return &SetupReturn{
					Paper:    papers[0],
					Category: categories[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ReorderPaperQuestionsRequest {
				return &proto.ReorderPaperQuestionsRequest{
					PaperHash:      setupData.Paper.Hash,
					CategoryId:     int64(setupData.Category.CategoryID),
					QuestionHashes: []string{"nonexistent1", "nonexistent2"},
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Questions from different category",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
					{
						PaperID:    papers[0].ID,
						CategoryID: 2,
					},
				})

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 1,
						CategoryID: categories[0].CategoryID,
					},
					{
						PaperID:    papers[0].ID,
						QuestionID: 2,
						CategoryID: categories[1].CategoryID, // Different category
					},
				})

				return &SetupReturn{
					Paper:     papers[0],
					Category:  categories[0],
					Questions: questions,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ReorderPaperQuestionsRequest {
				return &proto.ReorderPaperQuestionsRequest{
					CategoryId:     int64(setupData.Category.CategoryID),
					QuestionHashes: []string{"q_hash_1", "q_hash_2"},
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.ReorderPaperQuestions)
		})
	}
}

func TestGetPaperQuestionsMeta(t *testing.T) {
	md := metadata.MD{
		constants.X_EXAM_API_TOKEN: []string{env.EXAM_API_TOKEN},
	}

	type SetupReturn struct {
		Paper      models.Paper
		Questions  []models.PaperQuestion
		Categories []models.PaperCategory
	}

	testCases := []test.TestCase[*proto.PaperRequest, *proto.PaperQuestionsMeta, *SetupReturn]{
		{
			Name:     "Get questions metadata from paper",
			Metadata: md,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				categories := createTestCategories(t, []models.PaperCategory{
					{
						PaperID:    papers[0].ID,
						CategoryID: 1,
					},
					{
						PaperID:    papers[0].ID,
						CategoryID: 2,
					},
				})

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 1,
						CategoryID: categories[0].CategoryID,
						MaxScore:   10,
					},
					{
						PaperID:    papers[0].ID,
						QuestionID: 2,
						CategoryID: categories[1].CategoryID,
						MaxScore:   15,
					},
				})

				return &SetupReturn{
					Paper:      papers[0],
					Questions:  questions,
					Categories: categories,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: setupData.Paper.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.PaperQuestionsMeta, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Questions), len(resp.Questions))
				for i, q := range resp.Questions {
					assert.Equal(t, int64(setupData.Questions[i].QuestionID), q.Id)
					assert.Equal(t, int64(setupData.Questions[i].CategoryID), q.CategoryId)
					assert.Equal(t, int32(setupData.Questions[i].Order), q.Order)
					assert.Equal(t, int32(setupData.Questions[i].MaxScore), q.MaxScore)
				}
			},
		},
		{
			Name:     "Paper with no questions",
			Metadata: md,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)
				return &SetupReturn{
					Paper:     papers[0],
					Questions: []models.PaperQuestion{},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: setupData.Paper.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.PaperQuestionsMeta, setupData *SetupReturn) {
				assert.Empty(t, resp.Questions)
			},
		},
		{
			Name:     "Non-existent paper hash",
			Metadata: md,
			GetRequest: func(setupData *SetupReturn) *proto.PaperRequest {
				return &proto.PaperRequest{
					PaperHash: "nonexistent",
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.PaperQuestionsMeta, setupData *SetupReturn) {
				assert.Empty(t, resp.Questions)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetPaperQuestionsMeta)
		})
	}
}

func TestGetBoilerplate(t *testing.T) {
	type SetupReturn struct {
		Paper    models.Paper
		Question models.PaperQuestion
	}

	testCases := []test.TestCase[*proto.GetPaperBoilerplateRequest, *proto.BoilerplateResponse, *SetupReturn]{
		{
			Name:     "Get boilerplate for existing question",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 3, // Using questionID 3 which is a coding question in mock
						CategoryID: 1,
						MaxScore:   10,
					},
				})

				return &SetupReturn{
					Paper:    papers[0],
					Question: questions[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.GetPaperBoilerplateRequest {
				return &proto.GetPaperBoilerplateRequest{
					QuestionHash: "q_hash_3", // Static hash from mock
					LanguageId:   1,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.BoilerplateResponse, setupData *SetupReturn) {
				assert.NotEmpty(t, resp.Code)
			},
		},
		{
			Name:     "Non-existent question hash",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)
				return &SetupReturn{Paper: papers[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.GetPaperBoilerplateRequest {
				return &proto.GetPaperBoilerplateRequest{
					QuestionHash: "nonexistent",
					LanguageId:   1,
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetPaperBoilerplate)
		})
	}
}

func TestUpsertTestCases(t *testing.T) {
	type SetupReturn struct {
		Paper    models.Paper
		Question models.PaperQuestion
	}

	testCases := []test.TestCase[*proto.UpsertPaperTestCasesRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name:     "Upsert test cases for coding question",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 3, // QuestionID 3 is a coding question in mock
						CategoryID: 1,
						MaxScore:   10,
					},
				})

				return &SetupReturn{
					Paper:    papers[0],
					Question: questions[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpsertPaperTestCasesRequest {
				return &proto.UpsertPaperTestCasesRequest{
					PaperHash:    setupData.Paper.Hash,
					QuestionHash: "q_hash_3", // Static hash from mock
					TestCases: []*proto.UpsertTestCase{
						{
							Inputs: []string{"1", "2"},
							Output: "3",
							Hidden: false,
						},
						{
							Inputs: []string{"4", "5"},
							Output: "9",
							Hidden: true,
						},
					},
				}
			},
			ExpectedCode: codes.OK,
		},
		{
			Name:     "Non-existent question hash",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)
				return &SetupReturn{Paper: papers[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpsertPaperTestCasesRequest {
				return &proto.UpsertPaperTestCasesRequest{
					PaperHash:    setupData.Paper.Hash,
					QuestionHash: "nonexistent",
					TestCases: []*proto.UpsertTestCase{
						{
							Inputs: []string{"1", "2"},
							Output: "3",
						},
					},
				}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name:     "Empty test cases",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				papers := createDefaultTestPapers(t, 1)

				questions := createTestQuestions(t, []models.PaperQuestion{
					{
						PaperID:    papers[0].ID,
						QuestionID: 3,
						CategoryID: 1,
						MaxScore:   10,
					},
				})

				return &SetupReturn{
					Paper:    papers[0],
					Question: questions[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpsertPaperTestCasesRequest {
				return &proto.UpsertPaperTestCasesRequest{
					PaperHash:    setupData.Paper.Hash,
					QuestionHash: "q_hash_3",
					TestCases:    []*proto.UpsertTestCase{},
				}
			},
			ExpectedCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.UpsertPaperTestCases)
		})
	}
}
