package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/env"
	"pariksha/paper/internal/modules"
)

func main() {
	port := env.PAPER_SERVER_PORT
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	server, intc, cleanup := modules.New()
	defer cleanup()

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(intc...))
	defer grpcServer.Stop()

	proto.RegisterPaperServer(grpcServer, server)

	log.Printf("Paper gRPC server is running on port %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
