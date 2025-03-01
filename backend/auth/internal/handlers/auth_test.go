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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
	lis        *bufconn.Listener
	ctx        context.Context
	containers *testContainers
	client     proto.AuthServiceClient
)

type testContainers struct {
	postgres   testcontainers.Container
	sessionsDb testcontainers.Container
	rabbitmq   testcontainers.Container
	cleanup    func()
}

func setupContainers() *testContainers {
	ctx = context.Background()

	// Start Postgres container
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     env.DB_USER,
				"POSTGRES_PASSWORD": env.DB_PASS,
				"POSTGRES_DB":       env.DB_NAME,
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections"),
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("Failed to setup container: %v", err)
	}

	// Start Sessions DB container
	sessionsDb, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     env.SESSIONS_DB_USER,
				"POSTGRES_PASSWORD": env.SESSIONS_DB_PASS,
				"POSTGRES_DB":       env.SESSIONS_DB_NAME,
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections"),
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

	// Get mapped ports
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")
	sessionsPort, _ := sessionsDb.MappedPort(ctx, "5432")
	rabbitPort, _ := rabbitmq.MappedPort(ctx, "5672")

	// Initialize connections with container
	err = db.InitDB(
		"localhost",
		pgPort.Port(),
		env.DB_USER,
		env.DB_PASS,
		env.DB_NAME,
		"disable",
	)
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}

	err = db.InitSessionsDB(
		"localhost",
		sessionsPort.Port(),
		env.SESSIONS_DB_USER,
		env.SESSIONS_DB_PASS,
		env.SESSIONS_DB_NAME,
		"disable",
	)
	if err != nil {
		log.Fatalf("Failed to initialize Sessions DB: %v", err)
	}

	err = services.InitRabbitMQ("localhost", rabbitPort.Port())
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}

	cleanup := func() {
		services.CloseRabbit()
		pgContainer.Terminate(ctx)
		sessionsDb.Terminate(ctx)
		rabbitmq.Terminate(ctx)
	}

	return &testContainers{
		postgres:   pgContainer,
		sessionsDb: sessionsDb,
		rabbitmq:   rabbitmq,
		cleanup:    cleanup,
	}
}

func clearTables(t *testing.T) {
	tables := []string{"users", "otps"}
	for _, table := range tables {
		err := db.DB.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error
		require.NoError(t, err)
	}
	err := db.Sessions.Exec("DELETE FROM sessions").Error
	require.NoError(t, err)
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

func createTestOTP(t *testing.T, email string, purpose int, expired bool) {
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

	// https://stackoverflow.com/questions/78485578/how-to-use-the-bufconn-package-with-grpc-newclient
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
	containers = setupContainers()

	// Setup gRPC server and client
	srv, conn := setupGrpcServer()
	defer func() {
		conn.Close()
		srv.Stop()
	}()
	client = proto.NewAuthServiceClient(conn)

	// Run tests
	code := m.Run()

	// Cleanup
	containers.cleanup()
	os.Exit(code)
}

func TestSignUp(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.SignUpRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T)
	}{
		{
			name: "Successful signup",
			req: &proto.SignUpRequest{
				Email:    "new@example.com",
				Password: "newPass123",
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T) {
				var user models.User
				err := db.DB.Where("email = ?", "new@example.com").First(&user).Error
				assert.NoError(t, err)
				assert.False(t, user.Verified)

				var otp models.Otp
				err = db.DB.Where("email = ? AND purpose = ?", "new@example.com", constants.OTP_PURPOSE_SIGNUP).First(&otp).Error
				assert.NoError(t, err)
				assert.NotEmpty(t, otp.OTP)
				assert.True(t, otp.OTPExpiresAt.After(time.Now()))
			},
		},
		{
			name: "Email already exists and verified",
			setup: func(t *testing.T) {
				createTestUser(t, "existing@example.com", true)
			},
			req: &proto.SignUpRequest{
				Email:    "existing@example.com",
				Password: "pass123",
			},
			expectedCode: codes.AlreadyExists,
		},
		{
			name: "Unverified user signup again",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
			},
			req: &proto.SignUpRequest{
				Email:    testUnverifiedEmail,
				Password: "newPass123",
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T) {
				var otp models.Otp
				err := db.DB.Where("email = ? AND purpose = ?", testUnverifiedEmail, constants.OTP_PURPOSE_SIGNUP).First(&otp).Error
				assert.NoError(t, err)
				assert.NotEmpty(t, otp.OTP)
			},
		},
		{
			name: "Missing email",
			req: &proto.SignUpRequest{
				Password: "pass123",
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Missing password",
			req: &proto.SignUpRequest{
				Email: "test@example.com",
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			resp, err := client.SignUp(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t)
				}
			}

			clearTables(t)
		})
	}
}

