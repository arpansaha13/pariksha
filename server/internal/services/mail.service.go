package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	rabbit "github.com/rabbitmq/amqp091-go"

	"github.com/arpansaha13/common/pkg/constants"
	"github.com/arpansaha13/common/pkg/types"
)

type mailService struct{}

var MailService mailService

func (*mailService) SendVerificationMail(payload *types.MailRequestVerification) {
	pushToMailQueue(constants.MAIL_TYPE_VERIFICATION, payload)
}

func (*mailService) SendLoginOtpMail(payload *types.MailRequestLoginOtp) {
	pushToMailQueue(constants.MAIL_TYPE_LOGIN_OTP, payload)
}

func (*mailService) SendForgotPasswordMail(payload *types.MailRequestForgotPassword) {
	pushToMailQueue(constants.MAIL_TYPE_FORGOT_PASSWORD, payload)
}

func (*mailService) SendResetPasswordMail(payload *types.MailRequestResetPassword) {
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
