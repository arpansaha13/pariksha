package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/arpansaha13/auth/internal/models"
	"github.com/arpansaha13/common/pkg/constants"
)

var Sessions *gorm.DB

func init() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("SESSIONS_DB_HOST"),
		os.Getenv("SESSIONS_DB_USER"),
		os.Getenv("SESSIONS_DB_PASS"),
		os.Getenv("SESSIONS_DB_NAME"),
		os.Getenv("SESSIONS_DB_PORT"),
		os.Getenv("SESSIONS_DB_SSLMODE"))

	var err error
	Sessions, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal(err)
	}

	// Only run auto-migrations in development environment
	if os.Getenv("GO_ENV") == constants.GO_ENV_DEV || os.Getenv("GO_ENV") == constants.GO_ENV_DOCKER_DEV {
		autoMigrateSessionsDb()
	}
}

func autoMigrateSessionsDb() {
	err := Sessions.AutoMigrate(
		&models.Session{},
	)

	if err != nil {
		log.Fatal("Failed to auto-migrate database:", err)
		os.Exit(1)
	}
}
