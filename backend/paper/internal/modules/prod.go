package modules

import (
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/logging"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/env"
	"pariksha/paper/internal/controller"
	"pariksha/paper/internal/interservice"
	"pariksha/paper/internal/middleware"
	"pariksha/paper/internal/repository"
	"pariksha/paper/internal/service"
)

// Prod initializes modules for production environment with file-based migrations
func Prod() (proto.PaperServer, []grpc.UnaryServerInterceptor, func()) {
	dbInst := initProdDB()
	questionIntSvc := initProdQuestionInterservice()

	// Initialize logger and logging interceptor for the service
	baseLogger := logging.InitLogger(env.GO_ENV)
	loggingInterceptor := logging.NewLoggingInterceptor(baseLogger)

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
		loggingInterceptor,
		middleware.PaperAuthInterceptor(paperPermRepo),
		middleware.DeletePaperAuthInterceptor(paperPermRepo),
		middleware.ExamServiceAuthInterceptor(),
	}

	// Initialize controllers
	server := controllers.NewServer(paperSvc, categorySvc, questionSvc)

	cleanup := func() {
		sqlDB, err := dbInst.DB()
		if err != nil {
			log.Panicf("could not get sqlDB")
		}

		sqlDB.Close()
		baseLogger.Sync()
	}

	return server, intc, cleanup
}

func initProdDB() *gorm.DB {
	gormConfig := config.GetDefaultGormConfig()

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
		nil,
	)
	if err != nil {
		log.Panicf("failed to initialize paper database: %v", err)
	}

	return dbInst
}

func initProdQuestionInterservice() *interservice.Question {
	addr := fmt.Sprintf("%s:%s", env.QUESTION_SERVER_HOST, env.QUESTION_SERVER_PORT)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Panicf("Failed to connect to question service: %v", err)
	}

	client := proto.NewQuestionClient(conn)
	qIntSvc := interservice.NewQuestion(conn, client)

	return qIntSvc
}
