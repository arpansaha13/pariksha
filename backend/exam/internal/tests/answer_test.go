package tests

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
	"pariksha/common/pkg/types"
)

func TestGetParticipantAnswers(t *testing.T) {
	type SetupReturn struct {
		Participant *models.ExamParticipant
	}

	testCases := []test.TestCase[*proto.ParticipantRequest, *proto.AnswerList, *SetupReturn]{
		{
			Name:     "Success - Get multiple answers with evaluation data",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
				}})

				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{
						QuestionID: 2,
						MaxScore:   10,
					},
					{
						QuestionID: 1,
						MaxScore:   5,
					},
				})

				// Create answers with different evaluation states
				rawAnswer1 := json.RawMessage(`{"text": "Answer 1"}`)
				answer1 := models.Answer{
					ExamParticipantID: participants[0].ID,
					QuestionID:        questions[0].QuestionID,
					Answer:            &rawAnswer1,
					ScoreAwarded:      8,
					Evaluated:         true,
				}
				require.NoError(t, dbInst.Create(&answer1).Error)

				rawAnswer2 := json.RawMessage(`{"optionIndex": 1}`)
				answer2 := models.Answer{
					ExamParticipantID: participants[0].ID,
					QuestionID:        questions[1].QuestionID,
					Answer:            &rawAnswer2,
					ScoreAwarded:      0, // Not evaluated yet
					Evaluated:         false,
				}
				require.NoError(t, dbInst.Create(&answer2).Error)

				return &SetupReturn{Participant: &participants[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantRequest {
				return &proto.ParticipantRequest{ParticipantId: int64(setupData.Participant.ID)}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.AnswerList, setupData *SetupReturn) {
				require.Equal(t, 2, len(resp.Answers))

				answer1 := resp.Answers[0]
				assert.EqualValues(t, 1, answer1.Order)
				assert.EqualValues(t, proto.QuestionType_SUBJECTIVE, answer1.QuestionType)
				assert.EqualValues(t, 10, answer1.MaxScore)

				var answerData1 struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(answer1.Answer, &answerData1))
				assert.Equal(t, "Answer 1", answerData1.Text)

				answer2 := resp.Answers[1]
				assert.EqualValues(t, 2, answer2.Order)
				assert.EqualValues(t, proto.QuestionType_MCQ, answer2.QuestionType)
				assert.EqualValues(t, 5, answer2.MaxScore)

				var answerData2 struct {
					OptionIndex int `json:"optionIndex"`
				}
				require.NoError(t, json.Unmarshal(answer2.Answer, &answerData2))
				assert.Equal(t, 1, answerData2.OptionIndex)
			},
		},
		{
			Name:     "Fail - Participant not found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{Participant: &models.ExamParticipant{ID: 9999}}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantRequest {
				return &proto.ParticipantRequest{ParticipantId: int64(setupData.Participant.ID)}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name:     "Fail - No answers found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{{Status: constants.PARTICIPANT_STATUS_INVITED}})
				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &SetupReturn{Participant: &participant}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantRequest {
				return &proto.ParticipantRequest{ParticipantId: int64(setupData.Participant.ID)}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name:     "Success - Get answers as evaluator",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ?", exam.ID).First(&participant).Error)

				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{
						QuestionID: 2,
						MaxScore:   10,
					},
					{
						QuestionID: 1,
						MaxScore:   5,
					},
				})

				rawAnswer1 := json.RawMessage(`{"text": "Answer 1"}`)
				answer1 := models.Answer{
					ExamParticipantID: participant.ID,
					QuestionID:        questions[0].QuestionID,
					Answer:            &rawAnswer1,
					ScoreAwarded:      8,
					Evaluated:         true,
				}
				require.NoError(t, dbInst.Create(&answer1).Error)
				rawAnswer2 := json.RawMessage(`{"optionIndex": 1}`)
				answer2 := models.Answer{
					ExamParticipantID: participant.ID,
					QuestionID:        questions[1].QuestionID,
					Answer:            &rawAnswer2,
					ScoreAwarded:      0,
					Evaluated:         false,
				}
				require.NoError(t, dbInst.Create(&answer2).Error)
				return &SetupReturn{Participant: &participant}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantRequest {
				return &proto.ParticipantRequest{ParticipantId: int64(setupData.Participant.ID)}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.AnswerList, setupData *SetupReturn) {
				require.Equal(t, 2, len(resp.Answers))
				answer1 := resp.Answers[0]
				assert.EqualValues(t, 1, answer1.Order)
				assert.EqualValues(t, proto.QuestionType_SUBJECTIVE, answer1.QuestionType)
				assert.EqualValues(t, 10, answer1.MaxScore)
				var answerData1 struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(answer1.Answer, &answerData1))
				assert.Equal(t, "Answer 1", answerData1.Text)
				answer2 := resp.Answers[1]
				assert.EqualValues(t, 2, answer2.Order)
				assert.EqualValues(t, proto.QuestionType_MCQ, answer2.QuestionType)
				assert.EqualValues(t, 5, answer2.MaxScore)
				var answerData2 struct {
					OptionIndex int `json:"optionIndex"`
				}
				require.NoError(t, json.Unmarshal(answer2.Answer, &answerData2))
				assert.Equal(t, 1, answerData2.OptionIndex)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetParticipantAnswers)
		})
	}
}

