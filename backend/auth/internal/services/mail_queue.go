package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	"pariksha/auth/internal/config/env"
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/utils"
)

var mailQueueClient *asynq.Client

// InitMailQueue initializes Redis connection for Asynq with given host and port
func InitMailQueue(host, port string) error {
	redisAddr := fmt.Sprintf("%s:%s", host, port)
	mailQueueClient = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

	// Test connection
	if err := mailQueueClient.Close(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	// Recreate client for actual use
	mailQueueClient = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return nil
}

func init() {
	// Skip initialization if in test environment
	if env.GO_ENV == constants.GO_ENV_TEST {
		return
	}

	// Initialize with environment variables for non-test environments
	err := InitMailQueue(env.MAIL_QUEUE_HOST, env.MAIL_QUEUE_PORT)
	utils.FailOnError(err, "Failed to initialize Mail Queue")
}

func CloseMailQueue() {
	if mailQueueClient != nil {
		mailQueueClient.Close()
	}
}

type mailService struct{}

var MailService mailService

// SendVerificationMail enqueues a verification mail task.
func (*mailService) SendVerificationMail(payload *structs.MailRequestVerification) {
	pushToMailQueue(constants.MAIL_QUEUE_TASK_SEND_VERIFICATION, payload)
}

// SendLoginOtpMail enqueues a login OTP mail task.
func (*mailService) SendLoginOtpMail(payload *structs.MailRequestLoginOtp) {
	pushToMailQueue(constants.MAIL_QUEUE_TASK_SEND_LOGIN_OTP, payload)
}

// SendForgotPasswordMail enqueues a forgot password mail task.
func (*mailService) SendForgotPasswordMail(payload *structs.MailRequestForgotPassword) {
	pushToMailQueue(constants.MAIL_QUEUE_TASK_SEND_FORGOT_PASSWORD, payload)
}

// SendResetPasswordMail enqueues a reset password mail task.
func (*mailService) SendResetPasswordMail(payload *structs.MailRequestResetPassword) {
	pushToMailQueue(constants.MAIL_QUEUE_TASK_SEND_RESET_PASSWORD, payload)
}

// pushToMailQueue marshals the payload and enqueues it as an Asynq task.
func pushToMailQueue(taskType string, payload any) {
	taskBytes, err := json.Marshal(payload)
	if err != nil {
		log.Default().Printf("Failed to marshal mail payload: %v", err)
		return
	}

	task := asynq.NewTask(taskType, taskBytes)
	info, err := mailQueueClient.Enqueue(task, asynq.Queue(constants.MAIL_QUEUE_NAME))
	if err != nil {
		log.Default().Printf("Failed to enqueue mail task: %v", err)
		return
	}

	if env.GO_ENV != constants.GO_ENV_TEST {
		log.Default().Printf("Enqueued mail task: id=%s queue=%s", info.ID, info.Queue)
	}
}
