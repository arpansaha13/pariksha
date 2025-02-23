package env

import (
	"log"
	"os"
	"strconv"

	"github.com/arpansaha13/common/pkg/utils"
)

var (
	API_PORT                 string
	OTP_EXPIRES_IN_MINUTES   int
	SESSION_EXPIRES_IN_HOURS int
	SESSION_COOKIE_NAME      string
	CSRFTOKEN_COOKIE_NAME    string
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
	AUTH_SERVER_HOST string
	AUTH_SERVER_PORT string
)

func init() {
	requiredEnvVars := []string{
		"GO_ENV",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASS",
		"DB_NAME",
		"DB_SSLMODE",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s is not set", envVar)
		}
	}

	API_PORT = utils.GetEnvWithDefault("API_PORT", "4000")
	OTP_EXPIRES_IN_MINUTES, _ = strconv.Atoi(utils.GetEnvWithDefault("OTP_EXPIRES_IN_MINUTES", "15"))
	SESSION_EXPIRES_IN_HOURS, _ = strconv.Atoi(utils.GetEnvWithDefault("SESSION_EXPIRES_IN_HOURS", "24"))
	SESSION_COOKIE_NAME = utils.GetEnvWithDefault("SESSION_COOKIE_NAME", "token")
	CSRFTOKEN_COOKIE_NAME = utils.GetEnvWithDefault("CSRFTOKEN_COOKIE_NAME", "csrftoken")

	DB_HOST = os.Getenv("DB_HOST")
	DB_PORT = os.Getenv("DB_PORT")
	DB_USER = os.Getenv("DB_USER")
	DB_PASS = os.Getenv("DB_PASS")
	DB_NAME = os.Getenv("DB_NAME")
	DB_SSLMODE = os.Getenv("DB_SSLMODE")

	AUTH_SERVER_HOST = os.Getenv("AUTH_SERVER_HOST")
	AUTH_SERVER_PORT = os.Getenv("AUTH_SERVER_PORT")
}
