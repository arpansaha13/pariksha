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
	SESSIONS_DB_HOST    string
	SESSIONS_DB_PORT    string
	SESSIONS_DB_USER    string
	SESSIONS_DB_PASS    string
	SESSIONS_DB_NAME    string
	SESSIONS_DB_SSLMODE string
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

	SESSIONS_DB_USER = os.Getenv("SESSIONS_DB_USER")
	SESSIONS_DB_PASS = os.Getenv("SESSIONS_DB_PASS")
	SESSIONS_DB_NAME = os.Getenv("SESSIONS_DB_NAME")
	CRON_INTERVAL_HOURS, _ = strconv.Atoi(utils.GetEnvWithDefault("CRON_INTERVAL_HOURS", "1"))

	if GO_ENV != constants.GO_ENV_TEST {
		SESSIONS_DB_HOST = os.Getenv("SESSIONS_DB_HOST")
		SESSIONS_DB_PORT = os.Getenv("SESSIONS_DB_PORT")
		SESSIONS_DB_SSLMODE = os.Getenv("SESSIONS_DB_SSLMODE")
	}
}

func getRequiredEnvVars() []string {
	baseEnvVars := []string{
		"GO_ENV",
		"SESSIONS_DB_USER",
		"SESSIONS_DB_PASS",
		"SESSIONS_DB_NAME",
		"CRON_INTERVAL_HOURS",
	}

	if os.Getenv("GO_ENV") != constants.GO_ENV_TEST {
		// require these vars only when not in test environment
		additionalEnvVars := []string{
			"SESSIONS_DB_HOST",
			"SESSIONS_DB_PORT",
			"SESSIONS_DB_SSLMODE",
		}
		baseEnvVars = append(baseEnvVars, additionalEnvVars...)
	}

	return baseEnvVars
}
