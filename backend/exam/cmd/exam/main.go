package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/config/env"
	"pariksha/exam/internal/handlers"
	"pariksha/exam/internal/interceptors"
	"pariksha/exam/internal/services"
)

func main() {
	port := env.EXAM_SERVER_PORT
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.GeneralExamAuthInterceptor(),
			interceptors.DeleteExamsAuthInterceptor(),
			interceptors.EndExamInterceptor(),
		),
	)
	proto.RegisterExamServer(grpcServer, &handlers.ExamServer{})

	log.Printf("Exam gRPC server is running on port %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	defer closeConnections()
}

func closeConnections() {
	sqlDb, _ := db.DB.DB()
	sqlDb.Close()

	services.CloseExamQueue()
}
