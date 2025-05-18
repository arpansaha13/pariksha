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
	GO_ENV                                 string
	AUTH_SERVER_PORT                       string
	JWT_SECRET_KEY                         string
	OTP_EXPIRES_IN_MINUTES                 int
	SESSION_EXPIRES_IN_HOURS               int
	SESSION_COOKIE_NAME                    string
	CSRFTOKEN_COOKIE_NAME                  string
	OTP_EXPIRES_IN_MINUTES_FORGOT_PASSWORD int
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
	RABBIT_SERVER_HOST string
	RABBIT_SERVER_PORT string
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

	OTP_EXPIRES_IN_MINUTES, _ = strconv.Atoi(utils.GetEnvWithDefault("OTP_EXPIRES_IN_MINUTES", "15"))
	SESSION_EXPIRES_IN_HOURS, _ = strconv.Atoi(utils.GetEnvWithDefault("SESSION_EXPIRES_IN_HOURS", "24"))
	SESSION_COOKIE_NAME = utils.GetEnvWithDefault("SESSION_COOKIE_NAME", "token")
	CSRFTOKEN_COOKIE_NAME = utils.GetEnvWithDefault("CSRFTOKEN_COOKIE_NAME", "csrftoken")
	OTP_EXPIRES_IN_MINUTES_FORGOT_PASSWORD, _ = strconv.Atoi(utils.GetEnvWithDefault("OTP_EXPIRES_IN_MINUTES_FORGOT_PASSWORD", "30"))

	USERS_DB_USER = os.Getenv("USERS_DB_USER")
	USERS_DB_PASS = os.Getenv("USERS_DB_PASS")
	USERS_DB_NAME = os.Getenv("USERS_DB_NAME")

	if GO_ENV != constants.GO_ENV_TEST {
		USERS_DB_HOST = os.Getenv("USERS_DB_HOST")
		USERS_DB_PORT = os.Getenv("USERS_DB_PORT")
		USERS_DB_SSLMODE = os.Getenv("USERS_DB_SSLMODE")

		RABBIT_SERVER_HOST = os.Getenv("RABBIT_SERVER_HOST")
		RABBIT_SERVER_PORT = os.Getenv("RABBIT_SERVER_PORT")
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
			"RABBIT_SERVER_HOST",
			"RABBIT_SERVER_PORT",
		}
		baseEnvVars = append(baseEnvVars, additionalEnvVars...)
	}

	return baseEnvVars
}
