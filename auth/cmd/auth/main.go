package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/arpansaha13/common/pkg/proto"

	"github.com/arpansaha13/auth/internal/config/db"
	"github.com/arpansaha13/auth/internal/config/env"
	"github.com/arpansaha13/auth/internal/handlers"
	"github.com/arpansaha13/auth/internal/services"
)

func main() {
	port := env.AUTH_SERVER_PORT
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterAuthServiceServer(grpcServer, &handlers.AuthServer{})

	log.Printf("Auth gRPC server is running on port %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	defer closeConnections()
}

func closeConnections() {
	sqlDb, _ := db.DB.DB()
	sqlDb.Close()

	sessionsSqlDb, _ := db.Sessions.DB()
	sessionsSqlDb.Close()

	services.CloseRabbit()
}
