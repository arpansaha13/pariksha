package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"pariksha/common/pkg/proto"
	"pariksha/engine/internal/config/env"
	"pariksha/engine/internal/modules"
)

func main() {
	port := env.ENGINE_SERVER_PORT
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	server, cleanup, err := modules.New()
	if err != nil {
		log.Fatalf("failed to create engine server: %v", err)
	}
	defer cleanup()

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	proto.RegisterEngineServer(grpcServer, server)

	log.Printf("Engine gRPC server is running on port %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
