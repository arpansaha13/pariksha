package env

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/arpansaha13/common/pkg/utils"
)

var (
	API_PORT                               string
	OTP_EXPIRES_IN_MINUTES                 int
	SESSION_EXPIRES_IN_HOURS               int
	SESSION_COOKIE_NAME                    string
	CSRFTOKEN_COOKIE_NAME                  string
	OTP_EXPIRES_IN_MINUTES_FORGOT_PASSWORD int
	MAIL_SERVER_HOST                       string
	MAIL_SERVER_PORT                       string
)

func init() {
	if os.Getenv("GO_ENV") == "development" && godotenv.Load() != nil {
		log.Fatalf("Error loading .env file")
	}

	requiredEnvVars := []string{
		"GO_ENV",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASS",
		"DB_NAME",
		"JWT_SECRET_KEY",
		"MAIL_SERVER_PORT",
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
	OTP_EXPIRES_IN_MINUTES_FORGOT_PASSWORD, _ = strconv.Atoi(utils.GetEnvWithDefault("OTP_EXPIRES_IN_MINUTES_FORGOT_PASSWORD", "30"))

	MAIL_SERVER_HOST = utils.GetEnvWithDefault("MAIL_SERVER_HOST", "")
	MAIL_SERVER_PORT = os.Getenv("MAIL_SERVER_PORT")
}
