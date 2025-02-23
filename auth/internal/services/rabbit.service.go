package services

import (
	"fmt"

	rabbit "github.com/rabbitmq/amqp091-go"

	"github.com/arpansaha13/auth/internal/config/env"
	"github.com/arpansaha13/common/pkg/constants"
	"github.com/arpansaha13/common/pkg/utils"
)

var rabbitConn *rabbit.Connection
var rabbitCh *rabbit.Channel
var mailQueue rabbit.Queue

func init() {
	var err error
	var rabbitAddr = env.RABBIT_SERVER_HOST + ":" + env.RABBIT_SERVER_PORT

	rabbitConn, err = rabbit.Dial(fmt.Sprintf("amqp://guest:guest@%s/", rabbitAddr))
	utils.FailOnError(err, "Failed to connect to RabbitMQ")

	rabbitCh, err = rabbitConn.Channel()
	utils.FailOnError(err, "Failed to open a channel")

	mailQueue, err = rabbitCh.QueueDeclare(
		constants.RABBIT_MAIL_QUEUE_NAME,
		false,
		false,
		false,
		false,
		nil,
	)
	utils.FailOnError(err, "Failed to declare a queue")
}

func CloseRabbit() {
	rabbitConn.Close()
	rabbitCh.Close()
}
