package models

import (
	"time"

	"github.com/google/uuid"
)

type UnverifiedUser struct {
	Hash         uuid.UUID `gorm:"primaryKey;type:uuid"`
	OTP          string
	OTPExpiresAt time.Time
	Email        string
	Password     string
}

func (UnverifiedUser) TableName() string {
	return "unverified_users"
}
