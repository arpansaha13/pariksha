package tests

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/generate"
	"pariksha/exam/internal/interservice"
)

const (
	defaultUserID types.UserID = 1
	paperHash     string       = "some-string-for-now"
)

var defaultMetadata metadata.MD = metadata.MD{
	"user_id": []string{"1"},
}

// Set in TestMain
var (
	dbInst         *gorm.DB
	questionIntSvc *interservice.Question
)

// createTestExams creates multiple exam entries with default values for missing fields
func createTestExams(t *testing.T, exams []models.Exam) []models.Exam {
	result := make([]models.Exam, len(exams))

	for i, exam := range exams {
		// Validate required createdBy field
		require.NotEqual(t, types.UserID(0), exam.CreatedBy, "CreatedBy is required")

		// Set default values if not provided
		if exam.Title == "" {
			exam.Title = "Test Exam " + strconv.Itoa(i+1)
		}
		if exam.Type == 0 {
			exam.Type = constants.EXAM_ACCESS_TYPE_LINK
		}
		if exam.MaxCandidatesCount == 0 {
			exam.MaxCandidatesCount = 10
		}
		if exam.PaperHash == "" {
			exam.PaperHash = paperHash
		}
		if exam.DurationMinutes == 0 {
			exam.DurationMinutes = 60
		}
		if exam.ParticipantCounts == nil {
			exam.ParticipantCounts = []byte(`{"unattended":0,"invited":0,"started":0,"ended":0}`)
		}

		// Create exam
		require.NoError(t, dbInst.Create(&exam).Error)

		// Generate and store hash
		exam.Hash = generate.HMACHash(int64(exam.ID))
		require.NoError(t, dbInst.Model(&exam).Update("hash", exam.Hash).Error)

		// Create permission
		permission := models.ExamPermission{
			ExamID: exam.ID,
			UserID: exam.CreatedBy,
		}
		permission.SetWrite()
		permission.SetEvaluate()
		require.NoError(t, dbInst.Create(&permission).Error)

		result[i] = exam
	}

	return result
}

func createDefaultTestExams(t *testing.T, createdBy types.UserID, count int8) []models.Exam {
	require.Greater(t, count, int8(0), "invalid count argument to createDefaultTestExams")

	exams := make([]models.Exam, count)
	for i := range exams {
		exams[i] = models.Exam{
			CreatedBy: createdBy,
		}
	}

	exams = createTestExams(t, exams)
	return exams
}

func createDefaultTestExam(t *testing.T, createdBy types.UserID) models.Exam {
	exams := createDefaultTestExams(t, createdBy, 1)
	return exams[0]
}

// createTestExamParticipants creates exam participants and updates exam counts
func createTestExamParticipants(t *testing.T, exam *models.Exam, participants []models.ExamParticipant) []models.ExamParticipant {
	counts, err := exam.GetParticipantCounts()
	require.NoError(t, err)

	// Update counts based on participant status
	for i := range participants {
		// Set exam ID if not set
		if participants[i].ExamID == 0 {
			participants[i].ExamID = exam.ID
		}
		// Set UserID to defaultUserID if not specified
		if participants[i].UserID == 0 {
			participants[i].UserID = defaultUserID
		}

		// Update counts based on status
		switch participants[i].Status {
		case constants.PARTICIPANT_STATUS_INVITED:
			counts.Invited++
		case constants.PARTICIPANT_STATUS_STARTED:
			counts.Started++
		case constants.PARTICIPANT_STATUS_ENDED:
			counts.Ended++
		case constants.PARTICIPANT_STATUS_UNATTENDED:
			counts.Unattended++
		}
	}

	// Save participants
	require.NoError(t, dbInst.Create(&participants).Error)

	// Create permissions for all participants
	permissions := make([]models.ExamPermission, len(participants))
	for i := range participants {
		permissions[i] = models.ExamPermission{
			ExamID: exam.ID,
			UserID: participants[i].UserID,
		}
		permissions[i].SetParticipate()
	}

	require.NoError(t, dbInst.Create(&permissions).Error)

	// Update exam counts
	newCounts, err := json.Marshal(counts)
	require.NoError(t, err)

	exam.ParticipantCounts = newCounts
	require.NoError(t, dbInst.Save(&exam).Error)

	return participants
}

func createTestAnswer(t *testing.T, examParticipant *models.ExamParticipant, questionID types.QuestionID) models.Answer {
	rawAnswer := json.RawMessage(`{"text": "Test Answer"}`)
	answer := models.Answer{
		ExamParticipantID: examParticipant.ID,
		QuestionID:        questionID,
		Answer:            &rawAnswer,
		ScoreAwarded:      5,
		Evaluated:         true,
	}
	require.NoError(t, dbInst.Create(&answer).Error)
	return answer
}

// createTestExamQuestions creates exam questions with provided data, using defaults for missing fields
// The Order field is auto-filled based on slice index and should not be provided in input questions
func createTestExamQuestions(t *testing.T, examID types.ExamID, questions []models.ExamQuestion) []models.ExamQuestion {
	result := make([]models.ExamQuestion, len(questions))

	for i := range questions {
		// Validate that Order is not set in input
		require.Zero(t, questions[i].Order, "Order field should not be provided, it will be auto-filled")

		// Set required fields if not provided
		if questions[i].ExamID == 0 {
			questions[i].ExamID = examID
		}
		if questions[i].CategoryID == 0 {
			questions[i].CategoryID = 10
		}
		if questions[i].CategoryID == 0 {
			questions[i].CategoryID = 10
		}
		if questions[i].QuestionID == 0 {
			questions[i].QuestionID = types.QuestionID(i + 1)
		}
		if questions[i].MaxScore == 0 {
			questions[i].MaxScore = 10
		}

		// Set order based on index
		questions[i].Order = int16(i + 1)

		require.NoError(t, dbInst.Create(&questions[i]).Error)
		result[i] = questions[i]
	}

	return result
}

func getQuestionIdForHash(questionHash string) int64 {
	hashesList := []string{questionHash}
	questionIDs, err := questionIntSvc.GetQuestionIDsByHashes(hashesList)
	if err != nil || len(questionIDs) == 0 {
		panic("getQuestionIdForHash: could not find question ID for hash: " + questionHash)
	}
	return int64(questionIDs[0])
}

func getQuestionHashForId(questionID types.QuestionID) string {
	idsList := []types.QuestionID{questionID}
	questionHashes, err := questionIntSvc.GetQuestionHashesByIds(idsList)
	if err != nil || len(questionHashes) == 0 {
		panic("getQuestionHashForId: could not find question hash for ID: " + strconv.FormatInt(int64(questionID), 10))
	}
	return questionHashes[0]
}
