package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/generate"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/common/pkg/utils/testrunner"
	"pariksha/exam/internal/config/db"
)

func TestGetUserExams(t *testing.T) {
	tests := []getUserExamsTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Get user exams",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				// Create multiple exams for the user
				createDefaultTestExams(t, typedUserID, 2)
				return nil
			},
			validate: func(t *testing.T, resp *proto.ExamList) {
				assert.EqualValues(t, 2, len(resp.Exams))
				for i, exam := range resp.Exams {
					assert.EqualValues(t, userID, exam.CreatedBy)
					assert.Equal(t, fmt.Sprintf("Test Exam %d", i+1), exam.Title)
					assert.Equal(t, constants.EXAM_ACCESS_TYPE_LINK, exam.Type)
				}
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - No exams",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				return nil
			},
			validate: func(t *testing.T, resp *proto.ExamList) {
				assert.EqualValues(t, 0, len(resp.Exams))
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Only get own exams",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				createDefaultTestExam(t, typedUserID)
				createDefaultTestExam(t, 2) // Different user's exam
				return nil
			},
			validate: func(t *testing.T, resp *proto.ExamList) {
				assert.EqualValues(t, 1, len(resp.Exams))
				assert.Equal(t, userID, resp.Exams[0].CreatedBy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&emptypb.Empty{},
				client.GetUserExams,
				tt.validate,
			)
		})
	}
}

