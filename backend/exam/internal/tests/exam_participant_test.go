package tests

import (
	"testing"
	"time"

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

func TestGetExamParticipants(t *testing.T) {
	tests := []ExamSetupTestCase[*proto.ParticipantList]{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get exam participants",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, typedUserID)
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_INVITED},
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ParticipantList) {
				assert.EqualValues(t, 2, len(resp.Participants))
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_INVITED, int16(resp.Participants[0].Status))
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_STARTED, int16(resp.Participants[1].Status))
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - No participants",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, typedUserID)
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ParticipantList) {
				assert.EqualValues(t, 0, len(resp.Participants))
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Exam not found",
				userID:       typedUserID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) *models.Exam {
				return &models.Exam{ID: 9999}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.ExamRequest{ExamHash: exam.ExamHash.Hash},
				client.GetExamParticipants,
				tt.validate,
			)
		})
	}
}

func TestAddExamParticipant(t *testing.T) {
	tests := []ExamParticipantRequestTestCase[*proto.ParticipantResponse, *proto.AddParticipantRequest]{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Add participant",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, typedUserID)
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			request: &proto.AddParticipantRequest{
				UserId: 2, // Use hardcoded participant ID
			},
			validate: func(t *testing.T, examID types.ExamID, resp *proto.ParticipantResponse) {
				assert.EqualValues(t, 2, resp.UserId)
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_INVITED, resp.Status)

				// Check if exam participant counts were updated
				var exam models.Exam
				require.NoError(t, db.DB.First(&exam, examID).Error)
				counts, err := exam.GetParticipantCounts()
				require.NoError(t, err)
				assert.EqualValues(t, 1, counts.Invited, "Invited count should be updated")
				assert.EqualValues(t, 0, counts.Started)
				assert.EqualValues(t, 0, counts.Ended)
				assert.EqualValues(t, 0, counts.Unattended)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Max candidates limit reached",
				userID:       typedUserID,
				expectedCode: codes.FailedPrecondition,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, typedUserID)
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				exam.MaxCandidatesCount = 1
				require.NoError(t, db.DB.Save(&exam).Error)

				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				require.NoError(t, err)
				return &exam
			},
			request: &proto.AddParticipantRequest{
				UserId: 2,
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Cannot add participant to exam with access-type LINK",
				userID:       typedUserID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, typedUserID)
				exam.Type = constants.EXAM_ACCESS_TYPE_LINK
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			request: &proto.AddParticipantRequest{
				UserId: 2,
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Exam not found",
				userID:       typedUserID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) *models.Exam {
				return &models.Exam{ID: 9999}
			},
			request: &proto.AddParticipantRequest{
				UserId: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam := tt.setup(t)
			tt.request.ExamHash = exam.ExamHash.Hash

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.AddExamParticipant(ctx, tt.request)

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, exam.ID, resp)
		})
	}
}

func TestRemoveExamParticipant(t *testing.T) {
	tests := []ExamParticipantPairTestCase{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Remove participant",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createTestExam(t, typedUserID)
				exam.StartsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &exam, &participant
			},
			validate: func(t *testing.T, examID types.ExamID, participantID types.ParticipantID) {
				var count int64
				db.DB.Model(&models.ExamParticipant{}).Where("id = ?", participantID).Count(&count)
				assert.EqualValues(t, 0, count)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Cannot remove after exam started",
				userID:       typedUserID,
				expectedCode: codes.FailedPrecondition,
			},
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createTestExam(t, typedUserID)
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &exam, &participant
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Non-existent participant",
				userID:       typedUserID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createTestExam(t, typedUserID)
				exam.StartsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam, &models.ExamParticipant{ID: 9999}
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Non-existent exam",
				userID:       typedUserID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				participant := &models.ExamParticipant{ID: 1}
				return &models.Exam{ID: 9999}, participant
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam, participant := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.RemoveExamParticipant(ctx, &proto.RemoveParticipantRequest{
				ExamHash:      exam.ExamHash.Hash,
				ParticipantId: int64(participant.ID),
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, exam.ID, participant.ID)
			}
		})
	}
}

func TestGetExamParticipant(t *testing.T) {
	tests := []ExamSetupTestCase[*proto.GetExamParticipantResponse]{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get participant with timing data",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2) // Created by different user
				startTime := time.Now().Add(-1 * time.Hour)
				scheduledEndTime := startTime.Add(2 * time.Hour)

				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: typedUserID, Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)

				participant.StartedAt.Time = startTime
				participant.StartedAt.Valid = true
				participant.ScheduledEndTime.Time = scheduledEndTime
				participant.ScheduledEndTime.Valid = true
				require.NoError(t, db.DB.Save(&participant).Error)

				return &exam
			},
			validate: func(t *testing.T, resp *proto.GetExamParticipantResponse) {
				assert.NotZero(t, resp.ParticipantId)
				assert.NotNil(t, resp.StartedAt)
				assert.NotNil(t, resp.ScheduledEndTime)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get participant without timing data",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2)
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: typedUserID, Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				require.NoError(t, err)
				return &exam
			},
			validate: func(t *testing.T, resp *proto.GetExamParticipantResponse) {
				assert.NotZero(t, resp.ParticipantId)
				assert.Nil(t, resp.StartedAt)
				assert.Nil(t, resp.ScheduledEndTime)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Fail - User is not a participant",
				userID:       typedUserID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2)
				return &exam
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Fail - Exam not found",
				userID:       typedUserID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) *models.Exam {
				return &models.Exam{ID: 9999}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.GetExamParticipantRequest{ExamHash: exam.ExamHash.Hash},
				client.GetExamParticipant,
				tt.validate,
			)
		})
	}
}

func TestGetParticipantById(t *testing.T) {
	tests := []ParticipantTestCase[*proto.ParticipantResponse]{
		{
			BaseTestCase: BaseTestCase{
				name:         "Success - Get participant details",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createTestExam(t, typedUserID)
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				require.NoError(t, db.DB.Save(&exam).Error)

				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &participant
			},
			validate: func(t *testing.T, resp *proto.ParticipantResponse) {
				require.NotNil(t, resp)
				assert.NotZero(t, resp.ParticipantId)
				assert.Equal(t, int32(constants.PARTICIPANT_STATUS_STARTED), resp.Status)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Fail - Participant not found",
				userID:       typedUserID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) *models.ExamParticipant {
				return &models.ExamParticipant{ID: 9999}
			},
			validate: func(t *testing.T, resp *proto.ParticipantResponse) {
				assert.Nil(t, resp)
			},
		},
		{
			BaseTestCase: BaseTestCase{
				name:         "Fail - No evaluate permission",
				userID:       2, // Different user without evaluate permission
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createTestExam(t, typedUserID)
				err := createTestExamParticipants(t, &exam, []TestParticipantData{
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &participant
			},
			validate: func(t *testing.T, resp *proto.ParticipantResponse) {
				assert.Nil(t, resp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			participant := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.ParticipantRequest{ParticipantId: int64(participant.ID)},
				client.GetParticipantById,
				tt.validate,
			)
		})
	}
}
