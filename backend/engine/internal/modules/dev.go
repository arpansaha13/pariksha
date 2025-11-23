package modules

import (
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/logging"
	"pariksha/common/pkg/proto"
	"pariksha/engine/internal/config/env"
	"pariksha/engine/internal/controllers"
	"pariksha/engine/internal/interservice"
	"pariksha/engine/internal/runner"
	"pariksha/engine/internal/services"
)

func Dev() (*engineServer, []grpc.UnaryServerInterceptor, func(), error) {
	questionIntSvc := initQuestionInterservice()

	// Initialize Kubernetes runner
	kubernetesRunner, err := runner.NewKubernetes()
	if err != nil {
		return nil, nil, nil, status.Errorf(codes.Internal, "failed to create Kubernetes runner: %v", err)
	}

	// Initialize services
	engineSvc := services.NewEngine(questionIntSvc, kubernetesRunner)

	// Initialize controllers
	server := engineServer{
		engineCtrl: controllers.NewEngine(engineSvc),
	}

	// Initialize logger and logging interceptor
	baseLogger := logging.InitLogger(env.GO_ENV)
	loggingInterceptor := logging.NewLoggingInterceptor(baseLogger)

	// Create interceptors slice
	intc := []grpc.UnaryServerInterceptor{
		loggingInterceptor,
	}

	cleanup := func() {
		questionIntSvc.Close()
		baseLogger.Sync()
	}

	return &server, intc, cleanup, nil
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
