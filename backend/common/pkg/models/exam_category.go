package models

import (
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
)

type ExamCategory struct {
	ID         types.CategoryID `gorm:"primaryKey;autoIncrement;type:bigint"`
	ExamID     types.ExamID     `gorm:"type:bigint;not null"`
	CategoryID types.CategoryID `gorm:"type:bigint;not null"`
	Order      int16            `gorm:"type:smallint;not null;default:0"`
	Exam       Exam             `gorm:"foreignKey:ExamID"`
}

func (ExamCategory) TableName() string {
	return constants.TABLE_EXAM_CATEGORIES
}
