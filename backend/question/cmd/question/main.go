package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/constants"
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

	if env.GO_ENV != constants.GO_ENV_TEST {
		err = initDB()
		if err != nil {
			log.Fatal(err.Error())
		}

		controllers.Init()
	}

	grpcServer := grpc.NewServer()
	proto.RegisterQuestionServer(grpcServer, &controllers.QuestionServer{})

	log.Printf("Question gRPC server is running on port %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	defer db.Close()
}

func initDB() error {
	gormConfig := config.GetDevEnvGormConfig()
	if env.GO_ENV == constants.GO_ENV_PROD {
		gormConfig = config.GetDefaultGormConfig()
	}

	migrator := &db.AutoMigrator{}

	err := db.Init(
		&config.GormDsnImpl{
			Host:     env.QUESTION_DB_HOST,
			Port:     env.QUESTION_DB_PORT,
			User:     env.QUESTION_DB_USER,
			Password: env.QUESTION_DB_PASS,
			Dbname:   env.QUESTION_DB_NAME,
			Sslmode:  env.QUESTION_DB_SSLMODE,
		},
		gormConfig,
		migrator,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	return nil
}
