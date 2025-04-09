package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
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
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/config/env"
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

	// Start RabbitMQ container
	rabbitContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "rabbitmq:3-management-alpine",
			ExposedPorts: []string{"5672/tcp"},
			WaitingFor:   wait.ForLog("Server startup complete"),
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("Failed to setup RabbitMQ container: %v", err)
	}

	// Get container hosts and mapped ports
	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")
	rabbitHost, _ := rabbitContainer.Host(ctx)
	rabbitPort, _ := rabbitContainer.MappedPort(ctx, "5672")

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

	// Initialize RabbitMQ connection
	err = services.InitRabbitMQ(rabbitHost, rabbitPort.Port())
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}

	return func() {
		services.CloseExamQueue()
		pgContainer.Terminate(ctx)
		rabbitContainer.Terminate(ctx)
	}
}

func clearTables(t *testing.T) {
	tables := []string{
		"exam_participants",
		"exams",
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

func createContextWithUserID(userID int64) context.Context {
	md := metadata.New(map[string]string{
		"user_id": strconv.FormatInt(userID, 10),
	})
	return metadata.NewOutgoingContext(context.Background(), md)
}

func createContextWithMetadata(mdMap map[string]string) context.Context {
	md := metadata.New(mdMap)
	return metadata.NewOutgoingContext(context.Background(), md)
}

func setupGrpcServer() (*grpc.Server, *grpc.ClientConn) {
	lis = bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	proto.RegisterExamServiceServer(srv, &ExamServer{})

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

func createTestExam(t *testing.T, createdBy int64) models.Exam {
	exam := models.Exam{
		Title:              "Test Exam",
		CreatedBy:          createdBy,
		Type:               "PRIVATE",
		MaxCandidatesCount: 10,
		PaperID:            1,
		ParticipantCounts:  []byte(`{"unattended":0,"invited":0,"started":0,"ended":0}`),
	}
	require.NoError(t, db.DB.Create(&exam).Error)
	return exam
}

func createTestExamParticipants(t *testing.T, exam *models.Exam, participants []struct {
	UserID int64
	Status int
}) error {
	examParticipants := make([]models.ExamParticipant, len(participants))
	counts, err := exam.GetParticipantCounts()
	if err != nil {
		return err
	}

	// Create participants and update counts
	for i, p := range participants {
		examParticipants[i] = models.ExamParticipant{
			ExamID: exam.ID,
			UserID: p.UserID,
			Status: p.Status,
		}

		// Update counts based on status
		switch p.Status {
		case constants.PARTICIPANT_STATUS_INVITED:
			counts.Invited++
		case constants.PARTICIPANT_STATUS_STARTED:
			counts.Started++
		case constants.PARTICIPANT_STATUS_ENDED:
			counts.Ended++
		case constants.PARTICIPANT_STATUS_UNATTENDED:
			counts.Unattended++
		}
	}

	// Save participants
	if err := db.DB.Create(&examParticipants).Error; err != nil {
		return err
	}

	// Update exam counts
	exam.ParticipantCounts, err = json.Marshal(counts)
	if err != nil {
		return err
	}

	return db.DB.Save(&exam).Error
}

func createTestAnswer(t *testing.T, examParticipant *models.ExamParticipant, questionID int64) models.Answer {
	answer := models.Answer{
		ExamParticipantID: examParticipant.ID,
		QuestionID:        questionID,
		Answer:            sql.NullString{String: "Test Answer", Valid: true},
		Comments:          sql.NullString{String: "Test Comment", Valid: true},
		ScoreAwarded:      5,
		Evaluated:         true,
	}
	require.NoError(t, db.DB.Create(&answer).Error)
	return answer
}
