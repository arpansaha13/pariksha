package main

import (
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/utils"
	"pariksha/workers/mail/internal/config/env"
	"pariksha/workers/mail/internal/handlers"
)

func main() {
	redisAddr := fmt.Sprintf("%s:%s", env.MAIL_QUEUE_HOST, env.MAIL_QUEUE_PORT)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				constants.MAIL_QUEUE_NAME: 3,
			}},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(constants.MAIL_QUEUE_TASK_SEND_LOGIN_OTP, handlers.SendLoginOtpMail)
	mux.HandleFunc(constants.MAIL_QUEUE_TASK_SEND_FORGOT_PASSWORD, handlers.SendForgotPasswordMail)
	mux.HandleFunc(constants.MAIL_QUEUE_TASK_SEND_RESET_PASSWORD, handlers.SendResetPasswordMail)
	mux.HandleFunc(constants.MAIL_QUEUE_TASK_SEND_VERIFICATION, handlers.SendVerificationMail)

	if err := srv.Run(mux); err != nil {
		utils.FailOnError(err, "Failed to run asynq server")
	}

	log.Printf(" [*] Running exam questions worker. To exit press CTRL+C")
}
