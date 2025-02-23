package models

import (
	"time"

	"github.com/google/uuid"
)

// Note: CsrfToken varchar length should match with `constants.CSRF_TOKEN_LENGTH`

type Session struct {
	Key       uuid.UUID `gorm:"type:uuid;primaryKey"`
	Token     string    `gorm:"type:varchar(255);not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CsrfToken string    `gorm:"type:varchar(32);not null"`
}

func (Session) TableName() string {
	return "sessions"
}
