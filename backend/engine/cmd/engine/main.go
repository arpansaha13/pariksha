package main

import (
	"fmt"
	"log"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"pariksha/common/pkg/logging"
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

	server, intc, cleanup, err := modules.Dev()
	if err != nil {
		log.Fatalf("failed to create engine server: %v", err)
	}
	defer cleanup()

	// logger is initialized in the modules; get the package logger
	baseLogger := logging.GetLogger()

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(intc...))
	defer grpcServer.Stop()

	proto.RegisterEngineServer(grpcServer, server)

	baseLogger.Info("Engine gRPC server is running", zap.String("port", port))
	if err := grpcServer.Serve(lis); err != nil {
		baseLogger.Fatal("failed to serve", zap.Error(err))
	}
}
