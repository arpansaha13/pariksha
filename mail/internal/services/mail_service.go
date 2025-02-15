package services

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/arpansaha13/mail/internal/api"
	"github.com/arpansaha13/mail/internal/config/env"
)

type MailServiceServer struct {
	api.UnimplementedMailServiceServer
}

func (s *MailServiceServer) SendVerificationMail(ctx context.Context, req *api.SendVerificationMailRequest) (*api.MailResponse, error) {
	template := createVerificationMailTemplate(req.To, req.Otp, int(req.ExpiresInMinutes))

	return sendMail(req.To, template)
}

func (s *MailServiceServer) SendForgotPasswordMail(ctx context.Context, req *api.SendForgotPasswordMailRequest) (*api.MailResponse, error) {
	template := createForgotPasswordMailTemplate(req.To, req.Otp, int(req.ExpiresInMinutes))

	return sendMail(req.To, template)
}

func (s *MailServiceServer) SendLoginOtpMail(ctx context.Context, req *api.SendLoginOtpMailRequest) (*api.MailResponse, error) {
	template := createLoginOtpMailTemplate(req.To, req.Otp, int(req.ExpiresInMinutes))

	return sendMail(req.To, template)
}

func (s *MailServiceServer) SendResetPasswordSuccessMail(ctx context.Context, req *api.SendResetPasswordSuccessMailRequest) (*api.MailResponse, error) {
	template := createResetPasswordSuccessMailTemplate(req.To)

	return sendMail(req.To, template)
}

func sendMail(to string, content string) (*api.MailResponse, error) {
	user := env.SMTP_USER
	from := env.SMTP_FROM
	password := env.SMTP_PASSWORD
	smtpHost := env.SMTP_HOST
	smtpPort := env.SMTP_PORT

	auth := smtp.PlainAuth("", user, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(content))
	if err != nil {
		fmt.Println(err)
		return &api.MailResponse{Success: false, Error: "failed to send email"}, fmt.Errorf("failed to send email: %w", err)
	}

	return &api.MailResponse{Success: true}, nil
}

func createVerificationMailTemplate(to string, otp string, expiresInMinutes int) string {
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
		"%s\r\n", env.SMTP_NAME, env.SMTP_FROM, to, subject, htmlBody)
}

func createLoginOtpMailTemplate(to string, otp string, expiresInMinutes int) string {
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
		"%s\r\n", env.SMTP_NAME, env.SMTP_FROM, to, subject, htmlBody)
}

func createForgotPasswordMailTemplate(to string, otp string, expiresInMinutes int) string {
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
    `, otp, expiresInMinutes)

	return fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", env.SMTP_NAME, env.SMTP_FROM, to, subject, htmlBody)
}

func createResetPasswordSuccessMailTemplate(to string) string {
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

	return fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", env.SMTP_NAME, env.SMTP_FROM, to, subject, htmlBody)
}
