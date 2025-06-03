package models

import (
	"database/sql"

	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
)

type QuestionCategory struct {
	ID        int64          `gorm:"primaryKey;type:bigint"`
	PaperID   sql.NullInt64  `gorm:"type:bigint"`
	Name      string         `gorm:"type:varchar(255);not null"`
	Order     int16          `gorm:"type:smallint;not null"`
	Locked    bool           `gorm:"not null;default:false"`
	Paper     Paper          `gorm:"foreignKey:PaperID"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (QuestionCategory) TableName() string {
	return constants.TABLE_CATEGORIES
}
