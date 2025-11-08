package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/generate"
	"pariksha/common/pkg/utils/ptr"
)

func TestGetUserExams(t *testing.T) {
	type SetupReturn struct {
		Exams []models.Exam
	}

	testCases := []test.TestCase[*emptypb.Empty, *proto.ExamList, *SetupReturn]{
		{
			Name:     "Success - Get user exams",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exams := createDefaultTestExams(t, defaultUserID, 2)
				return &SetupReturn{Exams: exams}
			},
			GetRequest: func(setupData *SetupReturn) *emptypb.Empty {
				return &emptypb.Empty{}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamList, setupData *SetupReturn) {
				assert.EqualValues(t, 2, len(resp.Exams))
				for i, exam := range resp.Exams {
					assert.EqualValues(t, defaultUserID, exam.CreatedBy)
					assert.Equal(t, fmt.Sprintf("Test Exam %d", i+1), exam.Title)
					assert.EqualValues(t, proto.ExamType_LINK, exam.Type)
				}
			},
		},
		{
			Name:     "Success - No exams",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{}
			},
			GetRequest: func(setupData *SetupReturn) *emptypb.Empty {
				return &emptypb.Empty{}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamList, setupData *SetupReturn) {
				assert.EqualValues(t, 0, len(resp.Exams))
			},
		},
		{
			Name:     "Success - Only get own exams",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exams := []models.Exam{
					createDefaultTestExam(t, defaultUserID),
					createDefaultTestExam(t, 2), // Different user's exam
				}
				return &SetupReturn{Exams: exams}
			},
			GetRequest: func(setupData *SetupReturn) *emptypb.Empty {
				return &emptypb.Empty{}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamList, setupData *SetupReturn) {
				assert.EqualValues(t, 1, len(resp.Exams))
				assert.EqualValues(t, defaultUserID, resp.Exams[0].CreatedBy)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetUserExams)
		})
	}
}

