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

// SendResetPasswordMail handles the reset password mail Asynq task.
func SendResetPasswordMail(ctx context.Context, task *asynq.Task) error {
	var payload structs.MailRequestResetPassword
	err := json.Unmarshal(task.Payload(), &payload)

	if err != nil {
		log.Default().Println(err)
		return nil
	}

	subject := "Password Reset Successful"

	htmlBody := `
		<!DOCTYPE html>
		<html>
			<head>
					<meta charset="UTF-8">
			</head>
			<body>
					<p>Your password has been successfully reset.</p>
					<p>If you did not request this change, please secure your account immediately.</p>
			</body>
		</html>
	`

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
