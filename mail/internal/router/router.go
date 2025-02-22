package router

import (
	"log"

	rabbit "github.com/rabbitmq/amqp091-go"

	"github.com/arpansaha13/common/pkg/constants"
	"github.com/arpansaha13/mail/internal/handlers"
)

func RouteMailRequest(d rabbit.Delivery) {
	mailType := d.Type
	logger := log.Default()

	switch mailType {
	case constants.MAIL_TYPE_VERIFICATION:
		handlers.SendVerificationMail(d.Body)
		break
	case constants.MAIL_TYPE_FORGOT_PASSWORD:
		handlers.SendForgotPasswordMail(d.Body)
		break
	case constants.MAIL_TYPE_LOGIN_OTP:
		handlers.SendLoginOtpMail(d.Body)
		break
	case constants.MAIL_TYPE_RESET_PASSWORD:
		handlers.SendResetPasswordMail(d.Body)
		break
	default:
		logger.Println("Invalid Mail Type: ", mailType)
	}
}
