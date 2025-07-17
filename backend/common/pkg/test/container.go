package test

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PgContainerEnv struct {
	PgUser     string
	PgPassword string
	PgDbname   string
}

func StartPgContainer(ctx context.Context, env *PgContainerEnv) (testcontainers.Container, error) {
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15.10-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     env.PgUser,
				"POSTGRES_PASSWORD": env.PgPassword,
				"POSTGRES_DB":       env.PgDbname,
			},
			WaitingFor: wait.ForAll(
				wait.ForLog("database system is ready to accept connections"),
				wait.ForExposedPort(),
			),
		},
		Started: true,
	})

	if err != nil {
		return nil, err
	}

	return pgContainer, nil
}
