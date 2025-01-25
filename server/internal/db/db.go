package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/arpansaha13/pariksha/internal/models"
)

var DB *gorm.DB

/* Make sure env variables are loaded before initializing db */
func Init() {
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
		os.Exit(1)
	}

	// Only run auto-migrations in development environment
	if os.Getenv("GO_ENV") == "development" {
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
		&models.UnverifiedUser{},
	)

	if err != nil {
		log.Fatal("Failed to auto-migrate database:", err)
		os.Exit(1)
	}
}
