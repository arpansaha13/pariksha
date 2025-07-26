package tests

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
)

func TestGetExamParticipants(t *testing.T) {
	type SetupReturn struct {
		Exam         models.Exam
		Participants []models.ExamParticipant
	}

	testCases := []test.TestCase[*proto.ExamRequest, *proto.ParticipantList, *SetupReturn]{
		{
			Name:     "Success - Get exam participants",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{
						UserID: 2,
						Status: constants.PARTICIPANT_STATUS_INVITED,
					},
					{
						UserID: 3,
						Status: constants.PARTICIPANT_STATUS_STARTED,
					},
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
			Validate: func(t *testing.T, resp *proto.ParticipantList, setupData *SetupReturn) {
				assert.EqualValues(t, len(setupData.Participants), len(resp.Participants))
				for i, p := range resp.Participants {
					assert.EqualValues(t, setupData.Participants[i].Status, p.Status)
					assert.EqualValues(t, setupData.Participants[i].UserID, p.UserId)
				}
			},
		},
		{
			Name:     "Success - No participants",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				return &SetupReturn{
					Exam:         exam,
					Participants: []models.ExamParticipant{},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ParticipantList, setupData *SetupReturn) {
				assert.EqualValues(t, 0, len(resp.Participants))
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
			test.Runner(t, tc, client.GetExamParticipants)
		})
	}
}

func TestGetExamParticipant(t *testing.T) {
	type SetupReturn struct {
		Exam        models.Exam
		Participant models.ExamParticipant
	}

	testCases := []test.TestCase[*proto.GetExamParticipantRequest, *proto.GetExamParticipantResponse, *SetupReturn]{
		{
			Name:     "Success - Get participant with timing data",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				startTime := time.Now().Add(-1 * time.Hour)
				scheduledEndTime := startTime.Add(2 * time.Hour)

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
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

				return &SetupReturn{
					Exam:        exam,
					Participant: participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.GetExamParticipantRequest {
				return &proto.GetExamParticipantRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.GetExamParticipantResponse, setupData *SetupReturn) {
				assert.NotZero(t, resp.ParticipantId)
				require.NotNil(t, resp.StartedAt)
				require.NotNil(t, resp.ScheduledEndTime)
				assert.Equal(t, setupData.Participant.StartedAt.Time.Unix(), resp.StartedAt.AsTime().Unix())
				assert.Equal(t, setupData.Participant.ScheduledEndTime.Time.Unix(), resp.ScheduledEndTime.AsTime().Unix())
			},
		},
		{
			Name:     "Success - Get participant without timing data",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{
						Status: constants.PARTICIPANT_STATUS_INVITED,
					},
				})
				return &SetupReturn{
					Exam:        exam,
					Participant: participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.GetExamParticipantRequest {
				return &proto.GetExamParticipantRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.GetExamParticipantResponse, setupData *SetupReturn) {
				assert.NotZero(t, resp.ParticipantId)
				assert.Nil(t, resp.StartedAt)
				assert.Nil(t, resp.ScheduledEndTime)
			},
		},
		{
			Name:     "Fail - User is not a participant",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				return &SetupReturn{
					Exam: exam,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.GetExamParticipantRequest {
				return &proto.GetExamParticipantRequest{
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
			GetRequest: func(setupData *SetupReturn) *proto.GetExamParticipantRequest {
				return &proto.GetExamParticipantRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetExamParticipant)
		})
	}
}

func TestAddExamParticipant(t *testing.T) {
	type SetupReturn struct {
		Exam         models.Exam
		Participants []models.ExamParticipant
	}

	testCases := []test.TestCase[*proto.AddParticipantRequest, *proto.ParticipantResponse, *SetupReturn]{
		{
			Name:     "Success - Add participant",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				require.NoError(t, dbInst.Save(&exam).Error)
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.AddParticipantRequest {
				return &proto.AddParticipantRequest{
					ExamHash: setupData.Exam.Hash,
					UserId:   2,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ParticipantResponse, setupData *SetupReturn) {
				assert.EqualValues(t, 2, resp.UserId)
				assert.EqualValues(t, constants.PARTICIPANT_STATUS_INVITED, resp.Status)

				var exam models.Exam
				require.NoError(t, dbInst.First(&exam, setupData.Exam.ID).Error)
				counts, err := exam.GetParticipantCounts()
				require.NoError(t, err)
				assert.EqualValues(t, 1, counts.Invited)
				assert.EqualValues(t, 0, counts.Started)
				assert.EqualValues(t, 0, counts.Ended)
				assert.EqualValues(t, 0, counts.Unattended)
			},
		},
		{
			Name:     "Fail - Max candidates limit reached",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				exam.MaxCandidatesCount = 1
				require.NoError(t, dbInst.Save(&exam).Error)

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					UserID: 3,
					Status: constants.PARTICIPANT_STATUS_INVITED,
				}})

				return &SetupReturn{
					Exam:         exam,
					Participants: participants,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.AddParticipantRequest {
				return &proto.AddParticipantRequest{
					ExamHash: setupData.Exam.Hash,
					UserId:   2,
				}
			},
			ExpectedCode: codes.FailedPrecondition,
		},
		{
			Name:     "Fail - Cannot add participant to LINK exam",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				exam.Type = constants.EXAM_ACCESS_TYPE_LINK
				require.NoError(t, dbInst.Save(&exam).Error)
				return &SetupReturn{Exam: exam}
			},
			GetRequest: func(setupData *SetupReturn) *proto.AddParticipantRequest {
				return &proto.AddParticipantRequest{
					ExamHash: setupData.Exam.Hash,
					UserId:   2,
				}
			},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name:     "Fail - Exam not found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					Exam: models.Exam{ID: 9999},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.AddParticipantRequest {
				return &proto.AddParticipantRequest{
					ExamHash: setupData.Exam.Hash,
					UserId:   2,
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.AddExamParticipant)
		})
	}
}

