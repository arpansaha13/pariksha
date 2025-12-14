package modules

import (
	"github.com/docker/docker/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/mocks"
	"pariksha/engine/internal/controllers"
	"pariksha/engine/internal/interservice"
	"pariksha/engine/internal/runner"
	"pariksha/engine/internal/services"
)

func NewMock() (*engineServer, *interservice.Question, error) {
	mockQuestionIntSvc := mockQuestionInterservice()

	// Ensure DOCKER_API_VERSION = 1.46
	// Error response from daemon: client version 1.48 is too new. Maximum supported API version is 1.46
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "failed to create Docker client: %v", err)
	}

	// Initialize runners
	nodeRunner := runner.NewNode(cli)

	// Initialize services
	engineSvc := services.NewEngine(mockQuestionIntSvc, nodeRunner)

	// Initialize controllers
	server := &engineServer{
		engineCtrl: controllers.NewEngine(engineSvc),
	}

	return server, mockQuestionIntSvc, nil
}

func mockQuestionInterservice() *interservice.Question {
	mockClient := &mocks.QuestionClient{}
	qIntSvc := interservice.NewQuestion(nil, mockClient)
	return qIntSvc
}
