package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/models"
)

func setupTestDB(t *testing.T) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		"localhost",
		"postgres",
		"postgres",
		"pariksha_test",
		"5433",
	)

	silentLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
		},
	)

	var err error
	db.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: silentLogger,
	})

	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.DB.AutoMigrate(
		&models.User{},
		&models.UnverifiedUser{},
		&models.Session{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
}

func teardownTestDB(t *testing.T) {
	sqlDB, err := db.DB.DB()
	if err != nil {
		t.Errorf("Failed to get underlying DB: %v", err)
		return
	}

	// Drop all tables
	db.DB.Exec("DROP SCHEMA public CASCADE;")
	db.DB.Exec("CREATE SCHEMA public;")

	sqlDB.Close()
}

func TestLogin(t *testing.T) {
	setupTestDB(t)

	t.Cleanup(func() {
		teardownTestDB(t)
	})

	// Create users
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testPass123"), bcrypt.DefaultCost)
	testUser := models.User{
		Email:    "test@example.com",
		Password: sql.NullString{String: string(hashedPassword), Valid: true},
		Username: "testUser",
	}
	guestUser := models.User{
		Email:   "guest@example.com",
		IsGuest: true,
	}

	db.DB.Create(&testUser)
	db.DB.Create(&guestUser)

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
	setupTestDB(t)

	t.Cleanup(func() {
		teardownTestDB(t)
	})

	// Create users
	existingUser := models.User{
		Email:    "existing@example.com",
		Password: sql.NullString{String: "expired123", Valid: true},
	}
	guestUser := models.User{
		Email:   "guest@example.com",
		IsGuest: true,
	}

	db.DB.Create(&existingUser)
	db.DB.Create(&guestUser)

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
	setupTestDB(t)

	t.Cleanup(func() {
		teardownTestDB(t)
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
