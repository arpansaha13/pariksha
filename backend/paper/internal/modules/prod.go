package modules

import (
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/config/env"
	"pariksha/paper/internal/controllers"
	"pariksha/paper/internal/interceptors"
	"pariksha/paper/internal/interservice"
	"pariksha/paper/internal/repositories"
	"pariksha/paper/internal/services"
)

// Prod initializes modules for production environment with file-based migrations
func Prod() (*paperServer, []grpc.UnaryServerInterceptor, func()) {
	dbInst := initProdDB()
	questionIntSvc := initProdQuestionInterservice()

	// Initialize repositories
	paperRepo := repositories.NewPaper(dbInst)
	paperCatRepo := repositories.NewPaperCategory(dbInst)
	paperPermRepo := repositories.NewPaperPermission(dbInst)
	paperQuestRepo := repositories.NewPaperQuestion(dbInst)

	// Initialize services
	paperSvc := services.NewPaper(paperRepo, paperCatRepo, paperPermRepo, paperQuestRepo, questionIntSvc)
	questionSvc := services.NewQuestion(paperRepo, paperQuestRepo, questionIntSvc)
	categorySvc := services.NewCategory(paperRepo, paperCatRepo, paperQuestRepo, questionIntSvc)

	// Initialize interceptors
	intc := []grpc.UnaryServerInterceptor{
		interceptors.PaperAuthInterceptor(paperPermRepo),
		interceptors.DeletePaperAuthInterceptor(paperPermRepo),
		interceptors.ExamServiceAuthInterceptor(),
	}

	// Initialize controllers
	server := &paperServer{
		paperCtrl:    controllers.NewPaper(paperSvc),
		categoryCtrl: controllers.NewCategory(categorySvc),
		questionCtrl: controllers.NewQuestion(questionSvc),
	}

	cleanup := func() {
		sqlDB, err := dbInst.DB()
		if err != nil {
			log.Panicf("could not get sqlDB")
		}

		sqlDB.Close()
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
