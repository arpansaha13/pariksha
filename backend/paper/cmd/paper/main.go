package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/config/env"
	"pariksha/paper/internal/controllers"
	"pariksha/paper/internal/interceptors"
)

func main() {
	port := env.PAPER_SERVER_PORT
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
			interceptors.ExamServiceAuthInterceptor(),
			interceptors.PaperAuthInterceptor(),
			interceptors.DeletePaperAuthInterceptor(),
		),
	)
	proto.RegisterPaperServer(grpcServer, &controllers.PaperServer{})

	log.Printf("Paper gRPC server is running on port %s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	defer db.Close()
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
			Host:     env.PAPER_DB_HOST,
			Port:     env.PAPER_DB_PORT,
			User:     env.PAPER_DB_USER,
			Password: env.PAPER_DB_PASS,
			Dbname:   env.PAPER_DB_NAME,
			Sslmode:  env.PAPER_DB_SSLMODE,
		},
		gormConfig,
		migrator,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	return nil
}
