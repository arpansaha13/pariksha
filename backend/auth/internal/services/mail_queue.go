package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	rabbit "github.com/rabbitmq/amqp091-go"

	"pariksha/auth/internal/config/env"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/utils"
)

var (
	rabbitConn *rabbit.Connection
	rabbitCh   *rabbit.Channel
	mailQueue  rabbit.Queue
)

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

type mailService struct{}

var MailService mailService

func (*mailService) SendVerificationMail(payload *structs.MailRequestVerification) {
	pushToMailQueue(constants.MAIL_TYPE_VERIFICATION, payload)
}

func (*mailService) SendLoginOtpMail(payload *structs.MailRequestLoginOtp) {
	pushToMailQueue(constants.MAIL_TYPE_LOGIN_OTP, payload)
}

func (*mailService) SendForgotPasswordMail(payload *structs.MailRequestForgotPassword) {
	pushToMailQueue(constants.MAIL_TYPE_FORGOT_PASSWORD, payload)
}

func (*mailService) SendResetPasswordMail(payload *structs.MailRequestResetPassword) {
	pushToMailQueue(constants.MAIL_TYPE_RESET_PASSWORD, payload)
}

func pushToMailQueue(mailType string, payload any) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	byteArray, err := json.Marshal(payload)

	if err != nil {
		log.Default().Println("Failed to marshal message: ", err)
	}

	err = rabbitCh.PublishWithContext(
		ctx,
		"",
		mailQueue.Name,
		false,
		false,
		rabbit.Publishing{
			ContentType: "text/plain",
			Body:        byteArray,
			Type:        mailType,
		},
	)

	if err != nil {
		log.Default().Println("Failed to publish a message: ", err)
	}
}
