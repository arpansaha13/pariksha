package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Key       uuid.UUID `gorm:"type:uuid;primaryKey"`
	Token     string    `gorm:"type:varchar(255);not null"`
	ExpiresAt time.Time `gorm:"not null"`
}

func (Session) TableName() string {
	return "sessions"
}
