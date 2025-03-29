package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/server/internal/config/env"
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
	DB, err = gorm.Open(postgres.Open(dsn), config.ConfigureLogger(env.GO_ENV))

	if err != nil {
		log.Fatal(err)
	}

	// Only run auto-migrations in development environment
	if env.GO_ENV == constants.GO_ENV_DEV || env.GO_ENV == constants.GO_ENV_DOCKER_DEV {
		autoMigrate()
	}
}

func autoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Exam{},
		&models.Answer{},
		&models.ExamParticipant{},
		&models.Otp{},
	)

	if err != nil {
		log.Fatal("Failed to auto-migrate database:", err)
		os.Exit(1)
	}
}
