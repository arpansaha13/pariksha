package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"pariksha/common/pkg/constants"
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

	var server proto.QuestionServer
	var cleanup func()

	// Choose between dev and prod modules based on environment
	if env.GO_ENV == constants.GO_ENV_PROD {
		server, cleanup = modules.Prod()
	} else {
		server, cleanup = modules.Dev()
	}

	grpcServer := grpc.NewServer()
	proto.RegisterQuestionServer(grpcServer, server)

	log.Printf("Question gRPC server is running on port %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	defer cleanup()
	defer grpcServer.Stop()
}
