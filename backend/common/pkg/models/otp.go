package models

import (
	"time"
)

// Note: OTP varchar length should be >= `constants.VERIFICATION_OTP_LENGTH`

type Otp struct {
	Email        string    `gorm:"primaryKey;type:varchar(255)"`
	OTP          string    `gorm:"type:varchar(6);not null"`
	OTPExpiresAt time.Time `gorm:"not null"`
	Purpose      int16     `gorm:"type:smallint;not null"`
}

func (Otp) TableName() string {
	return "otps"
}