func TestCreateExam(t *testing.T) {
	type SetupReturn struct{}

	validStartTime := time.Now().Add(24 * time.Hour)
	validEndTime := time.Now().Add(48 * time.Hour)

	testCases := []test.TestCase[*proto.CreateExamRequest, *proto.ExamResponse, *SetupReturn]{
		{
			Name:     "Success - Create exam with no type specified",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.CreateExamRequest {
				return &proto.CreateExamRequest{
					Title:              "New Exam",
					StartsAt:           timestamppb.New(validStartTime),
					EndsAt:             timestamppb.New(validEndTime),
					MaxCandidatesCount: 50,
					PaperHash:          paperHash,
					DurationMinutes:    120,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamResponse, setupData *SetupReturn) {
				assert.NotEmpty(t, resp.ExamHash)
				assert.Equal(t, "New Exam", resp.Title)
				assert.EqualValues(t, defaultUserID, resp.CreatedBy)
				assert.EqualValues(t, proto.ExamType_LINK, resp.Type)
				assert.EqualValues(t, 50, resp.MaxCandidatesCount)
				assert.EqualValues(t, 120, resp.DurationMinutes)

				var exam models.Exam
				require.NoError(t, dbInst.Where("hash = ?", resp.ExamHash).Take(&exam).Error)
				assert.EqualValues(t, proto.ExamType_LINK, exam.Type)
				assert.EqualValues(t, 120, exam.DurationMinutes)
			},
		},
		{
			Name:     "Success - Create exam with INVITE type",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.CreateExamRequest {
				return &proto.CreateExamRequest{
					Title:              "New Exam",
					StartsAt:           timestamppb.New(validStartTime),
					EndsAt:             timestamppb.New(validEndTime),
					MaxCandidatesCount: 50,
					Type:               proto.ExamType_INVITE.Enum(),
					PaperHash:          paperHash,
					DurationMinutes:    60,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamResponse, setupData *SetupReturn) {
				assert.EqualValues(t, proto.ExamType_INVITE, resp.Type)
				assert.EqualValues(t, 60, resp.DurationMinutes)

				var exam models.Exam
				require.NoError(t, dbInst.Where("hash = ?", resp.ExamHash).Take(&exam).Error)
				assert.EqualValues(t, proto.ExamType_INVITE, exam.Type)
			},
		},
		{
			Name:     "Invalid - End time before start time",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.CreateExamRequest {
				return &proto.CreateExamRequest{
					Title:              "Invalid Exam",
					StartsAt:           timestamppb.New(validEndTime),   // Swapped times
					EndsAt:             timestamppb.New(validStartTime), // Swapped times
					MaxCandidatesCount: 50,
					Type:               proto.ExamType_INVITE.Enum(),
					PaperHash:          paperHash,
					DurationMinutes:    60,
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Invalid - Start time in past",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.CreateExamRequest {
				return &proto.CreateExamRequest{
					Title:              "Invalid Exam",
					StartsAt:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
					EndsAt:             timestamppb.New(validEndTime),
					MaxCandidatesCount: 50,
					Type:               proto.ExamType_INVITE.Enum(),
					PaperHash:          paperHash,
					DurationMinutes:    60,
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Invalid - Zero max candidates",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.CreateExamRequest {
				return &proto.CreateExamRequest{
					Title:              "Invalid Exam",
					StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
					EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
					MaxCandidatesCount: 0,
					Type:               proto.ExamType_INVITE.Enum(),
					PaperHash:          paperHash,
					DurationMinutes:    60,
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Invalid - Unknown exam type",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.CreateExamRequest {
				return &proto.CreateExamRequest{
					Title:              "Invalid Exam",
					StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
					EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
					MaxCandidatesCount: 50,
					Type:               proto.ExamType_UNKNOWN_EXAM_TYPE.Enum(),
					PaperHash:          paperHash,
					DurationMinutes:    60,
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Invalid - Zero duration minutes",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.CreateExamRequest {
				return &proto.CreateExamRequest{
					Title:              "Invalid Exam",
					StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
					EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
					MaxCandidatesCount: 50,
					Type:               proto.ExamType_LINK.Enum(),
					PaperHash:          paperHash,
					DurationMinutes:    0,
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Invalid - Missing duration minutes",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.CreateExamRequest {
				return &proto.CreateExamRequest{
					Title:              "Invalid Exam",
					StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
					EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
					MaxCandidatesCount: 50,
					Type:               proto.ExamType_LINK.Enum(),
					PaperHash:          paperHash,
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Invalid - Negative duration minutes",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.CreateExamRequest {
				return &proto.CreateExamRequest{
					Title:              "Invalid Exam",
					StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
					EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
					MaxCandidatesCount: 50,
					Type:               proto.ExamType_LINK.Enum(),
					PaperHash:          paperHash,
					DurationMinutes:    -30,
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Invalid - Duration exceeds maximum allowed",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.CreateExamRequest {
				return &proto.CreateExamRequest{
					Title:              "Invalid Exam",
					StartsAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
					EndsAt:             timestamppb.New(time.Now().Add(48 * time.Hour)),
					MaxCandidatesCount: 50,
					Type:               proto.ExamType_LINK.Enum(),
					PaperHash:          paperHash,
					DurationMinutes:    int32(constants.MAX_EXAM_DURATION_MINUTES + 1),
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.CreateExam)
		})
	}
}

func TestUpdateExam(t *testing.T) {
	type SetupReturn struct {
		Exam models.Exam
	}

	testCases := []test.TestCase[*proto.UpdateExamRequest, *proto.ExamResponse, *SetupReturn]{
		{
			Name:     "Success - Update all fields",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: defaultUserID,
					StartsAt:  time.Now().Add(24 * time.Hour),
					EndsAt:    time.Now().Add(48 * time.Hour),
				}})
				return &SetupReturn{Exam: exams[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateExamRequest {
				return &proto.UpdateExamRequest{
					ExamHash: setupData.Exam.Hash,
					Title:    ptr.String("Updated Exam"),
					StartsAt: timestamppb.New(time.Now().Add(48 * time.Hour)),
					EndsAt:   timestamppb.New(time.Now().Add(72 * time.Hour)),
					Type:     proto.ExamType_LINK.Enum(),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamResponse, setupData *SetupReturn) {
				var exam models.Exam
				require.NoError(t, dbInst.First(&exam, setupData.Exam.ID).Error)

				assert.Equal(t, "Updated Exam", exam.Title)
				assert.Equal(t, proto.ExamType_LINK, exam.Type)
			},
		},
		{
			Name:     "Fail - Cannot update StartsAt time after exam has started",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: defaultUserID,
					StartsAt:  time.Now().Add(-1 * time.Hour),
					EndsAt:    time.Now().Add(1 * time.Hour),
				}})
				return &SetupReturn{Exam: exams[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateExamRequest {
				return &proto.UpdateExamRequest{
					ExamHash: setupData.Exam.Hash,
					StartsAt: timestamppb.New(time.Now().Add(1 * time.Hour)),
				}
			},
			Validate: func(t *testing.T, resp *proto.ExamResponse, setupData *SetupReturn) {
				var exam models.Exam
				require.NoError(t, dbInst.First(&exam, setupData.Exam.ID).Error)

				// Ensure StartsAt is not updated
				assert.Equal(t, exam.StartsAt.Unix(), exam.StartsAt.Unix())
			},
			ExpectedCode: codes.FailedPrecondition,
		},
		{
			Name:     "Fail - End time before start time",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: defaultUserID,
					StartsAt:  time.Now().Add(24 * time.Hour),
					EndsAt:    time.Now().Add(48 * time.Hour),
				}})
				return &SetupReturn{Exam: exams[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateExamRequest {
				return &proto.UpdateExamRequest{
					ExamHash: setupData.Exam.Hash,
					EndsAt:   timestamppb.New(time.Now().Add(12 * time.Hour)),
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Fail - Exam not found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					Exam: models.Exam{
						ID:   999,
						Hash: "non_existent_hash",
					},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateExamRequest {
				return &proto.UpdateExamRequest{
					ExamHash: setupData.Exam.Hash,
					Title:    ptr.String("Updated Exam"),
				}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name:     "Fail - Not authorized",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2) // Created by different user
				return &SetupReturn{
					Exam: exam,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateExamRequest {
				return &proto.UpdateExamRequest{
					ExamHash: setupData.Exam.Hash,
					Title:    ptr.String("Updated Exam"),
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Success - Update duration",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: defaultUserID,
					StartsAt:  time.Now().Add(24 * time.Hour),
					EndsAt:    time.Now().Add(48 * time.Hour),
				}})
				return &SetupReturn{Exam: exams[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateExamRequest {
				return &proto.UpdateExamRequest{
					ExamHash:        setupData.Exam.Hash,
					DurationMinutes: ptr.Int32(180),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamResponse, setupData *SetupReturn) {
				var exam models.Exam
				require.NoError(t, dbInst.First(&exam, setupData.Exam.ID).Error)
				assert.EqualValues(t, 180, exam.DurationMinutes)
			},
		},

		{
			Name:     "Success - Update title after exam has ended",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: defaultUserID,
					StartsAt:  time.Now().Add(-2 * time.Hour),
					EndsAt:    time.Now().Add(-1 * time.Hour),
				}})
				return &SetupReturn{Exam: exams[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateExamRequest {
				title := "Updated Exam"
				return &proto.UpdateExamRequest{
					ExamHash: setupData.Exam.Hash,
					Title:    &title,
				}
			},
			Validate: func(t *testing.T, resp *proto.ExamResponse, setupData *SetupReturn) {
				var exam models.Exam
				require.NoError(t, dbInst.First(&exam, setupData.Exam.ID).Error)
				assert.Equal(t, "Updated Exam", exam.Title)
			},
			ExpectedCode: codes.OK,
		},
		{
			Name:     "Success - Update EndsAt after exam has started",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: defaultUserID,
					StartsAt:  time.Now().Add(-1 * time.Hour), // Started 1 hour ago
					EndsAt:    time.Now().Add(1 * time.Hour),  // Originally ends in 1 hour
				}})
				return &SetupReturn{Exam: exams[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateExamRequest {
				return &proto.UpdateExamRequest{
					ExamHash: setupData.Exam.Hash,
					EndsAt:   timestamppb.New(time.Now().Add(2 * time.Hour)), // Extend to 2 hours from now
				}
			},
			Validate: func(t *testing.T, resp *proto.ExamResponse, setupData *SetupReturn) {
				exam := setupData.Exam
				var updated models.Exam
				require.NoError(t, dbInst.First(&updated, exam.ID).Error)

				// Verify that EndsAt was updated but other fields remain unchanged
				assert.Equal(t, exam.Title, updated.Title)
				assert.Equal(t, exam.StartsAt.Unix(), updated.StartsAt.Unix())
				assert.Equal(t, exam.Type, updated.Type)
				assert.Greater(t, updated.EndsAt.Unix(), exam.EndsAt.Unix())
			},
			ExpectedCode: codes.OK,
		},
		{
			Name:     "Fail - Update duration exceeds maximum",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy: defaultUserID,
					StartsAt:  time.Now().Add(24 * time.Hour),
					EndsAt:    time.Now().Add(48 * time.Hour),
				}})
				return &SetupReturn{Exam: exams[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateExamRequest {
				return &proto.UpdateExamRequest{
					ExamHash:        setupData.Exam.Hash,
					DurationMinutes: ptr.Int32(int32(constants.MAX_EXAM_DURATION_MINUTES + 1)),
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Fail - Update duration after exam has started",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exams := createTestExams(t, []models.Exam{{
					CreatedBy:       defaultUserID,
					StartsAt:        time.Now().Add(-1 * time.Hour), // Started 1 hour ago
					EndsAt:          time.Now().Add(1 * time.Hour),  // Ends in 1 hour
					DurationMinutes: 60,
				}})
				return &SetupReturn{Exam: exams[0]}
			},
			GetRequest: func(setupData *SetupReturn) *proto.UpdateExamRequest {
				return &proto.UpdateExamRequest{
					ExamHash:        setupData.Exam.Hash,
					DurationMinutes: ptr.Int32(120), // Try to update to 2 hours
				}
			},
			ExpectedCode: codes.FailedPrecondition,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.UpdateExam)
		})
	}
}

func TestEndExam(t *testing.T) {
	type SetupReturn struct {
		Exam        models.Exam
		Participant models.ExamParticipant
	}

	testCases := []test.TestCase[*proto.EndExamRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name:     "Success - End exam",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, dbInst.Save(&exam).Error)

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_STARTED},
				})

				return &SetupReturn{
					Exam:        exam,
					Participant: participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.EndExamRequest {
				return &proto.EndExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				var exam models.Exam
				require.NoError(t, dbInst.First(&exam, setupData.Exam.ID).Error)
				counts, err := exam.GetParticipantCounts()
				require.NoError(t, err)

				assert.EqualValues(t, 0, counts.Started)
				assert.EqualValues(t, 1, counts.Ended)

				var participant models.ExamParticipant
				require.NoError(t, dbInst.First(&participant, setupData.Participant.ID).Error)
				assert.Equal(t, constants.PARTICIPANT_STATUS_ENDED, participant.Status)
				assert.True(t, participant.EndedAt.Valid)
			},
		},
		{
			Name:     "Fail - Participant not found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.Type = proto.ExamType_INVITE
				require.NoError(t, dbInst.Save(&exam).Error)
				return &SetupReturn{
					Exam:        exam,
					Participant: models.ExamParticipant{ID: 1},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.EndExamRequest {
				return &proto.EndExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Fail - Exam not found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					Exam:        models.Exam{ID: 9999},
					Participant: models.ExamParticipant{},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.EndExamRequest {
				return &proto.EndExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name:     "Fail - Exam not started yet",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.StartsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, dbInst.Save(&exam).Error)

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})

				return &SetupReturn{
					Exam:        exam,
					Participant: participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.EndExamRequest {
				return &proto.EndExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.FailedPrecondition,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.EndExam)
		})
	}
}

func TestStartExam(t *testing.T) {
	type SetupReturn struct {
		Exam models.Exam
	}

	testCases := []test.TestCase[*proto.StartExamRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name:     "Success - Start INVITE exam as invited participant",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.Type = proto.ExamType_INVITE
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, dbInst.Save(&exam).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.StartExamRequest {
				return &proto.StartExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				var exam models.Exam
				require.NoError(t, dbInst.First(&exam, setupData.Exam.ID).Error)
				counts, err := exam.GetParticipantCounts()
				require.NoError(t, err)

				assert.EqualValues(t, 0, counts.Invited)
				assert.EqualValues(t, 1, counts.Started)
				assert.EqualValues(t, 0, counts.Ended)

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ? AND user_id = ?", exam.ID, defaultUserID).First(&participant).Error)
				assert.Equal(t, constants.PARTICIPANT_STATUS_STARTED, participant.Status)
				assert.True(t, participant.StartedAt.Valid)
				assert.True(t, participant.ScheduledEndTime.Valid)
			},
		},
		{
			Name:     "Success - Start LINK exam as new participant",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.Type = proto.ExamType_LINK
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, dbInst.Save(&exam).Error)
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.StartExamRequest {
				return &proto.StartExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				var exam models.Exam
				require.NoError(t, dbInst.First(&exam, setupData.Exam.ID).Error)
				counts, err := exam.GetParticipantCounts()
				require.NoError(t, err)

				assert.EqualValues(t, 0, counts.Invited)
				assert.EqualValues(t, 1, counts.Started)
				assert.EqualValues(t, 0, counts.Ended)

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ? AND user_id = ?", exam.ID, defaultUserID).First(&participant).Error)
				assert.Equal(t, constants.PARTICIPANT_STATUS_STARTED, participant.Status)
			},
		},
		{
			Name:     "Fail - Exam not started yet",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.StartsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, dbInst.Save(&exam).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.StartExamRequest {
				return &proto.StartExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.FailedPrecondition,
		},
		{
			Name:     "Fail - Exam already ended",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.StartsAt = time.Now().Add(-2 * time.Hour)
				exam.EndsAt = time.Now().Add(-1 * time.Hour)
				require.NoError(t, dbInst.Save(&exam).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.StartExamRequest {
				return &proto.StartExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.FailedPrecondition,
		},
		{
			Name:     "Fail - Already started exam",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, dbInst.Save(&exam).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.StartExamRequest {
				return &proto.StartExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.FailedPrecondition,
		},
		{
			Name:     "Fail - Participant not added in INVITE exam",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.Type = proto.ExamType_INVITE
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, dbInst.Save(&exam).Error)
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.StartExamRequest {
				return &proto.StartExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Success - Start LINK exam with existing participant entry",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2) // Created by different user
				exam.Type = proto.ExamType_LINK
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				exam.EndsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, dbInst.Save(&exam).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.StartExamRequest {
				return &proto.StartExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				var updated models.Exam
				require.NoError(t, dbInst.First(&updated, setupData.Exam.ID).Error)
				counts, err := updated.GetParticipantCounts()
				require.NoError(t, err)

				assert.EqualValues(t, 0, counts.Invited)
				assert.EqualValues(t, 1, counts.Started)
				assert.EqualValues(t, 0, counts.Ended)

				var participant models.ExamParticipant
				require.NoError(t, dbInst.Where("exam_id = ? AND user_id = ?", setupData.Exam.ID, defaultUserID).First(&participant).Error)
				assert.Equal(t, constants.PARTICIPANT_STATUS_STARTED, participant.Status)
				assert.True(t, participant.StartedAt.Valid)
				assert.True(t, participant.ScheduledEndTime.Valid)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.StartExam)
		})
	}
}

func TestGetExamQuestions(t *testing.T) {
	type SetupReturn struct {
		Exam      models.Exam
		Questions []models.ExamQuestion
	}

	testCases := []test.TestCase[*proto.ExamRequest, *proto.ExamQuestionsResponse, *SetupReturn]{
		{
			Name:     "Success - Get questions as registered participant",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{QuestionID: 1},
					{QuestionID: 2, MaxScore: 5},
				})
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_STARTED},
				})
				return &SetupReturn{
					Exam:      exam,
					Questions: questions,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamQuestionsResponse, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Questions), len(resp.Questions))
				for i, q := range resp.Questions {
					assert.Equal(t, int64(setupData.Questions[i].QuestionID), getQuestionIdForHash(q.QuestionHash))
					assert.Equal(t, int64(setupData.Questions[i].CategoryID), q.CategoryId)
					assert.Equal(t, int32(setupData.Questions[i].Order), q.Order)
					assert.Equal(t, int32(setupData.Questions[i].MaxScore), q.MaxScore)
				}
			},
		},
		{
			Name:     "Fail - Unregistered participant cannot access questions",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.Type = proto.ExamType_LINK
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{MaxScore: 5},
				})
				return &SetupReturn{
					Exam:      exam,
					Questions: questions,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Success - Get questions as evaluator",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				questions := createTestExamQuestions(t, exam.ID, []models.ExamQuestion{
					{MaxScore: 10},
					{MaxScore: 5},
				})

				permission := models.ExamPermission{
					ExamID: exam.ID,
					UserID: defaultUserID,
				}
				permission.SetEvaluate()
				require.NoError(t, dbInst.Create(&permission).Error)

				return &SetupReturn{
					Exam:      exam,
					Questions: questions,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamQuestionsResponse, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Questions), len(resp.Questions))
				for i, q := range resp.Questions {
					assert.Equal(t, int64(setupData.Questions[i].QuestionID), getQuestionIdForHash(q.QuestionHash))
					assert.Equal(t, int64(setupData.Questions[i].CategoryID), q.CategoryId)
					assert.Equal(t, int32(setupData.Questions[i].Order), q.Order)
					assert.Equal(t, int32(setupData.Questions[i].MaxScore), q.MaxScore)
				}
			},
		},
		{
			Name:     "Fail - Exam not found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					Exam: models.Exam{ID: 9999},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetExamQuestions)
		})
	}
}

