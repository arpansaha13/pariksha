package db

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pariksha/auth/internal/config/env"
	"pariksha/common/pkg/config"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
)

var Sessions *gorm.DB

// InitSessionsDB initializes the sessions database connection with given parameters
func InitSessionsDB(host, port, user, password, dbname, sslmode string) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	var err error
	Sessions, err = gorm.Open(postgres.Open(dsn), config.ConfigureLogger(env.GO_ENV))

	if err != nil {
		return fmt.Errorf("failed to connect to sessions database: %v", err)
	}

	// Only run auto-migrations in development environment
	if env.GO_ENV == constants.GO_ENV_DEV || env.GO_ENV == constants.GO_ENV_DOCKER_DEV || env.GO_ENV == constants.GO_ENV_TEST {
		if err := autoMigrateSessionsDB(); err != nil {
			return err
		}
	}

	return nil
}

// AutoMigrateSessionsDB runs migrations for the sessions database
func autoMigrateSessionsDB() error {
	err := Sessions.AutoMigrate(
		&models.Session{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate sessions database: %v", err)
	}
	return nil
}

func init() {
	// Skip initialization if in test environment
	if env.GO_ENV == constants.GO_ENV_TEST {
		return
	}

	// Initialize with environment variables for non-test environments
	err := InitSessionsDB(
		env.SESSIONS_DB_HOST,
		env.SESSIONS_DB_PORT,
		env.SESSIONS_DB_USER,
		env.SESSIONS_DB_PASS,
		env.SESSIONS_DB_NAME,
		env.SESSIONS_DB_SSLMODE,
	)
	if err != nil {
		log.Fatal("Failed to initialize sessions database:", err)
	}
}
