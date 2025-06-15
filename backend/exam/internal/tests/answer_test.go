package tests

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/testrunner"
	"pariksha/exam/internal/config/db"
)

func TestGetParticipantAnswers(t *testing.T) {
	tests := []getParticipantAnswersTestCase{
		{
			baseTestCase: baseTestCase{
				name: "Success - Get multiple answers with evaluation data",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createDefaultTestExam(t, 2)
				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
				}})

				questions := createTestExamQuestions(t, &exam, []models.ExamQuestion{
					{
						Type: proto.QuestionType_SUBJECTIVE,
					},
					{
						Type:     proto.QuestionType_MCQ,
						MaxScore: 5,
					},
				})

				// Create answers with different evaluation states
				rawAnswer1 := json.RawMessage(`{"text": "Answer 1"}`)
				answer1 := models.Answer{
					ExamParticipantID: participants[0].ID,
					QuestionID:        questions[0].QuestionID,
					Answer:            &rawAnswer1,
					ScoreAwarded:      8,
					Comments:          sql.NullString{String: "Good answer", Valid: true},
					Evaluated:         true,
				}
				require.NoError(t, db.DB.Create(&answer1).Error)

				rawAnswer2 := json.RawMessage(`{"optionIndex": 1}`)
				answer2 := models.Answer{
					ExamParticipantID: participants[0].ID,
					QuestionID:        questions[1].QuestionID,
					Answer:            &rawAnswer2,
					ScoreAwarded:      0, // Not evaluated yet
					Evaluated:         false,
				}
				require.NoError(t, db.DB.Create(&answer2).Error)

				return &participants[0]
			},
			validate: func(t *testing.T, resp *proto.AnswerList) {
				require.Equal(t, 2, len(resp.Answers))

				// Questions should be ordered by the order field
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
			baseTestCase: baseTestCase{
				name: "Fail - Participant not found",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.NotFound,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) *models.ExamParticipant {
				return &models.ExamParticipant{ID: 9999}
			},
		},
		{
			baseTestCase: baseTestCase{
				name: "Fail - No answers found",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.NotFound,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createDefaultTestExam(t, 2) // Created by different user
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &participant
			},
		},
		{
			baseTestCase: baseTestCase{
				name: "Success - Get answers as evaluator",
				metadata: map[string]string{
					"user_id": "2", // Using exam creator's ID (has EVALUATE permission)
				},
				expectedCode: codes.OK,
				userID:       2,
			},
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createDefaultTestExam(t, 2) // Created by user with ID 2, gets EVALUATE permission
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_INVITED}, // Different user as participant
				})

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				questions := createTestExamQuestions(t, &exam, []models.ExamQuestion{
					{
						Type:     proto.QuestionType_SUBJECTIVE,
						MaxScore: 10,
					},
					{
						Type:     proto.QuestionType_MCQ,
						MaxScore: 5,
					},
				})

				// Create answers with different evaluation states
				rawAnswer1 := json.RawMessage(`{"text": "Answer 1"}`)
				answer1 := models.Answer{
					ExamParticipantID: participant.ID,
					QuestionID:        questions[0].QuestionID,
					Answer:            &rawAnswer1,
					ScoreAwarded:      8,
					Comments:          sql.NullString{String: "Good answer", Valid: true},
					Evaluated:         true,
				}
				require.NoError(t, db.DB.Create(&answer1).Error)

				rawAnswer2 := json.RawMessage(`{"optionIndex": 1}`)
				answer2 := models.Answer{
					ExamParticipantID: participant.ID,
					QuestionID:        questions[1].QuestionID,
					Answer:            &rawAnswer2,
					ScoreAwarded:      0,
					Evaluated:         false,
				}
				require.NoError(t, db.DB.Create(&answer2).Error)

				return &participant
			},
			validate: func(t *testing.T, resp *proto.AnswerList) {
				require.Equal(t, 2, len(resp.Answers))

				// Questions should be ordered by the order field
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			participant := tt.setup(t)

			ctx := createContextWithMetadata(tt.metadata)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.ParticipantRequest{ParticipantId: int64(participant.ID)},
				client.GetParticipantAnswers,
				tt.validate,
			)
		})
	}
}

