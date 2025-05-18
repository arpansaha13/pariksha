package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"pariksha/auth/internal/config/db"
	"pariksha/auth/internal/config/env"
	"pariksha/auth/internal/services"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
)

const (
	bufSize = 1024 * 1024

	testVerifiedEmail   = "verified@example.com"
	testUnverifiedEmail = "unverified@example.com"

	validOTP   = "123456"
	invalidOTP = "654321"
)

var (
	lis    *bufconn.Listener
	ctx    context.Context
	client proto.AuthServiceClient
)

func setupContainers() func() {
	ctx = context.Background()

	// Start Postgres container
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15.10-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     env.USERS_DB_USER,
				"POSTGRES_PASSWORD": env.USERS_DB_PASS,
				"POSTGRES_DB":       env.USERS_DB_NAME,
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

	// Start RabbitMQ container
	rabbitmq, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "rabbitmq:3-management-alpine",
			ExposedPorts: []string{"5672/tcp"},
			WaitingFor:   wait.ForLog("Server startup complete"),
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("Failed to setup container: %v", err)
	}

	// Get mapped host and ports
	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")
	rabbitHost, _ := rabbitmq.Host(ctx)
	rabbitPort, _ := rabbitmq.MappedPort(ctx, "5672")

	// Initialize connections with container
	err = db.InitDB(
		pgHost,
		pgPort.Port(),
		env.USERS_DB_USER,
		env.USERS_DB_PASS,
		env.USERS_DB_NAME,
		"disable",
	)
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}

	err = services.InitRabbitMQ(rabbitHost, rabbitPort.Port())
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}

	cleanup := func() {
		services.CloseRabbit()
		pgContainer.Terminate(ctx)
		rabbitmq.Terminate(ctx)
	}

	return cleanup
}

func clearTables(t *testing.T) {
	tables := []string{constants.TABLE_USERS, constants.TABLE_OTPS}
	for _, table := range tables {
		err := db.DB.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error
		require.NoError(t, err)
	}
	err := db.DB.Exec("DELETE FROM " + constants.TABLE_SESSIONS).Error
	require.NoError(t, err)
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func setupGrpcServer() (*grpc.Server, *grpc.ClientConn) {
	lis = bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	proto.RegisterAuthServiceServer(srv, &AuthServer{})

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}

	return srv, conn
}

func TestMain(m *testing.M) {
	cleanup := setupContainers()

	// Setup gRPC server
	srv, conn := setupGrpcServer()
	defer func() {
		conn.Close()
		srv.Stop()
	}()

	// Setup gRPC client
	client = proto.NewAuthServiceClient(conn)

	// Run tests
	code := m.Run()

	// Cleanup
	cleanup()
	os.Exit(code)
}

func createTestUser(t *testing.T, email string, verified bool) models.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testPass123"), bcrypt.DefaultCost)
	user := models.User{
		Email:    email,
		Username: strings.Split(email, "@")[0],
		Password: sql.NullString{String: string(hashedPassword), Valid: true},
		Verified: verified,
	}
	err := db.DB.Create(&user).Error
	require.NoError(t, err)
	return user
}

func createTestOTP(t *testing.T, email string, purpose int16, expired bool) {
	duration := 15 * time.Minute
	if expired {
		duration = -15 * time.Minute
	}

	otp := models.Otp{
		Email:        email,
		OTP:          validOTP,
		OTPExpiresAt: time.Now().Add(duration),
		Purpose:      purpose,
	}
	err := db.DB.Create(&otp).Error
	require.NoError(t, err)
}
