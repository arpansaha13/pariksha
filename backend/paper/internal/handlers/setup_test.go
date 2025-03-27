package handlers

import (
	"context"
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

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/config/env"
)

const (
	bufSize                         = 1024 * 1024
	userID                          = 1
	defaultPaperCategoryName string = "Category 1"
)

var (
	lis       *bufconn.Listener
	ctx       context.Context
	client    proto.PaperServiceClient
	testUsers map[int]models.User
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

func createTestUsers(t *testing.T) {
	testUsers = make(map[int]models.User)

	users := []models.User{
		{
			ID:       1,
			Username: "testuser1",
			Email:    "test1@example.com",
			Verified: true,
		},
		{
			ID:       2,
			Username: "testuser2",
			Email:    "test2@example.com",
			Verified: true,
		},
	}

	for _, user := range users {
		err := db.DB.Create(&user).Error
		require.NoError(t, err)
		testUsers[user.ID] = user
	}
}

func clearTables(t *testing.T) {
	tables := []string{
		"questions",
		"question_categories",
		"paper_ownerships",
		"papers",
	}

	for i := range tables {
		err := db.DB.Exec("TRUNCATE TABLE " + tables[i] + " CASCADE").Error
		require.NoError(t, err)
	}
}

func createTestPaper(t *testing.T, userID int) models.Paper {
	paper := models.Paper{
		Title:           "Test Paper",
		MaxScore:        0,
		DurationMinutes: 60,
	}
	err := db.DB.Create(&paper).Error
	require.NoError(t, err)

	ownership := models.PaperOwnership{
		UserID:  userID,
		PaperID: paper.ID,
		Type:    constants.PAPER_OWNERSHIP_TYPE_OWNER,
	}
	err = db.DB.Create(&ownership).Error
	require.NoError(t, err)

	category := models.QuestionCategory{
		PaperID: paper.ID,
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

func createContextWithUserID(userID int32) context.Context {
	md := metadata.New(map[string]string{
		"user_id": strconv.Itoa(int(userID)),
	})
	return metadata.NewOutgoingContext(context.Background(), md)
}

func setupGrpcServer() (*grpc.Server, *grpc.ClientConn) {
	lis = bufconn.Listen(bufSize)
	srv := grpc.NewServer()
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

	createTestUsers(&testing.T{})

	code := m.Run()

	if err := db.DB.Exec("TRUNCATE TABLE users CASCADE").Error; err != nil {
		log.Printf("Failed to cleanup users: %v", err)
	}

	cleanup()
	os.Exit(code)
}