func TestCreateExam(t *testing.T) {
	tests := []createExamTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Create exam with no type specified",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			request: &proto.CreateExamRequest{
				Title:              "New Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 50,
				PaperHash:          paperHash,
				DurationMinutes:    120,
			},
			validate: func(t *testing.T, resp *proto.ExamResponse) {
				assert.NotEmpty(t, resp.ExamHash)
				assert.Equal(t, "New Exam", resp.Title)
				assert.Equal(t, userID, resp.CreatedBy)
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_LINK, resp.Type)
				assert.EqualValues(t, 50, resp.MaxCandidatesCount)
				assert.EqualValues(t, 120, resp.DurationMinutes)

				var exam models.Exam
				require.NoError(t, db.DB.Where("hash = ?", resp.ExamHash).Take(&exam).Error)
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_LINK, exam.Type)
				assert.EqualValues(t, 120, exam.DurationMinutes)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Create exam with explicit LINK type",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			request: &proto.CreateExamRequest{
				Title:              "New Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               ptr.String(constants.EXAM_ACCESS_TYPE_LINK),
				PaperHash:          paperHash,
				DurationMinutes:    90,
			},
			validate: func(t *testing.T, resp *proto.ExamResponse) {
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_LINK, resp.Type)
				assert.EqualValues(t, 90, resp.DurationMinutes)

				var exam models.Exam
				require.NoError(t, db.DB.Where("hash = ?", resp.ExamHash).Take(&exam).Error)
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_LINK, exam.Type)
				assert.EqualValues(t, 90, exam.DurationMinutes)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Create exam with INVITE type",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			request: &proto.CreateExamRequest{
				Title:              "New Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               ptr.String(constants.EXAM_ACCESS_TYPE_INVITE),
				PaperHash:          paperHash,
				DurationMinutes:    60,
			},
			validate: func(t *testing.T, resp *proto.ExamResponse) {
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_INVITE, resp.Type)
				assert.EqualValues(t, 60, resp.DurationMinutes)

				var exam models.Exam
				require.NoError(t, db.DB.Where("hash = ?", resp.ExamHash).Take(&exam).Error)
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_INVITE, exam.Type)
				assert.EqualValues(t, 60, exam.DurationMinutes)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Invalid - End time before start time",
				userID:       typedUserID,
				expectedCode: codes.InvalidArgument,
			},
			request: &proto.CreateExamRequest{
				Title:              "Invalid Exam",
				StartsAt:           timestamppb.New(time.Now().Add(48 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(24 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               ptr.String(constants.EXAM_ACCESS_TYPE_INVITE),
				PaperHash:          paperHash,
				DurationMinutes:    60,
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Invalid - Start time in past",
				userID:       typedUserID,
				expectedCode: codes.InvalidArgument,
			},
			request: &proto.CreateExamRequest{
				Title:              "Invalid Exam",
				StartsAt:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(24 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               ptr.String(constants.EXAM_ACCESS_TYPE_INVITE),
				PaperHash:          paperHash,
				DurationMinutes:    60,
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Invalid - Zero max candidates",
				userID:       typedUserID,
				expectedCode: codes.InvalidArgument,
			},
			request: &proto.CreateExamRequest{
				Title:              "Invalid Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 0,
				Type:               ptr.String(constants.EXAM_ACCESS_TYPE_INVITE),
				PaperHash:          paperHash,
				DurationMinutes:    60,
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Invalid - Unknown exam type",
				userID:       typedUserID,
				expectedCode: codes.InvalidArgument,
			},
			request: &proto.CreateExamRequest{
				Title:              "Invalid Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               ptr.String("UNKNOWN"),
				PaperHash:          paperHash,
				DurationMinutes:    60,
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Invalid - Zero duration minutes",
				userID:       typedUserID,
				expectedCode: codes.InvalidArgument,
			},
			request: &proto.CreateExamRequest{
				Title:              "Invalid Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               ptr.String(constants.EXAM_ACCESS_TYPE_LINK),
				PaperHash:          paperHash,
				DurationMinutes:    0,
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Invalid - Missing duration minutes",
				userID:       typedUserID,
				expectedCode: codes.InvalidArgument,
			},
			request: &proto.CreateExamRequest{
				Title:              "Invalid Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               ptr.String(constants.EXAM_ACCESS_TYPE_LINK),
				PaperHash:          paperHash,
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Invalid - Negative duration minutes",
				userID:       typedUserID,
				expectedCode: codes.InvalidArgument,
			},
			request: &proto.CreateExamRequest{
				Title:              "Invalid Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               ptr.String(constants.EXAM_ACCESS_TYPE_LINK),
				PaperHash:          paperHash,
				DurationMinutes:    -30,
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Invalid - Duration exceeds maximum allowed",
				userID:       typedUserID,
				expectedCode: codes.InvalidArgument,
			},
			request: &proto.CreateExamRequest{
				Title:              "Invalid Exam",
				StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
				EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
				MaxCandidatesCount: 50,
				Type:               ptr.String(constants.EXAM_ACCESS_TYPE_LINK),
				PaperHash:          paperHash,
				DurationMinutes:    int32(constants.MAX_EXAM_DURATION_MINUTES + 1),
			},
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
	examType := constants.EXAM_ACCESS_TYPE_LINK

	tests := []updateExamTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Update all fields",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: typedUserID,
					StartsAt:  time.Now().Add(24 * time.Hour),
					EndsAt:    time.Now().Add(48 * time.Hour),
				}})
				return &exams[0]
			},
			request: &proto.UpdateExamRequest{
				Title:    &title,
				StartsAt: timestamppb.New(time.Now().Add(48 * time.Hour)),
				EndsAt:   timestamppb.New(time.Now().Add(72 * time.Hour)),
				Type:     &examType,
			},
			validate: func(t *testing.T, exam *models.Exam) {
				var updated models.Exam
				require.NoError(t, db.DB.First(&updated, exam.ID).Error)

				assert.Equal(t, "Updated Exam", updated.Title)
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_LINK, updated.Type)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Cannot update StartsAt time after exam has started",
				userID:       typedUserID,
				expectedCode: codes.FailedPrecondition,
			},
			setup: func(t *testing.T) *models.Exam {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: typedUserID,
					StartsAt:  time.Now().Add(-1 * time.Hour),
					EndsAt:    time.Now().Add(1 * time.Hour),
				}})
				return &exams[0]
			},
			request: &proto.UpdateExamRequest{
				StartsAt: timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			validate: func(t *testing.T, exam *models.Exam) {
				var updated models.Exam
				require.NoError(t, db.DB.First(&updated, exam.ID).Error)

				// Ensure StartsAt is not updated
				assert.Equal(t, exam.StartsAt.Unix(), updated.StartsAt.Unix())
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - End time before start time",
				userID:       typedUserID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Exam {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: typedUserID,
					StartsAt:  time.Now().Add(24 * time.Hour),
					EndsAt:    time.Now().Add(48 * time.Hour),
				}})
				return &exams[0]
			},
			request: &proto.UpdateExamRequest{
				EndsAt: timestamppb.New(time.Now().Add(12 * time.Hour)),
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Exam not found",
				userID:       typedUserID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) *models.Exam {
				return &models.Exam{
					ID:   9999,
					Hash: generate.HMACHash(9999),
				}
			},
			request: &proto.UpdateExamRequest{
				Title: &title,
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Not authorized",
				userID:       typedUserID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2) // Created by different user
				return &exam
			},
			request: &proto.UpdateExamRequest{
				Title: &title,
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Update duration",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: typedUserID,
					StartsAt:  time.Now().Add(24 * time.Hour),
					EndsAt:    time.Now().Add(48 * time.Hour),
				}})
				return &exams[0]
			},
			request: &proto.UpdateExamRequest{
				ExamHash:        "", // Will be set in test
				DurationMinutes: ptr.Int32(180),
			},
			validate: func(t *testing.T, exam *models.Exam) {
				var updated models.Exam
				require.NoError(t, db.DB.First(&updated, exam.ID).Error)
				assert.EqualValues(t, 180, updated.DurationMinutes)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Update title after exam has ended",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: typedUserID,
					StartsAt:  time.Now().Add(-2 * time.Hour),
					EndsAt:    time.Now().Add(-1 * time.Hour),
				}})
				return &exams[0]
			},
			request: &proto.UpdateExamRequest{
				Title: &title,
			},
			validate: func(t *testing.T, exam *models.Exam) {
				var updated models.Exam
				require.NoError(t, db.DB.First(&updated, exam.ID).Error)
				assert.Equal(t, "Updated Exam", updated.Title)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Update non-title field after exam has ended",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy:       typedUserID,
					StartsAt:        time.Now().Add(-2 * time.Hour),
					EndsAt:          time.Now().Add(-1 * time.Hour),
					DurationMinutes: 60,
				}})
				return &exams[0]
			},
			request: &proto.UpdateExamRequest{
				Title:           &title,
				DurationMinutes: ptr.Int32(180),
			},
			validate: func(t *testing.T, exam *models.Exam) {
				var updated models.Exam
				require.NoError(t, db.DB.First(&updated, exam.ID).Error)

				// Verify that all fields other than title remain unchanged
				assert.Equal(t, title, updated.Title)
				assert.Equal(t, exam.DurationMinutes, updated.DurationMinutes)
				assert.Equal(t, exam.StartsAt.Unix(), updated.StartsAt.Unix())
				assert.Equal(t, exam.EndsAt.Unix(), updated.EndsAt.Unix())
				assert.Equal(t, exam.Type, updated.Type)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Update EndsAt after exam has started",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: typedUserID,
					StartsAt:  time.Now().Add(-1 * time.Hour), // Started 1 hour ago
					EndsAt:    time.Now().Add(1 * time.Hour),  // Originally ends in 1 hour
				}})
				return &exams[0]
			},
			request: &proto.UpdateExamRequest{
				EndsAt: timestamppb.New(time.Now().Add(2 * time.Hour)), // Extend to 2 hours from now
			},
			validate: func(t *testing.T, exam *models.Exam) {
				var updated models.Exam
				require.NoError(t, db.DB.First(&updated, exam.ID).Error)

				// Verify that EndsAt was updated but other fields remain unchanged
				assert.Equal(t, exam.Title, updated.Title)
				assert.Equal(t, exam.StartsAt.Unix(), updated.StartsAt.Unix())
				assert.Equal(t, exam.Type, updated.Type)
				assert.Greater(t, updated.EndsAt.Unix(), exam.EndsAt.Unix())
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Update duration exceeds maximum",
				userID:       typedUserID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) *models.Exam {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: typedUserID,
					StartsAt:  time.Now().Add(24 * time.Hour),
					EndsAt:    time.Now().Add(48 * time.Hour),
				}})
				return &exams[0]
			},
			request: &proto.UpdateExamRequest{
				DurationMinutes: ptr.Int32(int32(constants.MAX_EXAM_DURATION_MINUTES + 1)),
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Update duration after exam has started",
				userID:       typedUserID,
				expectedCode: codes.FailedPrecondition,
			},
			setup: func(t *testing.T) *models.Exam {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy:       typedUserID,
					StartsAt:        time.Now().Add(-1 * time.Hour), // Started 1 hour ago
					EndsAt:          time.Now().Add(1 * time.Hour),  // Ends in 1 hour
					DurationMinutes: 60,
				}})
				return &exams[0]
			},
			request: &proto.UpdateExamRequest{
				DurationMinutes: ptr.Int32(120), // Try to update to 2 hours
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam := tt.setup(t)
			tt.request.ExamHash = exam.Hash

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
	tests := []endExamTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - End exam",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createDefaultTestExam(t, 2)
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_STARTED},
				})

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &exam, &participant
			},
			validate: func(t *testing.T, examID types.ExamID, participantID types.ParticipantID) {
				var exam models.Exam
				require.NoError(t, db.DB.First(&exam, examID).Error)
				counts, err := exam.GetParticipantCounts()
				require.NoError(t, err)

				assert.EqualValues(t, 0, counts.Started)
				assert.EqualValues(t, 1, counts.Ended)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.First(&participant, participantID).Error)
				assert.Equal(t, constants.PARTICIPANT_STATUS_ENDED, participant.Status)
				assert.True(t, participant.EndedAt.Valid)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Participant not found",
				userID:       typedUserID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createDefaultTestExam(t, 2)
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam, &models.ExamParticipant{ID: 1}
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Exam not found",
				userID:       typedUserID,
				expectedCode: codes.NotFound,
			},
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				return &models.Exam{ID: 9999}, &models.ExamParticipant{}
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Exam not started yet",
				userID:       typedUserID,
				expectedCode: codes.FailedPrecondition,
			},
			setup: func(t *testing.T) (*models.Exam, *models.ExamParticipant) {
				exam := createDefaultTestExam(t, 2)
				exam.StartsAt = time.Now().Add(1 * time.Hour) // Future start time
				require.NoError(t, db.DB.Save(&exam).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ?", exam.ID).First(&participant).Error)
				return &exam, &participant
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam, participant := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.EndExam(ctx, &proto.EndExamRequest{
				ExamHash: exam.Hash,
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
	tests := []startExamTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Start INVITE exam as invited participant",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2) // Created by different user
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				return &exam
			},
			duration: 60,
			validate: func(t *testing.T, exam *models.Exam) {
				var updated models.Exam
				require.NoError(t, db.DB.First(&updated, exam.ID).Error)
				counts, err := updated.GetParticipantCounts()
				require.NoError(t, err)

				assert.EqualValues(t, 0, counts.Invited)
				assert.EqualValues(t, 1, counts.Started)
				assert.EqualValues(t, 0, counts.Ended)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)
				assert.Equal(t, constants.PARTICIPANT_STATUS_STARTED, participant.Status)
				assert.True(t, participant.StartedAt.Valid)
				assert.True(t, participant.ScheduledEndTime.Valid)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Start LINK exam as new participant",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)
				exam.Type = constants.EXAM_ACCESS_TYPE_LINK
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			duration: 60,
			validate: func(t *testing.T, exam *models.Exam) {
				var updated models.Exam
				require.NoError(t, db.DB.First(&updated, exam.ID).Error)
				counts, err := updated.GetParticipantCounts()
				require.NoError(t, err)

				assert.EqualValues(t, 0, counts.Invited)
				assert.EqualValues(t, 1, counts.Started)
				assert.EqualValues(t, 0, counts.Ended)

				var participant models.ExamParticipant
				err = db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error
				require.NoError(t, err)
				assert.Equal(t, constants.PARTICIPANT_STATUS_STARTED, participant.Status)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Exam not started yet",
				userID:       typedUserID,
				expectedCode: codes.FailedPrecondition,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)
				exam.StartsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				return &exam
			},
			duration: 60,
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Exam already ended",
				userID:       typedUserID,
				expectedCode: codes.FailedPrecondition,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)
				exam.StartsAt = time.Now().Add(-2 * time.Hour)
				exam.EndsAt = time.Now().Add(-1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				return &exam
			},
			duration: 60,
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Already started exam",
				userID:       typedUserID,
				expectedCode: codes.FailedPrecondition,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				return &exam
			},
			duration: 60,
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Participant not added in INVITE exam",
				userID:       typedUserID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			duration: 60,
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Start LINK exam with existing participant entry",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2) // Created by different user
				exam.Type = constants.EXAM_ACCESS_TYPE_LINK
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, db.DB.Save(&exam).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				return &exam
			},
			duration: 60,
			validate: func(t *testing.T, exam *models.Exam) {
				var updated models.Exam
				require.NoError(t, db.DB.First(&updated, exam.ID).Error)
				counts, err := updated.GetParticipantCounts()
				require.NoError(t, err)

				assert.EqualValues(t, 0, counts.Invited)
				assert.EqualValues(t, 1, counts.Started)
				assert.EqualValues(t, 0, counts.Ended)

				var participant models.ExamParticipant
				require.NoError(t, db.DB.Where("exam_id = ? AND user_id = ?", exam.ID, userID).First(&participant).Error)
				assert.Equal(t, constants.PARTICIPANT_STATUS_STARTED, participant.Status)
				assert.True(t, participant.StartedAt.Valid)
				assert.True(t, participant.ScheduledEndTime.Valid)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.StartExam(ctx, &proto.StartExamRequest{
				ExamHash: exam.Hash,
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

func TestGetExamQuestions(t *testing.T) {
	tests := []getExamQuestionsTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Get questions as registered participant",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2) // Created by different user

				createTestExamQuestions(t, &exam, []models.ExamQuestion{
					{
						MaxScore: 10,
					},
					{
						MaxScore: 5,
					},
				})

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamQuestionsResponse) {
				require.EqualValues(t, 2, len(resp.Questions))
				expectedQuestions := []struct {
					questionId int64
					categoryId int64
					order      int32
					maxScore   int32
				}{
					{questionId: 1, categoryId: 10, order: 1, maxScore: 10},
					{questionId: 2, categoryId: 10, order: 2, maxScore: 5},
				}
				for i, q := range resp.Questions {
					expected := expectedQuestions[i]
					assert.Equal(t, expected.questionId, getQuestionIdForHash(q.QuestionHash))
					assert.Equal(t, expected.categoryId, q.CategoryId)
					assert.Equal(t, expected.order, q.Order)
					assert.Equal(t, expected.maxScore, q.MaxScore)
				}
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Unregistered participant cannot access questions",
				userID:       typedUserID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)
				exam.Type = constants.EXAM_ACCESS_TYPE_LINK // Even with LINK type
				createTestExamQuestions(t, &exam, []models.ExamQuestion{
					{
						MaxScore: 5,
					},
				})
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
		{
			baseTestCase: baseTestCase{
				name:         "Success - Get questions as evaluator",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2) // Created by different user

				createTestExamQuestions(t, &exam, []models.ExamQuestion{
					{MaxScore: 10},
					{MaxScore: 5},
				})

				// Create evaluator permission
				permission := models.ExamPermission{
					ExamID: exam.ID,
					UserID: typedUserID,
				}
				permission.SetEvaluate()
				require.NoError(t, db.DB.Create(&permission).Error)

				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamQuestionsResponse) {
				require.EqualValues(t, 2, len(resp.Questions))
				expectedQuestions := []struct {
					questionId int64
					categoryId int64
					order      int32
					maxScore   int32
				}{
					{questionId: 1, categoryId: 10, order: 1, maxScore: 10},
					{questionId: 2, categoryId: 10, order: 2, maxScore: 5},
				}
				for i, q := range resp.Questions {
					expected := expectedQuestions[i]
					assert.Equal(t, expected.questionId, getQuestionIdForHash(q.QuestionHash))
					assert.Equal(t, expected.categoryId, q.CategoryId)
					assert.Equal(t, expected.order, q.Order)
					assert.Equal(t, expected.maxScore, q.MaxScore)
				}
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
				client.GetExamQuestions,
				tt.validate,
			)
		})
	}
}