func TestGetAnswerForExam(t *testing.T) {
	type SetupReturn struct {
		Exam       *models.Exam
		QuestionID types.QuestionID
	}

	testCases := []test.TestCase[*proto.GetAnswerRequest, *proto.AnswerMinimalResponse, *SetupReturn]{
		{
			Name:     "Success - Get single answer",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
				}})

				answer := createTestAnswer(t, &participants[0], 1)
				return &SetupReturn{Exam: &exam, QuestionID: answer.QuestionID}
			},
			GetRequest: func(setupData *SetupReturn) *proto.GetAnswerRequest {
				return &proto.GetAnswerRequest{
					ExamHash:     setupData.Exam.Hash,
					QuestionHash: getQuestionHashForId(setupData.QuestionID),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.AnswerMinimalResponse, setupData *SetupReturn) {
				var answerData struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(resp.Answer, &answerData))
				assert.Equal(t, "Test Answer", answerData.Text)
				assert.NotZero(t, resp.AnswerId)
			},
		},
		{
			Name:     "Fail - User not a participant",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				return &SetupReturn{Exam: &exam, QuestionID: 1}
			},
			GetRequest: func(setupData *SetupReturn) *proto.GetAnswerRequest {
				return &proto.GetAnswerRequest{
					ExamHash:     setupData.Exam.Hash,
					QuestionHash: getQuestionHashForId(setupData.QuestionID),
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Success - Answer not found returns empty response",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				return &SetupReturn{Exam: &exam, QuestionID: 1}
			},
			GetRequest: func(setupData *SetupReturn) *proto.GetAnswerRequest {
				return &proto.GetAnswerRequest{
					ExamHash:     setupData.Exam.Hash,
					QuestionHash: "q_hash_1",
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.AnswerMinimalResponse, setupData *SetupReturn) {
				assert.EqualValues(t, 1, int(setupData.QuestionID))
				assert.Zero(t, resp.AnswerId)
				assert.Nil(t, resp.Answer)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetAnswerForExam)
		})
	}
}

