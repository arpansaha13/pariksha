package services

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/proto"
	"pariksha/server/internal/config/env"
)

var (
	paperService     *PaperService
	paperServiceOnce sync.Once
)

type PaperService struct {
	client proto.PaperClient
	conn   *grpc.ClientConn
}

func GetPaperService() *PaperService {
	paperServiceOnce.Do(func() {
		paperService = &PaperService{}
		paperService.connect()
	})
	return paperService
}

func (s *PaperService) connect() {
	addr := fmt.Sprintf("%s:%s", env.PAPER_SERVER_HOST, env.PAPER_SERVER_PORT)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to paper service: %v", err)
	}

	s.conn = conn
	s.client = proto.NewPaperClient(conn)
}

func (s *PaperService) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func init() {
	GetPaperService()
}

func (s *PaperService) Client() proto.PaperClient {
	return s.client
}

func (s *PaperService) CreateMetadata(userID int64) context.Context {
	md := metadata.New(map[string]string{
		"user_id": strconv.FormatInt(userID, 10),
	})
	return metadata.NewOutgoingContext(context.Background(), md)
}
