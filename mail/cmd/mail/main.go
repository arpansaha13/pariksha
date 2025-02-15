package main

import (
	"fmt"
	"log"
	"net"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"github.com/arpansaha13/mail/internal/api"
	"github.com/arpansaha13/mail/internal/config/env"
	"github.com/arpansaha13/mail/internal/services"
)

func main() {
	if godotenv.Load() != nil {
		log.Fatalf("Error loading .env file")
	}

	port := env.MAIL_SERVER_PORT
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	api.RegisterMailServiceServer(grpcServer, &services.MailServiceServer{})

	log.Printf("Mail gRPC server is running on port %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
