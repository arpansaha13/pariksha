package tests

import (
	"context"
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
	"pariksha/paper/internal/modules"
)

var (
	client proto.PaperClient
	dbInst *gorm.DB
)

func TestMain(m *testing.M) {
	mockServer, mockDbInst, mockIntc, cleanup := modules.NewMock()
	defer cleanup()
	dbInst = mockDbInst

	srv, conn := setupGrpcServer(mockServer, mockIntc)
	defer func() {
		conn.Close()
		srv.Stop()
	}()

	client = proto.NewPaperClient(conn)

	code := m.Run()

	os.Exit(code)
}

func setupGrpcServer(server proto.PaperServer, mockIntc []grpc.UnaryServerInterceptor) (*grpc.Server, *grpc.ClientConn) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(mockIntc...))
	proto.RegisterPaperServer(srv, server)

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
		constants.TABLE_PAPER_QUESTIONS,
		constants.TABLE_PAPER_CATEGORIES,
		constants.TABLE_PAPER_PERMISSIONS,
		constants.TABLE_PAPERS,
	}

	for i := range tables {
		err := dbInst.Exec("TRUNCATE TABLE " + string(tables[i]) + " CASCADE").Error
		require.NoError(t, err)
	}
}
