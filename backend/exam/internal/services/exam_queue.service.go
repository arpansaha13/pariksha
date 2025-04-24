package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/config/env"
)

var asynqClient *asynq.Client

// InitExamQueue initializes Redis connection for Asynq with given host and port
func InitExamQueue(host, port string) error {
	redisAddr := fmt.Sprintf("%s:%s", host, port)
	asynqClient = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

	// Test connection
	if err := asynqClient.Close(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	// Recreate client for actual use
	asynqClient = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return nil
}

func init() {
	// Skip initialization if in test environment
	if env.GO_ENV == constants.GO_ENV_TEST {
		return
	}

	// Initialize with environment variables for non-test environments
	err := InitExamQueue(env.EXAM_QUEUE_HOST, env.EXAM_QUEUE_PORT)
	utils.FailOnError(err, "Failed to initialize Exam Queue")
}

func CloseExamQueue() {
	if asynqClient != nil {
		asynqClient.Close()
	}
}

func EnqueuePrepareQuestons(payload types.ExamQueuePayload) {
	taskBytes, err := json.Marshal(payload)
	if err != nil {
		log.Default().Printf("Failed to marshal payload: %v", err)
		return
	}

	task := asynq.NewTask(constants.EXAM_QUEUE_TASK_PREPARE_QUESTIONS, taskBytes)

	info, err := asynqClient.Enqueue(task)
	if err != nil {
		log.Default().Printf("Failed to enqueue task: %v", err)
		return
	}

	log.Default().Printf("Enqueued task: id=%s queue=%s", info.ID, info.Queue)
}

func EnqueueAutoEndExam(payload types.AutoEndExamPayload, scheduledEndTime time.Time) {
	taskBytes, err := json.Marshal(payload)
	if err != nil {
		log.Default().Printf("Failed to marshal auto-end payload: %v", err)
		return
	}

	task := asynq.NewTask(constants.EXAM_QUEUE_TASK_AUTO_END, taskBytes)

	// Schedule the task at the exact scheduledEndTime
	info, err := asynqClient.Enqueue(task, asynq.ProcessAt(scheduledEndTime))
	if err != nil {
		log.Default().Printf("Failed to enqueue auto-end task: %v", err)
		return
	}

	log.Default().Printf("Enqueued auto-end task: id=%s queue=%s at=%v", info.ID, info.Queue, scheduledEndTime)
}
