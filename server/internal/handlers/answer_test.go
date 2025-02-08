package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/middlewares"
	"github.com/arpansaha13/pariksha/internal/models"
	testUtils "github.com/arpansaha13/pariksha/internal/utils/test"
)

func TestCreateAnswers(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	user := testUtils.CreateTestUser(t, &models.User{
		Verified: true,
	})
	participantUser := testUtils.CreateTestUser(t, &models.User{
		Email:    "participant@example.com",
		Verified: true,
	})
	paper := testUtils.CreateTestPaper(t, &models.Paper{})
	testUtils.CreateTestPaperOwnership(t, &models.PaperOwnership{
		UserID:  user.ID,
		PaperID: paper.ID,
	})
	question1 := testUtils.CreateTestQuestion(t, &models.Question{
		PaperID: paper.ID,
		Type:    constants.QUESTION_TYPE_MCQ,
	})
	question2 := testUtils.CreateTestQuestion(t, &models.Question{
		PaperID: paper.ID,
		Type:    constants.QUESTION_TYPE_MCQ,
	})
	exam := testUtils.CreateTestExam(t, &models.Exam{
		StartsAt:  time.Now().Add(-1 * time.Hour),
		EndsAt:    time.Now().Add(1 * time.Hour),
		CreatedBy: user.ID,
		PaperID:   paper.ID,
	})

	testUtils.CreateTestExamParticipant(t, &models.ExamParticipant{
		ExamID: exam.ID,
		UserID: participantUser.ID,
		Status: constants.PARTICIPANT_STATUS_STARTED,
		StartedAt: sql.NullTime{
			Time:  time.Now().Add(2 * time.Minute),
			Valid: true,
		},
		ScheduledEndTime: sql.NullTime{
			Time:  time.Now().Add((time.Duration(paper.DurationMinutes + 2)) * time.Minute),
			Valid: true,
		},
	})

	tests := []struct {
		name             string
		answerDTOs       []dtos.AnswerDTO
		expectedStatus   int
		expectedResponse map[string]int
	}{
		{
			name: "Successful answer submission",
			answerDTOs: []dtos.AnswerDTO{
				{
					Answer:      "Answer 1",
					SubmittedAt: time.Now().Add(3 * time.Minute),
					QuestionID:  question1.ID,
				},
			},
			expectedStatus: http.StatusCreated,
			expectedResponse: map[string]int{
				"totalCount":   1,
				"skippedCount": 0,
			},
		},
		{
			name: "Answer submitted after scheduled end time",
			answerDTOs: []dtos.AnswerDTO{
				{
					Answer:      "Answer 1",
					SubmittedAt: time.Now().Add(2 * time.Hour),
					QuestionID:  question1.ID,
				},
			},
			expectedStatus: http.StatusOK,
			expectedResponse: map[string]int{
				"totalCount":   1,
				"skippedCount": 1,
			},
		},
		{
			name: "Mixed valid and invalid answers",
			answerDTOs: []dtos.AnswerDTO{
				{
					Answer:      "Answer 1",
					SubmittedAt: time.Now().Add(3 * time.Minute),
					QuestionID:  question1.ID,
				},
				{
					Answer:      "Answer 2",
					SubmittedAt: time.Now().Add(2 * time.Hour),
					QuestionID:  question2.ID,
				},
			},
			expectedStatus: http.StatusCreated,
			expectedResponse: map[string]int{
				"totalCount":   2,
				"skippedCount": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			examID := strconv.Itoa(exam.ID)
			body, _ := json.Marshal(tt.answerDTOs)
			req, _ := http.NewRequest("POST", "/exams/"+examID+"/answers", bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"examId": examID})
			req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, participantUser.ID))

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(CreateAnswers)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response map[string]int
			json.NewDecoder(rr.Body).Decode(&response)
			assert.Equal(t, tt.expectedResponse, response)
		})
	}
}

func TestGetParticipantAnswers(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	user := testUtils.CreateTestUser(t, &models.User{
		Verified: true,
	})
	participantUser := testUtils.CreateTestUser(t, &models.User{
		Email:    "participant@example.com",
		Verified: true,
	})
	paper := testUtils.CreateTestPaper(t, &models.Paper{})
	testUtils.CreateTestPaperOwnership(t, &models.PaperOwnership{
		UserID:  user.ID,
		PaperID: paper.ID,
	})
	question1 := testUtils.CreateTestQuestion(t, &models.Question{
		PaperID: paper.ID,
		Type:    constants.QUESTION_TYPE_MCQ,
	})
	question2 := testUtils.CreateTestQuestion(t, &models.Question{
		PaperID: paper.ID,
		Type:    constants.QUESTION_TYPE_MCQ,
	})
	exam := testUtils.CreateTestExam(t, &models.Exam{
		StartsAt:  time.Now().Add(-1 * time.Hour),
		EndsAt:    time.Now().Add(1 * time.Hour),
		CreatedBy: user.ID,
		PaperID:   paper.ID,
	})
	examParticipant := testUtils.CreateTestExamParticipant(t, &models.ExamParticipant{
		ExamID: exam.ID,
		UserID: participantUser.ID,
	})

	answer1 := testUtils.CreateTestAnswer(t, &models.Answer{
		ExamParticipantID: examParticipant.ID,
		QuestionID:        question1.ID,
		Answer:            sql.NullString{String: "Answer 1", Valid: true},
	})
	answer2 := testUtils.CreateTestAnswer(t, &models.Answer{
		ExamParticipantID: examParticipant.ID,
		QuestionID:        question2.ID,
		Answer:            sql.NullString{String: "Answer 2", Valid: true},
	})

	tests := []struct {
		name            string
		examID          string
		participantID   string
		expectedStatus  int
		expectedAnswers []dtos.AnswerResponse
	}{
		{
			name:           "Successful retrieval of answers",
			examID:         strconv.Itoa(exam.ID),
			participantID:  strconv.Itoa(examParticipant.ID),
			expectedStatus: http.StatusOK,
			expectedAnswers: []dtos.AnswerResponse{
				{
					ID:                answer1.ID,
					ExamParticipantID: answer1.ExamParticipantID,
					Answer:            answer1.Answer.String,
					Comments:          answer1.Comments.String,
					ScoreAwarded:      answer1.ScoreAwarded,
					QuestionID:        answer1.QuestionID,
				},
				{
					ID:                answer2.ID,
					ExamParticipantID: answer2.ExamParticipantID,
					Answer:            answer2.Answer.String,
					Comments:          answer2.Comments.String,
					ScoreAwarded:      answer2.ScoreAwarded,
					QuestionID:        answer2.QuestionID,
				},
			},
		},
		{
			name:            "No answers found",
			examID:          strconv.Itoa(exam.ID),
			participantID:   "999",
			expectedStatus:  http.StatusNotFound,
			expectedAnswers: []dtos.AnswerResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/exams/"+tt.examID+"/participants/"+tt.participantID+"/answers", nil)
			req = mux.SetURLVars(req, map[string]string{"examId": tt.examID, "participantId": tt.participantID})

			w := httptest.NewRecorder()

			GetParticipantAnswers(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var answers []dtos.AnswerResponse
				json.NewDecoder(w.Body).Decode(&answers)
				assert.Equal(t, tt.expectedAnswers, answers)
			}
		})
	}
}
