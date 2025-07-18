package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
	"pariksha/common/pkg/types"
	"pariksha/question/internal/config/db"
	"pariksha/question/internal/models"
)

func TestCreateQuestion(t *testing.T) {
	testCases := []test.TestCase[*proto.CreateQuestionRequest, *proto.CreateQuestionResponse, map[string]any]{
		{
			Name: "Valid MCQ question",
			GetRequest: func(setupData map[string]any) *proto.CreateQuestionRequest {
				return &proto.CreateQuestionRequest{
					Type: proto.QuestionType_MCQ,
					RawQuestion: []byte(`{
						"statement": "What is 2+2?",
						"options": ["3", "4", "5", "6"]
					}`),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.CreateQuestionResponse, setupData map[string]any) {
				assert.Greater(t, resp.Id, int64(0))
				assert.NotEmpty(t, resp.Hash)
			},
		},
		{
			Name: "Valid subjective question",
			GetRequest: func(setupData map[string]any) *proto.CreateQuestionRequest {
				return &proto.CreateQuestionRequest{
					Type: proto.QuestionType_SUBJECTIVE,
					RawQuestion: []byte(`{
						"statement": "Explain Newton's laws of motion"
					}`),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.CreateQuestionResponse, setupData map[string]any) {
				assert.Greater(t, resp.Id, int64(0))
				assert.NotEmpty(t, resp.Hash)
			},
		},
		{
			Name: "Valid coding question",
			GetRequest: func(setupData map[string]any) *proto.CreateQuestionRequest {
				return &proto.CreateQuestionRequest{
					Type: proto.QuestionType_CODING,
					RawQuestion: []byte(`{
						"title": "Add two numbers",
						"statement": "Write a function to add two numbers",
						"input_definitions": [
							{"variable_name": "a", "type": 1},
							{"variable_name": "b", "type": 1}
						],
						"output_definition": {"type": 1}
					}`),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.CreateQuestionResponse, setupData map[string]any) {
				assert.Greater(t, resp.Id, int64(0))
				assert.NotEmpty(t, resp.Hash)
			},
		},
		{
			Name: "Invalid question type",
			GetRequest: func(setupData map[string]any) *proto.CreateQuestionRequest {
				return &proto.CreateQuestionRequest{
					Type:        proto.QuestionType(99),
					RawQuestion: []byte(`{}`),
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name: "Invalid MCQ format",
			GetRequest: func(setupData map[string]any) *proto.CreateQuestionRequest {
				return &proto.CreateQuestionRequest{
					Type: proto.QuestionType_MCQ,
					RawQuestion: []byte(`{
						"statement": "What is 2+2?",
						"options": []
					}`),
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name: "Invalid subjective format",
			GetRequest: func(setupData map[string]any) *proto.CreateQuestionRequest {
				return &proto.CreateQuestionRequest{
					Type: proto.QuestionType_SUBJECTIVE,
					RawQuestion: []byte(`{
						"invalid_property_name": "Explain Newton's laws",
					}`),
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name: "Invalid coding format",
			GetRequest: func(setupData map[string]any) *proto.CreateQuestionRequest {
				return &proto.CreateQuestionRequest{
					Type: proto.QuestionType_CODING,
					RawQuestion: []byte(`{
						"statement": "Add two numbers",
						"input_definitions": [],
						"output_definition": {"type": 1}
					}`),
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name: "Invalid JSON format",
			GetRequest: func(setupData map[string]any) *proto.CreateQuestionRequest {
				return &proto.CreateQuestionRequest{
					Type:        proto.QuestionType_MCQ,
					RawQuestion: []byte(`{invalid json}`),
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.CreateQuestion)
		})
	}
}

func TestUpdateQuestion(t *testing.T) {
	type SetupReturn struct {
		QuestionID types.QuestionID
	}

	testCases := []test.TestCase[*proto.UpdateQuestionRequest, *proto.UpdateQuestionResponse, *SetupReturn]{
		{
			Name: "Update MCQ question when not in use",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{
							"statement": "What is 2+2?",
							"options": ["3", "4", "5", "6"]
						}`),
						Type:          proto.QuestionType_MCQ,
						PaperIndegree: 1,
						ExamIndegree:  0,
					},
				})
				return &SetupReturn{
					QuestionID: questions[0].ID,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateQuestionRequest {
				return &proto.UpdateQuestionRequest{
					Hash: questionIdToHashMap[setupData.QuestionID],
					RawQuestion: []byte(`{
						"statement": "What is 3+3?",
						"options": ["4", "5", "6", "7"]
					}`),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.UpdateQuestionResponse, setupData *SetupReturn) {
				var question models.Question
				err := db.DB.Take(&question, setupData.QuestionID).Error
				require.NoError(t, err)
				assert.Contains(t, string(question.Question), "3+3")
			},
		},
		{
			Name: "Update question type when not in use",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{
							"statement": "What is 2+2?",
							"options": ["3", "4", "5", "6"]
						}`),
						Type:          proto.QuestionType_MCQ,
						PaperIndegree: 1,
						ExamIndegree:  0,
					},
				})
				return &SetupReturn{
					QuestionID: questions[0].ID,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateQuestionRequest {
				return &proto.UpdateQuestionRequest{
					Hash: questionIdToHashMap[setupData.QuestionID],
					Type: proto.QuestionType_SUBJECTIVE.Enum(),
					RawQuestion: []byte(`{
						"statement": "Explain what is 2+2 and why?"
					}`),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.UpdateQuestionResponse, setupData *SetupReturn) {
				var question models.Question
				err := db.DB.Take(&question, setupData.QuestionID).Error
				require.NoError(t, err)
				assert.Equal(t, proto.QuestionType_SUBJECTIVE, question.Type)
			},
		},
		{
			Name: "Create new version when updating question in use",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{
							"statement": "What is 2+2?",
							"options": ["3", "4", "5", "6"]
						}`),
						Type:          proto.QuestionType_MCQ,
						PaperIndegree: 1,
						ExamIndegree:  1, // Question in use
					},
				})
				return &SetupReturn{
					QuestionID: questions[0].ID,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateQuestionRequest {
				return &proto.UpdateQuestionRequest{
					Hash: questionIdToHashMap[setupData.QuestionID],
					RawQuestion: []byte(`{
						"statement": "What is 3+3?",
						"options": ["4", "5", "6", "7"]
					}`),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.UpdateQuestionResponse, setupData *SetupReturn) {
				// Verify original question exists unchanged
				var originalQuestion models.Question
				err := db.DB.Take(&originalQuestion, setupData.QuestionID).Error
				require.NoError(t, err)
				assert.Contains(t, string(originalQuestion.Question), "2+2")
				assert.Equal(t, int32(1), originalQuestion.ExamIndegree)

				// Verify new version exists
				var newQuestion models.Question
				err = db.DB.Take(&newQuestion, resp.Id).Error
				require.NoError(t, err)
				assert.Contains(t, string(newQuestion.Question), "3+3")
				assert.Equal(t, int32(0), newQuestion.ExamIndegree)
				assert.NotEqual(t, originalQuestion.Hash, newQuestion.Hash)
			},
		},
		{
			Name: "Invalid hash",
			GetRequest: func(setupData *SetupReturn) *proto.UpdateQuestionRequest {
				return &proto.UpdateQuestionRequest{
					Hash: "nonexistent",
					RawQuestion: []byte(`{
						"statement": "What is 2+2?",
						"options": ["3", "4", "5", "6"]
					}`),
				}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name: "Invalid question format",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{
							"statement": "What is 2+2?",
							"options": ["3", "4", "5", "6"]
						}`),
						Type:          proto.QuestionType_MCQ,
						PaperIndegree: 1,
						ExamIndegree:  0,
					},
				})
				return &SetupReturn{
					QuestionID: questions[0].ID,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateQuestionRequest {
				return &proto.UpdateQuestionRequest{
					Hash: questionIdToHashMap[setupData.QuestionID],
					RawQuestion: []byte(`{
						"statement": "What is 2+2?",
						"options": []
					}`),
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name: "Missing raw_question when changing type",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{
							"statement": "What is 2+2?",
							"options": ["3", "4", "5", "6"]
						}`),
						Type:          proto.QuestionType_MCQ,
						PaperIndegree: 1,
						ExamIndegree:  0,
					},
				})
				return &SetupReturn{
					QuestionID: questions[0].ID,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateQuestionRequest {
				return &proto.UpdateQuestionRequest{
					Hash: questionIdToHashMap[setupData.QuestionID],
					Type: proto.QuestionType_SUBJECTIVE.Enum(),
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.UpdateQuestion)
		})
	}
}

func TestGetQuestionsByIds(t *testing.T) {
	type SetupReturn struct {
		Questions []models.Question
	}

	testCases := []test.TestCase[*proto.QuestionIdsRequest, *proto.GetQuestionsResponse, *SetupReturn]{
		{
			Name: "Get multiple questions by IDs",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{"statement": "Q1", "options": ["1", "2", "3", "4"]}`),
						Type:     proto.QuestionType_MCQ,
					},
					{
						Question: []byte(`{"statement": "Q2"}`),
						Type:     proto.QuestionType_SUBJECTIVE,
					},
					{
						Question: []byte(`{
							"title": "Q3",
							"statement": "Add numbers",
							"input_definitions": [{"variable_name": "a", "type": 1}],
							"output_definition": {"type": 1}
						}`),
						Type: proto.QuestionType_CODING,
					},
				})
				return &SetupReturn{Questions: questions}
			},
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				ids := make([]int64, len(setupData.Questions))
				for i, q := range setupData.Questions {
					ids[i] = int64(q.ID)
				}
				return &proto.QuestionIdsRequest{Ids: ids}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.GetQuestionsResponse, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Questions), len(resp.Questions))
				for i, q := range resp.Questions {
					assert.Equal(t, int64(setupData.Questions[i].ID), q.Id)
					assert.Equal(t, setupData.Questions[i].Hash, q.Hash)
					assert.Equal(t, setupData.Questions[i].Type, q.Type)
					assert.JSONEq(t, string(setupData.Questions[i].Question), string(q.RawQuestion))
				}
			},
		},
		{
			Name: "Non-existent question ID",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{"statement": "Q1", "options": ["1", "2", "3", "4"]}`),
						Type:     proto.QuestionType_MCQ,
					},
				})
				return &SetupReturn{Questions: questions}
			},
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				return &proto.QuestionIdsRequest{
					Ids: []int64{int64(setupData.Questions[0].ID), 99999},
				}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name: "Empty IDs list",
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				return &proto.QuestionIdsRequest{Ids: []int64{}}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetQuestionsByIds)
		})
	}
}

func TestGetQuestionsByHashes(t *testing.T) {
	type SetupReturn struct {
		Questions []models.Question
	}

	testCases := []test.TestCase[*proto.QuestionHashesRequest, *proto.GetQuestionsResponse, *SetupReturn]{
		{
			Name: "Get multiple questions by hashes",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{"statement": "Q1", "options": ["1", "2", "3", "4"]}`),
						Type:     proto.QuestionType_MCQ,
					},
					{
						Question: []byte(`{"statement": "Q2"}`),
						Type:     proto.QuestionType_SUBJECTIVE,
					},
				})
				return &SetupReturn{Questions: questions}
			},
			GetRequest: func(setupData *SetupReturn) *proto.QuestionHashesRequest {
				hashes := make([]string, len(setupData.Questions))
				for i, q := range setupData.Questions {
					hashes[i] = q.Hash
				}
				return &proto.QuestionHashesRequest{Hashes: hashes}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.GetQuestionsResponse, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Questions), len(resp.Questions))
				for i, q := range resp.Questions {
					assert.Equal(t, int64(setupData.Questions[i].ID), q.Id)
					assert.Equal(t, setupData.Questions[i].Hash, q.Hash)
					assert.Equal(t, setupData.Questions[i].Type, q.Type)
					assert.JSONEq(t, string(setupData.Questions[i].Question), string(q.RawQuestion))
				}
			},
		},
		{
			Name: "Non-existent hash",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{"statement": "Q1", "options": ["1", "2", "3", "4"]}`),
						Type:     proto.QuestionType_MCQ,
					},
				})
				return &SetupReturn{Questions: questions}
			},
			GetRequest: func(setupData *SetupReturn) *proto.QuestionHashesRequest {
				return &proto.QuestionHashesRequest{
					Hashes: []string{setupData.Questions[0].Hash, "nonexistent"},
				}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name: "Empty hashes list",
			GetRequest: func(setupData *SetupReturn) *proto.QuestionHashesRequest {
				return &proto.QuestionHashesRequest{Hashes: []string{}}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetQuestionsByHashes)
		})
	}
}

func TestGetQuestionsMetaByIds(t *testing.T) {
	type SetupReturn struct {
		Questions []models.Question
	}

	testCases := []test.TestCase[*proto.QuestionIdsRequest, *proto.QuestionsMetaResponse, *SetupReturn]{
		{
			Name: "Get metadata for multiple questions",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{"statement": "Q1", "options": ["1", "2", "3", "4"]}`),
						Type:     proto.QuestionType_MCQ,
					},
					{
						Question: []byte(`{"statement": "Q2"}`),
						Type:     proto.QuestionType_SUBJECTIVE,
					},
				})
				return &SetupReturn{Questions: questions}
			},
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				ids := make([]int64, len(setupData.Questions))
				for i, q := range setupData.Questions {
					ids[i] = int64(q.ID)
				}
				return &proto.QuestionIdsRequest{Ids: ids}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.QuestionsMetaResponse, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Questions), len(resp.Meta))
				for i, meta := range resp.Meta {
					assert.Equal(t, int64(setupData.Questions[i].ID), meta.Id)
					assert.Equal(t, setupData.Questions[i].Hash, meta.Hash)
					assert.Equal(t, setupData.Questions[i].Type, meta.Type)
					assert.JSONEq(t, string(setupData.Questions[i].Question), string(meta.RawQuestion))
				}
			},
		},
		{
			Name: "Empty IDs list",
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				return &proto.QuestionIdsRequest{Ids: []int64{}}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetQuestionsMetaByIds)
		})
	}
}

