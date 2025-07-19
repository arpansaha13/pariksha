package modules

import (
	"log"

	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/constants"
	"pariksha/question/internal/config/db"
	"pariksha/question/internal/config/env"
	"pariksha/question/internal/controllers"
	"pariksha/question/internal/repositories"
	"pariksha/question/internal/services"
)

func New() (*questionServer, func()) {
	dbInst := initDB()

	// Initialize repositories
	questionRepo := repositories.NewQuestion(dbInst)
	categoryRepo := repositories.NewCategory(dbInst)
	boilerplateRepo := repositories.NewBoilerplate(dbInst)
	testcaseRepo := repositories.NewTestcase(dbInst)

	// Initialize services
	questionSvc := services.NewQuestion(questionRepo, boilerplateRepo, testcaseRepo)
	categorySvc := services.NewCategory(categoryRepo)

	// Initialize controllers
	server := &questionServer{
		questionCtrl: controllers.NewQuestion(questionSvc),
		categoryCtrl: controllers.NewCategory(categorySvc),
	}

	cleanup := func() {
		sqlDB, err := dbInst.DB()
		if err != nil {
			log.Panicf("could not get sqlDB")
		}

		sqlDB.Close()
	}

	return server, cleanup
}

func initDB() *gorm.DB {
	gormConfig := config.GetDevEnvGormConfig()
	if env.GO_ENV == constants.GO_ENV_PROD {
		gormConfig = config.GetDefaultGormConfig()
	}

	var dbInitializer config.DBInitializer = &config.PostgresInitializer{}
	dbInst, err := dbInitializer.Init(
		&config.GormDsnImpl{
			Host:     env.QUESTION_DB_HOST,
			Port:     env.QUESTION_DB_PORT,
			User:     env.QUESTION_DB_USER,
			Password: env.QUESTION_DB_PASS,
			Dbname:   env.QUESTION_DB_NAME,
			Sslmode:  env.QUESTION_DB_SSLMODE,
		},
		gormConfig,
		&db.AutoMigrator{},
	)
	if err != nil {
		log.Panicf("failed to initialize question database: %v", err)
	}

	return dbInst
}
