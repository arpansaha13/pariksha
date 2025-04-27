package handlers

import (
	"context"
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
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
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
				assert.Equal(t, 2, len(resp.Answers))
				for _, answer := range resp.Answers {
					var answerData struct {
						Text string `json:"text"`
					}
					require.NoError(t, json.Unmarshal(answer.Answer, &answerData))
					assert.Equal(t, "Test Answer", answerData.Text)
					assert.Equal(t, "Test Comment", answer.Comments)
					assert.Equal(t, int32(5), answer.ScoreAwarded)
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
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
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

func TestGetAnswer(t *testing.T) {
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
					Status int
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
			name: "Fail - Answer not found",
			setup: func(t *testing.T) (*models.Exam, int64) {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
				}{{UserID: userID, Status: constants.PARTICIPANT_STATUS_STARTED}})
				require.NoError(t, err)
				return &exam, 9999 // Non-existent question ID
			},
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
			},
			expectedCode: codes.NotFound,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam, questionId := tt.setup(t)

			ctx := createContextWithMetadata(tt.metadata)
			resp, err := client.GetAnswer(ctx, &proto.GetAnswerRequest{
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
					Status int
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
				require.NoError(t, json.Unmarshal(answer.Answer, &answerData))
				assert.Equal(t, "Test answer content", answerData.Text)
				assert.Equal(t, int64(1), answer.QuestionID)
			},
		},
		{
			name: "Success - Update existing answer",
			setup: func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest) {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
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
				require.NoError(t, json.Unmarshal(answer.Answer, &answerData))
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
					Status int
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
					Status int
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

func TestUpdateAnswerForEvaluation(t *testing.T) {
	maxScore := int32(10)
	newScore := int32(8)
	evaluated := true
	comments := "Good attempt"

	tests := []struct {
		name          string
		setup         func(t *testing.T) *models.Answer
		request       *proto.UpdateAnswerRequest
		metadata      map[string]string
		expectedCode  codes.Code
		validate      func(t *testing.T, answerId int64)
		questionScore int
	}{
		{
			name: "Success - Update all fields",
			setup: func(t *testing.T) *models.Answer {
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
				}{{UserID: userID, Status: constants.PARTICIPANT_STATUS_ENDED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				participant.ScoreAwarded = 5 // Initial score
				require.NoError(t, db.DB.Save(&participant).Error)

				answer := createTestAnswer(t, &participant, 1)
				answer.ScoreAwarded = 5 // Initial score
				answer.Evaluated = false
				require.NoError(t, db.DB.Save(&answer).Error)

				return &answer
			},
			request: &proto.UpdateAnswerRequest{
				NewScore:  &newScore,
				Evaluated: &evaluated,
				Comments:  &comments,
			},
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, answerId int64) {
				var answer models.Answer
				require.NoError(t, db.DB.First(&answer, answerId).Error)
				assert.Equal(t, int(newScore), answer.ScoreAwarded)
				assert.Equal(t, comments, answer.Comments.String)
				assert.True(t, answer.Evaluated)

				// Verify participant total score is updated
				var participant models.ExamParticipant
				require.NoError(t, db.DB.First(&participant, answer.ExamParticipantID).Error)
				assert.Equal(t, int(newScore), participant.ScoreAwarded)
			},
			questionScore: 10,
		},
		{
			name: "Success - Update only comments",
			setup: func(t *testing.T) *models.Answer {
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
				}{{UserID: userID, Status: constants.PARTICIPANT_STATUS_ENDED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				answer := createTestAnswer(t, &participant, 1)
				return &answer
			},
			request: &proto.UpdateAnswerRequest{
				Comments: &comments,
			},
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, answerId int64) {
				var answer models.Answer
				require.NoError(t, db.DB.First(&answer, answerId).Error)
				assert.Equal(t, comments, answer.Comments.String)
				assert.Equal(t, 5, answer.ScoreAwarded) // Original score unchanged
			},
			questionScore: 10,
		},
		{
			name: "Fail - Score exceeds max score",
			setup: func(t *testing.T) *models.Answer {
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
				}{{UserID: userID, Status: constants.PARTICIPANT_STATUS_ENDED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				answer := createTestAnswer(t, &participant, 1)
				return &answer
			},
			request: &proto.UpdateAnswerRequest{
				NewScore: &maxScore,
			},
			metadata: map[string]string{
				"question_score": "5", // Max score less than requested score
			},
			expectedCode:  codes.InvalidArgument,
			questionScore: 5,
		},
		{
			name: "Fail - Answer not found",
			setup: func(t *testing.T) *models.Answer {
				return &models.Answer{ID: 9999}
			},
			request: &proto.UpdateAnswerRequest{
				NewScore: &newScore,
			},
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
			},
			expectedCode:  codes.NotFound,
			questionScore: 10,
		},
		{
			name: "Fail - Missing question score metadata",
			setup: func(t *testing.T) *models.Answer {
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
				}{{UserID: userID, Status: constants.PARTICIPANT_STATUS_ENDED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				answer := createTestAnswer(t, &participant, 1)
				return &answer
			},
			request: &proto.UpdateAnswerRequest{
				NewScore: &newScore,
			},
			metadata:      map[string]string{},
			expectedCode:  codes.Internal,
			questionScore: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			answer := tt.setup(t)
			tt.request.AnswerId = answer.ID

			ctx := createContextWithMetadata(tt.metadata)
			_, err := client.UpdateAnswerForEvaluation(ctx, tt.request)

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, answer.ID)
			}
		})
	}
}

