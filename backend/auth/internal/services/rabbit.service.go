package services

import (
	"fmt"

	rabbit "github.com/rabbitmq/amqp091-go"

	"pariksha/auth/internal/config/env"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/utils"
)

var rabbitConn *rabbit.Connection
var rabbitCh *rabbit.Channel
var mailQueue rabbit.Queue

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

	mailQueue, err = rabbitCh.QueueDeclare(
		constants.RABBIT_MAIL_QUEUE_NAME,
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
	err := InitRabbitMQ(env.RABBIT_SERVER_HOST, env.RABBIT_SERVER_PORT)
	utils.FailOnError(err, "Failed to initialize RabbitMQ")
}

func CloseRabbit() {
	if rabbitConn != nil {
		rabbitConn.Close()
	}
	if rabbitCh != nil {
		rabbitCh.Close()
	}
}
