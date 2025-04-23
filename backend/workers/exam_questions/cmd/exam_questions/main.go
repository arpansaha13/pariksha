package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/utils"
	"pariksha/workers/exam_questions/internal/config/env"
	"pariksha/workers/exam_questions/internal/handlers"
)

func main() {
	redisAddr := fmt.Sprintf("%s:%s", env.EXAM_QUEUE_HOST, env.EXAM_QUEUE_PORT)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 10},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(constants.EXAM_QUEUE_TASK_START_EXAM, func(ctx context.Context, task *asynq.Task) error {
		handlers.PrepareExamQuestions(task.Payload())
		return nil
	})

	if err := srv.Run(mux); err != nil {
		utils.FailOnError(err, "Failed to run asynq server")
	}

	log.Printf(" [*] Running exam questions worker. To exit press CTRL+C")
}
