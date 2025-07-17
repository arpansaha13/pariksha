package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/config/env"
	"pariksha/exam/internal/controllers"
	"pariksha/exam/internal/interceptors"
	"pariksha/exam/internal/interservice"
)

func main() {
	port := env.EXAM_SERVER_PORT
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	err = initDB()
	if err != nil {
		log.Fatal(err.Error())
	}

	controllers.Init()

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.SingleQuestionHashInterceptor(),
			interceptors.GeneralExamAuthInterceptor(),
			interceptors.DeleteExamsAuthInterceptor(),
			interceptors.EndExamInterceptor(),
		),
	)
	proto.RegisterExamServer(grpcServer, &controllers.ExamServer{})

	log.Printf("Exam gRPC server is running on port %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	defer db.Close()
	defer interservice.CloseExamQueue()
}

func initDB() error {
	if env.GO_ENV == constants.GO_ENV_TEST {
		return nil
	}

	gormConfig := config.GetDevEnvGormConfig()
	if env.GO_ENV == constants.GO_ENV_PROD {
		gormConfig = config.GetDefaultGormConfig()
	}

	migrator := &db.AutoMigrator{}

	err := db.Init(
		&config.GormDsnImpl{
			Host:     env.DB_HOST,
			Port:     env.DB_PORT,
			User:     env.DB_USER,
			Password: env.DB_PASS,
			Dbname:   env.DB_NAME,
			Sslmode:  env.DB_SSLMODE,
		},
		gormConfig,
		migrator,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	return nil
}
