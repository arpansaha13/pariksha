package modules

import (
	"log"

	"google.golang.org/grpc"
	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/logging"
	"pariksha/question/internal/config/env"
	"pariksha/question/internal/controllers"
	"pariksha/question/internal/repositories"
	"pariksha/question/internal/services"
)

// Prod initializes modules for production environment with file-based migrations
func Prod() (*questionServer, []grpc.UnaryServerInterceptor, func()) {
	dbInst := initProdDB()

	// Initialize repositories
	questionRepo := repositories.NewQuestion(dbInst)
	categoryRepo := repositories.NewCategory(dbInst)
	boilerplateRepo := repositories.NewBoilerplate(dbInst)
	testcaseRepo := repositories.NewTestcase(dbInst)

	// Initialize services
	questionSvc := services.NewQuestion(questionRepo, boilerplateRepo, testcaseRepo)
	categorySvc := services.NewCategory(categoryRepo)

	// Initialize logger and logging interceptor
	baseLogger := logging.InitLogger(env.GO_ENV)
	loggingInterceptor := logging.NewLoggingInterceptor(baseLogger)

	// Initialize controllers
	server := &questionServer{
		questionCtrl: controllers.NewQuestion(questionSvc),
		categoryCtrl: controllers.NewCategory(categorySvc),
	}

	// Initialize interceptors
	intc := []grpc.UnaryServerInterceptor{
		loggingInterceptor,
	}

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
			Addr:     env.QUESTION_DB_ADDR,
			User:     env.QUESTION_DB_USER,
			Password: env.QUESTION_DB_PASS,
			Dbname:   env.QUESTION_DB_NAME,
			Sslmode:  env.QUESTION_DB_SSLMODE,
		},
		gormConfig,
		nil,
	)
	if err != nil {
		log.Panicf("failed to initialize question database: %v", err)
	}

	return dbInst
}
