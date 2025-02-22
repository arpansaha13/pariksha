package main

import (
	"fmt"
	"log"

	rabbit "github.com/rabbitmq/amqp091-go"

	"github.com/arpansaha13/common/pkg/constants"
	"github.com/arpansaha13/common/pkg/utils"
	"github.com/arpansaha13/mail/internal/config/env"
	"github.com/arpansaha13/mail/internal/router"
)

func main() {
	var rabbitAddr = env.RABBIT_SERVER_HOST + ":" + env.RABBIT_SERVER_PORT

	conn, err := rabbit.Dial(fmt.Sprintf("amqp://guest:guest@%s/", rabbitAddr))
	utils.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	utils.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		constants.RABBIT_MAIL_QUEUE_NAME,
		false,
		false,
		false,
		false,
		nil,
	)
	utils.FailOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	utils.FailOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			router.RouteMailRequest(d)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	<-forever
}