func TestUpsertAnswer(t *testing.T) {
	type SetupReturn struct {
		Exam        *models.Exam
		QuestionID  types.QuestionID
		Participant *models.ExamParticipant
		// Request     *proto.UpsertAnswersRequest
	}

	testCases := []test.TestCase[*proto.UpsertAnswersRequest, *proto.UpsertAnswersResponse, *SetupReturn]{
		{
			Name:     "Success - Create new answer",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{QuestionID: 2},
				})

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
					ScheduledEndTime: sql.NullTime{
						Time:  time.Now().Add(time.Hour),
						Valid: true,
					},
				}})

				return &SetupReturn{
					Exam:        &exam,
					QuestionID:  questions[0].QuestionID,
					Participant: &participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpsertAnswersRequest {
				return &proto.UpsertAnswersRequest{
					ExamHash: setupData.Exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: "q_hash_2",
						Answer:       []byte(`{"text": "Test answer content"}`),
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.UpsertAnswersResponse, setupData *SetupReturn) {
				assert.NotZero(t, resp.AnswerId)

				// Verify answer in database
				var answer models.Answer
				require.NoError(t, dbInst.First(&answer, resp.AnswerId).Error)
				var answerData struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(*answer.Answer, &answerData))
				assert.Equal(t, "Test answer content", answerData.Text)
				assert.EqualValues(t, setupData.QuestionID, answer.QuestionID)
			},
		},
		{
			Name:     "Success - Update existing answer",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{QuestionID: 2},
				})

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
					ScheduledEndTime: sql.NullTime{
						Time:  time.Now().Add(time.Hour),
						Valid: true,
					},
				}})

				createTestAnswer(t, &participants[0], questions[0].QuestionID)

				return &SetupReturn{
					Exam:        &exam,
					QuestionID:  questions[0].QuestionID,
					Participant: &participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpsertAnswersRequest {
				return &proto.UpsertAnswersRequest{
					ExamHash: setupData.Exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: "q_hash_2",
						Answer:       []byte(`{"text": "Updated answer content"}`),
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.UpsertAnswersResponse, setupData *SetupReturn) {
				assert.NotZero(t, resp.AnswerId)

				// Verify updated answer in database
				var answer models.Answer
				require.NoError(t, dbInst.First(&answer, resp.AnswerId).Error)
				var answerData struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(*answer.Answer, &answerData))
				assert.Equal(t, "Updated answer content", answerData.Text)
			},
		},
		{
			Name:     "Fail - Exam participant not found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				return &SetupReturn{
					Exam:        &exam,
					Participant: nil,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpsertAnswersRequest {
				return &proto.UpsertAnswersRequest{
					ExamHash: setupData.Exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: getQuestionHashForId(1),
						Answer:       []byte(`{"text": "Test answer"}`),
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Fail - Exam not started",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})

				return &SetupReturn{
					Exam:        &exam,
					Participant: nil,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpsertAnswersRequest {
				return &proto.UpsertAnswersRequest{
					ExamHash: setupData.Exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: getQuestionHashForId(1),
						Answer:       []byte(`{"text": "Test answer"}`),
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
			ExpectedCode: codes.FailedPrecondition,
		},
		{
			Name:     "Fail - Exam ended",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_ENDED},
				})

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ? AND user_id = ?", exam.ID, defaultUserID).First(&participant).Error)

				return &SetupReturn{
					Exam:        &exam,
					Participant: &participant,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpsertAnswersRequest {
				return &proto.UpsertAnswersRequest{
					ExamHash: setupData.Exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: "q_hash_2",
						Answer:       []byte(`{"text": "Test answer after exam ended"}`),
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
			ExpectedCode: codes.FailedPrecondition,
		},
		{
			Name:     "Success - Empty answer for SUBJECTIVE question",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{QuestionID: 2},
				})

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
					ScheduledEndTime: sql.NullTime{
						Time:  time.Now().Add(time.Hour),
						Valid: true,
					},
				}})

				return &SetupReturn{
					Exam:        &exam,
					QuestionID:  questions[0].QuestionID,
					Participant: &participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpsertAnswersRequest {
				return &proto.UpsertAnswersRequest{
					ExamHash: setupData.Exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: "q_hash_2",
						Answer:       []byte(`{"text": ""}`),
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.UpsertAnswersResponse, setupData *SetupReturn) {
				assert.NotZero(t, resp.AnswerId)

				var answer models.Answer
				require.NoError(t, dbInst.First(&answer, resp.AnswerId).Error)
				var answerData struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(*answer.Answer, &answerData))
				assert.Equal(t, "", answerData.Text)
			},
		},
		{
			Name:     "Success - Nil answer clears the answer",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{QuestionID: 2},
				})

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
					ScheduledEndTime: sql.NullTime{
						Time:  time.Now().Add(time.Hour),
						Valid: true,
					},
				}})

				// Create initial answer
				createTestAnswer(t, &participants[0], questions[0].QuestionID)

				return &SetupReturn{
					Exam:        &exam,
					QuestionID:  questions[0].QuestionID,
					Participant: &participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpsertAnswersRequest {
				return &proto.UpsertAnswersRequest{
					ExamHash: setupData.Exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: "q_hash_2",
						Answer:       nil, // Explicit nil answer
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.UpsertAnswersResponse, setupData *SetupReturn) {
				assert.NotZero(t, resp.AnswerId)

				var answer models.Answer
				require.NoError(t, dbInst.First(&answer, resp.AnswerId).Error)
				assert.Nil(t, answer.Answer)
			},
		},
		{
			Name:     "Fail - Empty MCQ answer is invalid",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{QuestionID: 1},
				})

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
					ScheduledEndTime: sql.NullTime{
						Time:  time.Now().Add(time.Hour),
						Valid: true,
					},
				}})

				// First create an answer
				createTestAnswer(t, &participants[0], questions[0].QuestionID)

				return &SetupReturn{
					Exam:        &exam,
					QuestionID:  questions[0].QuestionID,
					Participant: &participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpsertAnswersRequest {
				return &proto.UpsertAnswersRequest{
					ExamHash: setupData.Exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: "q_hash_1",
						Answer:       []byte(`{}`), // Empty answer object
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Fail - Nil optionIndex in MCQ answer is invalid",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{QuestionID: 1},
				})

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
					ScheduledEndTime: sql.NullTime{
						Time:  time.Now().Add(time.Hour),
						Valid: true,
					},
				}})

				createTestAnswer(t, &participants[0], questions[0].QuestionID)

				return &SetupReturn{
					Exam:        &exam,
					QuestionID:  questions[0].QuestionID,
					Participant: &participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpsertAnswersRequest {
				return &proto.UpsertAnswersRequest{
					ExamHash: setupData.Exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: "q_hash_1",
						Answer:       []byte(`{"optionIndex": null}`), // explicit null for optionIndex
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.UpsertAnswer)
		})
	}
}
