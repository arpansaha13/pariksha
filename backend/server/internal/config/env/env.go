package env

import (
	"log"
	"os"

	"pariksha/common/pkg/utils"
)

var (
	GO_ENV                string
	API_PORT              string
	ID_ENCRYPTION_KEY     string
	SESSION_COOKIE_NAME   string
	CSRFTOKEN_COOKIE_NAME string
)

var (
	AUTH_SERVER_HOST string
	AUTH_SERVER_PORT string
)

var (
	PAPER_SERVER_HOST string
	PAPER_SERVER_PORT string
)

var (
	EXAM_SERVER_HOST string
	EXAM_SERVER_PORT string
)

var (
	ENGINE_SERVER_HOST string
	ENGINE_SERVER_PORT string
)

var (
	CLIENT_URL string
)

func init() {
	requiredEnvVars := []string{
		"GO_ENV",
		"ID_ENCRYPTION_KEY",
		"CLIENT_URL",
		"AUTH_SERVER_HOST",
		"AUTH_SERVER_PORT",
		"PAPER_SERVER_HOST",
		"PAPER_SERVER_PORT",
		"EXAM_SERVER_HOST",
		"EXAM_SERVER_PORT",
		"ENGINE_SERVER_HOST",
		"ENGINE_SERVER_PORT",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s is not set", envVar)
		}
	}

	GO_ENV = os.Getenv("GO_ENV")
	API_PORT = utils.GetEnvWithDefault("API_PORT", "4000")
	ID_ENCRYPTION_KEY = os.Getenv("ID_ENCRYPTION_KEY")

	SESSION_COOKIE_NAME = utils.GetEnvWithDefault("SESSION_COOKIE_NAME", "token")
	CSRFTOKEN_COOKIE_NAME = utils.GetEnvWithDefault("CSRFTOKEN_COOKIE_NAME", "csrftoken")

	AUTH_SERVER_HOST = os.Getenv("AUTH_SERVER_HOST")
	AUTH_SERVER_PORT = os.Getenv("AUTH_SERVER_PORT")

	PAPER_SERVER_HOST = os.Getenv("PAPER_SERVER_HOST")
	PAPER_SERVER_PORT = os.Getenv("PAPER_SERVER_PORT")

	EXAM_SERVER_HOST = os.Getenv("EXAM_SERVER_HOST")
	EXAM_SERVER_PORT = os.Getenv("EXAM_SERVER_PORT")

	ENGINE_SERVER_HOST = os.Getenv("ENGINE_SERVER_HOST")
	ENGINE_SERVER_PORT = os.Getenv("ENGINE_SERVER_PORT")

	CLIENT_URL = os.Getenv("CLIENT_URL")
}
