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
	"pariksha/common/pkg/utils/testrunner"
	"pariksha/exam/internal/config/db"
)

func TestGetParticipantAnswers(t *testing.T) {
	tests := []ParticipantTestCase[*proto.AnswerList]{
		{
			BaseTestCase: BaseTestCase{
				name: "Success - Get multiple answers with evaluation data",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       userID,
			},
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: 1},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				// Create questions first
				q1 := createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 10,
					Order:      1,
					Type:       constants.QUESTION_TYPE_SUBJECTIVE,
					MaxScore:   10,
				})
				q2 := createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 2,
					CategoryID: 10,
					Order:      2,
					Type:       constants.QUESTION_TYPE_MCQ,
					MaxScore:   5,
				})

				// Create answers with different evaluation states
				rawAnswer1 := json.RawMessage(`{"text": "Answer 1"}`)
				answer1 := models.Answer{
					ExamParticipantID: participant.ID,
					QuestionID:        q1.QuestionID,
					Answer:            &rawAnswer1,
					ScoreAwarded:      8,
					Comments:          sql.NullString{String: "Good answer", Valid: true},
					Evaluated:         true,
				}
				require.NoError(t, db.DB.Create(&answer1).Error)

				rawAnswer2 := json.RawMessage(`{"optionIndex": 1}`)
				answer2 := models.Answer{
					ExamParticipantID: participant.ID,
					QuestionID:        q2.QuestionID,
					Answer:            &rawAnswer2,
					ScoreAwarded:      0, // Not evaluated yet
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
				assert.EqualValues(t, constants.QUESTION_TYPE_SUBJECTIVE, answer1.QuestionType)
				assert.EqualValues(t, 10, answer1.MaxScore)

				var answerData1 struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(answer1.Answer, &answerData1))
				assert.Equal(t, "Answer 1", answerData1.Text)

				answer2 := resp.Answers[1]
				assert.EqualValues(t, 2, answer2.Order)
				assert.EqualValues(t, constants.QUESTION_TYPE_MCQ, answer2.QuestionType)
				assert.EqualValues(t, 5, answer2.MaxScore)

				var answerData2 struct {
					OptionIndex int `json:"optionIndex"`
				}
				require.NoError(t, json.Unmarshal(answer2.Answer, &answerData2))
				assert.Equal(t, 1, answerData2.OptionIndex)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name: "Fail - Participant not found",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.NotFound,
				userID:       userID,
			},
			setup: func(t *testing.T) *models.ExamParticipant {
				return &models.ExamParticipant{ID: 9999}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name: "Fail - No answers found",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.NotFound,
				userID:       userID,
			},
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createTestExam(t, 2) // Created by different user
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: 1},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &participant
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name: "Success - Get answers as evaluator",
				metadata: map[string]string{
					"user_id": "2", // Using exam creator's ID (has EVALUATE permission)
				},
				expectedCode: codes.OK,
				userID:       2,
			},
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createTestExam(t, 2) // Created by user with ID 2, gets EVALUATE permission
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 3, Status: 1}, // Different user as participant
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				q1 := createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 10,
					Order:      1,
					Type:       constants.QUESTION_TYPE_SUBJECTIVE,
					MaxScore:   10,
				})
				q2 := createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 2,
					CategoryID: 10,
					Order:      2,
					Type:       constants.QUESTION_TYPE_MCQ,
					MaxScore:   5,
				})

				// Create answers with different evaluation states
				rawAnswer1 := json.RawMessage(`{"text": "Answer 1"}`)
				answer1 := models.Answer{
					ExamParticipantID: participant.ID,
					QuestionID:        q1.QuestionID,
					Answer:            &rawAnswer1,
					ScoreAwarded:      8,
					Comments:          sql.NullString{String: "Good answer", Valid: true},
					Evaluated:         true,
				}
				require.NoError(t, db.DB.Create(&answer1).Error)

				rawAnswer2 := json.RawMessage(`{"optionIndex": 1}`)
				answer2 := models.Answer{
					ExamParticipantID: participant.ID,
					QuestionID:        q2.QuestionID,
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
				assert.EqualValues(t, constants.QUESTION_TYPE_SUBJECTIVE, answer1.QuestionType)
				assert.EqualValues(t, 10, answer1.MaxScore)

				var answerData1 struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(answer1.Answer, &answerData1))
				assert.Equal(t, "Answer 1", answerData1.Text)

				answer2 := resp.Answers[1]
				assert.EqualValues(t, 2, answer2.Order)
				assert.EqualValues(t, constants.QUESTION_TYPE_MCQ, answer2.QuestionType)
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
				&proto.ParticipantRequest{ParticipantId: participant.ID},
				client.GetParticipantAnswers,
				tt.validate,
			)
		})
	}
}

