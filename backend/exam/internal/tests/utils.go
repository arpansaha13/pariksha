package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/generate"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/services/paper"
)

const (
	typedUserID types.UserID = 1 // Creator/admin user ID
	userID      int64        = 1 // Creator/admin user ID
	paperHash   string       = "some-string-for-now"
)

func createContextWithUserID(userID types.UserID) context.Context {
	md := metadata.New(map[string]string{
		"user_id": strconv.FormatInt(int64(userID), 10),
	})
	return metadata.NewOutgoingContext(context.Background(), md)
}

func createContextWithMetadata(mdMap map[string]string) context.Context {
	md := metadata.New(mdMap)
	return metadata.NewOutgoingContext(context.Background(), md)
}

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
		if exam.Type == "" {
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
		require.NoError(t, db.DB.Create(&exam).Error)

		// Generate and store hash
		exam.Hash = generate.HMACHash(int64(exam.ID))
		require.NoError(t, db.DB.Model(&exam).Update("hash", exam.Hash).Error)

		// Create permission
		permission := models.ExamPermission{
			ExamID: exam.ID,
			UserID: exam.CreatedBy,
		}
		permission.SetWrite()
		permission.SetEvaluate()
		require.NoError(t, db.DB.Create(&permission).Error)

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
		// Set UserID to typedUserID if not specified
		if participants[i].UserID == 0 {
			participants[i].UserID = typedUserID
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
	require.NoError(t, db.DB.Create(&participants).Error)

	// Create permissions for all participants
	permissions := make([]models.ExamPermission, len(participants))
	for i := range participants {
		permissions[i] = models.ExamPermission{
			ExamID: exam.ID,
			UserID: participants[i].UserID,
		}
		permissions[i].SetParticipate()
	}

	require.NoError(t, db.DB.Create(&permissions).Error)

	// Update exam counts
	newCounts, err := json.Marshal(counts)
	require.NoError(t, err)

	exam.ParticipantCounts = newCounts
	require.NoError(t, db.DB.Save(&exam).Error)

	return participants
}

func createTestAnswer(t *testing.T, examParticipant *models.ExamParticipant, questionID types.QuestionID) models.Answer {
	rawAnswer := json.RawMessage(`{"text": "Test Answer"}`)
	answer := models.Answer{
		ExamParticipantID: examParticipant.ID,
		QuestionID:        questionID,
		Answer:            &rawAnswer,
		Comments:          sql.NullString{String: "Test Comment", Valid: true},
		ScoreAwarded:      5,
		Evaluated:         true,
	}
	require.NoError(t, db.DB.Create(&answer).Error)
	return answer
}

// createTestExamQuestions creates exam questions with provided data, using defaults for missing fields
// The Order field is auto-filled based on slice index and should not be provided in input questions
func createTestExamQuestions(t *testing.T, exam *models.Exam, questions []models.ExamQuestion) []models.ExamQuestion {
	result := make([]models.ExamQuestion, len(questions))

	for i := range questions {
		// Validate that Order is not set in input
		require.Zero(t, questions[i].Order, "Order field should not be provided, it will be auto-filled")

		// Set required fields if not provided
		if questions[i].ExamID == 0 {
			questions[i].ExamID = exam.ID
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
		if questions[i].Type == 0 {
			questions[i].Type = proto.QuestionType_MCQ
		}

		// Set order based on index
		questions[i].Order = int16(i + 1)

		require.NoError(t, db.DB.Create(&questions[i]).Error)
		result[i] = questions[i]
	}

	return result
}

func getQuestionIdForHash(questionHash string) int64 {
	hashesList := []string{questionHash}
	questionIDs, _ := paper.FetchQuestionIdsForHashes(hashesList)
	return questionIDs[0]
}

func getQuestionHashForId(questionID types.QuestionID) string {
	idsList := make([]int64, 1)
	idsList[0] = int64(questionID)
	questionHashes, _ := paper.FetchQuestionHashesForIds(idsList)
	return questionHashes[0]
}