func TestGetExamCategories(t *testing.T) {
	type SetupReturn struct {
		Exam       models.Exam
		Categories []models.ExamCategory
	}

	testCases := []test.TestCase[*proto.ExamRequest, *proto.ExamCategoriesResponse, *SetupReturn]{
		{
			Name:     "Success - Get categories as registered participant",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				categories := []models.ExamCategory{
					{ExamID: exam.ID, CategoryID: 1},
					{ExamID: exam.ID, CategoryID: 2},
				}
				require.NoError(t, dbInst.Create(&categories).Error)

				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_STARTED},
				})

				return &SetupReturn{
					Exam:       exam,
					Categories: categories,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamCategoriesResponse, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Categories), len(resp.Categories))
				for i, c := range resp.Categories {
					assert.Equal(t, int64(setupData.Categories[i].CategoryID), c.CategoryId)
				}
			},
		},
		{
			Name:     "Fail - Unregistered participant cannot access categories",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.Type = proto.ExamType_LINK
				categories := []models.ExamCategory{
					{ExamID: exam.ID, CategoryID: 1},
					{ExamID: exam.ID, CategoryID: 2},
				}
				require.NoError(t, dbInst.Create(&categories).Error)

				return &SetupReturn{
					Exam:       exam,
					Categories: categories,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Success - Get categories as evaluator",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				categories := []models.ExamCategory{
					{ExamID: exam.ID, CategoryID: 1},
					{ExamID: exam.ID, CategoryID: 2},
				}
				require.NoError(t, dbInst.Create(&categories).Error)

				permission := models.ExamPermission{
					ExamID: exam.ID,
					UserID: defaultUserID,
				}
				permission.SetEvaluate()
				require.NoError(t, dbInst.Create(&permission).Error)

				return &SetupReturn{
					Exam:       exam,
					Categories: categories,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamCategoriesResponse, setupData *SetupReturn) {
				require.Equal(t, len(setupData.Categories), len(resp.Categories))
				for i, c := range resp.Categories {
					assert.Equal(t, int64(setupData.Categories[i].CategoryID), c.CategoryId)
				}
			},
		},
		{
			Name:     "Fail - Exam not found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					Exam: models.Exam{ID: 9999},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetExamCategories)
		})
	}
}

