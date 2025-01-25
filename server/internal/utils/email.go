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
func CreateVerificationMail(to string, otp string, linkHash string, expiresInMinutes int) string {
	smtpName := os.Getenv("SMTP_NAME")
	smtpFrom := os.Getenv("SMTP_FROM")
	subject := "Verify your email address"
	link := fmt.Sprintf("http://localhost:3000/auth/verification/%s", linkHash)

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
			<head>
				<meta charset="UTF-8">
			</head>
			<body>
				<p>Please use the OTP and verification link below to confirm your email.</p>
				<p>OTP: <strong>%s</strong></p>
				<p>Verification link: <a href="%s">%s</a></p>
				<p>The OTP will expire in <strong>%s minutes</strong>. If you did not request this email you can safely ignore it.</p>
			</body>
		</html>
`, otp, link, link, fmt.Sprintf("%d", expiresInMinutes))

	return fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", smtpName, smtpFrom, to, subject, htmlBody)
}
