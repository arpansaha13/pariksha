package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	"pariksha/common/pkg/structs"
	"pariksha/workers/mail/internal/config/env"
)

// SendVerificationMail handles the verification mail Asynq task.
func SendVerificationMail(ctx context.Context, task *asynq.Task) error {
	var payload structs.MailRequestVerification
	err := json.Unmarshal(task.Payload(), &payload)

	if err != nil {
		log.Default().Println(err)
		return nil
	}

	subject := "Verify your email address"

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
			<head>
				<meta charset="UTF-8">
			</head>
			<body>
				<p>Please use the OTP below to confirm your email.</p>
				<p>OTP: <strong>%s</strong></p>
				<p>The OTP will expire in <strong>%s minutes</strong>. If you did not request this email you can safely ignore it.</p>
			</body>
		</html>
	`, payload.Otp, fmt.Sprintf("%d", payload.ExpiresInMinutes))

	template := fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", env.SMTP_NAME, env.SMTP_FROM, payload.To, subject, htmlBody)

	sendMail(payload.To, template)
	return nil
}
