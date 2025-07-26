package modules

import (
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/config/db"
	"pariksha/exam/internal/config/env"
	"pariksha/exam/internal/controllers"
	"pariksha/exam/internal/interceptors"
	"pariksha/exam/internal/interservice"
	"pariksha/exam/internal/repositories"
	"pariksha/exam/internal/services"
)

func New() (*examServer, []grpc.UnaryServerInterceptor, func()) {
	dbInst := initDB()
	questionIntSvc := initQuestionInterservice()

	// Initialize repositories
	examRepo := repositories.NewExam(dbInst)
	answerRepo := repositories.NewAnswer(dbInst)
	questionRepo := repositories.NewQuestion(dbInst)
	categoryRepo := repositories.NewCategory(dbInst)
	participantRepo := repositories.NewParticipant(dbInst)
	permissionRepo := repositories.NewPermission(dbInst)

	// Initialize services
	examSvc := services.NewExam(examRepo, participantRepo, permissionRepo)
	answerSvc := services.NewAnswer(answerRepo, questionRepo, participantRepo, questionIntSvc)
	categorySvc := services.NewCategory(categoryRepo, questionIntSvc)
	questionSvc := services.NewQuestion(questionRepo, questionIntSvc)
	participantSvc := services.NewParticipant(participantRepo, permissionRepo)
	evaluationSvc := services.NewEvaluation(answerRepo, participantRepo, questionIntSvc)
	resultSvc := services.NewResult(participantRepo, answerRepo)

	// Initialize interceptors
	intc := []grpc.UnaryServerInterceptor{
		interceptors.SingleQuestionHashInterceptor(questionIntSvc),
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
	}

	return server, intc, cleanup
}

func initDB() *gorm.DB {
	gormConfig := config.GetDevEnvGormConfig()
	if env.GO_ENV == constants.GO_ENV_PROD {
		gormConfig = config.GetDefaultGormConfig()
	}

	var dbInitializer config.DBInitializer = &config.PostgresInitializer{}
	dbInst, err := dbInitializer.Init(
		&config.GormDsnImpl{
			Host:     env.EXAM_DB_HOST,
			Port:     env.EXAM_DB_PORT,
			User:     env.EXAM_DB_USER,
			Password: env.EXAM_DB_PASS,
			Dbname:   env.EXAM_DB_NAME,
			Sslmode:  env.EXAM_DB_SSLMODE,
		},
		gormConfig,
		&db.AutoMigrator{},
	)
	if err != nil {
		log.Panicf("failed to initialize exam database: %v", err)
	}

	return dbInst
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
