package services

import (
	"fmt"
	"log"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"pariksha/common/pkg/proto"
	"pariksha/server/internal/config/env"
)

var (
	engineService     *EngineService
	engineServiceOnce sync.Once
)

type EngineService struct {
	client proto.EngineClient
	conn   *grpc.ClientConn
}

func GetEngineService() *EngineService {
	engineServiceOnce.Do(func() {
		engineService = &EngineService{}
		engineService.connect()
	})
	return engineService
}

func (s *EngineService) connect() {
	addr := fmt.Sprintf("%s:%s", env.ENGINE_SERVER_HOST, env.ENGINE_SERVER_PORT)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to engine service: %v", err)
	}

	s.conn = conn
	s.client = proto.NewEngineClient(conn)
}

func (s *EngineService) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func init() {
	GetAuthService()
}

func (s *EngineService) Client() proto.EngineClient {
	return s.client
}
