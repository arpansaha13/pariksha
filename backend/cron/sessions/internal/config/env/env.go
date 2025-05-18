package env

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/utils"
)

var (
	GO_ENV              string
	CRON_INTERVAL_HOURS int
)

var (
	USERS_DB_HOST    string
	USERS_DB_PORT    string
	USERS_DB_USER    string
	USERS_DB_PASS    string
	USERS_DB_NAME    string
	USERS_DB_SSLMODE string
)

func init() {
	if os.Getenv("GO_ENV") == "" {
		if godotenv.Load("../../test.env") != nil {
			log.Fatalf("Error loading .env file")
		}
	}

	requiredEnvVars := getRequiredEnvVars()

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s is not set", envVar)
		}
	}

	GO_ENV = os.Getenv("GO_ENV")

	USERS_DB_USER = os.Getenv("USERS_DB_USER")
	USERS_DB_PASS = os.Getenv("USERS_DB_PASS")
	USERS_DB_NAME = os.Getenv("USERS_DB_NAME")
	CRON_INTERVAL_HOURS, _ = strconv.Atoi(utils.GetEnvWithDefault("CRON_INTERVAL_HOURS", "1"))

	if GO_ENV != constants.GO_ENV_TEST {
		USERS_DB_HOST = os.Getenv("USERS_DB_HOST")
		USERS_DB_PORT = os.Getenv("USERS_DB_PORT")
		USERS_DB_SSLMODE = os.Getenv("USERS_DB_SSLMODE")
	}
}

func getRequiredEnvVars() []string {
	baseEnvVars := []string{
		"GO_ENV",
		"USERS_DB_USER",
		"USERS_DB_PASS",
		"USERS_DB_NAME",
		"CRON_INTERVAL_HOURS",
	}

	if os.Getenv("GO_ENV") != constants.GO_ENV_TEST {
		// require these vars only when not in test environment
		additionalEnvVars := []string{
			"USERS_DB_HOST",
			"USERS_DB_PORT",
			"USERS_DB_SSLMODE",
		}
		baseEnvVars = append(baseEnvVars, additionalEnvVars...)
	}

	return baseEnvVars
}
