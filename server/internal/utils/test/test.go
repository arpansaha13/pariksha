package test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/middlewares"
	"github.com/arpansaha13/pariksha/internal/models"
)

func SetupTestDB(t *testing.T) {
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
		&models.Exam{},
		&models.Answer{},
		&models.ExamParticipant{},
		&models.Paper{},
		&models.PaperOwnership{},
		&models.Question{},
		&models.Session{},
		&models.UnverifiedUser{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
}

func TeardownTestDB(t *testing.T) {
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

func SetUserContext(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, middlewares.UserIDKey, userID)
}

func CreateTestUser(t *testing.T) models.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testPass123"), bcrypt.DefaultCost)

	user := models.User{
		Email:    "test@example.com",
		Password: sql.NullString{String: string(hashedPassword), Valid: true},
		Username: "testUser",
	}

	if err := db.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

func CreateGuestUser(t *testing.T) models.User {
	guestUser := models.User{
		Email:   "guest@example.com",
		IsGuest: true,
	}

	if err := db.DB.Create(&guestUser).Error; err != nil {
		t.Fatalf("Failed to create guest user: %v", err)
	}

	return guestUser
}

func CreateTestPaper(t *testing.T) models.Paper {
	paper := models.Paper{
		Title:           "Test Paper",
		DurationMinutes: 60,
	}
	if err := db.DB.Create(&paper).Error; err != nil {
		t.Fatalf("Failed to create test paper: %v", err)
	}
	return paper
}

func CreateTestExam(t *testing.T, userID int, paperID int) models.Exam {
	exam := models.Exam{
		Title:     "Test Exam",
		CreatedBy: userID,
		PaperID:   paperID,
	}

	if err := db.DB.Create(&exam).Error; err != nil {
		t.Fatalf("Failed to create test exam: %v", err)
	}

	return exam
}

func CreateTestExamParticipant(t *testing.T, userID int, examID int, status int) models.ExamParticipant {
	participant := models.ExamParticipant{
		ExamID: examID,
		UserID: userID,
		Status: status,
	}

	if err := db.DB.Create(&participant).Error; err != nil {
		t.Fatalf("Failed to create test exam: %v", err)
	}

	return participant
}
