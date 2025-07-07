package tests

import (
	"os"
	"testing"

	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils/generate"
	"pariksha/engine/internal/handlers"
	"pariksha/engine/internal/interservice"
)

var (
	server *handlers.EngineServer

	mockQuestionHash     string = generate.HMACHash(1)
	mockInputDefinitions []proto.InputDefinition
)

// mockPaperService replaces the actual paper service with test doubles
func mockPaperService() func() {
	originalFetchInputDefinitions := interservice.GetInputDefinitions

	// Mock FetchInputDefinitions
	interservice.GetInputDefinitions = func(questionHash string) ([]*proto.InputDefinition, error) {
		defs := make([]*proto.InputDefinition, len(mockInputDefinitions))
		for i := range mockInputDefinitions {
			defs[i] = &mockInputDefinitions[i]
		}
		// Set the mock before running the test
		return defs, nil
	}

	return func() {
		interservice.GetInputDefinitions = originalFetchInputDefinitions
	}
}

func TestMain(m *testing.M) {
	// Create tmp directory for test files
	if err := os.MkdirAll("tmp", 0755); err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.RemoveAll("tmp")

	os.Exit(code)
}

func setupTest(t *testing.T) proto.EngineServer {
	var err error
	server, err = handlers.NewEngineServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	return server
}
