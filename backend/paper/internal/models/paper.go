package models

import (
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
)

type Paper struct {
	ID              types.PaperID  `gorm:"primaryKey;type:bigint"`
	Title           string         `gorm:"type:varchar(255);not null;default:'Untitled Paper'"`
	DurationMinutes int16          `gorm:"type:smallint;not null;check:duration_minutes >= 0 AND duration_minutes <= 1440"`
	CreatedBy       types.UserID   `gorm:"type:bigint;not null"`
	Hash            string         `gorm:"type:varchar(64);uniqueIndex;not null"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (Paper) TableName() string {
	return constants.TABLE_PAPERS
}