func TestVerifySignup(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.VerificationRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T)
	}{
		{
			name: "Successful verification",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_SIGNUP, false)
			},
			req: &proto.VerificationRequest{
				Email: testUnverifiedEmail,
				Otp:   validOTP,
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T) {
				var user models.User
				err := db.DB.Where("email = ?", testUnverifiedEmail).First(&user).Error
				assert.NoError(t, err)
				assert.True(t, user.Verified)

				var otp models.Otp
				err = db.DB.Where("email = ?", testUnverifiedEmail).First(&otp).Error
				assert.Error(t, err)
			},
		},
		{
			name: "Expired OTP",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_SIGNUP, true)
			},
			req: &proto.VerificationRequest{
				Email: testUnverifiedEmail,
				Otp:   validOTP,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid OTP",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_SIGNUP, false)
			},
			req: &proto.VerificationRequest{
				Email: testUnverifiedEmail,
				Otp:   invalidOTP,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Missing email",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_SIGNUP, false)
			},
			req: &proto.VerificationRequest{
				Otp: validOTP,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Missing OTP",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_SIGNUP, false)
			},
			req: &proto.VerificationRequest{
				Email: testUnverifiedEmail,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Non-existent email",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_SIGNUP, false)
			},
			req: &proto.VerificationRequest{
				Email: "nonexistent@example.com",
				Otp:   validOTP,
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			resp, err := client.VerifySignup(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t)
				}
			}

			clearTables(t)
		})
	}
}

func TestForgotPassword(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.ForgotPasswordRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T)
	}{
		{
			name: "Success - OTP created",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
			},
			req: &proto.ForgotPasswordRequest{
				Email: testVerifiedEmail,
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T) {
				var otp models.Otp
				err := db.DB.Where("email = ? AND purpose = ?",
					testVerifiedEmail,
					constants.OTP_PURPOSE_FORGOT_PASSWORD,
				).First(&otp).Error
				assert.NoError(t, err)
				assert.NotEmpty(t, otp.OTP)
				assert.True(t, otp.OTPExpiresAt.After(time.Now()))
			},
		},
		{
			name: "Unverified user",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
			},
			req: &proto.ForgotPasswordRequest{
				Email: testUnverifiedEmail,
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Non-existent email",
			req: &proto.ForgotPasswordRequest{
				Email: "nonexistent@example.com",
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Missing email",
			req: &proto.ForgotPasswordRequest{
				Email: "",
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			resp, err := client.ForgotPassword(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t)
				}
			}

			clearTables(t)
		})
	}
}

func TestResetPassword(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.ResetPasswordRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T)
	}{
		{
			name: "Successful password reset",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
				createTestOTP(t, testVerifiedEmail, constants.OTP_PURPOSE_FORGOT_PASSWORD, false)
			},
			req: &proto.ResetPasswordRequest{
				Email:       testVerifiedEmail,
				OldPassword: "testPass123",
				NewPassword: "newPass123",
				Otp:         validOTP,
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T) {
				var user models.User
				err := db.DB.Where("email = ?", testVerifiedEmail).First(&user).Error
				assert.NoError(t, err)
				err = bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte("newPass123"))
				assert.NoError(t, err)

				var otp models.Otp
				err = db.DB.Where("email = ?", testVerifiedEmail).First(&otp).Error
				assert.Error(t, err)
			},
		},
		{
			name: "Expired OTP",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
				createTestOTP(t, testVerifiedEmail, constants.OTP_PURPOSE_FORGOT_PASSWORD, true)
			},
			req: &proto.ResetPasswordRequest{
				Email:       testVerifiedEmail,
				OldPassword: "testPass123",
				NewPassword: "newPass123",
				Otp:         validOTP,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Unverified user",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_FORGOT_PASSWORD, false)
			},
			req: &proto.ResetPasswordRequest{
				Email:       testUnverifiedEmail,
				OldPassword: "testPass123",
				NewPassword: "newPass123",
				Otp:         validOTP,
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Missing fields",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
			},
			req: &proto.ResetPasswordRequest{
				Email: testVerifiedEmail,
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			resp, err := client.ResetPassword(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t)
				}
			}

			clearTables(t)
		})
	}
}