func TestRemoveExamParticipant(t *testing.T) {
	type SetupReturn struct {
		Exam        models.Exam
		Participant models.ExamParticipant
	}

	testCases := []test.TestCase[*proto.RemoveParticipantRequest, *emptypb.Empty, *SetupReturn]{
		{
			Name:     "Success - Remove participant",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				exam.StartsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, dbInst.Save(&exam).Error)

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					UserID: 2,
					Status: constants.PARTICIPANT_STATUS_INVITED,
				}})
				return &SetupReturn{
					Exam:        exam,
					Participant: participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.RemoveParticipantRequest {
				return &proto.RemoveParticipantRequest{
					ExamHash:      setupData.Exam.Hash,
					ParticipantId: int64(setupData.Participant.ID),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *emptypb.Empty, setupData *SetupReturn) {
				var count int64
				dbInst.Model(&models.ExamParticipant{}).Where("id = ?", setupData.Participant.ID).Count(&count)
				assert.EqualValues(t, 0, count)
			},
		},
		{
			Name:     "Fail - Cannot remove after exam started",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				exam.StartsAt = time.Now().Add(-1 * time.Hour)
				require.NoError(t, dbInst.Save(&exam).Error)

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					UserID: 2,
					Status: constants.PARTICIPANT_STATUS_INVITED,
				}})
				return &SetupReturn{
					Exam:        exam,
					Participant: participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.RemoveParticipantRequest {
				return &proto.RemoveParticipantRequest{
					ExamHash:      setupData.Exam.Hash,
					ParticipantId: int64(setupData.Participant.ID),
				}
			},
			ExpectedCode: codes.FailedPrecondition,
		},
		{
			Name:     "Fail - Non-existent participant",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				exam.StartsAt = time.Now().Add(1 * time.Hour)
				require.NoError(t, dbInst.Save(&exam).Error)
				return &SetupReturn{
					Exam:        exam,
					Participant: models.ExamParticipant{ID: 9999},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.RemoveParticipantRequest {
				return &proto.RemoveParticipantRequest{
					ExamHash:      setupData.Exam.Hash,
					ParticipantId: 9999,
				}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name:     "Non-existent exam",
			Metadata: defaultMetadata,
			GetRequest: func(setupData *SetupReturn) *proto.RemoveParticipantRequest {
				return &proto.RemoveParticipantRequest{
					ExamHash:      "non_existent_hash",
					ParticipantId: 1,
				}
			},
			ExpectedCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.RemoveExamParticipant)
		})
	}
}

func TestGetParticipantById(t *testing.T) {
	type SetupReturn struct {
		Exam        models.Exam
		Participant models.ExamParticipant
	}

	testCases := []test.TestCase[*proto.ParticipantRequest, *proto.ParticipantResponse, *SetupReturn]{
		{
			Name:     "Success - Get participant details",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, defaultUserID)
				exam.Type = constants.EXAM_ACCESS_TYPE_INVITE
				require.NoError(t, dbInst.Save(&exam).Error)

				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					UserID: 2,
					Status: constants.PARTICIPANT_STATUS_STARTED,
				}})
				return &SetupReturn{
					Exam:        exam,
					Participant: participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantRequest {
				return &proto.ParticipantRequest{
					ParticipantId: int64(setupData.Participant.ID),
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ParticipantResponse, setupData *SetupReturn) {
				require.NotNil(t, resp)
				assert.NotZero(t, resp.ParticipantId)
				assert.Equal(t, int32(setupData.Participant.Status), resp.Status)
			},
		},
		{
			Name:     "Fail - Participant not found",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					Participant: models.ExamParticipant{ID: 9999},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantRequest {
				return &proto.ParticipantRequest{
					ParticipantId: int64(setupData.Participant.ID),
				}
			},
			ExpectedCode: codes.NotFound,
		},
		{
			Name:     "Fail - No evaluate permission",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2) // Different owner
				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{{
					UserID: 3,
					Status: constants.PARTICIPANT_STATUS_STARTED,
				}})
				return &SetupReturn{
					Exam:        exam,
					Participant: participants[0],
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ParticipantRequest {
				return &proto.ParticipantRequest{
					ParticipantId: int64(setupData.Participant.ID),
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetParticipantById)
		})
	}
}
