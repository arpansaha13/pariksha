package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"pariksha/common/pkg/proto"
	"pariksha/question/internal/config/db"
	"pariksha/question/internal/config/env"
	"pariksha/question/internal/controllers"
)

func main() {
	port := env.QUESTION_SERVER_PORT
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterQuestionServer(grpcServer, &controllers.QuestionServer{})

	log.Printf("Question gRPC server is running on port %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	defer closeConnections()
}

func closeConnections() {
	sqlDb, _ := db.DB.DB()
	sqlDb.Close()
}
