package tests

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils/testrunner"
	"pariksha/exam/internal/config/db"
)

func TestGetExamResults(t *testing.T) {
	tests := []ExamSetupTestCase[*proto.ExamResultsResponse]{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get results with all questions answered",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2) // Created by different user

				// Create questions for the exam
				questions := []models.ExamQuestion{
					{ExamID: exam.ID, QuestionID: 1, CategoryID: 10, Order: 1, MaxScore: 10},
					{ExamID: exam.ID, QuestionID: 2, CategoryID: 10, Order: 2, MaxScore: 5},
				}
				require.NoError(t, db.DB.Create(&questions).Error)

				// Create participant
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: constants.PARTICIPANT_STATUS_EVALUATED},
				})
				require.NoError(t, err)

				// Get participant to create answers
				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)

				// Create answers for both questions
				createTestAnswer(t, &participant, 1)
				createTestAnswer(t, &participant, 2)

				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamResultsResponse) {
				require.Len(t, resp.Results, 2)

				// Verify results are ordered correctly
				assert.EqualValues(t, 1, resp.Results[0].QuestionId)
				assert.EqualValues(t, 2, resp.Results[1].QuestionId)

				for _, result := range resp.Results {
					// Verify answer content
					var answerData struct {
						Text string `json:"text"`
					}
					require.NoError(t, json.Unmarshal(result.Answer, &answerData))
					assert.Equal(t, "Test Answer", answerData.Text)

					// Verify scores and comments
					assert.EqualValues(t, 5, result.ScoreAwarded)
					assert.Equal(t, "Test Comment", result.Comments)

					// Verify category and max score
					assert.EqualValues(t, 10, result.CategoryId)
					if result.QuestionId == 1 {
						assert.EqualValues(t, 10, result.MaxScore)
					} else {
						assert.EqualValues(t, 5, result.MaxScore)
					}
				}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get results with some unanswered questions",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2)

				// Create questions
				questions := []models.ExamQuestion{
					{ExamID: exam.ID, QuestionID: 1, CategoryID: 10, Order: 1, MaxScore: 10},
					{ExamID: exam.ID, QuestionID: 2, CategoryID: 10, Order: 2, MaxScore: 5},
				}
				require.NoError(t, db.DB.Create(&questions).Error)

				// Create participant
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: constants.PARTICIPANT_STATUS_EVALUATED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)

				// Create answer for only one question
				createTestAnswer(t, &participant, 1)

				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamResultsResponse) {
				require.Len(t, resp.Results, 2)

				// First question should have answer
				assert.NotNil(t, resp.Results[0].Answer)
				assert.EqualValues(t, 5, resp.Results[0].ScoreAwarded)
				assert.NotEmpty(t, resp.Results[0].Comments)

				// Second question should have nil answer and zero values
				assert.Nil(t, resp.Results[1].Answer)
				assert.Zero(t, resp.Results[1].ScoreAwarded)
				assert.Empty(t, resp.Results[1].Comments)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Fail - User is not a participant",
				userID:       userID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2)
				questions := []models.ExamQuestion{
					{ExamID: exam.ID, QuestionID: 1, CategoryID: 10, Order: 1, MaxScore: 10},
				}
				require.NoError(t, db.DB.Create(&questions).Error)
				return &exam
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Exam with no questions",
				userID:       userID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: constants.PARTICIPANT_STATUS_EVALUATED},
				})
				require.NoError(t, err)
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamResultsResponse) {
				assert.Empty(t, resp.Results)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam := tt.setup(t)

			ctx := createContextWithMetadata(map[string]string{
				"user_id": strconv.FormatInt(userID, 10),
			})

			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.ExamRequest{ExamId: exam.ID},
				client.GetExamResults,
				tt.validate,
			)
		})
	}
}
