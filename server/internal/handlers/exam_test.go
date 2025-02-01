package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/models"
	testUtils "github.com/arpansaha13/pariksha/internal/utils/test"
)

func TestCreateExam(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	user := testUtils.CreateTestUser(t)
	paper := testUtils.CreateTestPaper(t)

	tests := []struct {
		name           string
		examDto        dtos.CreateExamDto
		expectedStatus int
	}{
		{
			name: "Success",
			examDto: dtos.CreateExamDto{
				Title:              "Test Exam",
				StartsAt:           time.Now().Add(time.Hour),
				EndsAt:             time.Now().Add(2 * time.Hour),
				Type:               constants.EXAM_TYPE_INVITE,
				MaxCandidatesCount: 10,
				PaperID:            paper.ID,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid Paper ID",
			examDto: dtos.CreateExamDto{
				Title:              "Test Exam",
				StartsAt:           time.Now().Add(time.Hour),
				EndsAt:             time.Now().Add(2 * time.Hour),
				Type:               constants.EXAM_TYPE_INVITE,
				MaxCandidatesCount: 10,
				PaperID:            9999,
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.examDto)
			req := httptest.NewRequest("POST", "/exams", bytes.NewBuffer(body))
			req = req.WithContext(testUtils.SetUserContext(req.Context(), user.ID))
			w := httptest.NewRecorder()

			CreateExam(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var exam models.Exam
				if err := db.DB.Take(&exam, "title = ?", "Test Exam").Error; err != nil {
					t.Fatalf("Failed to fetch exam: %v", err)
				}
				// Verify exam entry in db
				if assert.Equal(t, "Test Exam", exam.Title) {
					assert.Equal(t, user.ID, exam.CreatedBy)
					assert.Equal(t, paper.ID, exam.PaperID)
				}
			}
		})
	}
}

func TestGetExamParticipants(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	user := testUtils.CreateTestUser(t)
	paper := testUtils.CreateTestPaper(t)
	exam := testUtils.CreateTestExam(t, user.ID, paper.ID)
	testUtils.CreateTestExamParticipant(t, user.ID, exam.ID, constants.PARTICIPANT_STATUS_INVITED)

	tests := []struct {
		name           string
		examID         string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Success",
			examID:         strconv.Itoa(exam.ID),
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Invalid Exam ID",
			examID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent Exam",
			examID:         "9999",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/exams/"+tt.examID+"/participants", nil)
			req = mux.SetURLVars(req, map[string]string{"examId": tt.examID})
			w := httptest.NewRecorder()

			GetExamParticipants(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []dtos.ExamParticipantResponse
				json.NewDecoder(w.Body).Decode(&response)
				assert.Equal(t, tt.expectedCount, len(response))
			}
		})
	}
}

func TestStartExam(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	user := testUtils.CreateTestUser(t)
	paper := testUtils.CreateTestPaper(t)
	exam := testUtils.CreateTestExam(t, user.ID, paper.ID)
	participant := testUtils.CreateTestExamParticipant(t, user.ID, exam.ID, constants.PARTICIPANT_STATUS_INVITED)

	tests := []struct {
		name           string
		examID         string
		participantID  string
		expectedStatus int
	}{
		{
			name:           "Success",
			examID:         strconv.Itoa(exam.ID),
			participantID:  strconv.Itoa(participant.ID),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Exam ID",
			examID:         "invalid",
			participantID:  strconv.Itoa(participant.ID),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Already Started",
			examID:         strconv.Itoa(exam.ID),
			participantID:  strconv.Itoa(participant.ID),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("PATCH", "/exams/"+tt.examID+"/participants/"+tt.participantID+"/start", nil)
			req = mux.SetURLVars(req, map[string]string{
				"examId":        tt.examID,
				"participantId": tt.participantID,
			})
			w := httptest.NewRecorder()

			StartExam(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var updatedParticipant models.ExamParticipant
				err := db.DB.Preload("Exam.Paper").Take(&updatedParticipant, tt.participantID).Error
				assert.NoError(t, err)

				// Verify status changed to STARTED
				assert.Equal(t, constants.PARTICIPANT_STATUS_STARTED, updatedParticipant.Status)

				// Verify StartedAt is set
				assert.True(t, updatedParticipant.StartedAt.Valid)

				// Verify ScheduledEndTime is set correctly
				expectedEndTime := updatedParticipant.StartedAt.Time.Add(
					time.Duration(updatedParticipant.Exam.Paper.DurationMinutes) * time.Minute,
				)
				assert.True(t, updatedParticipant.ScheduledEndTime.Valid)
				assert.Equal(t, expectedEndTime.Unix(), updatedParticipant.ScheduledEndTime.Time.Unix())
			}
		})
	}
}
