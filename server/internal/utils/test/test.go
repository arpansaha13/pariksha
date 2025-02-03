package test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/arpansaha13/pariksha/internal/config/validate"
	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/middlewares"
	"github.com/arpansaha13/pariksha/internal/models"
)

func SetupTestDB(t *testing.T) {
	if godotenv.Load("../../.env") != nil {
		log.Fatalf("Error loading .env file")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("TEST_DB_HOST"),
		os.Getenv("TEST_DB_USER"),
		os.Getenv("TEST_DB_PASS"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("TEST_DB_PORT"))

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

	validate.Init()

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

func CreateTestUser(t *testing.T, data *models.User) models.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testPass123"), bcrypt.DefaultCost)

	user := models.User{
		Email:    data.Email,
		Password: sql.NullString{String: string(hashedPassword), Valid: true},
	}

	if data.Email == "" {
		user.Email = "test@example.com"
	}

	user.Username = strings.Split(user.Email, "@")[0]

	if err := db.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

func CreateGuestUser(t *testing.T, data *models.User) models.User {
	guestUser := models.User{
		Email:   data.Email,
		IsGuest: true,
	}

	if data.Email == "" {
		guestUser.Email = "guest@example.com"
	}

	guestUser.Username = strings.Split(guestUser.Email, "@")[0]

	if err := db.DB.Create(&guestUser).Error; err != nil {
		t.Fatalf("Failed to create guest user: %v", err)
	}

	return guestUser
}

func CreateTestPaper(t *testing.T, data *models.Paper) models.Paper {
	paper := models.Paper{
		Title:           data.Title,
		MaxScore:        data.MaxScore,
		QuestionCounts:  data.QuestionCounts,
		DurationMinutes: 60,
	}

	if paper.Title == "" {
		paper.Title = "Test Paper"
	}
	if err := db.DB.Create(&paper).Error; err != nil {
		t.Fatalf("Failed to create test paper: %v", err)
	}
	return paper
}

func CreateTestPaperOwnership(t *testing.T, data *models.PaperOwnership) models.PaperOwnership {
	ownership := models.PaperOwnership{
		UserID:  data.UserID,
		PaperID: data.PaperID,
		Type:    data.Type,
	}
	if ownership.Type == "" {
		ownership.Type = constants.PAPER_OWNERSHIP_TYPE_OWNER
	}
	if err := db.DB.Create(&ownership).Error; err != nil {
		t.Fatalf("Failed to create test paper: %v", err)
	}
	return ownership
}

func CreateTestExam(t *testing.T, data *models.Exam) models.Exam {
	exam := models.Exam{
		Title:              data.Title,
		StartsAt:           data.StartsAt,
		EndsAt:             data.EndsAt,
		CreatedBy:          data.CreatedBy,
		PaperID:            data.PaperID,
		Type:               data.Type,
		ParticipantCounts:  data.ParticipantCounts,
		MaxCandidatesCount: data.MaxCandidatesCount,
	}

	if exam.Title == "" {
		exam.Title = "Test Exam"
	}

	if exam.Type == "" {
		exam.Type = constants.EXAM_TYPE_OPEN
	}

	if exam.MaxCandidatesCount == 0 {
		exam.MaxCandidatesCount = 10
	}

	if err := db.DB.Create(&exam).Error; err != nil {
		t.Fatalf("Failed to create test exam: %v", err)
	}

	return exam
}

func CreateTestExamParticipant(t *testing.T, data *models.ExamParticipant) models.ExamParticipant {
	participant := models.ExamParticipant{
		ExamID: data.ExamID,
		UserID: data.UserID,
		Status: data.Status,
	}

	if err := db.DB.Create(&participant).Error; err != nil {
		t.Fatalf("Failed to create test exam: %v", err)
	}

	return participant
}

func CreateTestQuestion(t *testing.T, data *models.Question) models.Question {
	var questionData json.RawMessage
	if data.Type == constants.QUESTION_TYPE_MCQ {
		questionData = json.RawMessage(`{
			"statement": "Test MCQ",
			"options": ["A", "B", "C", "D"]
		}`)
	} else {
		questionData = json.RawMessage(`{
			"statement": "Test Question"
		}`)
	}

	question := models.Question{
		PaperID:  data.PaperID,
		Question: questionData,
		Type:     data.Type,
		MaxScore: data.MaxScore,
		Tags:     json.RawMessage(`["test"]`),
	}

	if err := db.DB.Create(&question).Error; err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	return question
}
