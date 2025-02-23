package env

import (
	"log"
	"os"
	"strconv"

	"github.com/arpansaha13/common/pkg/utils"
)

var (
	AUTH_SERVER_PORT                       string
	JWT_SECRET_KEY                         string
	OTP_EXPIRES_IN_MINUTES                 int
	SESSION_EXPIRES_IN_HOURS               int
	SESSION_COOKIE_NAME                    string
	CSRFTOKEN_COOKIE_NAME                  string
	OTP_EXPIRES_IN_MINUTES_FORGOT_PASSWORD int
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
	SESSIONS_DB_HOST string
	SESSIONS_DB_PORT string
	SESSIONS_DB_USER string
	SESSIONS_DB_PASS string
	SESSIONS_DB_NAME string
)

var (
	RABBIT_SERVER_HOST string
	RABBIT_SERVER_PORT string
)

func init() {
	requiredEnvVars := []string{
		"AUTH_SERVER_PORT",
		"JWT_SECRET_KEY",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASS",
		"DB_NAME",
		"DB_SSLMODE",
		"SESSIONS_DB_HOST",
		"SESSIONS_DB_PORT",
		"SESSIONS_DB_USER",
		"SESSIONS_DB_PASS",
		"SESSIONS_DB_NAME",
		"RABBIT_SERVER_HOST",
		"RABBIT_SERVER_PORT",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s is not set", envVar)
		}
	}

	AUTH_SERVER_PORT = os.Getenv("AUTH_SERVER_PORT")
	JWT_SECRET_KEY = os.Getenv("JWT_SECRET_KEY")

	OTP_EXPIRES_IN_MINUTES, _ = strconv.Atoi(utils.GetEnvWithDefault("OTP_EXPIRES_IN_MINUTES", "15"))
	SESSION_EXPIRES_IN_HOURS, _ = strconv.Atoi(utils.GetEnvWithDefault("SESSION_EXPIRES_IN_HOURS", "24"))
	SESSION_COOKIE_NAME = utils.GetEnvWithDefault("SESSION_COOKIE_NAME", "token")
	CSRFTOKEN_COOKIE_NAME = utils.GetEnvWithDefault("CSRFTOKEN_COOKIE_NAME", "csrftoken")
	OTP_EXPIRES_IN_MINUTES_FORGOT_PASSWORD, _ = strconv.Atoi(utils.GetEnvWithDefault("OTP_EXPIRES_IN_MINUTES_FORGOT_PASSWORD", "30"))

	DB_HOST = os.Getenv("DB_HOST")
	DB_PORT = os.Getenv("DB_PORT")
	DB_USER = os.Getenv("DB_USER")
	DB_PASS = os.Getenv("DB_PASS")
	DB_NAME = os.Getenv("DB_NAME")
	DB_SSLMODE = os.Getenv("DB_SSLMODE")

	SESSIONS_DB_HOST = os.Getenv("SESSIONS_DB_HOST")
	SESSIONS_DB_PORT = os.Getenv("SESSIONS_DB_PORT")
	SESSIONS_DB_USER = os.Getenv("SESSIONS_DB_USER")
	SESSIONS_DB_PASS = os.Getenv("SESSIONS_DB_PASS")
	SESSIONS_DB_NAME = os.Getenv("SESSIONS_DB_NAME")

	RABBIT_SERVER_HOST = os.Getenv("RABBIT_SERVER_HOST")
	RABBIT_SERVER_PORT = os.Getenv("RABBIT_SERVER_PORT")
}
