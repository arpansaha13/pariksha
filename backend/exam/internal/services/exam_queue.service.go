package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	rabbit "github.com/rabbitmq/amqp091-go"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/exam/internal/config/env"
)

var rabbitConn *rabbit.Connection
var rabbitCh *rabbit.Channel
var examQueue rabbit.Queue

// InitRabbitMQ initializes RabbitMQ connection with given host and port
func InitRabbitMQ(host, port string) error {
	var err error
	var rabbitAddr = host + ":" + port

	rabbitConn, err = rabbit.Dial(fmt.Sprintf("amqp://guest:guest@%s/", rabbitAddr))
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	rabbitCh, err = rabbitConn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %v", err)
	}

	examQueue, err = rabbitCh.QueueDeclare(
		constants.RABBIT_EXAM_QUEUE_NAME,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %v", err)
	}

	return nil
}

func init() {
	// Skip initialization if in test environment
	if env.GO_ENV == constants.GO_ENV_TEST {
		return
	}

	// Initialize with environment variables for non-test environments
	err := InitRabbitMQ(env.EXAM_QUEUE_HOST, env.EXAM_QUEUE_PORT)
	utils.FailOnError(err, "Failed to initialize RabbitMQ")
}

func CloseExamQueue() {
	if rabbitConn != nil {
		rabbitConn.Close()
	}
	if rabbitCh != nil {
		rabbitCh.Close()
	}
}

func PushToExamQueue(payload types.ExamQueuePayload) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	byteArray, err := json.Marshal(payload)

	if err != nil {
		log.Default().Println("Failed to marshal message: ", err)
	}

	err = rabbitCh.PublishWithContext(
		ctx,
		"",
		examQueue.Name,
		false,
		false,
		rabbit.Publishing{
			ContentType: "text/plain",
			Body:        byteArray,
		},
	)

	if err != nil {
		log.Default().Println("Failed to publish a message: ", err)
	}
}
