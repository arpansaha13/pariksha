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
	paper := testUtils.CreateTestPaper(t, &models.Paper{})

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
	paper := testUtils.CreateTestPaper(t, &models.Paper{})
	exam := testUtils.CreateTestExam(t, &models.Exam{
		CreatedBy: user.ID,
		PaperID:   paper.ID,
	})
	testUtils.CreateTestExamParticipant(t, &models.ExamParticipant{
		ExamID: exam.ID,
		UserID: user.ID,
		Status: constants.PARTICIPANT_STATUS_INVITED,
	})

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
	paper := testUtils.CreateTestPaper(t, &models.Paper{})
	exam := testUtils.CreateTestExam(t, &models.Exam{
		CreatedBy: user.ID,
		PaperID:   paper.ID,
	})
	participant := testUtils.CreateTestExamParticipant(t, &models.ExamParticipant{
		ExamID: exam.ID,
		UserID: user.ID,
		Status: constants.PARTICIPANT_STATUS_INVITED,
	})

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

func TestAddExamParticipants(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	user := testUtils.CreateTestUser(t)
	paper := testUtils.CreateTestPaper(t, &models.Paper{})
	exam := testUtils.CreateTestExam(t, &models.Exam{
		StartsAt:  time.Now().Add(24 * time.Hour),
		EndsAt:    time.Now().Add(48 * time.Hour),
		CreatedBy: user.ID,
		PaperID:   paper.ID,
		Type:      constants.EXAM_TYPE_INVITE,
	})
	maxLimitExam := testUtils.CreateTestExam(t, &models.Exam{
		Title:              "Max Limit Exam",
		StartsAt:           time.Now().Add(time.Hour),
		EndsAt:             time.Now().Add(2 * time.Hour),
		CreatedBy:          user.ID,
		PaperID:            paper.ID,
		Type:               constants.EXAM_TYPE_INVITE,
		MaxCandidatesCount: 2,
	})

	tests := []struct {
		name           string
		examID         string
		participants   []dtos.AddExamParticipantDto
		expectedStatus int
		validateFunc   func(t *testing.T, examID int)
	}{
		{
			name:   "Success - Add registered user",
			examID: strconv.Itoa(exam.ID),
			participants: []dtos.AddExamParticipantDto{
				{UserID: user.ID},
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(t *testing.T, examID int) {
				// Verify participant was added
				var participant models.ExamParticipant
				err := db.DB.Where("exam_id = ? AND user_id = ?", examID, user.ID).First(&participant).Error
				assert.NoError(t, err)
				assert.Equal(t, constants.PARTICIPANT_STATUS_INVITED, participant.Status)

				// Verify count was updated
				var updatedExam models.Exam
				db.DB.First(&updatedExam, examID)
				counts, err := updatedExam.GetParticipantCounts()
				assert.NoError(t, err)
				assert.Equal(t, 1, counts.Invited)
			},
		},
		{
			name:   "Success - Add guest user",
			examID: strconv.Itoa(exam.ID),
			participants: []dtos.AddExamParticipantDto{
				{
					Email:     "guest@test.com",
					FirstName: "Guest",
					LastName:  "User",
				},
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(t *testing.T, examID int) {
				// Verify guest user was created
				var user models.User
				err := db.DB.Where("email = ?", "guest@test.com").First(&user).Error
				assert.NoError(t, err)
				assert.True(t, user.IsGuest)

				// Verify participant was added with correct status
				var participant models.ExamParticipant
				err = db.DB.Where("exam_id = ? AND user_id = ?", examID, user.ID).First(&participant).Error
				assert.NoError(t, err)
				assert.Equal(t, constants.PARTICIPANT_STATUS_INVITED, participant.Status)
			},
		},
		{
			name:   "Max Candidates Limit Reached",
			examID: strconv.Itoa(maxLimitExam.ID),
			participants: []dtos.AddExamParticipantDto{
				{
					UserID: user.ID,
				},
				{
					Email:     "test1@example.com",
					FirstName: "Test1",
					LastName:  "User1",
				},
				{
					Email:     "test2@example.com",
					FirstName: "Test2",
					LastName:  "User2",
				},
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(t *testing.T, examID int) {
				// Verify participants were added
				var count int64
				err := db.DB.Model(models.ExamParticipant{}).Where("exam_id = ?", examID).Count(&count).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(2), count) // 1 candidate omitted

				// Verify count was updated
				var updatedExam models.Exam
				db.DB.Take(&updatedExam, examID)
				counts, err := updatedExam.GetParticipantCounts()
				assert.NoError(t, err)
				assert.Equal(t, 2, counts.Invited)
			},
		},
		{
			name:   "Invalid exam ID",
			examID: "999999",
			participants: []dtos.AddExamParticipantDto{
				{UserID: user.ID},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Missing required fields",
			examID: strconv.Itoa(exam.ID),
			participants: []dtos.AddExamParticipantDto{
				{}, // No UserID or Email
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.participants)
			req := httptest.NewRequest("POST", "/exams/"+tt.examID+"/participants", bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"examId": tt.examID})
			w := httptest.NewRecorder()

			AddExamParticipants(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.validateFunc != nil {
				examID, _ := strconv.Atoi(tt.examID)
				tt.validateFunc(t, examID)
			}
		})
	}
}

func TestRemoveExamParticipant(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	user := testUtils.CreateTestUser(t)
	paper := testUtils.CreateTestPaper(t, &models.Paper{})

	participantCounts, _ := json.Marshal(models.ParticipantCount{
		Unattended: 0,
		Invited:    1,
		Started:    0,
		Ended:      0,
	})

	// Create exam that hasn't started
	futureExam := testUtils.CreateTestExam(t, &models.Exam{
		Title:             "Future Exam",
		StartsAt:          time.Now().Add(24 * time.Hour),
		EndsAt:            time.Now().Add(48 * time.Hour),
		CreatedBy:         user.ID,
		PaperID:           paper.ID,
		ParticipantCounts: participantCounts,
	})

	startedExam := testUtils.CreateTestExam(t, &models.Exam{
		Title:             "Started Exam",
		StartsAt:          time.Now().Add(-1 * time.Hour),
		EndsAt:            time.Now().Add(24 * time.Hour),
		CreatedBy:         user.ID,
		PaperID:           paper.ID,
		ParticipantCounts: participantCounts,
	})

	// Add participant to future exam
	futureExamParticipant := testUtils.CreateTestExamParticipant(t, &models.ExamParticipant{
		ExamID: futureExam.ID,
		UserID: user.ID,
		Status: constants.PARTICIPANT_STATUS_INVITED,
	})

	// Add participant to started exam
	startedExamParticipant := testUtils.CreateTestExamParticipant(t, &models.ExamParticipant{
		ExamID: startedExam.ID,
		UserID: user.ID,
		Status: constants.PARTICIPANT_STATUS_INVITED,
	})

	tests := []struct {
		name           string
		examID         string
		participantID  string
		expectedStatus int
		validateFunc   func(t *testing.T, examID, participantID string)
	}{
		{
			name:           "Success",
			examID:         strconv.Itoa(futureExam.ID),
			participantID:  strconv.Itoa(futureExamParticipant.ID),
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, examID, participantID string) {
				// Verify participant was removed
				var participant models.ExamParticipant
				err := db.DB.First(&participant, participantID).Error
				assert.Error(t, err) // Should not find participant

				// Verify count was decremented
				var exam models.Exam
				db.DB.First(&exam, examID)
				counts, err := exam.GetParticipantCounts()
				assert.NoError(t, err)
				assert.Equal(t, 0, counts.Invited)
			},
		},
		{
			name:           "Exam already started",
			examID:         strconv.Itoa(startedExam.ID),
			participantID:  strconv.Itoa(startedExamParticipant.ID),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid exam ID",
			examID:         "999999",
			participantID:  strconv.Itoa(futureExamParticipant.ID),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid participant ID",
			examID:         strconv.Itoa(futureExam.ID),
			participantID:  "999999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/exams/"+tt.examID+"/participants/"+tt.participantID, nil)
			req = mux.SetURLVars(req, map[string]string{
				"examId":        tt.examID,
				"participantId": tt.participantID,
			})
			w := httptest.NewRecorder()

			RemoveExamParticipant(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.validateFunc != nil {
				tt.validateFunc(t, tt.examID, tt.participantID)
			}
		})
	}
}