func TestGetQuestionHashesByIds(t *testing.T) {
	type SetupReturn struct {
		Questions []models.Question
	}

	testCases := []test.TestCase[*proto.QuestionIdsRequest, *proto.GetQuestionHashesByIdsResponse, *SetupReturn]{
		{
			Name: "Get hashes for multiple questions",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{"statement": "Q1", "options": ["1", "2", "3", "4"]}`),
						Type:     proto.QuestionType_MCQ,
					},
					{
						Question: []byte(`{"statement": "Q2"}`),
						Type:     proto.QuestionType_SUBJECTIVE,
					},
				})
				return &SetupReturn{Questions: questions}
			},
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				ids := make([]int64, len(setupData.Questions))
				for i, q := range setupData.Questions {
					ids[i] = int64(q.ID)
				}
				return &proto.QuestionIdsRequest{Ids: ids}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.GetQuestionHashesByIdsResponse, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Questions), len(resp.Hashes))
				for i, hash := range resp.Hashes {
					assert.Equal(t, setupData.Questions[i].Hash, hash)
				}
			},
		},
		{
			Name: "Empty IDs list",
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				return &proto.QuestionIdsRequest{Ids: []int64{}}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetQuestionHashesByIds)
		})
	}
}

func TestGetQuestionsMetaByHashes(t *testing.T) {
	type SetupReturn struct {
		Questions []models.Question
	}

	testCases := []test.TestCase[*proto.QuestionHashesRequest, *proto.QuestionsMetaResponse, *SetupReturn]{
		{
			Name: "Get metadata for multiple questions by hashes",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{"statement": "Q1", "options": ["1", "2", "3", "4"]}`),
						Type:     proto.QuestionType_MCQ,
					},
					{
						Question: []byte(`{"statement": "Q2"}`),
						Type:     proto.QuestionType_SUBJECTIVE,
					},
				})
				return &SetupReturn{Questions: questions}
			},
			GetRequest: func(setupData *SetupReturn) *proto.QuestionHashesRequest {
				hashes := make([]string, len(setupData.Questions))
				for i, q := range setupData.Questions {
					hashes[i] = q.Hash
				}
				return &proto.QuestionHashesRequest{Hashes: hashes}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.QuestionsMetaResponse, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Questions), len(resp.Meta))
				for i, meta := range resp.Meta {
					assert.Equal(t, int64(setupData.Questions[i].ID), meta.Id)
					assert.Equal(t, setupData.Questions[i].Hash, meta.Hash)
					assert.Equal(t, setupData.Questions[i].Type, meta.Type)
					assert.JSONEq(t, string(setupData.Questions[i].Question), string(meta.RawQuestion))
				}
			},
		},
		{
			Name: "Non-existent hash",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{"statement": "Q1", "options": ["1", "2", "3", "4"]}`),
						Type:     proto.QuestionType_MCQ,
					},
				})
				return &SetupReturn{Questions: questions}
			},
			GetRequest: func(setupData *SetupReturn) *proto.QuestionHashesRequest {
				return &proto.QuestionHashesRequest{
					Hashes: []string{setupData.Questions[0].Hash, "nonexistent"},
				}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name: "Empty hashes list",
			GetRequest: func(setupData *SetupReturn) *proto.QuestionHashesRequest {
				return &proto.QuestionHashesRequest{Hashes: []string{}}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetQuestionsMetaByHashes)
		})
	}
}

func TestGetQuestionIdsByHashes(t *testing.T) {
	type SetupReturn struct {
		Questions []models.Question
	}

	testCases := []test.TestCase[*proto.QuestionHashesRequest, *proto.GetQuestionIdsByHashesResponse, *SetupReturn]{
		{
			Name: "Get IDs for multiple questions",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{"statement": "Q1", "options": ["1", "2", "3", "4"]}`),
						Type:     proto.QuestionType_MCQ,
					},
					{
						Question: []byte(`{"statement": "Q2"}`),
						Type:     proto.QuestionType_SUBJECTIVE,
					},
				})
				return &SetupReturn{Questions: questions}
			},
			GetRequest: func(setupData *SetupReturn) *proto.QuestionHashesRequest {
				hashes := make([]string, len(setupData.Questions))
				for i, q := range setupData.Questions {
					hashes[i] = q.Hash
				}
				return &proto.QuestionHashesRequest{Hashes: hashes}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.GetQuestionIdsByHashesResponse, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Questions), len(resp.Ids))
				for i, id := range resp.Ids {
					assert.Equal(t, int64(setupData.Questions[i].ID), id)
				}
			},
		},
		{
			Name: "Empty hashes list",
			GetRequest: func(setupData *SetupReturn) *proto.QuestionHashesRequest {
				return &proto.QuestionHashesRequest{Hashes: []string{}}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetQuestionIdsByHashes)
		})
	}
}

