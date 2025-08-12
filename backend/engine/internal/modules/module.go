package modules

import (
	"fmt"
	"log"

	"github.com/docker/docker/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/proto"
	"pariksha/engine/internal/config/env"
	"pariksha/engine/internal/controllers"
	"pariksha/engine/internal/interservice"
	"pariksha/engine/internal/runner"
	"pariksha/engine/internal/services"
)

func New() (*engineServer, func(), error) {
	questionIntSvc := initQuestionInterservice()

	// Ensure DOCKER_API_VERSION = 1.46
	// Error response from daemon: client version 1.48 is too new. Maximum supported API version is 1.46
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "failed to create Docker client: %v", err)
	}

	// Initialize runners
	nodeRunner := runner.NewNode(cli)

	// Initialize services
	engineSvc := services.NewEngine(questionIntSvc, nodeRunner)

	// Initialize controllers
	server := engineServer{
		engineCtrl: controllers.NewEngine(engineSvc),
	}

	cleanup := func() {
		questionIntSvc.Close()
	}

	return &server, cleanup, nil
}

func initQuestionInterservice() *interservice.Question {
	addr := fmt.Sprintf("%s:%s", env.QUESTION_SERVER_HOST, env.QUESTION_SERVER_PORT)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Panicf("Failed to connect to question service: %v", err)
	}

	client := proto.NewQuestionClient(conn)
	qIntSvc := interservice.NewQuestion(conn, client)

	return qIntSvc
}