func TestGetAnswerForExam(t *testing.T) {
	tests := []getAnswerForExamTestCase{
		{
			baseTestCase: baseTestCase{
				name: "Success - Get single answer",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) (*models.Exam, types.QuestionID) {
				exam := createDefaultTestExam(t, 2)
				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
				}})

				answer := createTestAnswer(t, &participants[0], 1)
				return &exam, answer.QuestionID
			},
			validate: func(t *testing.T, resp *proto.AnswerMinimalResponse) {
				var answerData struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(resp.Answer, &answerData))
				assert.Equal(t, "Test Answer", answerData.Text)
				assert.NotZero(t, resp.AnswerId)
			},
		},
		{
			baseTestCase: baseTestCase{
				name: "Fail - User not a participant",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.PermissionDenied,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) (*models.Exam, types.QuestionID) {
				exam := createDefaultTestExam(t, 2)
				return &exam, 1
			},
		},
		{
			baseTestCase: baseTestCase{
				name: "Success - Answer not found returns empty response",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) (*models.Exam, types.QuestionID) {
				exam := createDefaultTestExam(t, 2)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				return &exam, 999 // Non-existent question ID
			},
			validate: func(t *testing.T, resp *proto.AnswerMinimalResponse) {
				assert.EqualValues(t, 999, getQuestionIdForHash(resp.QuestionHash))
				assert.Zero(t, resp.AnswerId)
				assert.Nil(t, resp.Answer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam, questionId := tt.setup(t)

			ctx := createContextWithMetadata(tt.metadata)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.GetAnswerRequest{
					ExamHash:     exam.Hash,
					QuestionHash: getQuestionHashForId(questionId),
				},
				client.GetAnswerForExam,
				tt.validate,
			)
		})
	}
}

