package models

import (
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
)

type Language struct {
	ID        types.LanguageID `gorm:"primaryKey;autoIncrement;type:smallint"`
	Slug      string           `gorm:"type:varchar(255);not null;unique"`
	Name      string           `gorm:"type:varchar(255);not null"`
	Extension string           `gorm:"type:varchar(16);not null"`
	Version   string           `gorm:"type:varchar(16);not null"`
	IsEnabled bool             `gorm:"type:boolean;not null;default:true"`
	// Runtime string `gorm:"type:varchar(16);not null"`
}

func (Language) TableName() string {
	return constants.TABLE_LANGUAGES
}
