package interservice

import (
	"fmt"
	"log"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"pariksha/common/pkg/proto"
	"pariksha/engine/internal/config/env"
)

var (
	pSvc     *paperService
	pSvcOnce sync.Once
)

type paperService struct {
	client proto.PaperClient
	conn   *grpc.ClientConn
}

func ensurePaperService() {
	pSvcOnce.Do(func() {
		pSvc = &paperService{}
		addr := fmt.Sprintf("%s:%s", env.PAPER_SERVER_HOST, env.PAPER_SERVER_PORT)
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect to paper service: %v", err)
		}

		pSvc.conn = conn
		pSvc.client = proto.NewPaperClient(conn)
	})
}

func init() {
	ensurePaperService()
}
