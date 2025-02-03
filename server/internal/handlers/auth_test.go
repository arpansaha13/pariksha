package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestLogin(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	testUser := testUtils.CreateTestUser(t, &models.User{})
	guestUser := testUtils.CreateGuestUser(t, &models.User{})

	tests := []struct {
		name           string
		loginDto       dtos.LoginDto
		expectedStatus int
	}{
		{
			name: "Successful login",
			loginDto: dtos.LoginDto{
				Email:    testUser.Email,
				Password: "testPass123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid password",
			loginDto: dtos.LoginDto{
				Email:    testUser.Email,
				Password: "wrongPass",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Non-existent user",
			loginDto: dtos.LoginDto{
				Email:    "nonexistent@example.com",
				Password: "testPass123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Guest user",
			loginDto: dtos.LoginDto{
				Email:    guestUser.Email,
				Password: "testPass123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.loginDto)
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, w.Header().Get("Set-Cookie"), "token=")
			}
		})
	}
}

func TestSignUp(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	existingUser := testUtils.CreateTestUser(t, &models.User{})
	guestUser := testUtils.CreateGuestUser(t, &models.User{})

	tests := []struct {
		name           string
		signUpDto      dtos.SignUpDto
		expectedStatus int
	}{
		{
			name: "Successful signup",
			signUpDto: dtos.SignUpDto{
				Email:    "new@example.com",
				Password: "newPass123",
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Email already exists",
			signUpDto: dtos.SignUpDto{
				Email:    existingUser.Email,
				Password: "existingPass123",
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Successful signup for guest user",
			signUpDto: dtos.SignUpDto{
				Email:    guestUser.Email,
				Password: "guestPass123",
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	// TODO: add test for the case when repeated signup is done for same email. it should update otp and expiry time

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.signUpDto)
			req := httptest.NewRequest("POST", "/auth/signup", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			SignUp(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusNoContent {
				var unverifiedUser models.UnverifiedUser
				result := db.DB.Where("email = ?", tt.signUpDto.Email).First(&unverifiedUser)
				assert.Nil(t, result.Error)
				assert.NotEmpty(t, unverifiedUser.OTP)
			}
		})
	}
}

func TestVerification(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	// Create unverified users
	validUnverifiedUser := models.UnverifiedUser{
		Hash:         "testHash12", // 10-character hash
		OTP:          "123456",
		Email:        "verify@example.com",
		Password:     "hashedPass",
		OTPExpiresAt: time.Now().Add(15 * time.Minute),
	}
	expiredUnverifiedUser := models.UnverifiedUser{
		Hash:         "expired123", // 10-character hash
		OTP:          "123457",
		Email:        "expired@example.com",
		Password:     "hashedPass",
		OTPExpiresAt: time.Now().Add(-10 * time.Minute),
	}

	db.DB.Create(&validUnverifiedUser)
	db.DB.Create(&expiredUnverifiedUser)

	tests := []struct {
		name           string
		hash           string
		verifyDto      dtos.VerificationDto
		expectedStatus int
	}{
		{
			name: "Invalid OTP",
			hash: validUnverifiedUser.Hash,
			verifyDto: dtos.VerificationDto{
				OTP: "wrong123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid or expired hash (wrong or deleted)",
			hash: "wrongHash",
			verifyDto: dtos.VerificationDto{
				OTP: validUnverifiedUser.OTP,
			},
			expectedStatus: constants.StatusInvalidToken,
		},
		{
			name: "Expired OTP",
			hash: expiredUnverifiedUser.Hash,
			verifyDto: dtos.VerificationDto{
				OTP: expiredUnverifiedUser.OTP,
			},
			expectedStatus: http.StatusUnauthorized,
		},
		// Note: `unverifiedUser` entry will be deleted after successful verification
		{
			name: "Successful verification",
			hash: validUnverifiedUser.Hash,
			verifyDto: dtos.VerificationDto{
				OTP: validUnverifiedUser.OTP,
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	// TODO: add test for verification of guest user

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.verifyDto)
			req := httptest.NewRequest("POST", fmt.Sprintf("/auth/verification/%s", tt.hash), bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"hash": tt.hash})
			w := httptest.NewRecorder()

			Verification(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusNoContent {
				// Verify user was created
				var user models.User
				result := db.DB.Where("email = ?", validUnverifiedUser.Email).First(&user)
				assert.Nil(t, result.Error)

				// Verify unverified user was deleted
				result = db.DB.Where("hash = ?", validUnverifiedUser.Hash).First(&models.UnverifiedUser{})
				assert.Error(t, result.Error)
			}
		})
	}
}
