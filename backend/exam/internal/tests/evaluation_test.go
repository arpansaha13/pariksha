package tests

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
	"pariksha/common/pkg/types"
)

func TestGetAnswerForEvaluation(t *testing.T) {
	type SetupReturn struct {
		Participant *models.ExamParticipant
		QuestionID  types.QuestionID
	}

	testCases := []test.TestCase[*proto.ParticipantQuestionRequest, *proto.AnswerMinimalResponse, *SetupReturn]{
		{
			Name:     "Success - Get answer as evaluator",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2) // Created by different user
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{
						QuestionID: 1,
						CategoryID: 1,
					},
				})

				// Create evaluator permission for test user
				permission := models.ExamPermission{
					ExamID: exam.ID,
					UserID: defaultUserID,
				}
				permission.SetEvaluate()
				require.NoError(t, dbInst.Create(&permission).Error)

				// Create participant and answer
				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_ENDED},
				})

				createTestAnswer(t, &participants[0], questions[0].QuestionID)

				return &SetupReturn{
					Participant: &participants[0],
					QuestionID:  questions[0].QuestionID,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantQuestionRequest {
				return &proto.ParticipantQuestionRequest{
					ParticipantId: int64(setupData.Participant.ID),
					QuestionHash:  "q_hash_1",
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.AnswerMinimalResponse, setupData *SetupReturn) {
				assert.NotZero(t, resp.AnswerId)
				assert.NotNil(t, resp.Answer)
				var answerData struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(resp.Answer, &answerData))
				assert.Equal(t, "Test Answer", answerData.Text)
			},
		},
		{
			Name:     "Fail - No evaluate permission",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2) // Created by different user

				// Create participant and answer
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_ENDED},
				})

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ?", exam.ID).First(&participant).Error)
				answer := createTestAnswer(t, &participant, 1)

				return &SetupReturn{
					Participant: &participant,
					QuestionID:  answer.QuestionID,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantQuestionRequest {
				return &proto.ParticipantQuestionRequest{
					ParticipantId: int64(setupData.Participant.ID),
					QuestionHash:  "q_hash_1",
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Success - Empty response for non-existent answer",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID) // Create as owner to get evaluate permission
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{QuestionID: 1},
				})

				// Create participant without answer
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED},
				})

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ?", exam.ID).First(&participant).Error)

				return &SetupReturn{
					Participant: &participant,
					QuestionID:  questions[0].ID,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantQuestionRequest {
				return &proto.ParticipantQuestionRequest{
					ParticipantId: int64(setupData.Participant.ID),
					QuestionHash:  "q_hash_1",
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.AnswerMinimalResponse, setupData *SetupReturn) {
				assert.Zero(t, resp.AnswerId)
				assert.Nil(t, resp.Answer)
				assert.EqualValues(t, "q_hash_1", resp.QuestionHash)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetAnswerForEvaluation)
		})
	}
}

func TestGetAnswerEvaluationData(t *testing.T) {
	type SetupReturn struct {
		Participant *models.ExamParticipant
		QuestionID  types.QuestionID
	}

	testCases := []test.TestCase[*proto.ParticipantQuestionRequest, *proto.GetAnswerEvaluationDataResponse, *SetupReturn]{
		{
			Name:     "Success - Get evaluation data as evaluator",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2) // Created by different user
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{
						MaxScore: 5,
					},
				})

				// Create evaluator permission for test user
				permission := models.ExamPermission{
					ExamID: exam.ID,
					UserID: defaultUserID,
				}
				permission.SetEvaluate()
				require.NoError(t, dbInst.Create(&permission).Error)

				// Create participant and answer
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_ENDED},
				})

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ?", exam.ID).First(&participant).Error)
				answer := createTestAnswer(t, &participant, questions[0].QuestionID)

				return &SetupReturn{
					Participant: &participant,
					QuestionID:  answer.QuestionID,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantQuestionRequest {
				return &proto.ParticipantQuestionRequest{
					ParticipantId: int64(setupData.Participant.ID),
					QuestionHash:  getQuestionHashForId(setupData.QuestionID),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.GetAnswerEvaluationDataResponse, setupData *SetupReturn) {
				assert.NotZero(t, resp.AnswerId)
				assert.EqualValues(t, 5, resp.ScoreAwarded)
			},
		},
		{
			Name:     "Fail - No evaluate permission",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2) // Created by different user

				// Create participant and answer
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_ENDED},
				})

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ?", exam.ID).First(&participant).Error)
				answer := createTestAnswer(t, &participant, 1)

				return &SetupReturn{
					Participant: &participant,
					QuestionID:  answer.QuestionID,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantQuestionRequest {
				return &proto.ParticipantQuestionRequest{
					ParticipantId: int64(setupData.Participant.ID),
					QuestionHash:  getQuestionHashForId(setupData.QuestionID),
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Fail - Exam not ended",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID) // Create as owner to get evaluate permission

				// Create participant with STARTED status
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_STARTED},
				})

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ?", exam.ID).First(&participant).Error)
				answer := createTestAnswer(t, &participant, 1)

				return &SetupReturn{
					Participant: &participant,
					QuestionID:  answer.QuestionID,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantQuestionRequest {
				return &proto.ParticipantQuestionRequest{
					ParticipantId: int64(setupData.Participant.ID),
					QuestionHash:  getQuestionHashForId(setupData.QuestionID),
				}
			},
			ExpectedCode: codes.FailedPrecondition,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetAnswerEvaluationData)
		})
	}
}

