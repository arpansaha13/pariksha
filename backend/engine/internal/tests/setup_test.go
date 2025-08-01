package tests

import (
	"context"
	"log"
	"net"
	"os"
	"testing"

	"pariksha/common/pkg/proto"
	"pariksha/engine/internal/interservice"
	"pariksha/engine/internal/modules"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var (
	client         proto.EngineClient
	questionIntSvc *interservice.Question
)

func TestMain(m *testing.M) {
	mockServer, mockQuestionIntSvc, err := modules.NewMock()
	if err != nil {
		log.Fatalf("Could not create mock engine module: %v", err)
	}

	questionIntSvc = mockQuestionIntSvc

	srv, conn := setupGrpcServer(mockServer)
	defer func() {
		conn.Close()
		srv.Stop()
	}()

	client = proto.NewEngineClient(conn)

	code := m.Run()

	os.Exit(code)
}

func setupGrpcServer(server proto.EngineServer) (*grpc.Server, *grpc.ClientConn) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	proto.RegisterEngineServer(srv, server)

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