func TestGetExamCategories(t *testing.T) {
	tests := []getExamCategoriesTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Get categories as registered participant",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2) // Created by different user
				categories := []models.ExamCategory{
					{ExamID: exam.ID, CategoryID: 1},
					{ExamID: exam.ID, CategoryID: 2},
				}
				require.NoError(t, db.DB.Create(&categories).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamCategoriesResponse) {
				require.EqualValues(t, 2, len(resp.Categories))
				expectedIDs := []int64{1, 2}
				for i, c := range resp.Categories {
					assert.Equal(t, expectedIDs[i], c.CategoryId)
				}
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Unregistered participant cannot access categories",
				userID:       typedUserID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)
				exam.Type = constants.EXAM_ACCESS_TYPE_LINK // Even with LINK type
				categories := []models.ExamCategory{
					{ExamID: exam.ID, CategoryID: 1},
					{ExamID: exam.ID, CategoryID: 2},
				}
				require.NoError(t, db.DB.Create(&categories).Error)
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
		{
			baseTestCase: baseTestCase{
				name:         "Success - Get categories as evaluator",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2) // Created by different user
				categories := []models.ExamCategory{
					{ExamID: exam.ID, CategoryID: 1},
					{ExamID: exam.ID, CategoryID: 2},
				}
				require.NoError(t, db.DB.Create(&categories).Error)

				// Create evaluator permission
				permission := models.ExamPermission{
					ExamID: exam.ID,
					UserID: typedUserID,
				}
				permission.SetEvaluate()
				require.NoError(t, db.DB.Create(&permission).Error)

				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamCategoriesResponse) {
				require.EqualValues(t, 2, len(resp.Categories))
				expectedIDs := []int64{1, 2}
				for i, c := range resp.Categories {
					assert.Equal(t, expectedIDs[i], c.CategoryId)
				}
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
				client.GetExamCategories,
				tt.validate,
			)
		})
	}
}

