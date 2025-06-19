package env

import (
	"log"
	"os"

	"pariksha/common/pkg/utils"
)

var (
	GO_ENV           string
	MAIL_SERVER_PORT string
	SMTP_NAME        string
	SMTP_USER        string
	SMTP_FROM        string
	SMTP_PASSWORD    string
	SMTP_HOST        string
	SMTP_PORT        string
)

var (
	MAIL_QUEUE_HOST string
	MAIL_QUEUE_PORT string
)

func init() {
	requiredEnvVars := []string{
		"SMTP_NAME",
		"SMTP_USER",
		"SMTP_FROM",
		"SMTP_PASSWORD",
		"SMTP_HOST",
		"SMTP_PORT",
		"MAIL_QUEUE_HOST",
		"MAIL_QUEUE_PORT",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s is not set", envVar)
		}
	}

	GO_ENV = os.Getenv("GO_ENV")
	MAIL_SERVER_PORT = utils.GetEnvWithDefault("MAIL_SERVER_PORT", "4010")
	SMTP_NAME = os.Getenv("SMTP_NAME")
	SMTP_USER = os.Getenv("SMTP_USER")
	SMTP_FROM = os.Getenv("SMTP_FROM")
	SMTP_PASSWORD = os.Getenv("SMTP_PASSWORD")
	SMTP_HOST = os.Getenv("SMTP_HOST")
	SMTP_PORT = os.Getenv("SMTP_PORT")

	MAIL_QUEUE_HOST = os.Getenv("MAIL_QUEUE_HOST")
	MAIL_QUEUE_PORT = os.Getenv("MAIL_QUEUE_PORT")
}
