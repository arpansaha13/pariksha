package modules

import (
	"context"
	"log"

	"github.com/docker/go-connections/nat"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/test"
	"pariksha/paper/internal/config/db"
	"pariksha/paper/internal/config/env"
	"pariksha/paper/internal/controllers"
	"pariksha/paper/internal/interceptors"
	"pariksha/paper/internal/interservice"
	"pariksha/paper/internal/mocks"
	"pariksha/paper/internal/repositories"
	"pariksha/paper/internal/services"
)

func NewMock() (*paperServer, *gorm.DB, []grpc.UnaryServerInterceptor, func()) {
	ctx := context.Background()
	mockQuestionIntSvc := mockQuestionInterservice()

	// Initialize database
	pgHost, pgPort, pgCleanup := setupPgContainer(ctx)
	dbInst := mockDB(pgHost, pgPort)

	// Initialize repositories
	paperRepo := repositories.NewPaper(dbInst)
	paperCatRepo := repositories.NewPaperCategory(dbInst)
	paperPermRepo := repositories.NewPaperPermission(dbInst)
	paperQuestRepo := repositories.NewPaperQuestion(dbInst)

	// Initialize services
	paperSvc := services.NewPaper(paperRepo, paperCatRepo, paperPermRepo, paperQuestRepo, mockQuestionIntSvc)
	questionSvc := services.NewQuestion(paperRepo, paperQuestRepo, mockQuestionIntSvc)
	categorySvc := services.NewCategory(paperRepo, paperCatRepo, paperQuestRepo, mockQuestionIntSvc)

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
		pgCleanup()
	}

	return server, dbInst, intc, cleanup
}

func mockDB(pgHost string, pgPort nat.Port) *gorm.DB {
	var dbInitializer config.DBInitializer = &config.PostgresInitializer{}
	dbInst, err := dbInitializer.Init(&config.GormDsnImpl{
		Host:     pgHost,
		Port:     pgPort.Port(),
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

func setupPgContainer(ctx context.Context) (string, nat.Port, func()) {
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

	pgCleanup := func() {
		pgContainer.Terminate(ctx)
	}

	return pgHost, pgPort, pgCleanup
}

func mockQuestionInterservice() *interservice.Question {
	mockClient := &mocks.QuestionClient{}
	qIntSvc := interservice.NewQuestion(nil, mockClient)
	return qIntSvc
}
