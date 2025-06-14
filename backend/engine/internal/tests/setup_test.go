package tests

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/engine/internal/config/db"
	"pariksha/engine/internal/config/env"
	"pariksha/engine/internal/handlers"
)

var (
	ctx    context.Context
	server *handlers.EngineServer
)

func setupContainers() func() {
	ctx = context.Background()

	// Start Postgres container
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15.10-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     env.PAPERS_DB_USER,
				"POSTGRES_PASSWORD": env.PAPERS_DB_PASS,
				"POSTGRES_DB":       env.PAPERS_DB_NAME,
			},
			WaitingFor: wait.ForAll(
				wait.ForLog("database system is ready to accept connections"),
				wait.ForExposedPort(),
			),
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("Failed to setup container: %v", err)
	}

	// Get container host and mapped port
	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")

	// Initialize DB connection
	err = db.InitDB(
		pgHost,
		pgPort.Port(),
		env.PAPERS_DB_USER,
		env.PAPERS_DB_PASS,
		env.PAPERS_DB_NAME,
		"disable",
	)
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}

	return func() {
		pgContainer.Terminate(ctx)
	}
}

func clearTables(t *testing.T) {
	tables := []string{
		constants.TABLE_QUESTIONS,
		constants.TABLE_CATEGORIES,
		constants.TABLE_PAPERS,
	}

	for _, table := range tables {
		err := db.Papers.Exec("TRUNCATE TABLE " + table + " CASCADE").Error
		require.NoError(t, err)
	}
}

func TestMain(m *testing.M) {
	cleanup := setupContainers()

	// Create tmp directory for test files
	if err := os.MkdirAll("tmp", 0755); err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	cleanup()
	os.RemoveAll("tmp")

	os.Exit(code)
}

func setupTest(t *testing.T) proto.EngineServer {
	clearTables(t)

	var err error
	server, err = handlers.NewEngineServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	return server
}
