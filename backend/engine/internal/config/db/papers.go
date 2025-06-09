package db

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/engine/internal/config/env"
)

var Papers *gorm.DB

// InitDB initializes the papers database connection with given parameters
func InitDB(host, port, user, password, dbname, sslmode string) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	var err error
	Papers, err = gorm.Open(postgres.Open(dsn), config.ConfigureLogger(env.GO_ENV))
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

func autoMigrateDB() error {
	err := Papers.AutoMigrate(
		&models.Paper{},
		&models.Question{},
		&models.QuestionCategory{},
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
		env.PAPERS_DB_HOST,
		env.PAPERS_DB_PORT,
		env.PAPERS_DB_USER,
		env.PAPERS_DB_PASS,
		env.PAPERS_DB_NAME,
		env.PAPERS_DB_SSLMODE,
	)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
}
