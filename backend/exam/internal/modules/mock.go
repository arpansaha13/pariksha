package modules

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/mocks"
	"pariksha/common/pkg/test"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/config/env"
	"pariksha/exam/internal/controllers"
	"pariksha/exam/internal/interceptors"
	"pariksha/exam/internal/interservice"
	"pariksha/exam/internal/repositories"
	"pariksha/exam/internal/services"
)

func NewMock() (*examServer, *gorm.DB, *interservice.Question, []grpc.UnaryServerInterceptor, func()) {
	ctx := context.Background()
	mockQuestionIntSvc := mockQuestionInterservice()

	redisCleanup := mockExamQueueInterservice(ctx)

	// Initialize database
	pgAddr, pgCleanup := setupPgContainer(ctx)
	dbInst := mockDB(pgAddr)

	// Initialize repositories
	examRepo := repositories.NewExam(dbInst)
	answerRepo := repositories.NewAnswer(dbInst)
	questionRepo := repositories.NewQuestion(dbInst)
	categoryRepo := repositories.NewCategory(dbInst)
	participantRepo := repositories.NewParticipant(dbInst)
	permissionRepo := repositories.NewPermission(dbInst)

	// Initialize services
	examSvc := services.NewExam(examRepo, participantRepo, permissionRepo)
	answerSvc := services.NewAnswer(answerRepo, questionRepo, participantRepo, mockQuestionIntSvc)
	categorySvc := services.NewCategory(categoryRepo, mockQuestionIntSvc)
	questionSvc := services.NewQuestion(questionRepo, mockQuestionIntSvc)
	participantSvc := services.NewParticipant(participantRepo, permissionRepo)
	evaluationSvc := services.NewEvaluation(answerRepo, participantRepo, mockQuestionIntSvc)
	resultSvc := services.NewResult(participantRepo, answerRepo)

	// Initialize interceptors
	intc := []grpc.UnaryServerInterceptor{
		interceptors.SingleQuestionHashInterceptor(mockQuestionIntSvc),
		interceptors.GeneralExamAuthInterceptor(examRepo, permissionRepo, dbInst),
		interceptors.DeleteExamsAuthInterceptor(permissionRepo),
		interceptors.EndExamInterceptor(participantRepo),
	}

	// Initialize controllers
	server := &examServer{
		examCtrl:        controllers.NewExam(examSvc),
		answerCtrl:      controllers.NewAnswer(answerSvc),
		categoryCtrl:    controllers.NewCategory(categorySvc),
		questionCtrl:    controllers.NewQuestion(questionSvc),
		participantCtrl: controllers.NewParticipant(participantSvc),
		evaluationCtrl:  controllers.NewEvaluation(evaluationSvc),
		resultCtrl:      controllers.NewResult(resultSvc),
	}

	cleanup := func() {
		sqlDB, err := dbInst.DB()
		if err != nil {
			log.Panicf("could not get sqlDB")
		}

		sqlDB.Close()
		pgCleanup()
		redisCleanup()
	}

	return server, dbInst, mockQuestionIntSvc, intc, cleanup
}

func mockDB(pgAddr string) *gorm.DB {
	var dbInitializer config.DBInitializer = &config.PostgresInitializer{}
	dbInst, err := dbInitializer.Init(&config.GormDsnImpl{
		Addr:     pgAddr,
		User:     env.EXAM_DB_USER,
		Password: env.EXAM_DB_PASS,
		Dbname:   env.EXAM_DB_NAME,
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
		PgUser:     env.EXAM_DB_USER,
		PgPassword: env.EXAM_DB_PASS,
		PgDbname:   env.EXAM_DB_NAME,
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

func mockExamQueueInterservice(ctx context.Context) func() {
	// Start Redis container
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections"),
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("Failed to setup Redis container: %v", err)
	}

	redisHost, _ := redisContainer.Host(ctx)
	redisPort, _ := redisContainer.MappedPort(ctx, "6379")

	// Initialize Redis connection
	err = interservice.InitExamQueue(redisHost, redisPort.Port())
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	return func() {
		redisContainer.Terminate(ctx)
	}
}
