package handlers

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/config/env"
)

const (
	bufSize = 1024 * 1024
	userID  = 1
)

var (
	lis       *bufconn.Listener
	ctx       context.Context
	client    proto.PaperServiceClient
	testUsers map[int]models.User // Store test users for reuse
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

	cleanup := func() {
		pgContainer.Terminate(ctx)
	}

	return cleanup
}

func createTestUsers(t *testing.T) {
	testUsers = make(map[int]models.User)

	users := []models.User{
		{
			ID:       1, // matches userID constant
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
	// Clear paper-related tables only, preserve users
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
		Name:    "Category 1",
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
	cleanup := setupContainer()

	// Setup gRPC server and client
	srv, conn := setupGrpcServer()
	defer func() {
		conn.Close()
		srv.Stop()
	}()

	client = proto.NewPaperServiceClient(conn)

	// Create test users before running tests
	createTestUsers(&testing.T{})

	// Run tests
	code := m.Run()

	// Cleanup users
	if err := db.DB.Exec("TRUNCATE TABLE users CASCADE").Error; err != nil {
		log.Printf("Failed to cleanup users: %v", err)
	}

	cleanup()
	os.Exit(code)
}

func TestGetUserPapers(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.PaperList)
	}{
		{
			name: "Success - Get user papers",
			setup: func(t *testing.T) {
				createTestPaper(t, int(userID))
				createTestPaper(t, int(userID))
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperList) {
				assert.Equal(t, 2, len(resp.Papers))
				for _, paper := range resp.Papers {
					assert.NotEmpty(t, paper.Title)
					assert.Equal(t, constants.PAPER_OWNERSHIP_TYPE_OWNER, paper.Ownership.Type)
				}
			},
		},
		{
			name:         "Success - No papers",
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperList) {
				assert.Equal(t, 0, len(resp.Papers))
			},
		},
		{
			name:         "Invalid user ID",
			userID:       0,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.GetUserPapers(ctx, &proto.Empty{})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)

			clearTables(t)
		})
	}
}

func TestCreatePaper(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.PaperResponse)
	}{
		{
			name:         "Success - Create paper",
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperResponse) {
				assert.NotZero(t, resp.Id)
				assert.Equal(t, "Untitled Paper", resp.Title)
				assert.Equal(t, constants.PAPER_OWNERSHIP_TYPE_OWNER, resp.Ownership.Type)

				// Verify default category was created
				var categories []models.QuestionCategory
				err := db.DB.Where("paper_id = ?", resp.Id).Find(&categories).Error
				require.NoError(t, err)
				assert.Equal(t, 1, len(categories))
				assert.Equal(t, "Category 1", categories[0].Name)
			},
		},
		{
			name:         "Invalid user ID",
			userID:       0,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.CreatePaper(ctx, &proto.Empty{})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)

			clearTables(t)
		})
	}
}

func TestUpdatePaper(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Paper
		userID       int32
		title        string
		expectedCode codes.Code
		validate     func(t *testing.T, paper *models.Paper)
	}{
		{
			name: "Success - Update title",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, int(userID))
				return &paper
			},
			userID:       userID,
			title:        "Updated Title",
			expectedCode: codes.OK,
			validate: func(t *testing.T, paper *models.Paper) {
				var updated models.Paper
				err := db.DB.First(&updated, paper.ID).Error
				require.NoError(t, err)
				assert.Equal(t, "Updated Title", updated.Title)
			},
		},
		{
			name: "Paper not found",
			setup: func(t *testing.T) *models.Paper {
				return &models.Paper{ID: 999}
			},
			userID:       userID,
			title:        "Updated Title",
			expectedCode: codes.NotFound,
		},
		{
			name: "Invalid user ID",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, int(userID))
				return &paper
			},
			userID:       0,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paper := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			_, err := client.UpdatePaper(ctx, &proto.UpdatePaperRequest{
				PaperId: int32(paper.ID),
				Title:   tt.title,
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, paper)
			}

			clearTables(t)
		})
	}
}

func TestGetPaper(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) *models.Paper
		userID       int32
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.PaperResponse)
	}{
		{
			name: "Success - Get owned paper",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, int(userID))
				return &paper
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperResponse) {
				assert.NotZero(t, resp.Id)
				assert.Equal(t, "Test Paper", resp.Title)
				assert.Equal(t, constants.PAPER_OWNERSHIP_TYPE_OWNER, resp.Ownership.Type)
				assert.Equal(t, int32(60), resp.DurationMinutes)
			},
		},
		{
			name: "Success - Get shared paper",
			setup: func(t *testing.T) *models.Paper {
				// Create paper owned by user 2
				paper := createTestPaper(t, 2)
				// Add shared access for test user
				ownership := models.PaperOwnership{
					UserID:  int(userID),
					PaperID: paper.ID,
					Type:    constants.PAPER_OWNERSHIP_TYPE_SHARED,
				}
				err := db.DB.Create(&ownership).Error
				require.NoError(t, err)
				return &paper
			},
			userID:       userID,
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.PaperResponse) {
				assert.NotZero(t, resp.Id)
				assert.Equal(t, constants.PAPER_OWNERSHIP_TYPE_SHARED, resp.Ownership.Type)
			},
		},
		{
			name: "Paper not found",
			setup: func(t *testing.T) *models.Paper {
				return &models.Paper{ID: 999}
			},
			userID:       userID,
			expectedCode: codes.NotFound,
		},
		{
			name: "No access to paper",
			setup: func(t *testing.T) *models.Paper {
				// Create paper owned by user 2 with no sharing
				paper := createTestPaper(t, 2)
				return &paper
			},
			userID:       userID,
			expectedCode: codes.NotFound,
		},
		{
			name: "Invalid user ID",
			setup: func(t *testing.T) *models.Paper {
				paper := createTestPaper(t, int(userID))
				return &paper
			},
			userID:       0,
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearTables(t)
			paper := tt.setup(t)

			ctx := createContextWithUserID(tt.userID)
			resp, err := client.GetPaper(ctx, &proto.PaperRequest{
				PaperId: int32(paper.ID),
			})

			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)
		})
	}
}