func TestGetBoilerplate(t *testing.T) {
	type SetupReturn struct {
		Question    models.Question
		Language    models.Language
		Boilerplate models.Boilerplate
	}

	testCases := []test.TestCase[*proto.GetBoilerplateRequest, *proto.BoilerplateResponse, *SetupReturn]{
		{
			Name: "Get existing boilerplate",
			Setup: func(t *testing.T) *SetupReturn {
				// Create test language
				language := models.Language{
					ID:        1,
					Slug:      constants.LangNode,
					Name:      "JavaScript (Node)",
					Extension: "js",
					Version:   "22",
					IsEnabled: true,
				}
				err := db.DB.Create(&language).Error
				require.NoError(t, err)

				// Create test question
				questions := createTestQuestions(t, []models.Question{
					{
						Question: []byte(`{
							"title": "Add Numbers",
							"statement": "Add two numbers",
							"input_definitions": [{"variable_name": "a", "type": 1}],
							"output_definition": {"type": 1}
						}`),
						Type: proto.QuestionType_CODING,
					},
				})

				// Create test boilerplate
				boilerplate := models.Boilerplate{
					QuestionID: questions[0].ID,
					LanguageID: language.ID,
					Code:       "function solve(a) {\n\n}",
				}
				err = db.DB.Create(&boilerplate).Error
				require.NoError(t, err)

				return &SetupReturn{
					Question:    questions[0],
					Language:    language,
					Boilerplate: boilerplate,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.GetBoilerplateRequest {
				return &proto.GetBoilerplateRequest{
					QuestionHash: setupData.Question.Hash,
					LanguageId:   int32(setupData.Language.ID),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.BoilerplateResponse, setupData *SetupReturn) {
				assert.Equal(t, setupData.Boilerplate.Code, resp.Code)
			},
		},
		{
			Name: "Non-existent question hash",
			Setup: func(t *testing.T) *SetupReturn {
				language := models.Language{
					ID:        1,
					Slug:      constants.LangNode,
					Name:      "JavaScript (Node)",
					Extension: "js",
					Version:   "22",
					IsEnabled: true,
				}
				err := db.DB.Create(&language).Error
				require.NoError(t, err)
				return &SetupReturn{Language: language}
			},
			GetRequest: func(setupData *SetupReturn) *proto.GetBoilerplateRequest {
				return &proto.GetBoilerplateRequest{
					QuestionHash: "nonexistent",
					LanguageId:   int32(setupData.Language.ID),
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetBoilerplate)
		})
	}
}

func TestIncPaperIndegreeByIds(t *testing.T) {
	type SetupReturn struct {
		Questions []models.Question
	}

	testCases := []test.TestCase[*proto.QuestionIdsRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name: "Increment paper indegree for multiple questions",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question:      []byte(`{"statement": "Q1"}`),
						Type:          proto.QuestionType_MCQ,
						PaperIndegree: 1,
					},
					{
						Question:      []byte(`{"statement": "Q2"}`),
						Type:          proto.QuestionType_SUBJECTIVE,
						PaperIndegree: 2,
					},
				})
				return &SetupReturn{Questions: questions}
			},
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				ids := make([]int64, len(setupData.Questions))
				for i, q := range setupData.Questions {
					ids[i] = int64(q.ID)
				}
				return &proto.QuestionIdsRequest{Ids: ids}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				for _, original := range setupData.Questions {
					var updated models.Question
					err := db.DB.Take(&updated, original.ID).Error
					require.NoError(t, err)
					assert.Equal(t, original.PaperIndegree+1, updated.PaperIndegree)
				}
			},
		},
		{
			Name: "Empty IDs list",
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				return &proto.QuestionIdsRequest{Ids: []int64{}}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.IncQuestionPaperIndegreeByIds)
		})
	}
}

