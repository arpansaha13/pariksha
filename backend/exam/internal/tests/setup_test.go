package tests

import (
	"context"
	"log"
	"net"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/config/env"
	"pariksha/exam/internal/handlers"
	"pariksha/exam/internal/interceptors"
	"pariksha/exam/internal/services"
)

const (
	bufSize       = 1024 * 1024
	userID  int64 = 1 // Creator/admin user ID
)

var (
	lis    *bufconn.Listener
	ctx    context.Context
	client proto.ExamServiceClient
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

	// Start Redis container
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections"),
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("Failed to setup Redis container: %v", err)
	}

	// Get container hosts and mapped ports
	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")
	redisHost, _ := redisContainer.Host(ctx)
	redisPort, _ := redisContainer.MappedPort(ctx, "6379")

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

	// Initialize Redis connection
	err = services.InitExamQueue(redisHost, redisPort.Port())
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	return func() {
		services.CloseExamQueue()
		pgContainer.Terminate(ctx)
		redisContainer.Terminate(ctx)
	}
}

func clearTables(t *testing.T) {
	tables := []string{
		"exam_participants",
		"exams",
		"permissions",
		"exam_questions",
		"exam_categories",
		"answers",
	}

	for _, table := range tables {
		err := db.DB.Exec("TRUNCATE TABLE " + table + " CASCADE").Error
		if err != nil {
			t.Fatalf("Failed to clear table %s: %v", table, err)
		}
	}
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func setupGrpcServer() (*grpc.Server, *grpc.ClientConn) {
	lis = bufconn.Listen(bufSize)
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.GeneralExamAuthInterceptor(),
			interceptors.DeleteExamsAuthInterceptor(),
			interceptors.EndExamInterceptor(),
		),
	)
	proto.RegisterExamServiceServer(srv, &handlers.ExamServer{})

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

	client = proto.NewExamServiceClient(conn)

	code := m.Run()

	cleanup()
	os.Exit(code)
}
