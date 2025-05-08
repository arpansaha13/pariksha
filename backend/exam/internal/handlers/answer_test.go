package handlers

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/db"
)

func TestGetParticipantAnswers(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.ExamParticipant
		metadata     map[string]string
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.AnswerList)
	}{
		{
			name: "Success - Get multiple answers",
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createTestExam(t, 2) // Created by different user
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{{UserID: userID, Status: 1}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				// Create multiple answers
				createTestAnswer(t, &participant, 1)
				createTestAnswer(t, &participant, 2)

				return &participant
			},
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.AnswerList) {
				assert.EqualValues(t, 2, len(resp.Answers))
				for _, answer := range resp.Answers {
					var answerData struct {
						Text string `json:"text"`
					}
					require.NoError(t, json.Unmarshal(answer.Answer, &answerData))
					assert.Equal(t, "Test Answer", answerData.Text)
					assert.Equal(t, "Test Comment", answer.Comments)
					assert.EqualValues(t, 5, answer.ScoreAwarded)
				}
			},
		},
		{
			name: "Fail - Participant not found",
			setup: func(t *testing.T) *models.ExamParticipant {
				return &models.ExamParticipant{ID: 9999}
			},
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Fail - No answers found",
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createTestExam(t, 2) // Created by different user
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{{UserID: userID, Status: 1}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &participant
			},
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			participant := tt.setup(t)

			ctx := createContextWithMetadata(tt.metadata)
			resp, err := client.GetParticipantAnswers(ctx, &proto.ParticipantRequest{
				ParticipantId: participant.ID,
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)
		})
	}
}

func TestGetAnswerForExam(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.Exam, int64)
		metadata     map[string]string
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.GetAnswerResponse)
	}{
		{
			name: "Success - Get single answer",
			setup: func(t *testing.T) (*models.Exam, int64) {
				exam := createTestExam(t, 2) // Created by different user
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{{UserID: userID, Status: constants.PARTICIPANT_STATUS_STARTED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				answer := createTestAnswer(t, &participant, 1)
				return &exam, answer.QuestionID
			},
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.GetAnswerResponse) {
				var answerData struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(resp.Answer, &answerData))
				assert.Equal(t, "Test Answer", answerData.Text)
				assert.NotZero(t, resp.Id)
			},
		},
		{
			name: "Fail - User not a participant",
			setup: func(t *testing.T) (*models.Exam, int64) {
				exam := createTestExam(t, 2)
				return &exam, 1
			},
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Success - Answer not found returns empty response",
			setup: func(t *testing.T) (*models.Exam, int64) {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{{UserID: userID, Status: constants.PARTICIPANT_STATUS_STARTED}})
				require.NoError(t, err)
				return &exam, 9999 // Non-existent question ID
			},
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.GetAnswerResponse) {
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
			resp, err := client.GetAnswerForExam(ctx, &proto.GetAnswerRequest{
				ExamId:     exam.ID,
				QuestionId: questionId,
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)
		})
	}
}

func TestUpsertAnswer(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest)
		userID       int64
		metadata     map[string]string
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.UpsertAnswersResponse)
	}{
		{
			name: "Success - Create new answer",
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2) // Created by different user
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{
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
			userID: userID,
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
				"question_type":  constants.QUESTION_TYPE_LONG,
			},
			expectedCode: codes.OK,
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
			name: "Success - Update existing answer",
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{
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
			userID: userID,
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
				"question_type":  constants.QUESTION_TYPE_LONG,
			},
			expectedCode: codes.OK,
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
			name: "Fail - Exam participant not found",
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
			userID: userID,
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
				"question_type":  constants.QUESTION_TYPE_LONG,
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Fail - Exam not started",
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{
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
			userID: userID,
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
				"question_type":  constants.QUESTION_TYPE_LONG,
			},
			expectedCode: codes.FailedPrecondition,
		},
		{
			name: "Fail - Exam ended",
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{
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
			userID: userID,
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
				"question_type":  constants.QUESTION_TYPE_LONG,
			},
			expectedCode: codes.FailedPrecondition,
		},
		{
			name: "Success - Empty answer for SHORT question",
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{
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
			userID: userID,
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
				"question_type":  constants.QUESTION_TYPE_SHORT,
			},
			expectedCode: codes.OK,
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
			name: "Success - Nil answer clears the answer",
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{
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
			userID: userID,
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
				"question_type":  constants.QUESTION_TYPE_MCQ,
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.UpsertAnswersResponse) {
				assert.NotZero(t, resp.AnswerId)

				var answer models.Answer
				require.NoError(t, db.DB.First(&answer, resp.AnswerId).Error)
				assert.Nil(t, answer.Answer)
			},
		},
		{
			name: "Fail - Empty MCQ answer is invalid",
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{
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
			userID: userID,
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
				"question_type":  constants.QUESTION_TYPE_MCQ,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Fail - Nil optionIndex in MCQ answer is invalid",
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{
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
			userID: userID,
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
				"question_type":  constants.QUESTION_TYPE_MCQ,
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			_, req := tt.setup(t)

			ctx := createContextWithMetadata(tt.metadata)
			resp, err := client.UpsertAnswer(ctx, req)

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)
		})
	}
}
