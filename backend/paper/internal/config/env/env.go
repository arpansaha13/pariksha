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
	DB_HOST    string
	DB_PORT    string
	DB_USER    string
	DB_PASS    string
	DB_NAME    string
	DB_SSLMODE string
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

	DB_USER = os.Getenv("DB_USER")
	DB_PASS = os.Getenv("DB_PASS")
	DB_NAME = os.Getenv("DB_NAME")

	if GO_ENV != constants.GO_ENV_TEST {
		DB_HOST = os.Getenv("DB_HOST")
		DB_PORT = os.Getenv("DB_PORT")
		DB_SSLMODE = os.Getenv("DB_SSLMODE")
	}
}

func getRequiredEnvVars() []string {
	baseEnvVars := []string{
		"GO_ENV",
		"DB_USER",
		"DB_PASS",
		"DB_NAME",
		"PAPER_SERVER_PORT",
	}

	if os.Getenv("GO_ENV") != constants.GO_ENV_TEST {
		additionalEnvVars := []string{
			"DB_HOST",
			"DB_PORT",
			"DB_SSLMODE",
			"QUESTION_SERVER_HOST",
			"QUESTION_SERVER_PORT",
			"EXAM_API_TOKEN",
		}
		baseEnvVars = append(baseEnvVars, additionalEnvVars...)
	}

	return baseEnvVars
}
