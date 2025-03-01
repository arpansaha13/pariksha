package types

type MailRequestVerification struct {
	To               string
	Otp              string
	ExpiresInMinutes int
}

type MailRequestLoginOtp struct {
	To               string
	Otp              string
	ExpiresInMinutes int
}

type MailRequestForgotPassword struct {
	To               string
	Otp              string
	ExpiresInMinutes int
}

type MailRequestResetPassword struct {
	To string
}
