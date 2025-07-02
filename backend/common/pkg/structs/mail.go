package structs

type MailRequestVerification struct {
	To               string
	Otp              string
	ExpiresInMinutes int16
}

type MailRequestLoginOtp struct {
	To               string
	Otp              string
	ExpiresInMinutes int16
}

type MailRequestForgotPassword struct {
	To               string
	Otp              string
	ExpiresInMinutes int16
}

type MailRequestResetPassword struct {
	To string
}
