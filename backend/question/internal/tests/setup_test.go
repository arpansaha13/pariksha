package tests

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
	"pariksha/question/internal/config/db"
	"pariksha/question/internal/config/env"
	"pariksha/question/internal/controllers"
)

const (
	bufSize = 1024 * 1024
)

var (
	lis    *bufconn.Listener
	ctx    context.Context
	client proto.QuestionClient
)

func TestMain(m *testing.M) {
	ctx = context.Background()
	cleanup := setupContainers()

	controllers.Init()

	srv, conn := setupGrpcServer()
	defer func() {
		conn.Close()
		srv.Stop()
	}()

	client = proto.NewQuestionClient(conn)

	code := m.Run()

	cleanup()
	os.Exit(code)
}

func setupContainers() func() {
	pgContainer, err := test.StartPgContainer(ctx, &test.PgContainerEnv{
		PgUser:     env.QUESTION_DB_USER,
		PgPassword: env.QUESTION_DB_PASS,
		PgDbname:   env.QUESTION_DB_NAME,
	})
	if err != nil {
		log.Fatalf("Failed to setup container: %v", err)
	}

	// Get mapped host and ports
	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")

	// Initialize connections with container
	err = db.Init(&config.GormDsnImpl{
		Host:     pgHost,
		Port:     pgPort.Port(),
		User:     env.QUESTION_DB_USER,
		Password: env.QUESTION_DB_PASS,
		Dbname:   env.QUESTION_DB_NAME,
		Sslmode:  "disable",
	},
		config.GetTestEnvGormConfig(),
		&db.AutoMigrator{},
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	cleanup := func() {
		pgContainer.Terminate(ctx)
	}

	return cleanup
}

func clearTables(t *testing.T) {
	tables := []string{
		constants.TABLE_BOILERPLATES,
		constants.TABLE_TEST_CASES,
		constants.TABLE_QUESTIONS,
		constants.TABLE_CATEGORIES,
		constants.TABLE_LANGUAGES,
	}
	for _, table := range tables {
		err := db.DB.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error
		require.NoError(t, err)
	}
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func setupGrpcServer() (*grpc.Server, *grpc.ClientConn) {
	lis = bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	proto.RegisterQuestionServer(srv, &controllers.QuestionServer{})

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
