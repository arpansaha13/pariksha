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
	ENGINE_API_TOKEN    string
)

var (
	PAPER_SERVER_HOST string
	PAPER_SERVER_PORT string
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
	ENGINE_API_TOKEN = os.Getenv("ENGINE_API_TOKEN")

	PAPER_SERVER_HOST = os.Getenv("PAPER_SERVER_HOST")
	PAPER_SERVER_PORT = os.Getenv("PAPER_SERVER_PORT")

	HOST_TMP_MOUNT_PATH = utils.GetEnvWithDefault("HOST_TMP_MOUNT_PATH", "")

	if GO_ENV != constants.GO_ENV_TEST {
	}
}

func getRequiredEnvVars() []string {
	baseEnvVars := []string{
		"GO_ENV",
		"ENGINE_SERVER_PORT",
		"DOCKER_API_VERSION", // used to create docker client with `client.FromEnv`
		"PAPER_SERVER_HOST",
		"PAPER_SERVER_PORT",
		"ENGINE_API_TOKEN",
	}

	if os.Getenv("GO_ENV") != constants.GO_ENV_TEST {
		// require these vars only when not in test environment
		additionalEnvVars := []string{}
		baseEnvVars = append(baseEnvVars, additionalEnvVars...)
	}

	return baseEnvVars
}
