package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"pariksha/common/pkg/constants"
)

var (
	GO_ENV           string
	EXAM_SERVER_PORT string
	EXAM_API_TOKEN   string
)

var (
	EXAM_DB_HOST    string
	EXAM_DB_PORT    string
	EXAM_DB_USER    string
	EXAM_DB_PASS    string
	EXAM_DB_NAME    string
	EXAM_DB_SSLMODE string
)

var (
	EXAM_QUEUE_HOST string
	EXAM_QUEUE_PORT string
)

var (
	PAPER_SERVER_HOST string
	PAPER_SERVER_PORT string
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

	EXAM_SERVER_PORT = os.Getenv("EXAM_SERVER_PORT")

	EXAM_DB_USER = os.Getenv("EXAM_DB_USER")
	EXAM_DB_PASS = os.Getenv("EXAM_DB_PASS")
	EXAM_DB_NAME = os.Getenv("EXAM_DB_NAME")

	if GO_ENV != constants.GO_ENV_TEST {
		EXAM_DB_HOST = os.Getenv("EXAM_DB_HOST")
		EXAM_DB_PORT = os.Getenv("EXAM_DB_PORT")
		EXAM_DB_SSLMODE = os.Getenv("EXAM_DB_SSLMODE")

		EXAM_API_TOKEN = os.Getenv("EXAM_API_TOKEN")

		EXAM_QUEUE_HOST = os.Getenv("EXAM_QUEUE_HOST")
		EXAM_QUEUE_PORT = os.Getenv("EXAM_QUEUE_PORT")

		PAPER_SERVER_HOST = os.Getenv("PAPER_SERVER_HOST")
		PAPER_SERVER_PORT = os.Getenv("PAPER_SERVER_PORT")

		QUESTION_SERVER_HOST = os.Getenv("QUESTION_SERVER_HOST")
		QUESTION_SERVER_PORT = os.Getenv("QUESTION_SERVER_PORT")
	}
}

func getRequiredEnvVars() []string {
	baseEnvVars := []string{
		"GO_ENV",
		"EXAM_SERVER_PORT",
		"EXAM_DB_USER",
		"EXAM_DB_PASS",
		"EXAM_DB_NAME",
	}

	if os.Getenv("GO_ENV") != constants.GO_ENV_TEST {
		additionalEnvVars := []string{
			"EXAM_DB_HOST",
			"EXAM_DB_PORT",
			"EXAM_DB_SSLMODE",
			"EXAM_API_TOKEN",
			"EXAM_QUEUE_HOST",
			"EXAM_QUEUE_PORT",
			"PAPER_SERVER_HOST",
			"PAPER_SERVER_PORT",
			"QUESTION_SERVER_HOST",
			"QUESTION_SERVER_PORT",
		}
		baseEnvVars = append(baseEnvVars, additionalEnvVars...)
	}

	return baseEnvVars
}