func TestUpdateAnswerForEvaluation(t *testing.T) {
	type SetupReturn struct {
		Answer *models.Answer
	}

	newScore := int32(8)
	evaluated := true
	exceedingScore := int32(15) // Exceeds max_score in exam_questions

	testCases := []test.TestCase[*proto.UpdateAnswerRequest, *proto.GetAnswerEvaluationDataResponse, *SetupReturn]{
		{
			Name:     "Success - Update all fields",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED},
				})

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ?", exam.ID).First(&participant).Error)
				participant.ScoreAwarded = 5 // Initial score
				require.NoError(t, dbInst.Save(&participant).Error)

				// Create exam question with max score of 10
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{
						QuestionID: 1,
						CategoryID: 1,
						MaxScore:   10,
					},
				})

				answer := createTestAnswer(t, &participant, questions[0].QuestionID)
				answer.ScoreAwarded = 5 // Initial score
				answer.Evaluated = false
				require.NoError(t, dbInst.Save(&answer).Error)

				return &SetupReturn{Answer: &answer}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateAnswerRequest {
				return &proto.UpdateAnswerRequest{
					AnswerId:  int64(setupData.Answer.ID),
					NewScore:  &newScore,
					Evaluated: &evaluated,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.GetAnswerEvaluationDataResponse, setupData *SetupReturn) {
				var answer models.Answer
				require.NoError(t, dbInst.First(&answer, setupData.Answer.ID).Error)
				assert.EqualValues(t, newScore, answer.ScoreAwarded)
				assert.True(t, answer.Evaluated)

				// Verify participant total score is updated
				var participant models.ExamParticipant
				require.NoError(t, dbInst.First(&participant, answer.ExamParticipantID).Error)
				assert.EqualValues(t, newScore, participant.ScoreAwarded)
			},
		},
		{
			Name:     "Fail - Answer not found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				// Create exam question even for non-existent answer to maintain consistency
				createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{
						MaxScore: 5,
					},
				})
				return &SetupReturn{
					Answer: &models.Answer{ID: 9999},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateAnswerRequest {
				return &proto.UpdateAnswerRequest{
					AnswerId: int64(setupData.Answer.ID),
					NewScore: &newScore,
				}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name:     "Fail - Score exceeds max score",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED},
				})

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ?", exam.ID).First(&participant).Error)

				// Create exam question with max score of 10
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{
						QuestionID: 1,
						CategoryID: 1,
						MaxScore:   10,
					},
				})

				answer := createTestAnswer(t, &participant, questions[0].QuestionID)
				return &SetupReturn{Answer: &answer}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateAnswerRequest {
				return &proto.UpdateAnswerRequest{
					AnswerId: int64(setupData.Answer.ID),
					NewScore: &exceedingScore, // Try to set score higher than max_score
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.UpdateAnswerForEvaluation)
		})
	}
}

func TestMarkParticipantAsEvaluated(t *testing.T) {
	type SetupReturn struct {
		Participant *models.ExamParticipant
	}

	testCases := []test.TestCase[*proto.ParticipantRequest, *proto.EvaluationStatusResponse, *SetupReturn]{
		{
			Name:     "Success - All answers evaluated",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED},
				})

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ?", exam.ID).First(&participant).Error)

				// Create answers that are all evaluated
				answer1 := createTestAnswer(t, &participant, 1)
				answer1.Evaluated = true
				answer2 := createTestAnswer(t, &participant, 2)
				answer2.Evaluated = true

				require.NoError(t, dbInst.Save(&answer1).Error)
				require.NoError(t, dbInst.Save(&answer2).Error)

				return &SetupReturn{Participant: &participant}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantRequest {
				return &proto.ParticipantRequest{
					ParticipantId: int64(setupData.Participant.ID),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.EvaluationStatusResponse, setupData *SetupReturn) {
				// Verify unevaluated count is 0
				assert.EqualValues(t, 0, resp.UnevaluatedCount)

				// Verify participant status changed to EVALUATED
				var updatedParticipant models.ExamParticipant
				require.NoError(t, dbInst.First(&updatedParticipant, setupData.Participant.ID).Error)
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_EVALUATED, updatedParticipant.Status)
			},
		},
		{
			Name:     "Success - Some answers unevaluated",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED},
				})

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ?", exam.ID).First(&participant).Error)

				// Create one evaluated and one unevaluated answer
				answer1 := createTestAnswer(t, &participant, 1)
				answer1.Evaluated = true
				answer2 := createTestAnswer(t, &participant, 2)
				answer2.Evaluated = false

				require.NoError(t, dbInst.Save(&answer1).Error)
				require.NoError(t, dbInst.Save(&answer2).Error)

				return &SetupReturn{Participant: &participant}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantRequest {
				return &proto.ParticipantRequest{
					ParticipantId: int64(setupData.Participant.ID),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.EvaluationStatusResponse, setupData *SetupReturn) {
				// Verify one answer is still unevaluated
				assert.EqualValues(t, 1, resp.UnevaluatedCount)

				// Verify participant status remains ENDED
				var updatedParticipant models.ExamParticipant
				require.NoError(t, dbInst.First(&updatedParticipant, setupData.Participant.ID).Error)
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_ENDED, updatedParticipant.Status)
			},
		},
		{
			Name:     "Fail - Participant not found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					Participant: &models.ExamParticipant{ID: 9999},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantRequest {
				return &proto.ParticipantRequest{
					ParticipantId: int64(setupData.Participant.ID),
				}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name:     "Fail - Exam not ended",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_STARTED},
				})

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &SetupReturn{Participant: &participant}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantRequest {
				return &proto.ParticipantRequest{
					ParticipantId: int64(setupData.Participant.ID),
				}
			},
			ExpectedCode: codes.FailedPrecondition,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.MarkParticipantAsEvaluated)
		})
	}
}
