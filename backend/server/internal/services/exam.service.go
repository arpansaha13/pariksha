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
	examService     *ExamService
	examServiceOnce sync.Once
)

type ExamService struct {
	client proto.ExamServiceClient
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
	s.client = proto.NewExamServiceClient(conn)
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

func (s *ExamService) Client() proto.ExamServiceClient {
	return s.client
}

func (s *ExamService) CreateMetadata(userID int) context.Context {
	md := metadata.New(map[string]string{
		"user_id": strconv.Itoa(userID),
	})
	return metadata.NewOutgoingContext(context.Background(), md)
}
