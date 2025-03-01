package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"pariksha/common/pkg/types"
	"pariksha/mail/internal/config/env"
)

func SendVerificationMail(body []byte) {
	var payload types.MailRequestVerification
	err := json.Unmarshal(body, &payload)

	if err != nil {
		log.Default().Println(err)
		return
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
}

func SendLoginOtpMail(body []byte) {
	var payload types.MailRequestLoginOtp
	err := json.Unmarshal(body, &payload)

	if err != nil {
		log.Default().Println(err)
		return
	}

	subject := "Login OTP"

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
			<head>
				<meta charset="UTF-8">
			</head>
			<body>
				<p>Here is your OTP to login to your account.</p>
				<p>OTP: <strong>%s</strong></p>
				<p>The OTP will expire in <strong>%d minutes</strong>. If you did not request this login, please secure your account.</p>
			</body>
		</html>
	`, payload.Otp, payload.ExpiresInMinutes)

	template := fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", env.SMTP_NAME, env.SMTP_FROM, payload.To, subject, htmlBody)

	sendMail(payload.To, template)
}

func SendForgotPasswordMail(body []byte) {
	var payload types.MailRequestLoginOtp
	err := json.Unmarshal(body, &payload)

	if err != nil {
		log.Default().Println(err)
		return
	}

	subject := "Reset Your Password"

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
			<head>
					<meta charset="UTF-8">
			</head>
			<body>
					<p>Here is your OTP to reset your password.</p>
					<p>OTP: <strong>%s</strong></p>
					<p>The OTP will expire in <strong>%d minutes</strong>. If you did not request this, please secure your account.</p>
			</body>
		</html>
	`, payload.Otp, payload.ExpiresInMinutes)

	template := fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", env.SMTP_NAME, env.SMTP_FROM, payload.To, subject, htmlBody)

	sendMail(payload.To, template)
}

func SendResetPasswordMail(body []byte) {
	var payload types.MailRequestResetPassword
	err := json.Unmarshal(body, &payload)

	if err != nil {
		log.Default().Println(err)
		return
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
}

func sendMail(to string, content string) error {
	// user := env.SMTP_USER
	// from := env.SMTP_FROM
	// password := env.SMTP_PASSWORD
	// smtpHost := env.SMTP_HOST
	// smtpPort := env.SMTP_PORT

	log.Default().Println(content)

	// auth := smtp.PlainAuth("", user, password, smtpHost)

	// err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(content))
	// if err != nil {
	// 	fmt.Println("Failed to send email: ", err)
	// 	return err
	// }

	return nil
}