func TestGetExamPermission(t *testing.T) {
	type SetupReturn struct {
		Exam models.Exam
	}

	testCases := []test.TestCase[*proto.ExamRequest, *proto.ExamPermissionResponse, *SetupReturn]{
		{
			Name:     "Success - Owner permissions",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamPermissionResponse, setupData *SetupReturn) {
				assert.True(t, resp.CanRead)
				assert.True(t, resp.CanWrite)
				assert.False(t, resp.CanParticipate)
				assert.True(t, resp.CanEvaluate)
				assert.Nil(t, resp.ParticipantStatus)
			},
		},
		{
			Name:     "Success - Participant permissions",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamPermissionResponse, setupData *SetupReturn) {
				assert.True(t, resp.CanRead)
				assert.False(t, resp.CanWrite)
				assert.True(t, resp.CanParticipate)
				assert.False(t, resp.CanEvaluate)
				require.NotNil(t, resp.ParticipantStatus)
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_INVITED, *resp.ParticipantStatus)
			},
		},
		{
			Name:     "Success - LINK exam default permissions",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.Type = proto.ExamType_LINK
				require.NoError(t, dbInst.Save(&exam).Error)
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamPermissionResponse, setupData *SetupReturn) {
				assert.True(t, resp.CanRead)
				assert.False(t, resp.CanWrite)
				assert.True(t, resp.CanParticipate)
				assert.False(t, resp.CanEvaluate)
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_INVITED, *resp.ParticipantStatus)
			},
		},
		{
			Name:     "Fail - Exam not found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					Exam: models.Exam{ID: 9999},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetExamPermission)
		})
	}
}

