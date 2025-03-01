package dtos

type LoginWithPasswordDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type SignUpDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type VerificationDto struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required"`
}

type LoginWithOtpDto struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordDto struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordDto struct {
	Email       string `json:"email" validate:"required,email"`
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required"`
	OTP         string `json:"otp" validate:"required"`
}
