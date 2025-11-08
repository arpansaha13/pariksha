package modules

import (
	"context"
	"fmt"
	"log"

	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/test"
	"pariksha/question/internal/config/db"
	"pariksha/question/internal/config/env"
	"pariksha/question/internal/controllers"
	"pariksha/question/internal/repositories"
	"pariksha/question/internal/services"
)

func NewMock() (*questionServer, *gorm.DB, func()) {
	ctx := context.Background()

	// Initialize database
	pgAddr, pgCleanup := setupPgContainer(ctx)
	dbInst := initMockDB(pgAddr)

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
		pgCleanup()
	}

	return server, dbInst, cleanup
}

func initMockDB(pgAddr string) *gorm.DB {
	var dbInitializer config.DBInitializer = &config.PostgresInitializer{}
	dbInst, err := dbInitializer.Init(&config.GormDsnImpl{
		Addr:     pgAddr,
		User:     env.QUESTION_DB_USER,
		Password: env.QUESTION_DB_PASS,
		Dbname:   env.QUESTION_DB_NAME,
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
		PgUser:     env.QUESTION_DB_USER,
		PgPassword: env.QUESTION_DB_PASS,
		PgDbname:   env.QUESTION_DB_NAME,
	})
	if err != nil {
		log.Fatalf("Failed to setup container: %v", err)
	}

	// Get mapped host and ports
	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")

	pgAddr := fmt.Sprintf("%s:%d", pgHost, pgPort.Int())

	cleanup := func() {
		pgContainer.Terminate(ctx)
	}

	return pgAddr, cleanup
}