func TestGetExam(t *testing.T) {
	type SetupReturn struct {
		Exam         models.Exam
		Participants []models.ExamParticipant
	}

	testCases := []test.TestCase[*proto.ExamRequest, *proto.ExamResponse, *SetupReturn]{
		{
			Name:     "Success - Get exam as owner",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamResponse, setupData *SetupReturn) {
				assert.Equal(t, setupData.Exam.Title, resp.Title)
				assert.Equal(t, int64(setupData.Exam.CreatedBy), resp.CreatedBy)
				assert.Equal(t, setupData.Exam.Type, resp.Type)
				assert.EqualValues(t, setupData.Exam.MaxCandidatesCount, resp.MaxCandidatesCount)
				assert.EqualValues(t, setupData.Exam.DurationMinutes, resp.DurationMinutes)

				assert.NotNil(t, resp.ParticipantCounts)
				assert.EqualValues(t, 0, resp.ParticipantCounts.Invited)
				assert.EqualValues(t, 0, resp.ParticipantCounts.Started)
				assert.EqualValues(t, 0, resp.ParticipantCounts.Ended)
			},
		},
		{
			Name:     "Success - Get exam as participant",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_INVITED},
				})
				return &SetupReturn{
					Exam:         exam,
					Participants: participants,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamResponse, setupData *SetupReturn) {
				assert.EqualValues(t, 2, resp.CreatedBy)
				assert.NotNil(t, resp.ParticipantCounts)
				assert.EqualValues(t, 1, resp.ParticipantCounts.Invited)
			},
		},
		{
			Name:     "Success - Get LINK exam as new participant",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.Type = proto.ExamType_LINK
				require.NoError(t, dbInst.Save(&exam).Error)
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamResponse, setupData *SetupReturn) {
				assert.EqualValues(t, proto.ExamType_LINK, resp.Type)
			},
		},
		{
			Name:     "Fail - No permission to access INVITE exam",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				exam.Type = proto.ExamType_INVITE
				require.NoError(t, dbInst.Save(&exam).Error)
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Fail - Exam not found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					Exam: models.Exam{ID: 9999},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetExam)
		})
	}
}

