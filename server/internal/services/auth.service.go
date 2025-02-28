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
	authService     *AuthService
	authServiceOnce sync.Once
)

type AuthService struct {
	client proto.AuthServiceClient
	conn   *grpc.ClientConn
}

// GetAuthService returns a singleton instance of AuthService
func GetAuthService() *AuthService {
	authServiceOnce.Do(func() {
		authService = &AuthService{}
		authService.connect()
	})
	return authService
}

func (s *AuthService) connect() {
	addr := fmt.Sprintf("%s:%s", env.AUTH_SERVER_HOST, env.AUTH_SERVER_PORT)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %v", err)
	}

	s.conn = conn
	s.client = proto.NewAuthServiceClient(conn)
}

// Close closes the gRPC connection
func (s *AuthService) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func init() {
	GetAuthService()
}

// Client returns the gRPC client for direct access if needed
func (s *AuthService) Client() proto.AuthServiceClient {
	return s.client
}
