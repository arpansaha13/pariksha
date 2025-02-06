package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	unverifiedUser := testUtils.CreateUnverifiedUser(t, &models.User{})

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
				Email:    unverifiedUser.Email,
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
	unverifiedUser := testUtils.CreateUnverifiedUser(t, &models.User{})

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
				Email:    unverifiedUser.Email,
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
				var unverifiedUser models.Otp
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

	// Create test user and OTP entry
	email := "verify@example.com"
	user := testUtils.CreateTestUser(t, &models.User{
		Email:    email,
		Verified: false,
	})

	validOtp := models.Otp{
		Email:        email,
		OTP:          "123456",
		OTPExpiresAt: time.Now().Add(15 * time.Minute),
		Purpose:      constants.OTP_PURPOSE_SIGNUP,
	}

	expiredOtp := models.Otp{
		Email:        "expired@example.com",
		OTP:          "123457",
		OTPExpiresAt: time.Now().Add(-10 * time.Minute),
		Purpose:      constants.OTP_PURPOSE_SIGNUP,
	}

	db.DB.Create(&validOtp)
	db.DB.Create(&expiredOtp)

	tests := []struct {
		name           string
		verifyDto      dtos.VerificationDto
		expectedStatus int
		validateFunc   func(t *testing.T)
	}{
		{
			name: "Success - Signup verification",
			verifyDto: dtos.VerificationDto{
				Email: email,
				OTP:   validOtp.OTP,
			},
			expectedStatus: http.StatusNoContent,
			validateFunc: func(t *testing.T) {
				var updatedUser models.User
				db.DB.First(&updatedUser, user.ID)
				assert.True(t, updatedUser.Verified)

				var otpEntry models.Otp
				result := db.DB.Where("email = ?", email).First(&otpEntry)
				assert.Error(t, result.Error) // OTP entry should be deleted
			},
		},
		{
			name: "Invalid OTP",
			verifyDto: dtos.VerificationDto{
				Email: email,
				OTP:   "wrong123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Expired OTP",
			verifyDto: dtos.VerificationDto{
				Email: "expired@example.com",
				OTP:   expiredOtp.OTP,
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid email",
			verifyDto: dtos.VerificationDto{
				Email: "nonexistent@example.com",
				OTP:   "123456",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.verifyDto)
			req := httptest.NewRequest("POST", "/auth/verification", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			Verification(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t)
			}
		})
	}
}
