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
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/question/internal/modules"
)

var (
	client proto.QuestionClient
	dbInst *gorm.DB
)

func TestMain(m *testing.M) {
	mockServer, mockDbInst, cleanup := modules.NewMock()
	dbInst = mockDbInst

	srv, conn := setupGrpcServer(mockServer)
	defer func() {
		conn.Close()
		srv.Stop()
	}()

	client = proto.NewQuestionClient(conn)

	code := m.Run()

	cleanup()
	os.Exit(code)
}

func setupGrpcServer(server proto.QuestionServer) (*grpc.Server, *grpc.ClientConn) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	proto.RegisterQuestionServer(srv, server)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

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

func clearTables(t *testing.T) {
	tables := []string{
		constants.TABLE_BOILERPLATES,
		constants.TABLE_TEST_CASES,
		constants.TABLE_QUESTIONS,
		constants.TABLE_CATEGORIES,
		constants.TABLE_LANGUAGES,
	}
	for _, table := range tables {
		err := dbInst.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error
		require.NoError(t, err)
	}
}
