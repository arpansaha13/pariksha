package handlers

import (
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

func TestGetUserExams(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.ExamList)
	}{
		{
			name: "Success - Get user exams",
			setup: func(t *testing.T) {
				// Create multiple exams for the user
				createTestExam(t, int(userID))
				createTestExam(t, int(userID))
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.ExamList) {
				assert.Equal(t, 2, len(resp.Exams))
				for _, exam := range resp.Exams {
					assert.Equal(t, int32(userID), exam.CreatedBy)
					assert.Equal(t, "Test Exam", exam.Title)
					assert.Equal(t, "PRIVATE", exam.Type)
				}
			},
		},
		{
			name:         "Success - No exams",
			setup:        func(t *testing.T) {},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.ExamList) {
				assert.Equal(t, 0, len(resp.Exams))
			},
		},
		{
			name: "Success - Only get own exams",
			setup: func(t *testing.T) {
				createTestExam(t, int(userID))
				createTestExam(t, 2) // Different user's exam
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.ExamList) {
				assert.Equal(t, 1, len(resp.Exams))
				assert.Equal(t, int32(userID), resp.Exams[0].CreatedBy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.GetUserExams(ctx, &proto.Empty{})

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

func TestCreateExam(t *testing.T) {
	examTypeOpen := constants.EXAM_ACCESS_TYPE_LINK
	examTypeInvite := constants.EXAM_ACCESS_TYPE_INVITE
	examTypeUnknown := "UNKNOWN"

	tests := []struct {
		name         string
		request      *proto.CreateExamRequest
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.ExamResponse)
	}{
		{
			name: "Success - Create exam with no type specified",
			request: &proto.CreateExamRequest{
				Title:              "New Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 50,
				PaperId:            1,
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.ExamResponse) {
				assert.NotZero(t, resp.Id)
				assert.Equal(t, "New Exam", resp.Title)
				assert.Equal(t, int32(userID), resp.CreatedBy)
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_LINK, resp.Type) // Should be LINK by default
				assert.Equal(t, int32(50), resp.MaxCandidatesCount)
				assert.Equal(t, int32(1), resp.PaperId)

				// Verify in database
				var exam models.Exam
				require.NoError(t, db.DB.First(&exam, resp.Id).Error)
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_LINK, exam.Type)
			},
		},
		{
			name: "Success - Create exam with explicit LINK type",
			request: &proto.CreateExamRequest{
				Title:              "New Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               &examTypeOpen,
				PaperId:            1,
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.ExamResponse) {
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_LINK, resp.Type)

				var exam models.Exam
				require.NoError(t, db.DB.First(&exam, resp.Id).Error)
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_LINK, exam.Type)
			},
		},
		{
			name: "Success - Create exam with INVITE type",
			request: &proto.CreateExamRequest{
				Title:              "New Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               &examTypeInvite,
				PaperId:            1,
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.ExamResponse) {
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_INVITE, resp.Type)

				var exam models.Exam
				require.NoError(t, db.DB.First(&exam, resp.Id).Error)
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_INVITE, exam.Type)
			},
		},
		{
			name: "Invalid - End time before start time",
			request: &proto.CreateExamRequest{
				Title:              "Invalid Exam",
				StartsAt:           timestamppb.New(time.Now().Add(48 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(24 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               &examTypeInvite,
				PaperId:            1,
			},
			userID:       userID,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid - Start time in past",
			request: &proto.CreateExamRequest{
				Title:              "Invalid Exam",
				StartsAt:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(24 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               &examTypeInvite,
				PaperId:            1,
			},
			userID:       userID,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid - Zero max candidates",
			request: &proto.CreateExamRequest{
				Title:              "Invalid Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 0,
				Type:               &examTypeInvite,
				PaperId:            1,
			},
			userID:       userID,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid - Unknown exam type",
			request: &proto.CreateExamRequest{
				Title:              "Invalid Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               &examTypeUnknown,
				PaperId:            1,
			},
			userID:       userID,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.CreateExam(ctx, tt.request)

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

func TestUpdateExam(t *testing.T) {
	title := "Updated Exam"
	maxCandidates := int32(100)
	examType := constants.EXAM_ACCESS_TYPE_LINK

	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Exam
		request      *proto.UpdateExamRequest
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, exam *models.Exam)
	}{
		{
			name: "Success - Update all fields",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, int(userID))
				exam.StartsAt = time.Now().Add(24 * time.Hour)
				exam.EndsAt = time.Now().Add(48 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			request: &proto.UpdateExamRequest{
				Title:              &title,
				StartsAt:           timestamppb.New(time.Now().Add(48 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(72 * time.Hour)),
				MaxCandidatesCount: &maxCandidates,
				Type:               &examType,
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, exam *models.Exam) {
				var updated models.Exam
				require.NoError(t, db.DB.First(&updated, exam.ID).Error)

				assert.Equal(t, "Updated Exam", updated.Title)
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_LINK, updated.Type)
				assert.Equal(t, 100, updated.MaxCandidatesCount)
			},
		},
		{
			name: "Fail - Cannot update started exam timing",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, int(userID))
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			request: &proto.UpdateExamRequest{
				StartsAt: timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			userID:       userID,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name: "Fail - Cannot update ended exam",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, int(userID))
				exam.StartsAt = time.Now().Add(-2 * time.Hour)
				exam.EndsAt = time.Now().Add(-1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			request: &proto.UpdateExamRequest{
				Title: &title,
			},
			userID:       userID,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name: "Fail - End time before start time",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, int(userID))
				exam.StartsAt = time.Now().Add(24 * time.Hour)
				exam.EndsAt = time.Now().Add(48 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			request: &proto.UpdateExamRequest{
				EndsAt: timestamppb.New(time.Now().Add(12 * time.Hour)),
			},
			userID:       userID,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Fail - Exam not found",
			setup: func(t *testing.T) *models.Exam {
				return &models.Exam{ID: 9999}
			},
			request: &proto.UpdateExamRequest{
				Title: &title,
			},
			userID:       userID,
			expectedCode: codes.NotFound,
		},
		{
			name: "Fail - Not authorized",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2) // Created by different user
				return &exam
			},
			request: &proto.UpdateExamRequest{
				Title: &title,
			},
			userID:       userID,
			expectedCode: codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam := tt.setup(t)
			tt.request.ExamId = int32(exam.ID)

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.UpdateExam(ctx, tt.request)

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			if tt.validate != nil {
				tt.validate(t, exam)
			}
		})
	}
}

