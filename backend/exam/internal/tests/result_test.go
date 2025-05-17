package tests

import (
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
				exam := createTestExam(t, 2)

				// Create participant
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: userID, Status: constants.PARTICIPANT_STATUS_EVALUATED},
				})
				require.NoError(t, err)

				// Get participant to create answers
				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)

				// Create answers
				createTestAnswer(t, &participant, 1)
				createTestAnswer(t, &participant, 2)

				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamResultsResponse) {
				require.Len(t, resp.Results, 2)

				for _, result := range resp.Results {
					assert.NotZero(t, result.AnswerId)
					assert.EqualValues(t, 5, result.ScoreAwarded)
					assert.Equal(t, "Test Comment", result.Comments)
				}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Exam with no answers",
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
		{
			BaseTestCase: BaseTestCase{
				name:         "Fail - User is not a participant",
				userID:       userID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2)
				return &exam
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