func TestDecPaperIndegreeByIds(t *testing.T) {
	type SetupReturn struct {
		Questions []models.Question
	}

	testCases := []test.TestCase[*proto.QuestionIdsRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name: "Decrement paper indegree for multiple questions",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question:      []byte(`{"statement": "Q1"}`),
						Type:          proto.QuestionType_MCQ,
						PaperIndegree: 2,
					},
					{
						Question:      []byte(`{"statement": "Q2"}`),
						Type:          proto.QuestionType_SUBJECTIVE,
						PaperIndegree: 3,
					},
				})
				return &SetupReturn{Questions: questions}
			},
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				ids := make([]int64, len(setupData.Questions))
				for i, q := range setupData.Questions {
					ids[i] = int64(q.ID)
				}
				return &proto.QuestionIdsRequest{Ids: ids}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				for _, original := range setupData.Questions {
					var updated models.Question
					err := db.DB.Take(&updated, original.ID).Error
					require.NoError(t, err)
					assert.Equal(t, original.PaperIndegree-1, updated.PaperIndegree)
				}
			},
		},
		{
			Name: "Empty IDs list",
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				return &proto.QuestionIdsRequest{Ids: []int64{}}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.DecQuestionPaperIndegreeByIds)
		})
	}
}

func TestIncExamIndegreeByIds(t *testing.T) {
	type SetupReturn struct {
		Questions []models.Question
	}

	testCases := []test.TestCase[*proto.QuestionIdsRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name: "Increment exam indegree for multiple questions",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question:     []byte(`{"statement": "Q1"}`),
						Type:         proto.QuestionType_MCQ,
						ExamIndegree: 0,
					},
					{
						Question:     []byte(`{"statement": "Q2"}`),
						Type:         proto.QuestionType_SUBJECTIVE,
						ExamIndegree: 1,
					},
				})
				return &SetupReturn{Questions: questions}
			},
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				ids := make([]int64, len(setupData.Questions))
				for i, q := range setupData.Questions {
					ids[i] = int64(q.ID)
				}
				return &proto.QuestionIdsRequest{Ids: ids}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				for _, original := range setupData.Questions {
					var updated models.Question
					err := db.DB.Take(&updated, original.ID).Error
					require.NoError(t, err)
					assert.Equal(t, original.ExamIndegree+1, updated.ExamIndegree)
				}
			},
		},
		{
			Name: "Empty IDs list",
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				return &proto.QuestionIdsRequest{Ids: []int64{}}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.IncQuestionExamIndegreeByIds)
		})
	}
}

