package main

import (
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/utils"
	"pariksha/workers/exam/internal/config/db"
	"pariksha/workers/exam/internal/config/env"
	"pariksha/workers/exam/internal/handlers"
)

func main() {
	err := initDB()
	if err != nil {
		log.Fatal(err.Error())
	}

	redisAddr := fmt.Sprintf("%s:%s", env.EXAM_QUEUE_HOST, env.EXAM_QUEUE_PORT)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				constants.EXAM_QUEUE_NAME: 3,
			}},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(constants.EXAM_QUEUE_TASK_PREPARE_QUESTIONS, handlers.PrepareExamQuestions)
	mux.HandleFunc(constants.EXAM_QUEUE_TASK_AUTO_END, handlers.AutoEndExam)
	mux.HandleFunc(constants.EXAM_QUEUE_TASK_DELETE_EXAMS, handlers.PostDeleteExamsCleanup)

	if err := srv.Run(mux); err != nil {
		utils.FailOnError(err, "Failed to run asynq server")
	}

	log.Printf("[*] Running exam questions worker. To exit press CTRL+C")

	defer db.Close()
}

func initDB() error {
	if env.GO_ENV == constants.GO_ENV_TEST {
		return nil
	}

	gormConfig := config.GetDevEnvGormConfig()
	if env.GO_ENV == constants.GO_ENV_PROD {
		gormConfig = config.GetDefaultGormConfig()
	}

	err := db.Init(
		&config.GormDsnImpl{
			Addr:     env.EXAM_DB_ADDR,
			User:     env.EXAM_DB_USER,
			Password: env.EXAM_DB_PASS,
			Dbname:   env.EXAM_DB_NAME,
			Sslmode:  env.EXAM_DB_SSLMODE,
		},
		gormConfig,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	return nil
}
