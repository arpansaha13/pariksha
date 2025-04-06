package env

import (
	"log"
	"os"
)

var GO_ENV string

var (
	EXAM_DB_HOST    string
	EXAM_DB_USER    string
	EXAM_DB_PASS    string
	EXAM_DB_NAME    string
	EXAM_DB_PORT    string
	EXAM_DB_SSLMODE string
)

var (
	PAPER_DB_HOST    string
	PAPER_DB_USER    string
	PAPER_DB_PASS    string
	PAPER_DB_NAME    string
	PAPER_DB_PORT    string
	PAPER_DB_SSLMODE string
)

var (
	EXAM_QUEUE_HOST string
	EXAM_QUEUE_PORT string
)

func init() {
	requiredEnvVars := []string{
		"GO_ENV",
		"EXAM_DB_HOST",
		"EXAM_DB_USER",
		"EXAM_DB_PASS",
		"EXAM_DB_NAME",
		"EXAM_DB_PORT",
		"EXAM_DB_SSLMODE",
		"PAPER_DB_HOST",
		"PAPER_DB_USER",
		"PAPER_DB_PASS",
		"PAPER_DB_NAME",
		"PAPER_DB_PORT",
		"PAPER_DB_SSLMODE",
		"EXAM_QUEUE_HOST",
		"EXAM_QUEUE_PORT",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s is not set", envVar)
		}
	}

	GO_ENV = os.Getenv("GO_ENV")

	EXAM_DB_HOST = os.Getenv("EXAM_DB_HOST")
	EXAM_DB_USER = os.Getenv("EXAM_DB_USER")
	EXAM_DB_PASS = os.Getenv("EXAM_DB_PASS")
	EXAM_DB_NAME = os.Getenv("EXAM_DB_NAME")
	EXAM_DB_PORT = os.Getenv("EXAM_DB_PORT")
	EXAM_DB_SSLMODE = os.Getenv("EXAM_DB_SSLMODE")

	PAPER_DB_HOST = os.Getenv("PAPER_DB_HOST")
	PAPER_DB_USER = os.Getenv("PAPER_DB_USER")
	PAPER_DB_PASS = os.Getenv("PAPER_DB_PASS")
	PAPER_DB_NAME = os.Getenv("PAPER_DB_NAME")
	PAPER_DB_PORT = os.Getenv("PAPER_DB_PORT")
	PAPER_DB_SSLMODE = os.Getenv("PAPER_DB_SSLMODE")

	EXAM_QUEUE_HOST = os.Getenv("EXAM_QUEUE_HOST")
	EXAM_QUEUE_PORT = os.Getenv("EXAM_QUEUE_PORT")
}