func TestDecExamIndegreeByIds(t *testing.T) {
	type SetupReturn struct {
		Questions []models.Question
	}

	testCases := []test.TestCase[*proto.QuestionIdsRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name: "Decrement exam indegree for multiple questions",
			Setup: func(t *testing.T) *SetupReturn {
				questions := createTestQuestions(t, []models.Question{
					{
						Question:     []byte(`{"statement": "Q1"}`),
						Type:         proto.QuestionType_MCQ,
						ExamIndegree: 1,
					},
					{
						Question:     []byte(`{"statement": "Q2"}`),
						Type:         proto.QuestionType_SUBJECTIVE,
						ExamIndegree: 2,
					},
				})
				return &SetupReturn{Questions: questions}
			},
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				ids := make([]int64, len(setupData.Questions))
				for i, q := range setupData.Questions {
					ids[i] = int64(q.ID)
				}
				return &proto.QuestionIdsRequest{Ids: ids}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				for _, original := range setupData.Questions {
					var updated models.Question
					err := db.DB.Take(&updated, original.ID).Error
					require.NoError(t, err)
					assert.Equal(t, original.ExamIndegree-1, updated.ExamIndegree)
				}
			},
		},
		{
			Name: "Empty IDs list",
			GetRequest: func(setupData *SetupReturn) *proto.QuestionIdsRequest {
				return &proto.QuestionIdsRequest{Ids: []int64{}}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.DecQuestionExamIndegreeByIds)
		})
	}
}
