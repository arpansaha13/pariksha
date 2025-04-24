package main

import (
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/utils"
	"pariksha/workers/exam/internal/config/env"
	"pariksha/workers/exam/internal/handlers"
)

func main() {
	redisAddr := fmt.Sprintf("%s:%s", env.EXAM_QUEUE_HOST, env.EXAM_QUEUE_PORT)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 10},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(constants.EXAM_QUEUE_TASK_PREPARE_QUESTIONS, handlers.PrepareExamQuestions)
	mux.HandleFunc(constants.EXAM_QUEUE_TASK_AUTO_END, handlers.AutoEndExam)

	if err := srv.Run(mux); err != nil {
		utils.FailOnError(err, "Failed to run asynq server")
	}

	log.Printf(" [*] Running exam questions worker. To exit press CTRL+C")
}
