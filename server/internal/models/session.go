package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Key       uuid.UUID `gorm:"primaryKey"`
	Token     string
	ExpiresAt time.Time
}

func (Session) TableName() string {
	return "sessions"
}
