package services

import (
	"context"
	"fmt"
	"log"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/logging"
	"pariksha/common/pkg/proto"
	"pariksha/gateway/internal/config/env"
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
	GetEngineService()
}

func (s *EngineService) Client() proto.EngineClient {
	return s.client
}

// CreateMetadata creates outgoing gRPC metadata including request_id (if present in ctx).
// It extracts request_id from the provided context (set by the HTTP gateway middleware) and appends
// it to the outgoing metadata so downstream services receive the correlation id.
func (s *EngineService) CreateMetadata(ctx context.Context) context.Context {
	mdMap := make(map[string]string)

	if reqID, ok := logging.GetRequestIDFromContext(ctx); ok && reqID != "" {
		mdMap["request_id"] = reqID
	}

	md := metadata.New(mdMap)
	return metadata.NewOutgoingContext(ctx, md)
}