func TestGetAnswerForExam(t *testing.T) {
	tests := []ExamTestCase[*proto.AnswerMinimalResponse]{
		{
			BaseTestCase: BaseTestCase{
				name: "Success - Get single answer",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       userID,
			},
			setup: func(t *testing.T) (*models.Exam, int64) {
				exam := createTestExam(t, 2) // Created by different user
				err := createTestExamParticipants(t, &exam, []TestParticipantData{{UserID: userID, Status: constants.PARTICIPANT_STATUS_STARTED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				answer := createTestAnswer(t, &participant, 1)
				return &exam, answer.QuestionID
			},
			validate: func(t *testing.T, resp *proto.AnswerMinimalResponse) {
				var answerData struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(resp.Answer, &answerData))
				assert.Equal(t, "Test Answer", answerData.Text)
				assert.NotZero(t, resp.Id)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name: "Fail - User not a participant",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.PermissionDenied,
				userID:       userID,
			},
			setup: func(t *testing.T) (*models.Exam, int64) {
				exam := createTestExam(t, 2)
				return &exam, 1
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name: "Success - Answer not found returns empty response",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       userID,
			},
			setup: func(t *testing.T) (*models.Exam, int64) {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []TestParticipantData{{UserID: userID, Status: constants.PARTICIPANT_STATUS_STARTED}})
				require.NoError(t, err)
				return &exam, 9999 // Non-existent question ID
			},
			validate: func(t *testing.T, resp *proto.AnswerMinimalResponse) {
				assert.EqualValues(t, 9999, resp.QuestionId)
				assert.Zero(t, resp.Id)
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
					ExamId:     exam.ID,
					QuestionId: questionId,
				},
				client.GetAnswerForExam,
				tt.validate,
			)
		})
	}
}

