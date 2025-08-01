package modules

import (
	"github.com/docker/docker/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/mocks"
	"pariksha/engine/internal/controllers"
	"pariksha/engine/internal/interservice"
	"pariksha/engine/internal/services"
)

func NewMock() (*engineServer, error) {
	mockQuestionIntSvc := mockQuestionInterservice()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create Docker client: %v", err)
	}

	// Initialize services
	engineSvc := services.NewEngine(cli, mockQuestionIntSvc)

	// Initialize controllers
	return &engineServer{
		engineCtrl: controllers.NewEngine(engineSvc),
	}, nil
}

func mockQuestionInterservice() *interservice.Question {
	mockClient := &mocks.QuestionClient{}
	qIntSvc := interservice.NewQuestion(nil, mockClient)
	return qIntSvc
}
