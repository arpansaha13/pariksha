package db

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/constants"
	"pariksha/workers/exam/internal/config/env"
)

var Papers *gorm.DB

// InitDB initializes the main database connection with given parameters
func InitPapersDB(host, port, user, password, dbname, sslmode string) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	var err error
	Papers, err = gorm.Open(postgres.Open(dsn), config.GormLogger(env.GO_ENV))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	return nil
}

func init() {
	// Skip initialization if in test environment
	if env.GO_ENV == constants.GO_ENV_TEST {
		return
	}

	// Initialize with environment variables for non-test environments
	err := InitPapersDB(
		env.PAPER_DB_HOST,
		env.PAPER_DB_PORT,
		env.PAPER_DB_USER,
		env.PAPER_DB_PASS,
		env.PAPER_DB_NAME,
		env.PAPER_DB_SSLMODE,
	)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
}
