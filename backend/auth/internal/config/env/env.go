package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"pariksha/common/pkg/constants"
)

var (
	GO_ENV           string
	AUTH_SERVER_PORT string
	JWT_SECRET_KEY   string
)

const (
	SESSION_EXPIRES_IN_HOURS               int16 = 24
	OTP_LOGIN_EXPIRES_IN_MINUTES           int16 = 15
	OTP_FORGOT_PASSWORD_EXPIRES_IN_MINUTES int16 = 15
)

var (
	USERS_DB_HOST    string
	USERS_DB_PORT    string
	USERS_DB_USER    string
	USERS_DB_PASS    string
	USERS_DB_NAME    string
	USERS_DB_SSLMODE string
)

var (
	MAIL_QUEUE_HOST string
	MAIL_QUEUE_PORT string
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
	AUTH_SERVER_PORT = os.Getenv("AUTH_SERVER_PORT")
	JWT_SECRET_KEY = os.Getenv("JWT_SECRET_KEY")

	USERS_DB_USER = os.Getenv("USERS_DB_USER")
	USERS_DB_PASS = os.Getenv("USERS_DB_PASS")
	USERS_DB_NAME = os.Getenv("USERS_DB_NAME")

	if GO_ENV != constants.GO_ENV_TEST {
		USERS_DB_HOST = os.Getenv("USERS_DB_HOST")
		USERS_DB_PORT = os.Getenv("USERS_DB_PORT")
		USERS_DB_SSLMODE = os.Getenv("USERS_DB_SSLMODE")

		MAIL_QUEUE_HOST = os.Getenv("MAIL_QUEUE_HOST")
		MAIL_QUEUE_PORT = os.Getenv("MAIL_QUEUE_PORT")
	}
}

func getRequiredEnvVars() []string {
	baseEnvVars := []string{
		"GO_ENV",
		"AUTH_SERVER_PORT",
		"JWT_SECRET_KEY",
		"USERS_DB_USER",
		"USERS_DB_PASS",
		"USERS_DB_NAME",
	}

	if os.Getenv("GO_ENV") != constants.GO_ENV_TEST {
		// require these vars only when not in test environment
		additionalEnvVars := []string{
			"USERS_DB_HOST",
			"USERS_DB_PORT",
			"USERS_DB_SSLMODE",
			"MAIL_QUEUE_HOST",
			"MAIL_QUEUE_PORT",
		}
		baseEnvVars = append(baseEnvVars, additionalEnvVars...)
	}

	return baseEnvVars
}
