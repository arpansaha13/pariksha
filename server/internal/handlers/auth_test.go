package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/arpansaha13/common/pkg/constants"
	"github.com/arpansaha13/common/pkg/models"
	"github.com/arpansaha13/common/pkg/utils"
	"github.com/arpansaha13/pariksha/internal/config/db"
	"github.com/arpansaha13/pariksha/internal/config/env"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/middlewares"
	testUtils "github.com/arpansaha13/pariksha/internal/utils/test"
)

func TestLogin(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	testUser := testUtils.CreateTestUser(t, &models.User{
		Verified: true,
	})
	unverifiedUser := testUtils.CreateTestUser(t, &models.User{
		Email:    "unverified@example.com",
		Verified: false,
	})

	tests := []struct {
		name           string
		loginDto       dtos.LoginWithPasswordDto
		expectedStatus int
	}{
		{
			name: "Successful login",
			loginDto: dtos.LoginWithPasswordDto{
				Email:    testUser.Email,
				Password: "testPass123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid password",
			loginDto: dtos.LoginWithPasswordDto{
				Email:    testUser.Email,
				Password: "wrongPass",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Non-existent user",
			loginDto: dtos.LoginWithPasswordDto{
				Email:    "nonexistent@example.com",
				Password: "testPass123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Guest user",
			loginDto: dtos.LoginWithPasswordDto{
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

			LoginWithPassword(w, req)

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

	existingUser := testUtils.CreateTestUser(t, &models.User{
		Verified: true,
	})
	unverifiedUser := testUtils.CreateTestUser(t, &models.User{
		Email:    "unverified@example.com",
		Verified: false,
	})

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
			name: "Successful signup for unverified user",
			signUpDto: dtos.SignUpDto{
				Email:    unverifiedUser.Email,
				Password: "pass123",
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Email already exists but unverified",
			signUpDto: dtos.SignUpDto{
				Email:    unverifiedUser.Email,
				Password: "existingPass123",
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

func TestVerifySignup(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	unverifiedEmail := "unverified@example.com"
	unverifiedUser := testUtils.CreateTestUser(t, &models.User{
		Email:    unverifiedEmail,
		Verified: false,
	})

	validSignupOtp := models.Otp{
		Email:        unverifiedEmail,
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

	db.DB.Create(&validSignupOtp)
	db.DB.Create(&expiredOtp)

	tests := []struct {
		name           string
		verifyDto      dtos.VerificationDto
		expectedStatus int
		validateFunc   func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Signup verification",
			verifyDto: dtos.VerificationDto{
				Email: unverifiedEmail,
				OTP:   validSignupOtp.OTP,
			},
			expectedStatus: http.StatusNoContent,
			validateFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				var updatedUser models.User
				db.DB.First(&updatedUser, unverifiedUser.ID)
				assert.True(t, updatedUser.Verified)

				var otpEntry models.Otp
				result := db.DB.Where("email = ?", unverifiedEmail).First(&otpEntry)
				assert.Error(t, result.Error) // OTP entry should be deleted
			},
		},
		{
			name: "Invalid OTP",
			verifyDto: dtos.VerificationDto{
				Email: unverifiedEmail,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.verifyDto)
			req := httptest.NewRequest("POST", "/auth/verification/signup", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			VerifySignup(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t, w)
			}
		})
	}
}

func TestVerifyLogin(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	verifiedEmail := "verified@example.com"
	unverifiedEmail := "unverified@example.com"

	verifiedUser := testUtils.CreateTestUser(t, &models.User{
		Email:    verifiedEmail,
		Verified: true,
	})

	testUtils.CreateTestUser(t, &models.User{
		Email:    unverifiedEmail,
		Verified: false,
	})

	validLoginOtp := models.Otp{
		Email:        verifiedEmail,
		OTP:          "654321",
		OTPExpiresAt: time.Now().Add(15 * time.Minute),
		Purpose:      constants.OTP_PURPOSE_LOGIN,
	}

	unverifiedLoginOtp := models.Otp{
		Email:        unverifiedEmail,
		OTP:          "111111",
		OTPExpiresAt: time.Now().Add(15 * time.Minute),
		Purpose:      constants.OTP_PURPOSE_LOGIN,
	}

	db.DB.Create(&validLoginOtp)
	db.DB.Create(&unverifiedLoginOtp)

	tests := []struct {
		name           string
		verifyDto      dtos.VerificationDto
		expectedStatus int
		validateFunc   func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Login verification",
			verifyDto: dtos.VerificationDto{
				Email: verifiedEmail,
				OTP:   validLoginOtp.OTP,
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Header().Get("Set-Cookie"), "token=")

				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, verifiedUser.Email, response["email"])
				assert.Equal(t, verifiedUser.Username, response["username"])

				var otpEntry models.Otp
				result := db.DB.Where("email = ?", verifiedEmail).First(&otpEntry)
				assert.Error(t, result.Error)

				cookie := w.Result().Cookies()[0]
				var session models.Session
				result = db.DB.Where("key = ?", cookie.Value).First(&session)
				assert.NoError(t, result.Error)
			},
		},
		{
			name: "Failed - Unverified user attempting login",
			verifyDto: dtos.VerificationDto{
				Email: unverifiedEmail,
				OTP:   unverifiedLoginOtp.OTP,
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid OTP",
			verifyDto: dtos.VerificationDto{
				Email: verifiedEmail,
				OTP:   "wrong123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.verifyDto)
			req := httptest.NewRequest("POST", "/auth/verification/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			VerifyLoginWithOtp(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t, w)
			}
		})
	}
}

func TestLoginWithOtp(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	verifiedUser := testUtils.CreateTestUser(t, &models.User{
		Email:    "verified@example.com",
		Verified: true,
	})

	tests := []struct {
		name           string
		loginOtpDto    dtos.LoginWithOtpDto
		expectedStatus int
		validateFunc   func(t *testing.T)
	}{
		{
			name: "Success - OTP created and sent",
			loginOtpDto: dtos.LoginWithOtpDto{
				Email: verifiedUser.Email,
			},
			expectedStatus: http.StatusNoContent,
			validateFunc: func(t *testing.T) {
				var otpEntry models.Otp
				result := db.DB.Where("email = ? AND purpose = ?",
					verifiedUser.Email,
					constants.OTP_PURPOSE_LOGIN,
				).First(&otpEntry)

				assert.NoError(t, result.Error)
				assert.Equal(t, constants.OTP_PURPOSE_LOGIN, otpEntry.Purpose)
				assert.NotEmpty(t, otpEntry.OTP)
				assert.True(t, otpEntry.OTPExpiresAt.After(time.Now()))
			},
		},
		{
			name: "Invalid email format",
			loginOtpDto: dtos.LoginWithOtpDto{
				Email: "invalid-email",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty email",
			loginOtpDto: dtos.LoginWithOtpDto{
				Email: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.loginOtpDto)
			req := httptest.NewRequest("POST", "/auth/login/otp", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			LoginWithOtp(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t)
			}
		})
	}
}

func TestForgotPassword(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	verifiedUser := testUtils.CreateTestUser(t, &models.User{
		Email:    "verified@example.com",
		Verified: true,
	})
	unverifiedUser := testUtils.CreateTestUser(t, &models.User{
		Email:    "unverified@example.com",
		Verified: false,
	})

	tests := []struct {
		name              string
		forgotPasswordDto dtos.ForgotPasswordDto
		expectedStatus    int
		validateFunc      func(t *testing.T)
	}{
		{
			name: "Success - OTP created and sent",
			forgotPasswordDto: dtos.ForgotPasswordDto{
				Email: verifiedUser.Email,
			},
			expectedStatus: http.StatusNoContent,
			validateFunc: func(t *testing.T) {
				var otpEntry models.Otp
				result := db.DB.Where("email = ? AND purpose = ?",
					verifiedUser.Email,
					constants.OTP_PURPOSE_FORGOT_PASSWORD,
				).First(&otpEntry)

				assert.NoError(t, result.Error)
				assert.Equal(t, constants.OTP_PURPOSE_FORGOT_PASSWORD, otpEntry.Purpose)
				assert.NotEmpty(t, otpEntry.OTP)
				assert.True(t, otpEntry.OTPExpiresAt.After(time.Now()))
			},
		},
		{
			name: "Email not found or not verified",
			forgotPasswordDto: dtos.ForgotPasswordDto{
				Email: unverifiedUser.Email,
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Invalid email format",
			forgotPasswordDto: dtos.ForgotPasswordDto{
				Email: "invalid-email",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty email",
			forgotPasswordDto: dtos.ForgotPasswordDto{
				Email: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.forgotPasswordDto)
			req := httptest.NewRequest("POST", "/auth/forgot-password", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			ForgotPassword(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t)
			}
		})
	}
}

func TestResetPassword(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	verifiedUser := testUtils.CreateTestUser(t, &models.User{
		Email:    "verified@example.com",
		Verified: true,
	})

	unverifiedUser := testUtils.CreateTestUser(t, &models.User{
		Email:    "unverified@example.com",
		Verified: false,
	})

	validOtp := models.Otp{
		Email:        verifiedUser.Email,
		OTP:          "123456",
		OTPExpiresAt: time.Now().Add(15 * time.Minute),
		Purpose:      constants.OTP_PURPOSE_FORGOT_PASSWORD,
	}

	expiredOtp := models.Otp{
		Email:        verifiedUser.Email,
		OTP:          "expiredOtp",
		OTPExpiresAt: time.Now().Add(-15 * time.Minute),
		Purpose:      constants.OTP_PURPOSE_FORGOT_PASSWORD,
	}

	db.DB.Create(&validOtp)
	db.DB.Create(&expiredOtp)

	tests := []struct {
		name             string
		resetPasswordDto dtos.ResetPasswordDto
		expectedStatus   int
		validateFunc     func(t *testing.T)
	}{
		{
			name: "Success - Password reset",
			resetPasswordDto: dtos.ResetPasswordDto{
				Email:       verifiedUser.Email,
				OldPassword: "oldPassword",
				NewPassword: "newPassword123",
				OTP:         validOtp.OTP,
			},
			expectedStatus: http.StatusNoContent,
			validateFunc: func(t *testing.T) {
				var user models.User
				err := db.DB.Where("email = ?", verifiedUser.Email).Take(&user).Error
				assert.NoError(t, err)

				err = bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte("newPassword123"))
				assert.NoError(t, err)

				var otpEntry models.Otp
				err = db.DB.Where("email = ?", verifiedUser.Email).Take(&otpEntry).Error
				assert.Error(t, err) // OTP entry should be deleted
			},
		},
		{
			name: "Invalid OTP",
			resetPasswordDto: dtos.ResetPasswordDto{
				Email:       verifiedUser.Email,
				OldPassword: "oldPassword",
				NewPassword: "newPassword123",
				OTP:         "wrongOtp",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid old password",
			resetPasswordDto: dtos.ResetPasswordDto{
				Email:       verifiedUser.Email,
				OldPassword: "wrongOldPassword",
				NewPassword: "newPassword123",
				OTP:         validOtp.OTP,
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid email",
			resetPasswordDto: dtos.ResetPasswordDto{
				Email:       "invalid@example.com",
				OldPassword: "oldPassword",
				NewPassword: "newPassword123",
				OTP:         validOtp.OTP,
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Expired OTP",
			resetPasswordDto: dtos.ResetPasswordDto{
				Email:       verifiedUser.Email,
				OldPassword: "oldPassword",
				NewPassword: "newPassword123",
				OTP:         expiredOtp.OTP,
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Unverified user",
			resetPasswordDto: dtos.ResetPasswordDto{
				Email:       unverifiedUser.Email,
				OldPassword: "oldPassword",
				NewPassword: "newPassword123",
				OTP:         validOtp.OTP,
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.resetPasswordDto)
			req := httptest.NewRequest("POST", "/auth/reset-password", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			ResetPassword(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateFunc != nil {
				tt.validateFunc(t)
			}
		})
	}
}

func TestCsrfMiddleware(t *testing.T) {
	testUtils.SetupTestDB(t)

	t.Cleanup(func() {
		testUtils.TeardownTestDB(t)
	})

	// Create a test user and session
	testUtils.CreateTestUser(t, &models.User{
		Verified: true,
	})
	sessionKey := uuid.New()
	csrfToken, _ := utils.GenerateBase64String(32)
	session := models.Session{
		Key:       sessionKey,
		Token:     "testToken",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CsrfToken: csrfToken,
	}
	db.DB.Create(&session)

	tests := []struct {
		name           string
		method         string
		csrfToken      string
		expectedStatus int
	}{
		{
			name:           "Valid CSRF token",
			method:         "POST",
			csrfToken:      csrfToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing CSRF token",
			method:         "POST",
			csrfToken:      "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid CSRF token",
			method:         "POST",
			csrfToken:      "invalidToken",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Safe method without CSRF token",
			method:         "GET",
			csrfToken:      "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/protected", nil)
			req.AddCookie(&http.Cookie{
				Name:  utils.GetEnvWithDefault("SESSION_COOKIE_NAME", env.SESSION_COOKIE_NAME),
				Value: sessionKey.String(),
			})
			if tt.csrfToken != "" {
				req.Header.Set("X-CSRFToken", tt.csrfToken)
			}
			w := httptest.NewRecorder()

			handler := middlewares.CsrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
