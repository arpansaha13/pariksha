package dtos

type LoginDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignUpDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type VerificationDto struct {
	OTP string `json:"otp"`
}