func TestMarkAsEvaluated(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.ExamParticipant
		expectedCode codes.Code
		validate     func(t *testing.T, participant *models.ExamParticipant, resp *proto.EvaluationStatusResponse)
	}{
		{
			name: "Success - All answers evaluated",
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
				}{{UserID: userID, Status: constants.PARTICIPANT_STATUS_ENDED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				// Create answers that are all evaluated
				answer1 := createTestAnswer(t, &participant, 1)
				answer1.Evaluated = true
				answer2 := createTestAnswer(t, &participant, 2)
				answer2.Evaluated = true

				require.NoError(t, db.DB.Save(&answer1).Error)
				require.NoError(t, db.DB.Save(&answer2).Error)

				return &participant
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, participant *models.ExamParticipant, resp *proto.EvaluationStatusResponse) {
				// Verify unevaluated count is 0
				assert.Equal(t, int32(0), resp.UnevaluatedCount)

				// Verify participant status changed to EVALUATED
				var updatedParticipant models.ExamParticipant
				require.NoError(t, db.DB.First(&updatedParticipant, participant.ID).Error)
				assert.Equal(t, constants.PARTICIPANT_STATUS_EVALUATED, updatedParticipant.Status)
			},
		},
		{
			name: "Success - Some answers unevaluated",
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
				}{{UserID: userID, Status: constants.PARTICIPANT_STATUS_ENDED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				// Create one evaluated and one unevaluated answer
				answer1 := createTestAnswer(t, &participant, 1)
				answer1.Evaluated = true
				answer2 := createTestAnswer(t, &participant, 2)
				answer2.Evaluated = false

				require.NoError(t, db.DB.Save(&answer1).Error)
				require.NoError(t, db.DB.Save(&answer2).Error)

				return &participant
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, participant *models.ExamParticipant, resp *proto.EvaluationStatusResponse) {
				// Verify one answer is still unevaluated
				assert.Equal(t, int32(1), resp.UnevaluatedCount)

				// Verify participant status remains ENDED
				var updatedParticipant models.ExamParticipant
				require.NoError(t, db.DB.First(&updatedParticipant, participant.ID).Error)
				assert.Equal(t, constants.PARTICIPANT_STATUS_ENDED, updatedParticipant.Status)
			},
		},
		{
			name: "Fail - Participant not found",
			setup: func(t *testing.T) *models.ExamParticipant {
				return &models.ExamParticipant{ID: 9999}
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Fail - Exam not ended",
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
				}{{UserID: userID, Status: constants.PARTICIPANT_STATUS_STARTED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &participant
			},
			expectedCode: codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			participant := tt.setup(t)

			resp, err := client.MarkAsEvaluated(context.Background(), &proto.ParticipantRequest{
				ParticipantId: participant.ID,
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, participant, resp)
		})
	}
}

func TestGetAnswerById(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Answer
		metadata     map[string]string
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.AnswerResponse)
	}{
		{
			name: "Success - Get answer by ID",
			setup: func(t *testing.T) *models.Answer {
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
				}{{UserID: userID, Status: 1}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				answer := createTestAnswer(t, &participant, 1)
				return &answer
			},
			metadata: map[string]string{
				"user_id":        strconv.FormatInt(userID, 10),
				"question_score": "10",
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.AnswerResponse) {
				var answerData struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(resp.Answer, &answerData))
				assert.Equal(t, "Test Answer", answerData.Text)
				assert.Equal(t, "Test Comment", resp.Comments)
				assert.Equal(t, int32(5), resp.ScoreAwarded)
				assert.NotZero(t, resp.Id)
				assert.NotZero(t, resp.ExamParticipantId)
				assert.NotZero(t, resp.QuestionId)
			},
		},
		{
			name: "Fail - Answer not found",
			setup: func(t *testing.T) *models.Answer {
				return &models.Answer{ID: 9999} // Non-existent answer ID
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
			answer := tt.setup(t)

			ctx := createContextWithMetadata(tt.metadata)
			resp, err := client.GetAnswerById(ctx, &proto.GetAnswerByIdRequest{
				AnswerId: answer.ID,
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