func TestEndExam(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (*models.Exam, *models.ExamParticipant)
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, examID, participantID int)
	}{
		{
			name: "Success - End exam",
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createTestExam(t, 2) // Created by different user
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				err := createTestExamParticipants(t, &exam, []struct {
					UserID int
					Status int
				}{
					{UserID: int(userID), Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &exam, &participant
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, examID, participantID int) {
				var exam models.Exam
				require.NoError(t, db.DB.First(&exam, examID).Error)
				counts, err := exam.GetParticipantCounts()
				require.NoError(t, err)

				assert.Equal(t, 0, counts.Started)
				assert.Equal(t, 1, counts.Ended)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.First(&participant, participantID).Error)
				assert.Equal(t, constants.PARTICIPANT_STATUS_ENDED, participant.Status)
				assert.True(t, participant.EndedAt.Valid)
			},
		},
		{
			name: "Fail - Participant not found",
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createTestExam(t, userID)
				return &exam, &models.ExamParticipant{ID: 9999}
			},
			userID:       userID,
			expectedCode: codes.NotFound,
		},
		{
			name: "Fail - Exam not found",
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				return &models.Exam{ID: 9999}, &models.ExamParticipant{}
			},
			userID:       userID,
			expectedCode: codes.NotFound,
		},
		{
			name: "Fail - Exam not started yet",
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createTestExam(t, 2)
				exam.StartsAt = time.Now().Add(1 * time.Hour) // Future start time
				require.NoError(t, db.DB.Save(&exam).Error)

				err := createTestExamParticipants(t, &exam, []struct {
					UserID int
					Status int
				}{
					{UserID: int(userID), Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				require.NoError(t, err)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &exam, &participant
			},
			userID:       userID,
			expectedCode: codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam, participant := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.EndExam(ctx, &proto.EndExamRequest{
				ExamId: int32(exam.ID),
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

func TestStartExam(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Exam
		userID       int32
		duration     int32
		expectedCode codes.Code
		validate     func(t *testing.T, exam *models.Exam)
	}{
		{
			name: "Success - Start private exam as invited participant",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2) // Created by different user
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				err := createTestExamParticipants(t, &exam, []struct {
					UserID int
					Status int
				}{
					{UserID: int(userID), Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				require.NoError(t, err)
				return &exam
			},
			userID:       userID,
			duration:     60,
			expectedCode: codes.OK,
			validate: func(t *testing.T, exam *models.Exam) {
				var updated models.Exam
				require.NoError(t, db.DB.First(&updated, exam.ID).Error)
				counts, err := updated.GetParticipantCounts()
				require.NoError(t, err)

				assert.Equal(t, 0, counts.Invited)
				assert.Equal(t, 1, counts.Started)
				assert.Equal(t, 0, counts.Ended)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)
				assert.Equal(t, constants.PARTICIPANT_STATUS_STARTED, participant.Status)
				assert.True(t, participant.StartedAt.Valid)
				assert.True(t, participant.ScheduledEndTime.Valid)
			},
		},
		{
			name: "Success - Start open exam as new participant",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2)
				exam.Type = constants.EXAM_ACCESS_TYPE_LINK
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			userID:       userID,
			duration:     60,
			expectedCode: codes.OK,
			validate: func(t *testing.T, exam *models.Exam) {
				var participant models.ExamParticipant
				err := db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error
				require.NoError(t, err)
				assert.Equal(t, constants.PARTICIPANT_STATUS_STARTED, participant.Status)
			},
		},
		{
			name: "Fail - Exam not started yet",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2)
				exam.StartsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				err := createTestExamParticipants(t, &exam, []struct {
					UserID int
					Status int
				}{
					{UserID: int(userID), Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				require.NoError(t, err)
				return &exam
			},
			userID:       userID,
			duration:     60,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name: "Fail - Exam already ended",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2)
				exam.StartsAt = time.Now().Add(-2 * time.Hour)
				exam.EndsAt = time.Now().Add(-1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				err := createTestExamParticipants(t, &exam, []struct {
					UserID int
					Status int
				}{
					{UserID: int(userID), Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				require.NoError(t, err)
				return &exam
			},
			userID:       userID,
			duration:     60,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name: "Fail - Already started exam",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2)
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				err := createTestExamParticipants(t, &exam, []struct {
					UserID int
					Status int
				}{
					{UserID: int(userID), Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				require.NoError(t, err)
				return &exam
			},
			userID:       userID,
			duration:     60,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name: "Fail - Participant not found in private exam",
			setup: func(t *testing.T) *models.Exam {
				exam := createTestExam(t, 2)
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			userID:       userID,
			duration:     60,
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.StartExam(ctx, &proto.StartExamRequest{
				ExamId:          int32(exam.ID),
				DurationMinutes: tt.duration,
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, exam)
			}
		})
	}
}
