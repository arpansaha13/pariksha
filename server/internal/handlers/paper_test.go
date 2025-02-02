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

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/models"
	testUtils "github.com/arpansaha13/pariksha/internal/utils/test"
)

func TestGetUserPapers(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	user := testUtils.CreateTestUser(t)

	paper1 := testUtils.CreateTestPaper(t, &models.Paper{
		Title: "Test Paper 1",
	})
	paper2 := testUtils.CreateTestPaper(t, &models.Paper{
		Title: "Test Paper 2",
	})
	testUtils.CreateTestPaperOwnership(t, &models.PaperOwnership{
		UserID:  user.ID,
		PaperID: paper1.ID,
	})
	testUtils.CreateTestPaperOwnership(t, &models.PaperOwnership{
		UserID:  user.ID,
		PaperID: paper2.ID,
	})

	tests := []struct {
		name           string
		userID         int
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Success - User has papers",
			userID:         user.ID,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Success - User has no papers",
			userID:         999,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/papers", nil)
			req = req.WithContext(testUtils.SetUserContext(req.Context(), tt.userID))
			w := httptest.NewRecorder()

			GetUserPapers(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []dtos.PaperResponse
				json.NewDecoder(w.Body).Decode(&response)
				assert.Equal(t, tt.expectedCount, len(response))
			}
		})
	}
}

func TestCreatePaper(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	user := testUtils.CreateTestUser(t)

	tests := []struct {
		name           string
		paperDto       dtos.CreatePaperDto
		expectedStatus int
		validateFunc   func(t *testing.T, response dtos.PaperResponse)
	}{
		{
			name: "Success",
			paperDto: dtos.CreatePaperDto{
				Title: "New Paper",
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(t *testing.T, response dtos.PaperResponse) {
				assert.NotZero(t, response.ID)
				assert.Equal(t, "New Paper", response.Title)
				assert.Equal(t, constants.PAPER_OWNERSHIP_TYPE_OWNER, response.PaperOwnership.Type)
			},
		},
		{
			name:           "Invalid request body",
			paperDto:       dtos.CreatePaperDto{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.paperDto)
			req := httptest.NewRequest("POST", "/papers", bytes.NewBuffer(body))
			req = req.WithContext(testUtils.SetUserContext(req.Context(), user.ID))
			w := httptest.NewRecorder()

			CreatePaper(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.validateFunc != nil {
				var response dtos.PaperResponse
				json.NewDecoder(w.Body).Decode(&response)
				tt.validateFunc(t, response)
			}
		})
	}
}

func TestUpdatePaper(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	testUtils.CreateTestUser(t)
	paper := testUtils.CreateTestPaper(t, &models.Paper{})

	tests := []struct {
		name           string
		paperID        string
		paperDto       dtos.UpdatePaperDto
		expectedStatus int
	}{
		{
			name:    "Success",
			paperID: strconv.Itoa(paper.ID),
			paperDto: dtos.UpdatePaperDto{
				Title: "Updated Paper",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Paper not found",
			paperID: "999",
			paperDto: dtos.UpdatePaperDto{
				Title: "Updated Paper",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty request body",
			paperID:        strconv.Itoa(paper.ID),
			paperDto:       dtos.UpdatePaperDto{},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.paperDto)
			req := httptest.NewRequest("PATCH", "/papers/"+tt.paperID, bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"id": tt.paperID})
			w := httptest.NewRecorder()

			UpdatePaper(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
