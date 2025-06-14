package tests

import (
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"testing"

	"google.golang.org/grpc/codes"
)

// baseTestCase contains common fields used across test cases
type baseTestCase struct {
	name         string
	metadata     map[string]string
	expectedCode codes.Code
	userID       types.UserID
}

type getExamResultsTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.Exam
	validate func(t *testing.T, resp *proto.ExamResultsResponse)
}

type getExamParticipantsTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.Exam
	validate func(t *testing.T, resp *proto.ParticipantList)
}

type getParticipantByIdTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.ExamParticipant
	validate func(t *testing.T, resp *proto.ParticipantResponse)
}

type getUserExamsTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.Exam
	validate func(t *testing.T, resp *proto.ExamList)
}

type createExamTestCase struct {
	baseTestCase
	request  *proto.CreateExamRequest
	validate func(t *testing.T, resp *proto.ExamResponse)
}

type updateExamTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.Exam
	request  *proto.UpdateExamRequest
	validate func(t *testing.T, exam *models.Exam)
}

type endExamTestCase struct {
	baseTestCase
	setup    func(t *testing.T) (*models.Exam, *models.ExamParticipant)
	validate func(t *testing.T, examID types.ExamID, participantID types.ParticipantID)
}

type startExamTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.Exam
	duration int32
	validate func(t *testing.T, exam *models.Exam)
}

type getExamQuestionsTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.Exam
	validate func(t *testing.T, resp *proto.ExamQuestionsResponse)
}

type deleteExamsTestCase struct {
	baseTestCase
	setup    func(t *testing.T) []models.Exam
	validate func(t *testing.T, exams []models.Exam)
}

type getParticipantAnswersTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.ExamParticipant
	validate func(t *testing.T, resp *proto.AnswerList)
}

type getAnswerForExamTestCase struct {
	baseTestCase
	setup    func(t *testing.T) (*models.Exam, types.QuestionID)
	validate func(t *testing.T, resp *proto.AnswerMinimalResponse)
}

type upsertAnswerTestCase struct {
	baseTestCase
	setup    func(t *testing.T) (*models.ExamParticipant, *proto.UpsertAnswersRequest)
	validate func(t *testing.T, resp *proto.UpsertAnswersResponse)
}

type getExamCategoriesTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.Exam
	validate func(t *testing.T, resp *proto.ExamCategoriesResponse)
}

type getExamPermissionTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.Exam
	validate func(t *testing.T, resp *proto.ExamPermissionResponse)
}

type getExamTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.Exam
	validate func(t *testing.T, resp *proto.ExamResponse)
}

type getExamParticipantTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.Exam
	validate func(t *testing.T, resp *proto.GetExamParticipantResponse)
}

type removeExamParticipantTestCase struct {
	baseTestCase
	setup    func(t *testing.T) (*models.Exam, *models.ExamParticipant)
	validate func(t *testing.T, examID types.ExamID, participantID types.ParticipantID)
}

type addExamParticipantTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.Exam
	request  *proto.AddParticipantRequest
	validate func(t *testing.T, examID types.ExamID, resp *proto.ParticipantResponse)
}

type getAnswerForEvaluationTestCase struct {
	baseTestCase
	setup    func(t *testing.T) (*models.ExamParticipant, types.QuestionID)
	validate func(t *testing.T, resp *proto.AnswerMinimalResponse)
}

type getAnswerEvaluationDataTestCase struct {
	baseTestCase
	setup    func(t *testing.T) (*models.ExamParticipant, types.QuestionID)
	validate func(t *testing.T, resp *proto.GetAnswerEvaluationDataResponse)
}

type updateAnswerForEvaluationTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.Answer
	request  *proto.UpdateAnswerRequest
	validate func(t *testing.T, answerId types.AnswerID)
}

type markParticipantAsEvaluatedTestCase struct {
	baseTestCase
	setup    func(t *testing.T) *models.ExamParticipant
	validate func(t *testing.T, participant *models.ExamParticipant, resp *proto.EvaluationStatusResponse)
}
