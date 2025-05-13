package models

import (
	"pariksha/common/pkg/constants"
)

type ExamCategory struct {
	ID         int64 `gorm:"primaryKey;type:bigint"`
	ExamID     int64 `gorm:"type:bigint;not null"`
	CategoryID int64 `gorm:"type:bigint;not null"`
	Order      int16 `gorm:"type:smallint;not null;default:0"`
	Exam       Exam  `gorm:"foreignKey:ExamID"`
}

func (ExamCategory) TableName() string {
	return constants.TABLE_EXAM_CATEGORIES
}
