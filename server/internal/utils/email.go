package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(to string, content string) error {
	user := os.Getenv("SMTP_USER")
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", user, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(content))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// CreateHTMLEmailTemplate generates an HTML formatted email message
func CreateVerificationMail(to string, otp string, expiresInMinutes int) string {
	smtpName := os.Getenv("SMTP_NAME")
	smtpFrom := os.Getenv("SMTP_FROM")
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
	`, otp, fmt.Sprintf("%d", expiresInMinutes))

	return fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", smtpName, smtpFrom, to, subject, htmlBody)
}

func CreateLoginOtpMail(to string, otp string, expiresInMinutes int) string {
	smtpName := os.Getenv("SMTP_NAME")
	smtpFrom := os.Getenv("SMTP_FROM")
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
	`, otp, expiresInMinutes)

	return fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", smtpName, smtpFrom, to, subject, htmlBody)
}
