package handlers

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
	"pariksha/exam/internal/config/db"
)

func TestGetExamParticipants(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Exam
		userID       int64
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.ParticipantList)
	}{
		{
			name: "Success - Get exam participants",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, userID)
				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
				}{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_INVITED},
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)
				return &exam
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.ParticipantList) {
				assert.Equal(t, 2, len(resp.Participants))
				assert.Equal(t, constants.PARTICIPANT_STATUS_INVITED, int(resp.Participants[0].Status))
				assert.Equal(t, constants.PARTICIPANT_STATUS_STARTED, int(resp.Participants[1].Status))
			},
		},
		{
			name: "Success - No participants",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, userID)
				return &exam
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.ParticipantList) {
				assert.Equal(t, 0, len(resp.Participants))
			},
		},
		{
			name: "Exam not found",
			setup: func(t *testing.T) *models.Exam {
				return &models.Exam{ID: 9999}
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.ParticipantList) {
				assert.Equal(t, 0, len(resp.Participants), "Should return empty array for non-existent exam")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.GetExamParticipants(ctx, &proto.ExamRequest{
				ExamId: exam.ID,
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

func TestAddExamParticipant(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Exam
		request      *proto.AddParticipantRequest
		userID       int64
		expectedCode codes.Code
		validate     func(t *testing.T, examID int64, resp *proto.ParticipantResponse)
	}{
		{
			name: "Success - Add participant",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, userID)
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			request: &proto.AddParticipantRequest{
				UserId: 2, // Use hardcoded participant ID
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, examID int64, resp *proto.ParticipantResponse) {
				assert.Equal(t, int64(2), resp.UserId)
				assert.Equal(t, int32(constants.PARTICIPANT_STATUS_INVITED), resp.Status)

				// Check if exam participant counts were updated
				var exam models.Exam
				require.NoError(t, db.DB.First(&exam, examID).Error)
				counts, err := exam.GetParticipantCounts()
				require.NoError(t, err)
				assert.Equal(t, 1, counts.Invited, "Invited count should be updated")
				assert.Equal(t, 0, counts.Started)
				assert.Equal(t, 0, counts.Ended)
				assert.Equal(t, 0, counts.Unattended)
			},
		},
		{
			name: "Max candidates limit reached",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, userID)
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				exam.MaxCandidatesCount = 1
				require.NoError(t, db.DB.Save(&exam).Error)

				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
				}{
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				require.NoError(t, err)
				return &exam
			},
			request: &proto.AddParticipantRequest{
				UserId: 2,
			},
			userID:       userID,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name: "Cannot add participant to exam with access-type LINK",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, userID)
				exam.Type = constants.EXAM_ACCESS_TYPE_LINK
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			request: &proto.AddParticipantRequest{
				UserId: 2,
			},
			userID:       userID,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Exam not found",
			setup: func(t *testing.T) *models.Exam {
				return &models.Exam{ID: 9999}
			},
			request: &proto.AddParticipantRequest{
				UserId: 2,
			},
			userID:       userID,
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam := tt.setup(t)
			tt.request.ExamId = exam.ID

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
	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.Exam, *models.ExamParticipant)
		userID       int64
		expectedCode codes.Code
		validate     func(t *testing.T, examID, participantID int64)
	}{
		{
			name: "Success - Remove participant",
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createTestExam(t, userID)
				exam.StartsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
				}{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &exam, &participant
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, examID, participantID int64) {
				var count int64
				db.DB.Model(&models.ExamParticipant{}).Where("id = ?", participantID).Count(&count)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			name: "Cannot remove after exam started",
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createTestExam(t, userID)
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				err := createTestExamParticipants(t, &exam, []struct {
					UserID int64
					Status int
				}{
					{UserID: 2, Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &exam, &participant
			},
			userID:       userID,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name: "Non-existent participant",
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createTestExam(t, userID)
				exam.StartsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam, &models.ExamParticipant{ID: 9999}
			},
			userID:       userID,
			expectedCode: codes.NotFound,
		},
		{
			name: "Non-existent exam",
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				participant := &models.ExamParticipant{ID: 1}
				return &models.Exam{ID: 9999}, participant
			},
			userID:       userID,
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam, participant := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.RemoveExamParticipant(ctx, &proto.RemoveParticipantRequest{
				ExamId:        exam.ID,
				ParticipantId: participant.ID,
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
