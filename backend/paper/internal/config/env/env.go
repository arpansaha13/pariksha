package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"pariksha/common/pkg/constants"
)

var (
	GO_ENV            string
	PAPER_SERVER_PORT string
	EXAM_API_TOKEN    string
)

var (
	PAPER_DB_HOST    string
	PAPER_DB_PORT    string
	PAPER_DB_USER    string
	PAPER_DB_PASS    string
	PAPER_DB_NAME    string
	PAPER_DB_SSLMODE string
)

var (
	QUESTION_SERVER_HOST string
	QUESTION_SERVER_PORT string
)

func init() {
	if os.Getenv("GO_ENV") == "" {
		if godotenv.Load("../../env/test.env") != nil {
			log.Fatalf("Error loading test.env file")
		}
	}

	requiredEnvVars := getRequiredEnvVars()

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s is not set", envVar)
		}
	}

	GO_ENV = os.Getenv("GO_ENV")

	PAPER_SERVER_PORT = os.Getenv("PAPER_SERVER_PORT")

	QUESTION_SERVER_HOST = os.Getenv("QUESTION_SERVER_HOST")
	QUESTION_SERVER_PORT = os.Getenv("QUESTION_SERVER_PORT")

	EXAM_API_TOKEN = os.Getenv("EXAM_API_TOKEN")

	PAPER_DB_USER = os.Getenv("PAPER_DB_USER")
	PAPER_DB_PASS = os.Getenv("PAPER_DB_PASS")
	PAPER_DB_NAME = os.Getenv("PAPER_DB_NAME")

	if GO_ENV != constants.GO_ENV_TEST {
		PAPER_DB_HOST = os.Getenv("PAPER_DB_HOST")
		PAPER_DB_PORT = os.Getenv("PAPER_DB_PORT")
		PAPER_DB_SSLMODE = os.Getenv("PAPER_DB_SSLMODE")
	}
}

func getRequiredEnvVars() []string {
	baseEnvVars := []string{
		"GO_ENV",
		"PAPER_DB_USER",
		"PAPER_DB_PASS",
		"PAPER_DB_NAME",
		"PAPER_SERVER_PORT",
	}

	if os.Getenv("GO_ENV") != constants.GO_ENV_TEST {
		additionalEnvVars := []string{
			"PAPER_DB_HOST",
			"PAPER_DB_PORT",
			"PAPER_DB_SSLMODE",
			"QUESTION_SERVER_HOST",
			"QUESTION_SERVER_PORT",
			"EXAM_API_TOKEN",
		}
		baseEnvVars = append(baseEnvVars, additionalEnvVars...)
	}

	return baseEnvVars
}
