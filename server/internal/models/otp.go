package models

import (
	"time"
)

// Note: OTP varchar length should match with `constants.VERIFICATION_HASH_LENGTH`

type Otp struct {
	Email        string    `gorm:"primaryKey;type:varchar(255)"`
	OTP          string    `gorm:"type:varchar(6);not null"`
	OTPExpiresAt time.Time `gorm:"not null"`
	Purpose      int       `gorm:"not null"`
}

func (Otp) TableName() string {
	return "otps"
}
