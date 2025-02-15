package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/models"
)

var DB *gorm.DB

func init() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"))

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal(err)
	}

	// Only run auto-migrations in development environment
	if os.Getenv("GO_ENV") == constants.GO_ENV_DEV || os.Getenv("GO_ENV") == constants.GO_ENV_DOCKER_DEV {
		autoMigrate()
	}
}

func autoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Exam{},
		&models.Answer{},
		&models.ExamParticipant{},
		&models.Paper{},
		&models.PaperOwnership{},
		&models.Question{},
		&models.Session{},
		&models.Otp{},
	)

	if err != nil {
		log.Fatal("Failed to auto-migrate database:", err)
		os.Exit(1)
	}
}
