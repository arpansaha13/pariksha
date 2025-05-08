package handlers

import (
	"bytes"
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/db"
)

func TestGetAnswerForEvaluation(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.Answer, *models.ExamParticipant)
		metadata     map[string]string
		request      *proto.GetAnswerForEvaluationRequest
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.AnswerResponse)
	}{
		{
			name: "Success - Get answer for evaluation",
			setup: func(t *testing.T) (*models.Answer, *models.ExamParticipant) {
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				answer := createTestAnswer(t, &participant, 1)
				return &answer, &participant
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.AnswerResponse) {
				assert.NotNil(t, resp)
				assert.True(t, bytes.Equal([]byte(`{"text":"Test Answer"}`), resp.Answer))
				assert.EqualValues(t, 5, resp.ScoreAwarded)
				assert.Equal(t, "Test Comment", resp.Comments)
			},
		},
		{
			name: "Success - Answer not found returns empty response",
			setup: func(t *testing.T) (*models.Answer, *models.ExamParticipant) {
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				return nil, &participant
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.AnswerResponse) {
				assert.NotNil(t, resp)
				assert.EqualValues(t, 0, resp.Id)
				assert.EqualValues(t, 0, resp.ExamParticipantId)
				assert.Empty(t, resp.Answer)
				assert.EqualValues(t, 0, resp.ScoreAwarded)
				assert.Empty(t, resp.Comments)
			},
		},
		{
			name: "Fail - Participant does not exist",
			setup: func(t *testing.T) (*models.Answer, *models.ExamParticipant) {
				return nil, &models.ExamParticipant{ID: 99999} // Non-existent ID
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Fail - Exam not ended",
			setup: func(t *testing.T) (*models.Answer, *models.ExamParticipant) {
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{{UserID: 2, Status: constants.PARTICIPANT_STATUS_STARTED}}) // Participant still in STARTED state
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				answer := createTestAnswer(t, &participant, 1)
				return &answer, &participant
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.FailedPrecondition,
		},
		{
			name: "Success - Question ID does not exist returns empty response",
			setup: func(t *testing.T) (*models.Answer, *models.ExamParticipant) {
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				// Create an answer for question ID 1
				answer := createTestAnswer(t, &participant, 1)
				return &answer, &participant
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			request: &proto.GetAnswerForEvaluationRequest{
				QuestionId: 9999, // Non-existent question ID
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.AnswerResponse) {
				assert.NotNil(t, resp)
				assert.EqualValues(t, 9999, resp.QuestionId)
				assert.EqualValues(t, 0, resp.Id)
				assert.EqualValues(t, 0, resp.ExamParticipantId)
				assert.Empty(t, resp.Answer)
				assert.EqualValues(t, 0, resp.ScoreAwarded)
				assert.Empty(t, resp.Comments)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			_, participant := tt.setup(t)

			ctx := createContextWithMetadata(tt.metadata)
			var req *proto.GetAnswerForEvaluationRequest

			if tt.request != nil {
				req = tt.request
				req.ParticipantId = participant.ID
			} else {
				req = &proto.GetAnswerForEvaluationRequest{
					ParticipantId: participant.ID,
					QuestionId:    1,
				}
			}

			resp, err := client.GetAnswerForEvaluation(ctx, req)

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

func TestUpdateAnswerForEvaluation(t *testing.T) {
	newScore := int32(8)
	evaluated := true
	comments := "Good attempt"
	exceedingScore := int32(15) // Exceeds max_score in exam_questions

	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Answer
		request      *proto.UpdateAnswerRequest
		metadata     map[string]string
		expectedCode codes.Code
		validate     func(t *testing.T, answerId int64)
	}{
		{
			name: "Success - Update all fields",
			setup: func(t *testing.T) *models.Answer {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{{UserID: userID, Status: constants.PARTICIPANT_STATUS_ENDED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				participant.ScoreAwarded = 5 // Initial score
				require.NoError(t, db.DB.Save(&participant).Error)

				// Create exam question with max score of 10
				createTestExamQuestion(t, &exam, 1, 10)

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
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, answerId int64) {
				var answer models.Answer
				require.NoError(t, db.DB.First(&answer, answerId).Error)
				assert.EqualValues(t, newScore, answer.ScoreAwarded)
				assert.Equal(t, comments, answer.Comments.String)
				assert.True(t, answer.Evaluated)

				// Verify participant total score is updated
				var participant models.ExamParticipant
				require.NoError(t, db.DB.First(&participant, answer.ExamParticipantID).Error)
				assert.EqualValues(t, newScore, participant.ScoreAwarded)
			},
		},
		{
			name: "Success - Update only comments",
			setup: func(t *testing.T) *models.Answer {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{{UserID: userID, Status: constants.PARTICIPANT_STATUS_ENDED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				// Create exam question with max score of 10
				createTestExamQuestion(t, &exam, 1, 10)

				answer := createTestAnswer(t, &participant, 1)
				return &answer
			},
			request: &proto.UpdateAnswerRequest{
				Comments: &comments,
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, answerId int64) {
				var answer models.Answer
				require.NoError(t, db.DB.First(&answer, answerId).Error)
				assert.Equal(t, comments, answer.Comments.String)
				assert.EqualValues(t, 5, answer.ScoreAwarded) // Original score unchanged
			},
		},
		{
			name: "Fail - Answer not found",
			setup: func(t *testing.T) *models.Answer {
				exam := createTestExam(t, 2)
				// Create exam question even for non-existent answer to maintain consistency
				createTestExamQuestion(t, &exam, 1, 10)
				return &models.Answer{ID: 9999}
			},
			request: &proto.UpdateAnswerRequest{
				NewScore: &newScore,
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Fail - Score exceeds max score",
			setup: func(t *testing.T) *models.Answer {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
				}{{UserID: userID, Status: constants.PARTICIPANT_STATUS_ENDED}})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				// Create exam question with max score of 10
				createTestExamQuestion(t, &exam, 1, 10)

				answer := createTestAnswer(t, &participant, 1)
				return &answer
			},
			request: &proto.UpdateAnswerRequest{
				NewScore: &exceedingScore, // Try to set score higher than max_score
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.InvalidArgument,
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

func TestMarkParticipantAsEvaluated(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.ExamParticipant
		expectedCode codes.Code
		validate     func(t *testing.T, participant *models.ExamParticipant, resp *proto.EvaluationStatusResponse)
	}{
		{
			name: "Success - All answers evaluated",
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
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
				assert.EqualValues(t, 0, resp.UnevaluatedCount)

				// Verify participant status changed to EVALUATED
				var updatedParticipant models.ExamParticipant
				require.NoError(t, db.DB.First(&updatedParticipant, participant.ID).Error)
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_EVALUATED, updatedParticipant.Status)
			},
		},
		{
			name: "Success - Some answers unevaluated",
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
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
				assert.EqualValues(t, 1, resp.UnevaluatedCount)

				// Verify participant status remains ENDED
				var updatedParticipant models.ExamParticipant
				require.NoError(t, db.DB.First(&updatedParticipant, participant.ID).Error)
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_ENDED, updatedParticipant.Status)
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
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int16
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

			resp, err := client.MarkParticipantAsEvaluated(context.Background(), &proto.ParticipantRequest{
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
