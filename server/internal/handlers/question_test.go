package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/server/internal/config/db"
	"pariksha/server/internal/dtos"
	testUtils "pariksha/server/internal/utils/test"
)

func TestGetPaperQuestions(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	paper := testUtils.CreateTestPaper(t, &models.Paper{})
	testUtils.CreateTestQuestion(t, &models.Question{
		PaperID:  paper.ID,
		Type:     constants.QUESTION_TYPE_MCQ,
		MaxScore: 5,
	})
	testUtils.CreateTestQuestion(t, &models.Question{
		PaperID:  paper.ID,
		Type:     constants.QUESTION_TYPE_SHORT,
		MaxScore: 5,
	})

	tests := []struct {
		name           string
		paperID        string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Success",
			paperID:        strconv.Itoa(paper.ID),
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Paper not found",
			paperID:        "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/papers/"+tt.paperID+"/questions", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.paperID})
			w := httptest.NewRecorder()

			GetPaperQuestions(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []dtos.QuestionResponse
				json.NewDecoder(w.Body).Decode(&response)
				assert.Equal(t, tt.expectedCount, len(response))
			}
		})
	}
}

func TestCreatePaperQuestions(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	paper := testUtils.CreateTestPaper(t, &models.Paper{})

	mcqQuestion := json.RawMessage(`{
        "statement": "Test MCQ",
        "options": ["A", "B", "C", "D"]
    }`)

	shortQuestion := json.RawMessage(`{
        "statement": "Test Short Question"
    }`)

	tests := []struct {
		name           string
		paperID        string
		questions      []dtos.CreateQuestionDto
		expectedStatus int
		validateFunc   func(t *testing.T, paperID string)
	}{
		{
			name:    "Success - Multiple question types",
			paperID: strconv.Itoa(paper.ID),
			questions: []dtos.CreateQuestionDto{
				{
					PaperID:       paper.ID,
					Question:      mcqQuestion,
					Type:          constants.QUESTION_TYPE_MCQ,
					Tags:          json.RawMessage(`["test"]`),
					MaxScore:      5,
					CorrectAnswer: "A",
				},
				{
					PaperID:  paper.ID,
					Question: shortQuestion,
					Type:     constants.QUESTION_TYPE_SHORT,
					Tags:     json.RawMessage(`["test"]`),
					MaxScore: 10,
				},
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(t *testing.T, paperID string) {
				// Verify questions were created
				var questions []models.Question
				db.DB.Where("paper_id = ?", paperID).Find(&questions)
				assert.Equal(t, 2, len(questions))

				// Verify paper score and counts updated
				var paper models.Paper
				db.DB.First(&paper, paperID)
				assert.Equal(t, 15, paper.MaxScore)

				counts, _ := paper.GetQuestionCounts()
				assert.Equal(t, 1, counts.MCQ)
				assert.Equal(t, 1, counts.Short)
			},
		},
		{
			name:    "Invalid question format",
			paperID: strconv.Itoa(paper.ID),
			questions: []dtos.CreateQuestionDto{
				{
					PaperID:  paper.ID,
					Question: json.RawMessage(`{invalid json`),
					Type:     constants.QUESTION_TYPE_MCQ,
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.questions)
			req := httptest.NewRequest("POST", "/papers/"+tt.paperID+"/questions", bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"id": tt.paperID})
			w := httptest.NewRecorder()

			CreatePaperQuestions(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.validateFunc != nil {
				tt.validateFunc(t, tt.paperID)
			}
		})
	}
}

func TestUpdateQuestion(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	paper := testUtils.CreateTestPaper(t, &models.Paper{
		MaxScore: 5,
	})
	question := testUtils.CreateTestQuestion(t, &models.Question{
		PaperID:  paper.ID,
		Type:     constants.QUESTION_TYPE_MCQ,
		MaxScore: 5,
	})

	tests := []struct {
		name           string
		questionID     string
		updateDto      dtos.UpdateQuestionDto
		expectedStatus int
		validateFunc   func(t *testing.T, questionID string)
	}{
		{
			name:       "Success - Update score",
			questionID: strconv.Itoa(question.ID),
			updateDto: dtos.UpdateQuestionDto{
				MaxScore: 10,
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, questionID string) {
				var updatedQuestion models.Question
				db.DB.First(&updatedQuestion, questionID)
				assert.Equal(t, 10, updatedQuestion.MaxScore)

				var paper models.Paper
				db.DB.First(&paper, updatedQuestion.PaperID)
				assert.Equal(t, 10, paper.MaxScore)
			},
		},
		{
			name:       "Question not found",
			questionID: "999",
			updateDto: dtos.UpdateQuestionDto{
				MaxScore: 10,
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.updateDto)
			req := httptest.NewRequest("PATCH", "/questions/"+tt.questionID, bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"id": tt.questionID})
			w := httptest.NewRecorder()

			UpdateQuestion(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.validateFunc != nil {
				tt.validateFunc(t, tt.questionID)
			}
		})
	}
}

func TestDeleteQuestion(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	paper := testUtils.CreateTestPaper(t, &models.Paper{
		MaxScore: 5,
		QuestionCounts: json.RawMessage(`{
			"mcq": 1,
			"short": 0,
			"long": 0
		}`),
	})
	question := testUtils.CreateTestQuestion(t, &models.Question{
		PaperID:  paper.ID,
		Type:     constants.QUESTION_TYPE_MCQ,
		MaxScore: 5,
	})

	tests := []struct {
		name           string
		questionID     string
		expectedStatus int
		validateFunc   func(t *testing.T, paperID int)
	}{
		{
			name:           "Success",
			questionID:     strconv.Itoa(question.ID),
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, paperID int) {
				// Verify question was deleted
				var count int64
				db.DB.Model(&models.Question{}).Where("id = ?", question.ID).Count(&count)
				assert.Equal(t, int64(0), count)

				// Verify paper counts updated
				var paper models.Paper
				db.DB.First(&paper, paperID)
				counts, _ := paper.GetQuestionCounts()
				assert.Equal(t, 0, counts.MCQ)
				assert.Equal(t, 0, paper.MaxScore)
			},
		},
		{
			name:           "Question not found",
			questionID:     "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/questions/"+tt.questionID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.questionID})
			w := httptest.NewRecorder()

			DeleteQuestion(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.validateFunc != nil {
				tt.validateFunc(t, paper.ID)
			}
		})
	}
}
