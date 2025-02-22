package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/arpansaha13/pariksha/internal/config/db"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/models"
	testUtils "github.com/arpansaha13/pariksha/internal/utils/test"
)

func TestGetUser(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	user := testUtils.CreateTestUser(t, &models.User{
		Username:  "testUser",
		FirstName: sql.NullString{String: "Test", Valid: true},
		LastName:  sql.NullString{String: "User", Valid: true},
	})

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedUser   models.User
	}{
		{
			name:           "Success",
			userID:         strconv.Itoa(user.ID),
			expectedStatus: http.StatusOK,
			expectedUser:   user,
		},
		{
			name:           "User not found",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/user/"+tt.userID, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.userID})
			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/user/{id}", GetUser).Methods("GET")
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var fetchedUser models.User
				json.NewDecoder(rr.Body).Decode(&fetchedUser)
				assert.Equal(t, tt.expectedUser.Username, fetchedUser.Username)
				assert.Equal(t, tt.expectedUser.FirstName.String, fetchedUser.FirstName.String)
				assert.Equal(t, tt.expectedUser.LastName.String, fetchedUser.LastName.String)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	user := testUtils.CreateTestUser(t, &models.User{
		Email:     "test@example.com",
		Username:  "testUser",
		FirstName: sql.NullString{String: "Test", Valid: true},
		LastName:  sql.NullString{String: "User", Valid: true},
	})

	existingUser := testUtils.CreateTestUser(t, &models.User{
		Email:    "existing@example.com",
		Username: "existingUser",
	})

	updatedUsername := "updatedUser"

	tests := []struct {
		name           string
		userID         string
		userDto        dtos.UpdateUserDto
		expectedStatus int
		validateFunc   func(t *testing.T, userID int, response map[string]interface{})
	}{
		{
			name:   "Success - Unique username",
			userID: strconv.Itoa(user.ID),
			userDto: dtos.UpdateUserDto{
				Username:  updatedUsername,
				FirstName: "Updated",
				LastName:  "User",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, userID int, response map[string]interface{}) {
				var updatedUser models.User
				db.DB.Take(&updatedUser, userID)
				assert.Equal(t, updatedUsername, updatedUser.Username)
				assert.Equal(t, "Updated", updatedUser.FirstName.String)
				assert.Equal(t, "User", updatedUser.LastName.String)

				updatedFields := response["updated_fields"].(map[string]interface{})
				notUpdatedFields := response["not_updated_fields"].(map[string]interface{})
				assert.Equal(t, updatedUsername, updatedFields["username"])
				assert.Equal(t, "Updated", updatedFields["first_name"])
				assert.Equal(t, "User", updatedFields["last_name"])
				assert.Empty(t, notUpdatedFields)
			},
		},
		{
			name:   "Username already taken",
			userID: strconv.Itoa(user.ID),
			userDto: dtos.UpdateUserDto{
				Username:  existingUser.Username,
				FirstName: "UpdatedAgain",
				LastName:  "UserAgain",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, userID int, response map[string]interface{}) {
				var updatedUser models.User
				db.DB.Take(&updatedUser, userID)
				assert.Equal(t, updatedUsername, updatedUser.Username) // Username should stay unchanged
				assert.Equal(t, "UpdatedAgain", updatedUser.FirstName.String)
				assert.Equal(t, "UserAgain", updatedUser.LastName.String)

				updatedFields := response["updated_fields"].(map[string]interface{})
				notUpdatedFields := response["not_updated_fields"].(map[string]interface{})
				assert.Equal(t, "UpdatedAgain", updatedFields["first_name"])
				assert.Equal(t, "UserAgain", updatedFields["last_name"])
				assert.NotZero(t, notUpdatedFields["Username"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.userDto)
			req, _ := http.NewRequest("PATCH", "/user/"+tt.userID, bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"id": tt.userID})
			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/user/{id}", UpdateUser).Methods("PATCH")
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.validateFunc != nil {
				var response map[string]interface{}
				json.NewDecoder(rr.Body).Decode(&response)
				tt.validateFunc(t, user.ID, response)
			}
		})
	}
}
