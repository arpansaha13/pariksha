package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
)

func TestGetExamResults(t *testing.T) {
	type SetupReturn struct {
		Exam        models.Exam
		Participant models.ExamParticipant
		Answers     []models.Answer
	}

	testCases := []test.TestCase[*proto.ExamRequest, *proto.ExamResultsResponse, *SetupReturn]{
		{
			Name:     "Success - Get results with all questions answered",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)

				// Create participant
				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_EVALUATED},
				})

				// Create answers
				var answers []models.Answer
				answers = append(answers, createTestAnswer(t, &participants[0], 1))
				answers = append(answers, createTestAnswer(t, &participants[0], 2))

				return &SetupReturn{
					Exam:        exam,
					Participant: participants[0],
					Answers:     answers,
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamResultsResponse, setupData *SetupReturn) {
				require.Len(t, resp.Results, len(setupData.Answers))

				for i, result := range resp.Results {
					assert.NotZero(t, result.AnswerId)
					assert.EqualValues(t, setupData.Answers[i].ScoreAwarded, result.ScoreAwarded)
				}
			},
		},
		{
			Name:     "Success - Exam with no answers",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				exam := createDefaultTestExam(t, 2)
				participants := createTestExamParticipants(t, &exam, []models.ExamParticipant{
					{Status: constants.PARTICIPANT_STATUS_EVALUATED},
				})
				return &SetupReturn{
					Exam:        exam,
					Participant: participants[0],
					Answers:     []models.Answer{},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.ExamResultsResponse, setupData *SetupReturn) {
				assert.Empty(t, resp.Results)
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
			GetRequest: func(setupData *SetupReturn) *proto.ExamRequest {
				return &proto.ExamRequest{
					ExamHash: setupData.Exam.Hash,
				}
			},
			ExpectedCode: codes.PermissionDenied,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clearTables(t)
			test.Runner(t, tc, client.GetExamResults)
		})
	}
}
