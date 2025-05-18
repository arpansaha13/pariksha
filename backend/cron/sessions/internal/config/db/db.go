package db

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/cron/sessions/internal/config/env"
)

var DB *gorm.DB

// InitDB initializes the main database connection with given parameters
func InitDB(host, port, user, password, dbname, sslmode string) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), config.ConfigureLogger(env.GO_ENV))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Only run auto-migrations in development and test environments
	if env.GO_ENV == constants.GO_ENV_DEV || env.GO_ENV == constants.GO_ENV_DOCKER_DEV || env.GO_ENV == constants.GO_ENV_TEST {
		if err := autoMigrateDB(); err != nil {
			return err
		}
	}

	return nil
}

// AutoMigrateDB runs migrations for the main database
func autoMigrateDB() error {
	err := DB.AutoMigrate(
		&models.Session{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate database: %v", err)
	}
	return nil
}

func init() {
	// Skip initialization if in test environment
	if env.GO_ENV == constants.GO_ENV_TEST {
		return
	}

	// Initialize with environment variables for non-test environments
	err := InitDB(
		env.USERS_DB_HOST,
		env.USERS_DB_PORT,
		env.USERS_DB_USER,
		env.USERS_DB_PASS,
		env.USERS_DB_NAME,
		env.USERS_DB_SSLMODE,
	)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
}