func TestGetExamPermission(t *testing.T) {
	tests := []getExamPermissionTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Owner permissions",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, typedUserID)
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamPermissionResponse) {
				assert.True(t, resp.CanRead)
				assert.True(t, resp.CanWrite)
				assert.False(t, resp.CanParticipate)
				assert.True(t, resp.CanEvaluate)
				assert.Nil(t, resp.ParticipantStatus)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Participant permissions",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2) // Created by different user
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamPermissionResponse) {
				assert.True(t, resp.CanRead)
				assert.False(t, resp.CanWrite)
				assert.True(t, resp.CanParticipate)
				assert.False(t, resp.CanEvaluate)
				require.NotNil(t, resp.ParticipantStatus)
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_INVITED, *resp.ParticipantStatus)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Non-owners and uninvited users get participate permission with INVITED status in LINK exam",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2) // Created by different user
				exam.Type = constants.EXAM_ACCESS_TYPE_LINK
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamPermissionResponse) {
				assert.True(t, resp.CanRead)
				assert.False(t, resp.CanWrite)
				assert.True(t, resp.CanParticipate)
				assert.False(t, resp.CanEvaluate)
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_INVITED, *resp.ParticipantStatus)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - LINK exam owner has normal permissions",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, typedUserID) // Created by test user
				exam.Type = constants.EXAM_ACCESS_TYPE_LINK
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamPermissionResponse) {
				assert.True(t, resp.CanRead)
				assert.True(t, resp.CanWrite)
				assert.False(t, resp.CanParticipate)
				assert.True(t, resp.CanEvaluate)
				assert.Nil(t, resp.ParticipantStatus)
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
		{
			baseTestCase: baseTestCase{
				name:         "Success - Check participant status in response",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2) // Created by different user
				// Add participants with different statuses
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{UserID: 3, Status: constants.PARTICIPANT_STATUS_INVITED},
					{Status: constants.PARTICIPANT_STATUS_STARTED}, // Our test user
					{UserID: 4, Status: constants.PARTICIPANT_STATUS_ENDED},
				})

				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamPermissionResponse) {
				assert.True(t, resp.CanRead)
				assert.False(t, resp.CanWrite)
				assert.True(t, resp.CanParticipate)
				assert.False(t, resp.CanEvaluate)
				require.NotNil(t, resp.ParticipantStatus)
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_STARTED, *resp.ParticipantStatus)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exam := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.GetExamPermission(ctx, &proto.ExamRequest{
				ExamHash: exam.Hash,
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

func TestGetExam(t *testing.T) {
	tests := []getExamTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Get exam as owner",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, typedUserID)
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamResponse) {
				assert.Equal(t, "Test Exam 1", resp.Title)
				assert.Equal(t, userID, resp.CreatedBy)
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_LINK, resp.Type)
				assert.EqualValues(t, 10, resp.MaxCandidatesCount)
				assert.EqualValues(t, 60, resp.DurationMinutes)

				// Validate participant counts
				assert.NotNil(t, resp.ParticipantCounts)
				assert.EqualValues(t, 0, resp.ParticipantCounts.Invited)
				assert.EqualValues(t, 0, resp.ParticipantCounts.Started)
				assert.EqualValues(t, 0, resp.ParticipantCounts.Ended)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Get exam as participant",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2) // Created by different user
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamResponse) {
				assert.EqualValues(t, 2, resp.CreatedBy) // Created by different user
				assert.NotNil(t, resp.ParticipantCounts)
				assert.EqualValues(t, 1, resp.ParticipantCounts.Invited)
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Get LINK exam as new participant",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)
				exam.Type = constants.EXAM_ACCESS_TYPE_LINK
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
			},
			validate: func(t *testing.T, resp *proto.ExamResponse) {
				assert.Equal(t, constants.EXAM_ACCESS_TYPE_LINK, resp.Type)
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
		{
			baseTestCase: baseTestCase{
				name:         "Fail - No permission to access INVITE exam",
				userID:       typedUserID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) *models.Exam {
				exam := createDefaultTestExam(t, 2)
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				require.NoError(t, db.DB.Save(&exam).Error)
				return &exam
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
				client.GetExam,
				tt.validate,
			)
		})
	}
}

