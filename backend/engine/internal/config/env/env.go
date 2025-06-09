package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/utils"
)

var (
	GO_ENV              string
	ENGINE_SERVER_PORT  string
	HOST_TMP_MOUNT_PATH string
)

var (
	PAPERS_DB_HOST    string
	PAPERS_DB_PORT    string
	PAPERS_DB_USER    string
	PAPERS_DB_PASS    string
	PAPERS_DB_NAME    string
	PAPERS_DB_SSLMODE string
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

	PAPERS_DB_USER = os.Getenv("PAPERS_DB_USER")
	PAPERS_DB_PASS = os.Getenv("PAPERS_DB_PASS")
	PAPERS_DB_NAME = os.Getenv("PAPERS_DB_NAME")
	HOST_TMP_MOUNT_PATH = utils.GetEnvWithDefault("HOST_TMP_MOUNT_PATH", "")

	if GO_ENV != constants.GO_ENV_TEST {
		PAPERS_DB_HOST = os.Getenv("PAPERS_DB_HOST")
		PAPERS_DB_PORT = os.Getenv("PAPERS_DB_PORT")
		PAPERS_DB_SSLMODE = os.Getenv("PAPERS_DB_SSLMODE")
	}
}

func getRequiredEnvVars() []string {
	baseEnvVars := []string{
		"GO_ENV",
		"ENGINE_SERVER_PORT",
		"DOCKER_API_VERSION", // used to create docker client with `client.FromEnv`
		"PAPERS_DB_USER",
		"PAPERS_DB_PASS",
		"PAPERS_DB_NAME",
	}

	if os.Getenv("GO_ENV") != constants.GO_ENV_TEST {
		// require these vars only when not in test environment
		additionalEnvVars := []string{
			"PAPERS_DB_HOST",
			"PAPERS_DB_PORT",
			"PAPERS_DB_SSLMODE",
		}
		baseEnvVars = append(baseEnvVars, additionalEnvVars...)
	}

	return baseEnvVars
}
