package models

import (
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
)

type Category struct {
	ID            types.CategoryID `gorm:"primaryKey;type:bigint"`
	Name          string           `gorm:"type:varchar(255);not null"`
	PaperIndegree int32            `gorm:"not null"`
	ExamIndegree  int32            `gorm:"not null"`
	DeletedAt     gorm.DeletedAt   `gorm:"index"`
}

func (Category) TableName() string {
	return constants.TABLE_CATEGORIES
}
