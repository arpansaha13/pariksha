package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Key       uuid.UUID `gorm:"type:uuid;primaryKey"`
	Token     string
	ExpiresAt time.Time
}

func (Session) TableName() string {
	return "sessions"
}
