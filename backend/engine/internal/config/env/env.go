package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	GO_ENV             string
	ENGINE_SERVER_PORT string
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
	ENGINE_SERVER_PORT = os.Getenv("ENGINE_SERVER_PORT")
}

func getRequiredEnvVars() []string {
	baseEnvVars := []string{
		"GO_ENV",
		"ENGINE_SERVER_PORT",
	}

	return baseEnvVars
}
