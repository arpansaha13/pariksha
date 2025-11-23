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

	"pariksha/common/pkg/logging"
	"pariksha/common/pkg/proto"
	"pariksha/server/internal/config/env"
)

var (
	examService     *ExamService
	examServiceOnce sync.Once
)

type ExamService struct {
	client proto.ExamClient
	conn   *grpc.ClientConn
}

func GetExamService() *ExamService {
	examServiceOnce.Do(func() {
		examService = &ExamService{}
		examService.connect()
	})
	return examService
}

func (s *ExamService) connect() {
	addr := fmt.Sprintf("%s:%s", env.EXAM_SERVER_HOST, env.EXAM_SERVER_PORT)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to exam service: %v", err)
	}

	s.conn = conn
	s.client = proto.NewExamClient(conn)
}

func (s *ExamService) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func init() {
	GetExamService()
}

func (s *ExamService) Client() proto.ExamClient {
	return s.client
}

// CreateMetadata creates outgoing gRPC metadata including user_id and request_id (if present in ctx).
// It extracts request_id from the provided context (set by the HTTP gateway middleware) and appends
// it to the outgoing metadata so downstream services receive the correlation id.
func (s *ExamService) CreateMetadata(ctx context.Context, userID int64) context.Context {
	mdMap := map[string]string{
		"user_id": strconv.FormatInt(userID, 10),
	}

	if reqID, ok := logging.GetRequestIDFromContext(ctx); ok && reqID != "" {
		mdMap["request_id"] = reqID
	}

	md := metadata.New(mdMap)
	return metadata.NewOutgoingContext(ctx, md)
}
