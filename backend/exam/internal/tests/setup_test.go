package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils/generate"
	"pariksha/common/pkg/utils/ptr"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/config/env"
	"pariksha/exam/internal/controllers"
	"pariksha/exam/internal/interceptors"
	"pariksha/exam/internal/interservice"
)

const (
	bufSize = 1024 * 1024
)

var (
	lis    *bufconn.Listener
	ctx    context.Context
	client proto.ExamClient
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
	err = interservice.InitExamQueue(redisHost, redisPort.Port())
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	return func() {
		interservice.CloseExamQueue()
		pgContainer.Terminate(ctx)
		redisContainer.Terminate(ctx)
	}
}

func clearTables(t *testing.T) {
	tables := []string{
		constants.TABLE_EXAM_PARTICIPANTS,
		constants.TABLE_EXAMS,
		constants.TABLE_EXAM_PERMISSIONS,
		constants.TABLE_EXAM_QUESTIONS,
		constants.TABLE_EXAM_CATEGORIES,
		constants.TABLE_ANSWERS,
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
			interceptors.SingleQuestionHashInterceptor(),
			interceptors.GeneralExamAuthInterceptor(),
			interceptors.DeleteExamsAuthInterceptor(),
			interceptors.EndExamInterceptor(),
		),
	)

	// Initialize all controllers before registering server
	controllers.InitializeHandlers()
	proto.RegisterExamServer(srv, &controllers.ExamServer{})

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

// mockPaperService replaces the actual paper service with test doubles
func mockPaperService() func() {
	originalFetchHashes := interservice.GetQuestionHashesByIds
	originalFetchIDs := interservice.GetQuestionIDsByHashes
	originalFetchQuestionsByIDs := interservice.GetQuestionsByIDs
	originalFetchQuestionByHash := interservice.GetQuestionByHash
	originalFetchCategoriesByIds := interservice.GetCategoriesByIDs

	// List of question IDs to use in tests
	testQuestionIDs := []types.QuestionID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 999}

	// Generate hash mappings at runtime
	idToHash := make(map[types.QuestionID]string)
	hashToID := make(map[string]types.QuestionID)

	// Sample raw question content
	rawQuestion := json.RawMessage(`{
		"title": "Mock Question",
		"description": "This is a mock question for testing",
		"options": ["A", "B", "C", "D"]
	}`)

	// Generate hashes for test question IDs
	for _, id := range testQuestionIDs {
		hash := fmt.Sprintf("q_%s", generate.HMACHash(int64(id)))
		idToHash[id] = hash
		hashToID[hash] = id
	}

	// Mock GetQuestionHashesByIds
	interservice.GetQuestionHashesByIds = func(questionIDs []types.QuestionID) ([]string, error) {
		hashes := make([]string, len(questionIDs))
		for i, id := range questionIDs {
			if hash, exists := idToHash[id]; exists {
				hashes[i] = hash
			} else {
				return nil, fmt.Errorf("question ID %d not found", id)
			}
		}
		return hashes, nil
	}

	// Mock GetQuestionIDsByHashes
	interservice.GetQuestionIDsByHashes = func(hashes []string) ([]types.QuestionID, error) {
		ids := make([]types.QuestionID, len(hashes))
		for i, hash := range hashes {
			if id, exists := hashToID[hash]; exists {
				ids[i] = id
			} else {
				return nil, fmt.Errorf("question hash %s not found", hash)
			}
		}
		return ids, nil
	}

	// Mock GetQuestionsByIDs
	interservice.GetQuestionsByIDs = func(typedQuestionIDs []types.QuestionID) ([]*proto.QuestionResponse, error) {
		questions := make([]*proto.QuestionResponse, len(typedQuestionIDs))
		for i, typedID := range typedQuestionIDs {
			hash, exists := idToHash[typedID]
			if !exists {
				return nil, fmt.Errorf("question ID %d not found", int64(typedID))
			}
			questions[i] = &proto.QuestionResponse{
				Id:          int64(typedID),
				Hash:        hash,
				RawQuestion: rawQuestion,
				Type:        proto.QuestionType_MCQ,
			}
		}
		return questions, nil
	}

	// Mock GetQuestionByHash
	interservice.GetQuestionByHash = func(questionHash string) (*proto.QuestionResponse, error) {
		return &proto.QuestionResponse{
			Hash:        questionHash,
			RawQuestion: rawQuestion,
			Type:        proto.QuestionType_MCQ,
			TestCases: []*proto.CodingQuestionTestCase{
				{
					Inputs:      []string{"1", "2"},
					Output:      "3",
					Explanation: ptr.String("First test case"),
					Hidden:      false,
				},
				{
					Inputs: []string{"4", "5"},
					Output: "9",
					Hidden: true,
				},
			},
		}, nil
	}

	// Mock GetCategoriesByIDs
	interservice.GetCategoriesByIDs = func(typedCategoryIDs []types.CategoryID) ([]*proto.CategoryResponse, error) {
		questions := make([]*proto.CategoryResponse, len(typedCategoryIDs))
		for i, typedID := range typedCategoryIDs {
			questions[i] = &proto.CategoryResponse{
				Id:   int64(typedID),
				Name: fmt.Sprintf("Category %d", i+1),
			}
		}
		return questions, nil
	}

	return func() {
		interservice.GetQuestionHashesByIds = originalFetchHashes
		interservice.GetQuestionIDsByHashes = originalFetchIDs
		interservice.GetQuestionsByIDs = originalFetchQuestionsByIDs
		interservice.GetQuestionByHash = originalFetchQuestionByHash
		interservice.GetCategoriesByIDs = originalFetchCategoriesByIds
	}
}

func TestMain(m *testing.M) {
	cleanup := setupContainer()
	paperServiceCleanup := mockPaperService()

	srv, conn := setupGrpcServer()
	defer func() {
		conn.Close()
		srv.Stop()
	}()

	client = proto.NewExamClient(conn)

	code := m.Run()

	paperServiceCleanup()
	cleanup()
	os.Exit(code)
}
