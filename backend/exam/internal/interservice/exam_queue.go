package interservice

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/config/env"
)

var examQueueClient *asynq.Client

// InitExamQueue initializes Redis connection for Asynq with given host and port
func InitExamQueue(host, port string) error {
	redisAddr := fmt.Sprintf("%s:%s", host, port)
	examQueueClient = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

	// Test connection
	if err := examQueueClient.Close(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	// Recreate client for actual use
	examQueueClient = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return nil
}

func init() {
	// Skip initialization if in test environment
	if env.GO_ENV == constants.GO_ENV_TEST {
		return
	}

	err := InitExamQueue(env.EXAM_QUEUE_HOST, env.EXAM_QUEUE_PORT)
	utils.FailOnError(err, "Failed to initialize Exam Queue")
}

func CloseExamQueue() {
	if examQueueClient != nil {
		examQueueClient.Close()
	}
}

// EnqueuePrepareQuestions enqueues a prepare questions task.
func EnqueuePrepareQuestions(payload structs.PrepareQuestionsPayload) {
	pushToExamQueue(constants.EXAM_QUEUE_TASK_PREPARE_QUESTIONS, payload)
}

// EnqueueAutoEndExam enqueues an auto-end exam task at the scheduled end time.
func EnqueueAutoEndExam(payload structs.AutoEndExamPayload, scheduledEndTime time.Time) {
	pushToExamQueue(constants.EXAM_QUEUE_TASK_AUTO_END, payload, asynq.ProcessAt(scheduledEndTime))
}

// EnqueuePostDeleteExamsCleanup enqueues a delete exams cleanup task.
func EnqueuePostDeleteExamsCleanup(examIds []types.ExamID) error {
	payload := structs.DeleteExamsPayload{
		ExamIDs: examIds,
	}
	pushToExamQueue(constants.EXAM_QUEUE_TASK_DELETE_EXAMS, payload)
	return nil
}

// pushToExamQueue marshals the payload and enqueues it as an Asynq task to the exam queue.
func pushToExamQueue(taskType string, payload any, opts ...asynq.Option) {
	taskBytes, err := json.Marshal(payload)
	if err != nil {
		log.Default().Printf("Failed to marshal exam queue payload: %v", err)
		return
	}
	options := append([]asynq.Option{asynq.Queue(constants.EXAM_QUEUE_NAME)}, opts...)
	task := asynq.NewTask(taskType, taskBytes)
	info, err := examQueueClient.Enqueue(task, options...)
	if err != nil {
		log.Default().Printf("Failed to enqueue exam queue task: %v", err)
		return
	}
	if env.GO_ENV != constants.GO_ENV_TEST {
		log.Default().Printf("Enqueued exam queue task: id=%s queue=%s", info.ID, info.Queue)
	}
}
