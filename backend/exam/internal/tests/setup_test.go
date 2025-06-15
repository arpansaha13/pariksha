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
	"pariksha/common/pkg/utils/generate"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/config/env"
	"pariksha/exam/internal/handlers"
	"pariksha/exam/internal/interceptors"
	"pariksha/exam/internal/services"
	"pariksha/exam/internal/services/paper"
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
	proto.RegisterExamServer(srv, &handlers.ExamServer{})

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
	originalFetchHashes := paper.FetchQuestionHashesForIds
	originalFetchIDs := paper.FetchQuestionIdsForHashes
	originalFetchQuestions := paper.FetchQuestionsByIds

	// List of question IDs to use in tests
	testQuestionIDs := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 999}

	// Generate hash mappings at runtime
	idToHash := make(map[int64]string)
	hashToID := make(map[string]int64)

	// Sample raw question content
	rawQuestion := json.RawMessage(`{
		"title": "Mock Question",
		"description": "This is a mock question for testing",
		"options": ["A", "B", "C", "D"]
	}`)

	// Generate hashes for test question IDs
	for _, id := range testQuestionIDs {
		hash := fmt.Sprintf("q_%s", generate.HMACHash(id))
		idToHash[id] = hash
		hashToID[hash] = id
	}

	// Mock FetchQuestionHashesForIds
	paper.FetchQuestionHashesForIds = func(questionIDs []int64) ([]string, error) {
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

	// Mock FetchQuestionIdsForHashes
	paper.FetchQuestionIdsForHashes = func(hashes []string) ([]int64, error) {
		ids := make([]int64, len(hashes))
		for i, hash := range hashes {
			if id, exists := hashToID[hash]; exists {
				ids[i] = id
			} else {
				return nil, fmt.Errorf("question hash %s not found", hash)
			}
		}
		return ids, nil
	}

	// Mock FetchQuestionsByIds
	paper.FetchQuestionsByIds = func(questionIDs []int64) ([]*proto.QuestionBatchItem, error) {
		questions := make([]*proto.QuestionBatchItem, len(questionIDs))
		for i, id := range questionIDs {
			hash, exists := idToHash[id]
			if !exists {
				return nil, fmt.Errorf("question ID %d not found", id)
			}
			questions[i] = &proto.QuestionBatchItem{
				QuestionId:   id,
				QuestionHash: hash,
				RawQuestion:  rawQuestion,
				MaxScore:     10,
				Type:         proto.QuestionType_MCQ,
			}
		}
		return questions, nil
	}

	return func() {
		paper.FetchQuestionHashesForIds = originalFetchHashes
		paper.FetchQuestionIdsForHashes = originalFetchIDs
		paper.FetchQuestionsByIds = originalFetchQuestions
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
