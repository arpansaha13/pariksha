package tests

import (
	"context"
	"log"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/config/env"
	"pariksha/paper/internal/handlers"
	"pariksha/paper/internal/interceptors"
)

const (
	bufSize              = 1024 * 1024
	userID  types.UserID = 1
)

var (
	lis    *bufconn.Listener
	ctx    context.Context
	client proto.PaperClient
)

func setupContainer() func() {
	ctx = context.Background()

	// Start Postgres container
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15.10-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     env.DB_USER,
				"POSTGRES_PASSWORD": env.DB_PASS,
				"POSTGRES_DB":       env.DB_NAME,
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
		env.DB_USER,
		env.DB_PASS,
		env.DB_NAME,
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
		constants.TABLE_BOILERPLATES,
		constants.TABLE_LANGUAGES,
		constants.TABLE_TEST_CASES,
		constants.TABLE_QUESTIONS,
		constants.TABLE_CATEGORIES,
		constants.TABLE_PAPER_PERMISSIONS,
		constants.TABLE_PAPERS,
	}

	for i := range tables {
		err := db.DB.Exec("TRUNCATE TABLE " + string(tables[i]) + " CASCADE").Error
		require.NoError(t, err)
	}
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func setupGrpcServer() (*grpc.Server, *grpc.ClientConn) {
	lis = bufconn.Listen(bufSize)
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.ExamServiceAuthInterceptor(),
			interceptors.PaperAuthInterceptor(),
			interceptors.DeletePaperAuthInterceptor(),
		),
	)
	proto.RegisterPaperServer(srv, &handlers.PaperServer{})

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}

	return srv, conn
}

func TestMain(m *testing.M) {
	cleanup := setupContainer()

	srv, conn := setupGrpcServer()
	defer func() {
		conn.Close()
		srv.Stop()
	}()

	client = proto.NewPaperClient(conn)

	code := m.Run()

	cleanup()
	os.Exit(code)
}
