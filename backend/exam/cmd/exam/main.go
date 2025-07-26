package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/env"
	"pariksha/exam/internal/modules"
)

func main() {
	port := env.EXAM_SERVER_PORT
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	server, intc, cleanup := modules.New()
	defer cleanup()

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(intc...))
	defer grpcServer.Stop()

	proto.RegisterExamServer(grpcServer, server)

	log.Printf("Exam gRPC server is running on port %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
