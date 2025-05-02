package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/config/env"
	"pariksha/paper/internal/handlers"
	"pariksha/paper/internal/interceptors"
)

func main() {
	port := env.PAPER_SERVER_PORT
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.PaperAuthInterceptor()),
	)
	proto.RegisterPaperServiceServer(grpcServer, &handlers.PaperServer{})

	log.Printf("Paper gRPC server is running on port %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	defer closeConnections()
}

func closeConnections() {
	sqlDb, _ := db.DB.DB()
	sqlDb.Close()
}
