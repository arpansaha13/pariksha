package env

import (
	"log"
	"os"

	"github.com/arpansaha13/common/pkg/utils"
)

var (
	MAIL_SERVER_PORT   string
	SMTP_NAME          string
	SMTP_USER          string
	SMTP_FROM          string
	SMTP_PASSWORD      string
	SMTP_HOST          string
	SMTP_PORT          string
	RABBIT_SERVER_HOST string
	RABBIT_SERVER_PORT string
)

func init() {
	requiredEnvVars := []string{
		"SMTP_NAME",
		"SMTP_USER",
		"SMTP_FROM",
		"SMTP_PASSWORD",
		"SMTP_HOST",
		"SMTP_PORT",
		"RABBIT_SERVER_HOST",
		"RABBIT_SERVER_PORT",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s is not set", envVar)
		}
	}

	MAIL_SERVER_PORT = utils.GetEnvWithDefault("MAIL_SERVER_PORT", "4010")
	SMTP_NAME = os.Getenv("SMTP_NAME")
	SMTP_USER = os.Getenv("SMTP_USER")
	SMTP_FROM = os.Getenv("SMTP_FROM")
	SMTP_PASSWORD = os.Getenv("SMTP_PASSWORD")
	SMTP_HOST = os.Getenv("SMTP_HOST")
	SMTP_PORT = os.Getenv("SMTP_PORT")

	RABBIT_SERVER_HOST = os.Getenv("RABBIT_SERVER_HOST")
	RABBIT_SERVER_PORT = os.Getenv("RABBIT_SERVER_PORT")
}