func TestUpsertAnswer(t *testing.T) {
	tests := []ParticipantRequestTestCase[*proto.UpsertAnswersResponse, *proto.UpsertAnswersRequest]{
		{
			BaseTestCase: BaseTestCase{
				name: "Success - Create new answer",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       userID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2) // Created by different user
				createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 10,
					MaxScore:   10,
					Type:       constants.QUESTION_TYPE_SUBJECTIVE,
				})

				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)

				// Set scheduled end time to 1 hour in the future
				participant.ScheduledEndTime = sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true}
				require.NoError(t, db.DB.Save(&participant).Error)

				return &participant, &proto.UpsertAnswersRequest{
					ExamId: exam.ID,
					Answer: &proto.Answer{
						QuestionId:  1,
						Answer:      []byte(`{"text": "Test answer content"}`),
						SubmittedAt: timestamppb.Now(),
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
			BaseTestCase: BaseTestCase{
				name: "Success - Update existing answer",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       userID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 10,
					MaxScore:   10,
					Type:       constants.QUESTION_TYPE_SUBJECTIVE,
				})

				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)

				// Set scheduled end time to 1 hour in the future
				participant.ScheduledEndTime = sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true}
				require.NoError(t, db.DB.Save(&participant).Error)

				// Create initial answer
				answer := createTestAnswer(t, &participant, 1)

				return &participant, &proto.UpsertAnswersRequest{
					ExamId: exam.ID,
					Answer: &proto.Answer{
						QuestionId:  answer.QuestionID,
						Answer:      []byte(`{"text": "Updated answer content"}`),
						SubmittedAt: timestamppb.Now(),
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
			BaseTestCase: BaseTestCase{
				name: "Fail - Exam participant not found",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.PermissionDenied,
				userID:       userID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				return nil, &proto.UpsertAnswersRequest{
					ExamId: exam.ID,
					Answer: &proto.Answer{
						QuestionId:  1,
						Answer:      []byte(`{"text": "Test answer"}`),
						SubmittedAt: timestamppb.Now(),
					},
				}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name: "Fail - Exam not started",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.FailedPrecondition,
				userID:       userID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				require.NoError(t, err)

				return nil, &proto.UpsertAnswersRequest{
					ExamId: exam.ID,
					Answer: &proto.Answer{
						QuestionId:  1,
						Answer:      []byte(`{"text": "Test answer"}`),
						SubmittedAt: timestamppb.Now(),
					},
				}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name: "Fail - Exam ended",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.FailedPrecondition,
				userID:       userID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: constants.PARTICIPANT_STATUS_ENDED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)

				return &participant, &proto.UpsertAnswersRequest{
					ExamId: exam.ID,
					Answer: &proto.Answer{
						QuestionId:  1,
						Answer:      []byte(`{"text": "Test answer after exam ended"}`),
						SubmittedAt: timestamppb.Now(),
					},
				}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name: "Success - Empty answer for SUBJECTIVE question",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       userID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 10,
					MaxScore:   10,
					Type:       constants.QUESTION_TYPE_SUBJECTIVE,
				})

				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)
				participant.ScheduledEndTime = sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true}
				require.NoError(t, db.DB.Save(&participant).Error)

				return &participant, &proto.UpsertAnswersRequest{
					ExamId: exam.ID,
					Answer: &proto.Answer{
						QuestionId:  1,
						Answer:      []byte(`{"text": ""}`),
						SubmittedAt: timestamppb.Now(),
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
			BaseTestCase: BaseTestCase{
				name: "Success - Nil answer clears the answer",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.OK,
				userID:       userID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 10,
					MaxScore:   10,
					Type:       constants.QUESTION_TYPE_SUBJECTIVE,
				})

				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)
				participant.ScheduledEndTime = sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true}
				require.NoError(t, db.DB.Save(&participant).Error)

				// First create an answer
				answer := createTestAnswer(t, &participant, 1)

				return &participant, &proto.UpsertAnswersRequest{
					ExamId: exam.ID,
					Answer: &proto.Answer{
						QuestionId:  answer.QuestionID,
						Answer:      nil, // Explicit nil answer
						SubmittedAt: timestamppb.Now(),
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
			BaseTestCase: BaseTestCase{
				name: "Fail - Empty MCQ answer is invalid",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.InvalidArgument,
				userID:       userID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 10,
					MaxScore:   10,
					Type:       constants.QUESTION_TYPE_MCQ,
				})

				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)
				participant.ScheduledEndTime = sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true}
				require.NoError(t, db.DB.Save(&participant).Error)

				// First create an answer
				answer := createTestAnswer(t, &participant, 1)

				return &participant, &proto.UpsertAnswersRequest{
					ExamId: exam.ID,
					Answer: &proto.Answer{
						QuestionId:  answer.QuestionID,
						Answer:      []byte(`{}`), // Empty answer object
						SubmittedAt: timestamppb.Now(),
					},
				}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name: "Fail - Nil optionIndex in MCQ answer is invalid",
				metadata: map[string]string{
					"user_id": strconv.FormatInt(userID, 10),
				},
				expectedCode: codes.InvalidArgument,
				userID:       userID,
			},
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 10,
					MaxScore:   10,
					Type:       constants.QUESTION_TYPE_MCQ,
				})

				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)
				participant.ScheduledEndTime = sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true}
				require.NoError(t, db.DB.Save(&participant).Error)

				// First create an answer
				answer := createTestAnswer(t, &participant, 1)

				return &participant, &proto.UpsertAnswersRequest{
					ExamId: exam.ID,
					Answer: &proto.Answer{
						QuestionId:  answer.QuestionID,
						Answer:      []byte(`{"optionIndex": null}`), // explicit null for optionIndex
						SubmittedAt: timestamppb.Now(),
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
