package models

import (
	"pariksha/common/pkg/constants"
)

type ExamQuestion struct {
	ID         int64  `gorm:"primaryKey;type:bigint"`
	ExamID     int64  `gorm:"type:bigint;not null"`
	Type       string `gorm:"type:varchar(20);not null;check:type IN ('MCQ', 'SHORT', 'LONG')"`
	QuestionID int64  `gorm:"type:bigint;not null"`
	CategoryID int64  `gorm:"type:bigint;not null"`
	Order      int16  `gorm:"type:smallint;not null"`
	MaxScore   int16  `gorm:"type:smallint;not null"`
	Exam       Exam   `gorm:"foreignKey:ExamID"`
}

func (ExamQuestion) TableName() string {
	return constants.TABLE_EXAM_QUESTIONS
}
