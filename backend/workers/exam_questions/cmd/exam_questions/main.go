package main

import (
	"fmt"
	"log"

	rabbit "github.com/rabbitmq/amqp091-go"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/utils"
	"pariksha/workers/exam_questions/internal/config/env"
	"pariksha/workers/exam_questions/internal/handlers"
)

func main() {
	var rabbitAddr = env.EXAM_QUEUE_HOST + ":" + env.EXAM_QUEUE_PORT

	conn, err := rabbit.Dial(fmt.Sprintf("amqp://guest:guest@%s/", rabbitAddr))
	utils.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	utils.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		constants.RABBIT_EXAM_QUEUE_NAME,
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
			handlers.PrepareExamQuestions(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	<-forever
}
