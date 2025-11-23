package main

import (
	"fmt"
	"log"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/logging"
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

	var server proto.ExamServer
	var intc []grpc.UnaryServerInterceptor
	var cleanup func()

	// Choose between dev and prod modules based on environment
	if env.GO_ENV == constants.GO_ENV_PROD {
		server, intc, cleanup = modules.Prod()
	} else {
		server, intc, cleanup = modules.Dev()
	}
	defer cleanup()

	// logger is initialized in the modules; get the package logger
	baseLogger := logging.GetLogger()

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(intc...))
	defer grpcServer.Stop()

	proto.RegisterExamServer(grpcServer, server)

	baseLogger.Info("Exam gRPC server is running", zap.String("port", port))
	if err := grpcServer.Serve(lis); err != nil {
		baseLogger.Fatal("failed to serve", zap.Error(err))
	}
}
