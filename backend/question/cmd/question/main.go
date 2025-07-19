package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"pariksha/common/pkg/proto"
	"pariksha/question/internal/config/env"
	"pariksha/question/internal/modules"
)

func main() {
	port := env.QUESTION_SERVER_PORT
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server, cleanup := modules.New()

	grpcServer := grpc.NewServer()
	proto.RegisterQuestionServer(grpcServer, server)

	log.Printf("Question gRPC server is running on port %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	defer cleanup()
	defer grpcServer.Stop()
}
