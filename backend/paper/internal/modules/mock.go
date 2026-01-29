package modules

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/mocks"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/config/env"
	"pariksha/paper/internal/controller"
	"pariksha/paper/internal/interservice"
	"pariksha/paper/internal/middleware"
	"pariksha/paper/internal/repository"
	"pariksha/paper/internal/service"
)

func NewMock() (proto.PaperServer, *gorm.DB, []grpc.UnaryServerInterceptor, func()) {
	ctx := context.Background()
	mockQuestionIntSvc := mockQuestionInterservice()

	// Initialize database
	pgAddr, pgCleanup := setupPgContainer(ctx)
	dbInst := mockDB(pgAddr)

	// Initialize repositories
	paperRepo := repository.NewPaper(dbInst)
	paperCatRepo := repository.NewPaperCategory(dbInst)
	paperPermRepo := repository.NewPaperPermission(dbInst)
	paperQuestRepo := repository.NewPaperQuestion(dbInst)

	// Initialize services
	paperSvc := service.NewPaper(paperRepo, paperCatRepo, paperPermRepo, paperQuestRepo, mockQuestionIntSvc)
	questionSvc := service.NewQuestion(paperRepo, paperQuestRepo, mockQuestionIntSvc)
	categorySvc := service.NewCategory(paperRepo, paperCatRepo, paperQuestRepo, mockQuestionIntSvc)

	// Initialize interceptors
	intc := []grpc.UnaryServerInterceptor{
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
		pgCleanup()
	}

	return server, dbInst, intc, cleanup
}

func mockDB(pgAddr string) *gorm.DB {
	var dbInitializer config.DBInitializer = &config.PostgresInitializer{}
	dbInst, err := dbInitializer.Init(&config.GormDsnImpl{
		Addr:     pgAddr,
		User:     env.PAPER_DB_USER,
		Password: env.PAPER_DB_PASS,
		Dbname:   env.PAPER_DB_NAME,
		Sslmode:  "disable",
	},
		config.GetTestEnvGormConfig(),
		&db.AutoMigrator{},
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	return dbInst
}

func setupPgContainer(ctx context.Context) (string, func()) {
	pgContainer, err := test.StartPgContainer(ctx, &test.PgContainerEnv{
		PgUser:     env.PAPER_DB_USER,
		PgPassword: env.PAPER_DB_PASS,
		PgDbname:   env.PAPER_DB_NAME,
	})
	if err != nil {
		log.Fatalf("Failed to setup container: %v", err)
	}

	// Get mapped host and ports
	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")

	pgAddr := fmt.Sprintf("%s:%d", pgHost, pgPort.Int())

	pgCleanup := func() {
		pgContainer.Terminate(ctx)
	}

	return pgAddr, pgCleanup
}

func mockQuestionInterservice() *interservice.Question {
	mockClient := &mocks.QuestionClient{}
	qIntSvc := interservice.NewQuestion(nil, mockClient)
	return qIntSvc
}