func TestDeleteExams(t *testing.T) {
	type SetupReturn struct {
		Exams []models.Exam
	}

	testCases := []test.TestCase[*proto.DeleteExamsRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name:     "Success - Delete multiple exams with hashes",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exams := createDefaultTestExams(t, defaultUserID, 2)

				return &SetupReturn{
					Exams: exams,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.DeleteExamsRequest {
				hashes := make([]string, len(setupData.Exams))
				for i, e := range setupData.Exams {
					hashes[i] = e.Hash
				}
				return &proto.DeleteExamsRequest{
					ExamHashes: hashes,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				examIDs := make([]types.ExamID, len(setupData.Exams))
				for i, e := range setupData.Exams {
					examIDs[i] = e.ID
				}

				// Verify exams are deleted
				var examCount int64
				err := dbInst.Model(&models.Exam{}).Where("id IN ?", examIDs).Count(&examCount).Error
				require.NoError(t, err)
				assert.Equal(t, int64(0), examCount)

				// Verify permissions are deleted
				var permCount int64
				err = dbInst.Model(&models.ExamPermission{}).Where("exam_id IN ?", examIDs).Count(&permCount).Error
				require.NoError(t, err)
				assert.Equal(t, int64(0), permCount)
			},
		},
		{
			Name:     "Fail - Attempt to delete exams created by another user",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2) // Created by a different user

				// Create read-only permission for the test user
				permission := models.ExamPermission{
					ExamID: exam.ID,
					UserID: defaultUserID,
				}
				permission.SetRead()
				require.NoError(t, dbInst.Create(&permission).Error)

				return &SetupReturn{
					Exams: []models.Exam{exam},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.DeleteExamsRequest {
				return &proto.DeleteExamsRequest{
					ExamHashes: []string{setupData.Exams[0].Hash},
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
		{
			Name:     "Success - Non-existent exam returns success",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					Exams: []models.Exam{{
						ID:   999,
						Hash: generate.HMACHash(999),
					}},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.DeleteExamsRequest {
				return &proto.DeleteExamsRequest{
					ExamHashes: []string{setupData.Exams[0].Hash},
				}
			},
			ExpectedCode: codes.OK,
		},
		{
			Name:     "Fail - Empty exam hashes list",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{}
			},
			GetRequest: func(setupData *SetupReturn) *proto.DeleteExamsRequest {
				return &proto.DeleteExamsRequest{
					ExamHashes: []string{},
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.DeleteExams)
		})
	}
}
