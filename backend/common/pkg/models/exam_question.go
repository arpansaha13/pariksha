package models

import (
	"pariksha/common/pkg/constants"
)

type ExamQuestion struct {
	ID         int64 `gorm:"primaryKey;type:bigint"`
	ExamID     int64 `gorm:"type:bigint;not null"`
	Type       int16 `gorm:"type:smallint;not null;check:type > 0 AND type <= 3"`
	QuestionID int64 `gorm:"type:bigint;not null"`
	CategoryID int64 `gorm:"type:bigint;not null"`
	Order      int16 `gorm:"type:smallint;not null"`
	MaxScore   int16 `gorm:"type:smallint;not null"`
	Exam       Exam  `gorm:"foreignKey:ExamID"`
}

func (ExamQuestion) TableName() string {
	return constants.TABLE_EXAM_QUESTIONS
}
