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

func createTestExam(t *testing.T, createdBy types.UserID) models.Exam {
	exam := models.Exam{
		Title:              "Test Exam",
		CreatedBy:          createdBy,
		Type:               constants.EXAM_ACCESS_TYPE_LINK,
		MaxCandidatesCount: 10,
		PaperID:            1,
		DurationMinutes:    60,
		ParticipantCounts:  []byte(`{"unattended":0,"invited":0,"started":0,"ended":0}`),
	}
	require.NoError(t, db.DB.Create(&exam).Error)

	// Create exam hash
	examHash := models.ExamHash{
		ID:   exam.ID,
		Hash: generate.HMACHash(int64(exam.ID)),
	}
	require.NoError(t, db.DB.Create(&examHash).Error)
	exam.ExamHash = examHash

	// Create permission
	permission := models.ExamPermission{
		ExamID: exam.ID,
		UserID: createdBy,
	}
	permission.SetWrite()
	permission.SetEvaluate()
	require.NoError(t, db.DB.Create(&permission).Error)

	return exam
}

func createTestExamParticipants(t *testing.T, exam *models.Exam, participants []TestParticipantData) error {
	examParticipants := make([]models.ExamParticipant, len(participants))
	counts, err := exam.GetParticipantCounts()
	if err != nil {
		return err
	}

	// Create participants and update counts
	for i, p := range participants {
		examParticipants[i] = models.ExamParticipant{
			ExamID: exam.ID,
			UserID: p.UserID,
			Status: p.Status,
		}

		// Update counts based on status
		switch p.Status {
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
	if err := db.DB.Create(&examParticipants).Error; err != nil {
		return err
	}

	// Create permissions for all participants
	permissions := make([]models.ExamPermission, len(participants))
	for i, p := range participants {
		permissions[i] = models.ExamPermission{
			ExamID: exam.ID,
			UserID: p.UserID,
		}
		permissions[i].SetParticipate()
	}

	if err := db.DB.Create(&permissions).Error; err != nil {
		return err
	}

	// Update exam counts
	exam.ParticipantCounts, err = json.Marshal(counts)
	if err != nil {
		return err
	}

	return db.DB.Save(&exam).Error
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

// createTestExamQuestion creates an exam question with provided data, using defaults for missing fields
func createTestExamQuestion(t *testing.T, exam *models.Exam, question models.ExamQuestion) models.ExamQuestion {
	// Set required fields if not provided
	if question.ExamID == 0 {
		question.ExamID = exam.ID
	}
	if question.Order == 0 {
		question.Order = 1
	}
	if question.MaxScore == 0 {
		question.MaxScore = 10
	}
	if question.Type == 0 {
		question.Type = proto.QuestionType_MCQ
	}
	require.NoError(t, db.DB.Create(&question).Error)
	return question
}
