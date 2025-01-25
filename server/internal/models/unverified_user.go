package models

import (
	"time"
)

type UnverifiedUser struct {
	Hash         string `gorm:"primaryKey;type:varchar(10)"`
	OTP          string
	OTPExpiresAt time.Time
	Email        string
	Password     string
}

func (UnverifiedUser) TableName() string {
	return "unverified_users"
}
