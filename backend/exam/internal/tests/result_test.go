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
	tests := []getExamResultsTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Get results with all questions answered",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)

				// Create participant
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_EVALUATED},
				})

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
				}
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Exam with no answers",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{
						Status: constants.PARTICIPANT_STATUS_EVALUATED,
					},
				})
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamResultsResponse) {
				assert.Empty(t, resp.Results)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - User is not a participant",
				userID:       typedUserID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)
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
				&proto.ExamRequest{ExamHash: exam.Hash},
				client.GetExamResults,
				tt.validate,
			)
		})
	}
}
