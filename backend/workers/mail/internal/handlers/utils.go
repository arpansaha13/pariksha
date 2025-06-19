package handlers

import (
	"fmt"
	"log"
	"net/smtp"

	"pariksha/common/pkg/constants"
	"pariksha/workers/mail/internal/config/env"
)

func sendMail(to string, content string) error {
	if env.GO_ENV == constants.GO_ENV_DEV || env.GO_ENV == constants.GO_ENV_TEST {
		log.Default().Println(content)
		return nil
	}

	user := env.SMTP_USER
	from := env.SMTP_FROM
	password := env.SMTP_PASSWORD
	smtpHost := env.SMTP_HOST
	smtpPort := env.SMTP_PORT

	auth := smtp.PlainAuth("", user, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(content))
	if err != nil {
		fmt.Println("Failed to send email: ", err)
		return err
	}

	return nil
}
