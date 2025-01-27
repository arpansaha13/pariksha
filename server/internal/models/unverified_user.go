package models

import (
	"time"
)

// Note: OTP varchar length should match with `constants.VERIFICATION_HASH_LENGTH`

type UnverifiedUser struct {
	Hash         string    `gorm:"primaryKey;type:varchar(10)"`
	OTP          string    `gorm:"type:varchar(6);not null"`
	OTPExpiresAt time.Time `gorm:"not null"`
	Email        string    `gorm:"not null"`
	Password     string    `gorm:"not null"`
}

func (UnverifiedUser) TableName() string {
	return "unverified_users"
}
