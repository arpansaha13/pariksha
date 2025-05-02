package handlers

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/config/env"
	"pariksha/paper/internal/interceptors"
)

const (
	bufSize                         = 1024 * 1024
	userID                   int64  = 1
	defaultPaperCategoryName string = "Category 1"
)

var (
	lis    *bufconn.Listener
	ctx    context.Context
	client proto.PaperServiceClient
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
		"questions",
		"question_categories",
		"permissions",
		"papers",
	}

	for i := range tables {
		err := db.DB.Exec("TRUNCATE TABLE " + tables[i] + " CASCADE").Error
		require.NoError(t, err)
	}
}

func createTestPaper(t *testing.T, userID int64) models.Paper {
	paper := models.Paper{
		Title:           "Test Paper",
		MaxScore:        0,
		DurationMinutes: 60,
		CreatedBy:       userID,
	}
	err := db.DB.Create(&paper).Error
	require.NoError(t, err)

	// Create permissions entry with write access
	permissions := models.PaperPermissions{
		UserID:  userID,
		PaperID: paper.ID,
	}
	permissions.SetWrite()
	err = db.DB.Create(&permissions).Error
	require.NoError(t, err)

	category := models.QuestionCategory{
		PaperID: sql.NullInt64{Int64: paper.ID, Valid: true},
		Name:    defaultPaperCategoryName,
		Order:   1,
	}
	err = db.DB.Create(&category).Error
	require.NoError(t, err)

	return paper
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func createContextWithUserID(userID int64) context.Context {
	md := metadata.New(map[string]string{
		"user_id": strconv.FormatInt(userID, 10),
	})
	return metadata.NewOutgoingContext(context.Background(), md)
}

func setupGrpcServer() (*grpc.Server, *grpc.ClientConn) {
	lis = bufconn.Listen(bufSize)
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.PaperAuthInterceptor()),
	)
	proto.RegisterPaperServiceServer(srv, &PaperServer{})

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

	client = proto.NewPaperServiceClient(conn)

	code := m.Run()

	cleanup()
	os.Exit(code)
}
