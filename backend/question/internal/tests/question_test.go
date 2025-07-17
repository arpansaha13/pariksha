package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
	"pariksha/common/pkg/types"
	"pariksha/question/internal/config/db"
	"pariksha/question/internal/models"
)

func TestCreateQuestion(t *testing.T) {
	testCases := []test.TestCase[*proto.CreateQuestionRequest, *proto.CreateQuestionResponse]{
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
	testCases := []test.TestCase[*proto.UpdateQuestionRequest, *proto.UpdateQuestionResponse]{
		{
			Name: "Update MCQ question when not in use",
			Setup: func(t *testing.T) map[string]any {
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
				return map[string]any{
					"questionId": questions[0].ID,
				}
			},
			GetRequest: func(setupData map[string]any) *proto.UpdateQuestionRequest {
				qid := setupData["questionId"].(types.QuestionID)
				return &proto.UpdateQuestionRequest{
					Hash: questionIdToHashMap[qid],
					RawQuestion: []byte(`{
						"statement": "What is 3+3?",
						"options": ["4", "5", "6", "7"]
					}`),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.UpdateQuestionResponse, setupData map[string]any) {
				var question models.Question
				err := db.DB.First(&question, setupData["questionId"]).Error
				require.NoError(t, err)
				assert.Contains(t, string(question.Question), "3+3")
			},
		},
		{
			Name: "Update question type when not in use",
			Setup: func(t *testing.T) map[string]any {
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
				return map[string]any{
					"questionId": questions[0].ID,
				}
			},
			GetRequest: func(setupData map[string]any) *proto.UpdateQuestionRequest {
				qid := setupData["questionId"].(types.QuestionID)
				return &proto.UpdateQuestionRequest{
					Hash: questionIdToHashMap[qid],
					Type: proto.QuestionType_SUBJECTIVE.Enum(),
					RawQuestion: []byte(`{
						"statement": "Explain what is 2+2 and why?"
					}`),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.UpdateQuestionResponse, setupData map[string]any) {
				var question models.Question
				err := db.DB.First(&question, setupData["questionId"]).Error
				require.NoError(t, err)
				assert.Equal(t, proto.QuestionType_SUBJECTIVE, question.Type)
			},
		},
		{
			Name: "Create new version when updating question in use",
			Setup: func(t *testing.T) map[string]any {
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
				return map[string]any{
					"originalId": questions[0].ID,
				}
			},
			GetRequest: func(setupData map[string]any) *proto.UpdateQuestionRequest {
				qid := setupData["originalId"].(types.QuestionID)
				return &proto.UpdateQuestionRequest{
					Hash: questionIdToHashMap[qid],
					RawQuestion: []byte(`{
						"statement": "What is 3+3?",
						"options": ["4", "5", "6", "7"]
					}`),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.UpdateQuestionResponse, setupData map[string]any) {
				// Verify original question exists unchanged
				var originalQuestion models.Question
				err := db.DB.First(&originalQuestion, setupData["originalId"]).Error
				require.NoError(t, err)
				assert.Contains(t, string(originalQuestion.Question), "2+2")
				assert.Equal(t, int32(1), originalQuestion.ExamIndegree)

				// Verify new version exists
				var newQuestion models.Question
				err = db.DB.First(&newQuestion, resp.Id).Error
				require.NoError(t, err)
				assert.Contains(t, string(newQuestion.Question), "3+3")
				assert.Equal(t, int32(0), newQuestion.ExamIndegree)
				assert.NotEqual(t, originalQuestion.Hash, newQuestion.Hash)
			},
		},
		{
			Name: "Invalid hash",
			GetRequest: func(setupData map[string]any) *proto.UpdateQuestionRequest {
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
			Setup: func(t *testing.T) map[string]any {
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
				return map[string]any{
					"questionId": questions[0].ID,
				}
			},
			GetRequest: func(setupData map[string]any) *proto.UpdateQuestionRequest {
				qid := setupData["questionId"].(types.QuestionID)
				return &proto.UpdateQuestionRequest{
					Hash: questionIdToHashMap[qid],
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
			Setup: func(t *testing.T) map[string]any {
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
				return map[string]any{
					"questionId": questions[0].ID,
				}
			},
			GetRequest: func(setupData map[string]any) *proto.UpdateQuestionRequest {
				qid := setupData["questionId"].(types.QuestionID)
				return &proto.UpdateQuestionRequest{
					Hash: questionIdToHashMap[qid],
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
