package tests

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/testrunner"
	"pariksha/exam/internal/config/db"
)

func TestGetAnswerForEvaluation(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.ExamParticipant, types.QuestionID)
		metadata     map[string]string
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.AnswerMinimalResponse)
	}{
		{
			name: "Success - Get answer as evaluator",
			setup: func(t *testing.T) (*models.ExamParticipant, types.QuestionID) {
				exam := createTestExam(t, 2) // Created by different user
				createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 1,
					MaxScore:   10,
					Type:       constants.QUESTION_TYPE_SUBJECTIVE,
				})

				// Create evaluator permission for test user
				permission := models.ExamPermission{
					ExamID: exam.ID,
					UserID: typedUserID,
				}
				permission.SetEvaluate()
				require.NoError(t, db.DB.Create(&permission).Error)

				// Create participant and answer
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_ENDED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				answer := createTestAnswer(t, &participant, 1)

				return &participant, answer.QuestionID
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.AnswerMinimalResponse) {
				assert.NotZero(t, resp.Id)
				assert.NotNil(t, resp.Answer)
				var answerData struct {
					Text string `json:"text"`
				}
				require.NoError(t, json.Unmarshal(resp.Answer, &answerData))
				assert.Equal(t, "Test Answer", answerData.Text)
			},
		},
		{
			name: "Fail - No evaluate permission",
			setup: func(t *testing.T) (*models.ExamParticipant, types.QuestionID) {
				exam := createTestExam(t, 2) // Created by different user

				// Create participant and answer
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_ENDED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				answer := createTestAnswer(t, &participant, 1)

				return &participant, answer.QuestionID
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Success - Empty response for non-existent answer",
			setup: func(t *testing.T) (*models.ExamParticipant, types.QuestionID) {
				exam := createTestExam(t, typedUserID) // Create as owner to get evaluate permission

				// Create participant without answer
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				return &participant, 999 // Non-existent question ID
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.AnswerMinimalResponse) {
				assert.Zero(t, resp.Id)
				assert.Nil(t, resp.Answer)
				assert.EqualValues(t, 999, resp.QuestionId)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			participant, questionId := tt.setup(t)

			ctx := createContextWithMetadata(tt.metadata)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.ParticipantQuestionRequest{
					ParticipantId: int64(participant.ID),
					QuestionId:    int64(questionId),
				},
				client.GetAnswerForEvaluation,
				tt.validate,
			)
		})
	}
}

func TestGetAnswerEvaluationData(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.ExamParticipant, types.QuestionID)
		metadata     map[string]string
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.GetAnswerEvaluationDataResponse)
	}{
		{
			name: "Success - Get evaluation data as evaluator",
			setup: func(t *testing.T) (*models.ExamParticipant, types.QuestionID) {
				exam := createTestExam(t, 2) // Created by different user
				createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 10,
					MaxScore:   5,
				})

				// Create evaluator permission for test user
				permission := models.ExamPermission{
					ExamID: exam.ID,
					UserID: typedUserID,
				}
				permission.SetEvaluate()
				require.NoError(t, db.DB.Create(&permission).Error)

				// Create participant and answer
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_ENDED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				answer := createTestAnswer(t, &participant, 1)

				return &participant, answer.QuestionID
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.GetAnswerEvaluationDataResponse) {
				assert.NotZero(t, resp.Id)
				assert.EqualValues(t, 5, resp.ScoreAwarded)
				assert.Equal(t, "Test Comment", resp.Comments)
			},
		},
		{
			name: "Fail - No evaluate permission",
			setup: func(t *testing.T) (*models.ExamParticipant, types.QuestionID) {
				exam := createTestExam(t, 2) // Created by different user

				// Create participant and answer
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_ENDED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				answer := createTestAnswer(t, &participant, 1)

				return &participant, answer.QuestionID
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Fail - Exam not ended",
			setup: func(t *testing.T) (*models.ExamParticipant, types.QuestionID) {
				exam := createTestExam(t, typedUserID) // Create as owner to get evaluate permission

				// Create participant with STARTED status
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				answer := createTestAnswer(t, &participant, 1)

				return &participant, answer.QuestionID
			},
			metadata: map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			},
			expectedCode: codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			participant, questionId := tt.setup(t)

			ctx := createContextWithMetadata(tt.metadata)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.ParticipantQuestionRequest{
					ParticipantId: int64(participant.ID),
					QuestionId:    int64(questionId),
				},
				client.GetAnswerEvaluationData,
				tt.validate,
			)
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
		validate     func(t *testing.T, answerId types.AnswerID)
	}{
		{
			name: "Success - Update all fields",
			setup: func(t *testing.T) *models.Answer {
				exam := createTestExam(t, typedUserID)
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				participant.ScoreAwarded = 5 // Initial score
				require.NoError(t, db.DB.Save(&participant).Error)

				// Create exam question with max score of 10
				createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 1,
					MaxScore:   10,
					Type:       constants.QUESTION_TYPE_MCQ,
				})

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
			validate: func(t *testing.T, answerId types.AnswerID) {
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
				exam := createTestExam(t, typedUserID)
				createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 10,
					MaxScore:   5,
				})

				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				// Create exam question with max score of 10
				createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 1,
					MaxScore:   10,
					Type:       constants.QUESTION_TYPE_MCQ,
				})

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
			validate: func(t *testing.T, answerId types.AnswerID) {
				var answer models.Answer
				require.NoError(t, db.DB.First(&answer, answerId).Error)
				assert.Equal(t, comments, answer.Comments.String)
				assert.EqualValues(t, 5, answer.ScoreAwarded) // Original score unchanged
			},
		},
		{
			name: "Fail - Answer not found",
			setup: func(t *testing.T) *models.Answer {
				exam := createTestExam(t, typedUserID)
				// Create exam question even for non-existent answer to maintain consistency
				createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 10,
					MaxScore:   5,
				})
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
				exam := createTestExam(t, typedUserID)
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)

				// Create exam question with max score of 10
				createTestExamQuestion(t, &exam, models.ExamQuestion{
					QuestionID: 1,
					CategoryID: 1,
					MaxScore:   10,
					Type:       constants.QUESTION_TYPE_MCQ,
				})

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
			tt.request.AnswerId = int64(answer.ID)

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
				exam := createTestExam(t, typedUserID)
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED},
				})
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
				exam := createTestExam(t, typedUserID)
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_ENDED},
				})
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
				exam := createTestExam(t, typedUserID)
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_STARTED},
				})
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

			ctx := createContextWithMetadata(map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			})
			resp, err := client.MarkParticipantAsEvaluated(ctx, &proto.ParticipantRequest{
				ParticipantId: int64(participant.ID),
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