func TestLoginWithPassword(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.LoginWithPasswordRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T, resp *proto.UserResponse, md metadata.MD)
	}{
		{
			name: "Successful login",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
			},
			req: &proto.LoginWithPasswordRequest{
				Email:    testVerifiedEmail,
				Password: "testPass123",
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T, resp *proto.UserResponse, md metadata.MD) {
				assert.Equal(t, testVerifiedEmail, resp.Email)
				assert.Equal(t, "verified", resp.Username)
				assert.NotEmpty(t, md.Get("session-key"))
				assert.NotEmpty(t, md.Get("csrf-token"))
				assert.NotEmpty(t, md.Get("expires-at"))

				sessionKey := md.Get("session-key")[0]
				var session models.Session
				err := db.Sessions.Where("key = ?", sessionKey).First(&session).Error
				assert.NoError(t, err)
				assert.NotEmpty(t, session.Token)
				assert.True(t, session.ExpiresAt.After(time.Now()))
			},
		},
		{
			name: "Missing credentials",
			req: &proto.LoginWithPasswordRequest{
				Email:    "",
				Password: "",
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid password",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
			},
			req: &proto.LoginWithPasswordRequest{
				Email:    testVerifiedEmail,
				Password: "wrongPass",
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Non-existent user",
			req: &proto.LoginWithPasswordRequest{
				Email:    "nonexistent@example.com",
				Password: "testPass123",
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Unverified user",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
			},
			req: &proto.LoginWithPasswordRequest{
				Email:    testUnverifiedEmail,
				Password: "testPass123",
			},
			expectedCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			var md metadata.MD
			resp, err := client.LoginWithPassword(context.Background(), tt.req, grpc.Header(&md))

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t, resp, md)
				}
			}

			clearTables(t)
		})
	}
}

func TestInitiateLoginWithOtp(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.LoginWithOtpRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T)
	}{
		{
			name: "Success - OTP created",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
			},
			req: &proto.LoginWithOtpRequest{
				Email: testVerifiedEmail,
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T) {
				var otpEntry models.Otp
				result := db.DB.Where("email = ? AND purpose = ?",
					testVerifiedEmail,
					constants.OTP_PURPOSE_LOGIN,
				).First(&otpEntry)
				assert.NoError(t, result.Error)
				assert.NotEmpty(t, otpEntry.OTP)
				assert.True(t, otpEntry.OTPExpiresAt.After(time.Now()))
			},
		},
		{
			name: "Missing email",
			req: &proto.LoginWithOtpRequest{
				Email: "",
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Unverified user",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
			},
			req: &proto.LoginWithOtpRequest{
				Email: testUnverifiedEmail,
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Non-existent user",
			req: &proto.LoginWithOtpRequest{
				Email: "nonexistent@example.com",
			},
			expectedCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			resp, err := client.InitiateLoginWithOtp(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t)
				}
			}

			clearTables(t)
		})
	}
}

func TestVerifyLoginOtp(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		req          *proto.VerificationRequest
		expectedCode codes.Code
		validateFunc func(t *testing.T, resp *proto.UserResponse, md metadata.MD)
	}{
		{
			name: "Success - Valid OTP",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
				createTestOTP(t, testVerifiedEmail, constants.OTP_PURPOSE_LOGIN, false)
			},
			req: &proto.VerificationRequest{
				Email: testVerifiedEmail,
				Otp:   validOTP,
			},
			expectedCode: codes.OK,
			validateFunc: func(t *testing.T, resp *proto.UserResponse, md metadata.MD) {
				assert.Equal(t, testVerifiedEmail, resp.Email)
				assert.NotEmpty(t, md.Get("session-key"))
				assert.NotEmpty(t, md.Get("csrf-token"))

				var otpEntry models.Otp
				result := db.DB.Where("email = ?", testVerifiedEmail).First(&otpEntry)
				assert.Error(t, result.Error)
			},
		},
		{
			name: "Invalid OTP",
			setup: func(t *testing.T) {
				createTestUser(t, testVerifiedEmail, true)
				createTestOTP(t, testVerifiedEmail, constants.OTP_PURPOSE_LOGIN, true)
			},
			req: &proto.VerificationRequest{
				Email: testVerifiedEmail,
				Otp:   invalidOTP,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Unverified user",
			setup: func(t *testing.T) {
				createTestUser(t, testUnverifiedEmail, false)
				createTestOTP(t, testUnverifiedEmail, constants.OTP_PURPOSE_LOGIN, false)
			},
			req: &proto.VerificationRequest{
				Email: testUnverifiedEmail,
				Otp:   validOTP,
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Missing fields",
			req: &proto.VerificationRequest{
				Email: "",
				Otp:   "",
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			var md metadata.MD
			resp, err := client.VerifyLoginOtp(context.Background(), tt.req, grpc.Header(&md))

			if tt.expectedCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateFunc != nil {
					tt.validateFunc(t, resp, md)
				}
			}

			clearTables(t)
		})
	}
}