func TestUpsertAnswer(t *testing.T) {
	tests := []upsertAnswerTestCase{
		{
			baseTestCase: baseTestCase{
				name: "Success - Create new answer",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createDefaultTestExam(t, 2)
				questions := createTestExamQuestions(t, &exam, []models.ExamQuestion{
					{
						Type:     proto.QuestionType_SUBJECTIVE,
						MaxScore: 10,
					},
				})

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
					ScheduledEndTime: sql.NullTime{
						Time:  time.Now().Add(time.Hour),
						Valid: true,
					},
				}})

				return &participants[0], &proto.UpsertAnswersRequest{
					ExamHash: exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: getQuestionHashForId(questions[0].QuestionID),
						Answer:       []byte(`{"text": "Test answer content"}`),
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
			validate: func(t *testing.T, resp *proto.UpsertAnswersResponse) {
				assert.NotZero(t, resp.AnswerId)

				// Verify answer in database
				var answer models.Answer
				require.NoError(t, db.DB.First(&answer, resp.AnswerId).Error)
				var answerData struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(*answer.Answer, &answerData))
				assert.Equal(t, "Test answer content", answerData.Text)
				assert.EqualValues(t, 1, answer.QuestionID)
			},
		},
		{
			baseTestCase: baseTestCase{
				name: "Success - Update existing answer",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createDefaultTestExam(t, 2)
				createTestExamQuestions(t, &exam, []models.ExamQuestion{
					{
						Type:     proto.QuestionType_SUBJECTIVE,
						MaxScore: 10,
					},
				})

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
					ScheduledEndTime: sql.NullTime{
						Time:  time.Now().Add(time.Hour),
						Valid: true,
					},
				}})

				// Create initial answer
				answer := createTestAnswer(t, &participants[0], 1)

				return &participants[0], &proto.UpsertAnswersRequest{
					ExamHash: exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: getQuestionHashForId(answer.QuestionID),
						Answer:       []byte(`{"text": "Updated answer content"}`),
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
			validate: func(t *testing.T, resp *proto.UpsertAnswersResponse) {
				assert.NotZero(t, resp.AnswerId)

				// Verify updated answer in database
				var answer models.Answer
				require.NoError(t, db.DB.First(&answer, resp.AnswerId).Error)
				var answerData struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(*answer.Answer, &answerData))
				assert.Equal(t, "Updated answer content", answerData.Text)
			},
		},
		{
			baseTestCase: baseTestCase{
				name: "Fail - Exam participant not found",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.PermissionDenied,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createDefaultTestExam(t, 2)
				return nil, &proto.UpsertAnswersRequest{
					ExamHash: exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: getQuestionHashForId(1),
						Answer:       []byte(`{"text": "Test answer"}`),
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
		},
		{
			baseTestCase: baseTestCase{
				name: "Fail - Exam not started",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.FailedPrecondition,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createDefaultTestExam(t, 2)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})

				return nil, &proto.UpsertAnswersRequest{
					ExamHash: exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: getQuestionHashForId(1),
						Answer:       []byte(`{"text": "Test answer"}`),
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
		},
		{
			baseTestCase: baseTestCase{
				name: "Fail - Exam ended",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.FailedPrecondition,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createDefaultTestExam(t, 2)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_ENDED},
				})

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)

				return &participant, &proto.UpsertAnswersRequest{
					ExamHash: exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: getQuestionHashForId(1),
						Answer:       []byte(`{"text": "Test answer after exam ended"}`),
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
		},
		{
			baseTestCase: baseTestCase{
				name: "Success - Empty answer for SUBJECTIVE question",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createDefaultTestExam(t, 2)
				questions := createTestExamQuestions(t, &exam, []models.ExamQuestion{
					{
						Type:     proto.QuestionType_SUBJECTIVE,
						MaxScore: 10,
					},
				})

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
					ScheduledEndTime: sql.NullTime{
						Time:  time.Now().Add(time.Hour),
						Valid: true,
					},
				}})

				return &participants[0], &proto.UpsertAnswersRequest{
					ExamHash: exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: getQuestionHashForId(questions[0].QuestionID),
						Answer:       []byte(`{"text": ""}`),
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
			validate: func(t *testing.T, resp *proto.UpsertAnswersResponse) {
				assert.NotZero(t, resp.AnswerId)

				var answer models.Answer
				require.NoError(t, db.DB.First(&answer, resp.AnswerId).Error)
				var answerData struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(*answer.Answer, &answerData))
				assert.Equal(t, "", answerData.Text)
			},
		},
		{
			baseTestCase: baseTestCase{
				name: "Success - Nil answer clears the answer",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createDefaultTestExam(t, 2)
				questions := createTestExamQuestions(t, &exam, []models.ExamQuestion{
					{
						Type:     proto.QuestionType_SUBJECTIVE,
						MaxScore: 10,
					},
				})

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
					ScheduledEndTime: sql.NullTime{
						Time:  time.Now().Add(time.Hour),
						Valid: true,
					},
				}})

				// Create initial answer
				answer := createTestAnswer(t, &participants[0], questions[0].QuestionID)

				return &participants[0], &proto.UpsertAnswersRequest{
					ExamHash: exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: getQuestionHashForId(answer.QuestionID),
						Answer:       nil, // Explicit nil answer
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
			validate: func(t *testing.T, resp *proto.UpsertAnswersResponse) {
				assert.NotZero(t, resp.AnswerId)

				var answer models.Answer
				require.NoError(t, db.DB.First(&answer, resp.AnswerId).Error)
				assert.Nil(t, answer.Answer)
			},
		},
		{
			baseTestCase: baseTestCase{
				name: "Fail - Empty MCQ answer is invalid",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.InvalidArgument,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createDefaultTestExam(t, 2)
				createTestExamQuestions(t, &exam, []models.ExamQuestion{
					{
						Type: proto.QuestionType_MCQ,
					},
				})

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
					ScheduledEndTime: sql.NullTime{
						Time:  time.Now().Add(time.Hour),
						Valid: true,
					},
				}})

				// First create an answer
				answer := createTestAnswer(t, &participants[0], 1)

				return &participants[0], &proto.UpsertAnswersRequest{
					ExamHash: exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: getQuestionHashForId(answer.QuestionID),
						Answer:       []byte(`{}`), // Empty answer object
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
		},
		{
			baseTestCase: baseTestCase{
				name: "Fail - Nil optionIndex in MCQ answer is invalid",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.InvalidArgument,
				userID:       typedUserID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createDefaultTestExam(t, 2)
				createTestExamQuestions(t, &exam, []models.ExamQuestion{
					{Type: proto.QuestionType_MCQ},
				})

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
					ScheduledEndTime: sql.NullTime{
						Time:  time.Now().Add(time.Hour),
						Valid: true,
					},
				}})

				// First create an answer
				answer := createTestAnswer(t, &participants[0], 1)

				return &participants[0], &proto.UpsertAnswersRequest{
					ExamHash: exam.Hash,
					Answer: &proto.Answer{
						QuestionHash: getQuestionHashForId(answer.QuestionID),
						Answer:       []byte(`{"optionIndex": null}`), // explicit null for optionIndex
						SubmittedAt:  timestamppb.Now(),
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			_, req := tt.setup(t)

			ctx := createContextWithMetadata(tt.metadata)
			testrunner.Runner(t, ctx, tt.expectedCode,
				req,
				client.UpsertAnswer,
				tt.validate,
			)
		})
	}
}
