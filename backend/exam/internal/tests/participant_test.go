package tests

import (
	"database/sql"
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
	tests := []getExamParticipantsTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Get exam participants",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, typedUserID)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{
						UserID: 2,
						Status: constants.PARTICIPANT_STATUS_INVITED,
					},
					{
						UserID: 3,
						Status: constants.PARTICIPANT_STATUS_STARTED,
					},
				})
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ParticipantList) {
				assert.EqualValues(t, 2, len(resp.Participants))
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_INVITED, int16(resp.Participants[0].Status))
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_STARTED, int16(resp.Participants[1].Status))
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - No participants",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, typedUserID)
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ParticipantList) {
				assert.EqualValues(t, 0, len(resp.Participants))
			},
		},
		{
			baseTestCase: baseTestCase{
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
				&proto.ExamRequest{ExamHash: exam.Hash},
				client.GetExamParticipants,
				tt.validate,
			)
		})
	}
}

func TestGetExamParticipant(t *testing.T) {
	tests := []getExamParticipantTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Get participant with timing data",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)
				startTime := time.Now().Add(-1 * time.Hour)
				scheduledEndTime := startTime.Add(2 * time.Hour)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					Status: constants.PARTICIPANT_STATUS_STARTED,
					StartedAt: sql.NullTime{
						Time:  startTime,
						Valid: true,
					},
					ScheduledEndTime: sql.NullTime{
						Time:  scheduledEndTime,
						Valid: true,
					},
				}})
				return &exam
			},
			validate: func(t *testing.T, resp *proto.GetExamParticipantResponse) {
				assert.NotZero(t, resp.ParticipantId)
				assert.NotNil(t, resp.StartedAt)
				assert.NotNil(t, resp.ScheduledEndTime)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Get participant without timing data",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{
						Status: constants.PARTICIPANT_STATUS_INVITED,
					},
				})
				return &exam
			},
			validate: func(t *testing.T, resp *proto.GetExamParticipantResponse) {
				assert.NotZero(t, resp.ParticipantId)
				assert.Nil(t, resp.StartedAt)
				assert.Nil(t, resp.ScheduledEndTime)
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
		{
			baseTestCase: baseTestCase{
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
				&proto.GetExamParticipantRequest{ExamHash: exam.Hash},
				client.GetExamParticipant,
				tt.validate,
			)
		})
	}
}

func TestAddExamParticipant(t *testing.T) {
	tests := []addExamParticipantTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Add participant",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, typedUserID)
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
			baseTestCase: baseTestCase{
				name:         "Max candidates limit reached",
				userID:       typedUserID,
				expectedCode: codes.FailedPrecondition,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, typedUserID)
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				exam.MaxCandidatesCount = 1
				require.NoError(t, db.DB.Save(&exam).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					UserID: 3,
					Status: constants.PARTICIPANT_STATUS_INVITED,
				}})
				return &exam
			},
			request: &proto.AddParticipantRequest{
				UserId: 2,
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Cannot add participant to exam with access-type LINK",
				userID:       typedUserID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, typedUserID)
				exam.Type = constants.EXAM_ACCESS_TYPE_LINK
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			request: &proto.AddParticipantRequest{
				UserId: 2,
			},
		},
		{
			baseTestCase: baseTestCase{
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
			tt.request.ExamHash = exam.Hash

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
	tests := []removeExamParticipantTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Remove participant",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createDefaultTestExam(t, typedUserID)
				exam.StartsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					UserID: 2,
					Status: constants.PARTICIPANT_STATUS_INVITED,
				}})
				return &exam, &participants[0]
			},
			validate: func(t *testing.T, examID types.ExamID, participantID types.ParticipantID) {
				var count int64
				db.DB.Model(&models.ExamParticipant{}).Where("id = ?", participantID).Count(&count)
				assert.EqualValues(t, 0, count)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Cannot remove after exam started",
				userID:       typedUserID,
				expectedCode: codes.FailedPrecondition,
			},
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createDefaultTestExam(t, typedUserID)
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					UserID: 2,
					Status: constants.PARTICIPANT_STATUS_INVITED,
				}})
				return &exam, &participants[0]
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Non-existent participant",
				userID:       typedUserID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createDefaultTestExam(t, typedUserID)
				exam.StartsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam, &models.ExamParticipant{ID: 9999}
			},
		},
		{
			baseTestCase: baseTestCase{
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
				ExamHash:      exam.Hash,
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

func TestGetParticipantById(t *testing.T) {
	tests := []getParticipantByIdTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Get participant details",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createDefaultTestExam(t, typedUserID)
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				require.NoError(t, db.DB.Save(&exam).Error)

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					UserID: 2,
					Status: constants.PARTICIPANT_STATUS_STARTED,
				}})
				return &participants[0]
			},
			validate: func(t *testing.T, resp *proto.ParticipantResponse) {
				require.NotNil(t, resp)
				assert.NotZero(t, resp.ParticipantId)
				assert.Equal(t, int32(constants.PARTICIPANT_STATUS_STARTED), resp.Status)
			},
		},
		{
			baseTestCase: baseTestCase{
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
			baseTestCase: baseTestCase{
				name:         "Fail - No evaluate permission",
				userID:       2, // Different user without evaluate permission
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.ExamParticipant {
				exam := createDefaultTestExam(t, typedUserID)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{
						UserID: 3,
						Status: constants.PARTICIPANT_STATUS_STARTED,
					},
				})

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
