package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"pariksha/common/pkg/constants"
)

var (
	GO_ENV               string
	QUESTION_SERVER_PORT string
)

var (
	EXAM_API_TOKEN   string
	PAPER_API_TOKEN  string
	ENGINE_API_TOKEN string
)

var (
	QUESTION_DB_HOST    string
	QUESTION_DB_PORT    string
	QUESTION_DB_USER    string
	QUESTION_DB_PASS    string
	QUESTION_DB_NAME    string
	QUESTION_DB_SSLMODE string
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

	QUESTION_SERVER_PORT = os.Getenv("QUESTION_SERVER_PORT")

	QUESTION_DB_USER = os.Getenv("QUESTION_DB_USER")
	QUESTION_DB_PASS = os.Getenv("QUESTION_DB_PASS")
	QUESTION_DB_NAME = os.Getenv("QUESTION_DB_NAME")

	if GO_ENV != constants.GO_ENV_TEST {
		QUESTION_DB_HOST = os.Getenv("QUESTION_DB_HOST")
		QUESTION_DB_PORT = os.Getenv("QUESTION_DB_PORT")
		QUESTION_DB_SSLMODE = os.Getenv("QUESTION_DB_SSLMODE")

		EXAM_API_TOKEN = os.Getenv("EXAM_API_TOKEN")
		PAPER_API_TOKEN = os.Getenv("PAPER_API_TOKEN")
		ENGINE_API_TOKEN = os.Getenv("ENGINE_API_TOKEN")
	}
}

func getRequiredEnvVars() []string {
	baseEnvVars := []string{
		"GO_ENV",
		"QUESTION_SERVER_PORT",
		"QUESTION_DB_USER",
		"QUESTION_DB_PASS",
		"QUESTION_DB_NAME",
	}

	if os.Getenv("GO_ENV") != constants.GO_ENV_TEST {
		additionalEnvVars := []string{
			"QUESTION_DB_HOST",
			"QUESTION_DB_PORT",
			"QUESTION_DB_SSLMODE",
			"EXAM_API_TOKEN",
			"PAPER_API_TOKEN",
			"ENGINE_API_TOKEN",
		}
		baseEnvVars = append(baseEnvVars, additionalEnvVars...)
	}

	return baseEnvVars
}
