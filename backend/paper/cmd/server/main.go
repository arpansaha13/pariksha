package main

import (
	"fmt"
	"log"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/logging"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/config/env"
	"pariksha/paper/internal/controller"
	"pariksha/paper/internal/interservice"
	"pariksha/paper/internal/middleware"
	"pariksha/paper/internal/repository"
	"pariksha/paper/internal/service"
)

func main() {
	port := env.PAPER_SERVER_PORT
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	baseLogger := logging.InitLogger(env.GO_ENV)
	defer baseLogger.Sync()

	dbInst := initDB()
	defer closeDB(dbInst)

	questionIntSvc := initQuestionInterservice()
	defer questionIntSvc.Close()

	// Initialize repositories
	paperRepo := repository.NewPaper(dbInst)
	paperCatRepo := repository.NewPaperCategory(dbInst)
	paperPermRepo := repository.NewPaperPermission(dbInst)
	paperQuestRepo := repository.NewPaperQuestion(dbInst)

	// Initialize services
	paperSvc := service.NewPaper(paperRepo, paperCatRepo, paperPermRepo, paperQuestRepo, questionIntSvc)
	questionSvc := service.NewQuestion(paperRepo, paperQuestRepo, questionIntSvc)
	categorySvc := service.NewCategory(paperRepo, paperCatRepo, paperQuestRepo, questionIntSvc)

	// Initialize interceptors
	intc := []grpc.UnaryServerInterceptor{
		logging.NewLoggingInterceptor(baseLogger),
		middleware.PaperAuthInterceptor(paperPermRepo),
		middleware.DeletePaperAuthInterceptor(paperPermRepo),
		middleware.ExamServiceAuthInterceptor(),
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(intc...))
	defer grpcServer.Stop()

	controllers.SetupPaperServer(grpcServer, paperSvc, categorySvc, questionSvc)

	baseLogger.Info("Paper gRPC server is running", zap.String("port", port))
	if err := grpcServer.Serve(lis); err != nil {
		baseLogger.Fatal("failed to serve", zap.Error(err))
	}
}

func initDB() *gorm.DB {
	gormConfig := config.GetDevEnvGormConfig()
	var migrator config.DBMigrator = &db.AutoMigrator{}
	if env.GO_ENV == constants.GO_ENV_PROD {
		gormConfig = config.GetDefaultGormConfig()
		migrator = nil
	}

	var dbInitializer config.DBInitializer = &config.PostgresInitializer{}
	dbInst, err := dbInitializer.Init(
		&config.GormDsnImpl{
			Addr:     env.PAPER_DB_ADDR,
			User:     env.PAPER_DB_USER,
			Password: env.PAPER_DB_PASS,
			Dbname:   env.PAPER_DB_NAME,
			Sslmode:  env.PAPER_DB_SSLMODE,
		},
		gormConfig,
		migrator,
	)
	if err != nil {
		log.Fatalf("failed to initialize paper database: %v", err)
	}

	return dbInst
}

func initQuestionInterservice() *interservice.Question {
	addr := fmt.Sprintf("%s:%s", env.QUESTION_SERVER_HOST, env.QUESTION_SERVER_PORT)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to question service: %v", err)
	}

	client := proto.NewQuestionClient(conn)
	return interservice.NewQuestion(conn, client)
}

func closeDB(dbInst *gorm.DB) {
	sqlDB, err := dbInst.DB()
	if err != nil {
		log.Panicf("could not get sqlDB")
	}
	sqlDB.Close()
}