func TestDeleteExams(t *testing.T) {
	tests := []deleteExamsTestCase{
		{
			baseTestCase: baseTestCase{
				name:         "Success - Delete multiple exams with hashes",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) []models.Exam {
				exams := createDefaultTestExams(t, typedUserID, 2)
				return exams
			},
			validate: func(t *testing.T, exams []models.Exam) {
				examIDs := make([]int64, len(exams))
				for i, e := range exams {
					examIDs[i] = int64(e.ID)
				}

				// Verify exams are deleted
				var examCount int64
				err := db.DB.Model(&models.Exam{}).Where("id IN ?", examIDs).Count(&examCount).Error
				require.NoError(t, err)
				assert.Equal(t, int64(0), examCount, "Exams should be deleted")
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Attempt to delete exams created by another user",
				userID:       typedUserID,
				expectedCode: codes.PermissionDenied,
			},
			setup: func(t *testing.T) []models.Exam {
				exam := createDefaultTestExam(t, 2) // Created by a different user

				// Create read-only permission for the test user
				permission := models.ExamPermission{
					ExamID: exam.ID,
					UserID: typedUserID,
				}
				permission.SetRead()
				require.NoError(t, db.DB.Create(&permission).Error)

				return []models.Exam{exam}
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Success - Non-existent exam returns success",
				userID:       typedUserID,
				expectedCode: codes.OK,
			},
			setup: func(t *testing.T) []models.Exam {
				return []models.Exam{
					{
						ID:   999, // Non-existent exam ID
						Hash: generate.HMACHash(999),
					},
				}
			},
		},
		{
			baseTestCase: baseTestCase{
				name:         "Fail - Empty exam IDs list",
				userID:       typedUserID,
				expectedCode: codes.InvalidArgument,
			},
			setup: func(t *testing.T) []models.Exam {
				return []models.Exam{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			exams := tt.setup(t)

			hashes := make([]string, len(exams))
			for i, e := range exams {
				hashes[i] = e.Hash
			}

			ctx := createContextWithUserID(tt.userID)
			testrunner.Runner(t, ctx, tt.expectedCode,
				&proto.DeleteExamsRequest{ExamHashes: hashes},
				client.DeleteExams,
				func(t *testing.T, _ *emptypb.Empty) {
					if tt.validate != nil {
						tt.validate(t, exams)
					}
				},
			)
		})
	}
}
