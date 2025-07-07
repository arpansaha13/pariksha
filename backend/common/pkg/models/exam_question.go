package models

import (
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
)

type ExamQuestion struct {
	ID         types.QuestionID `gorm:"primaryKey;type:bigint"`
	ExamID     types.ExamID     `gorm:"type:bigint;not null;uniqueIndex:uq_exam_question"`
	QuestionID types.QuestionID `gorm:"type:bigint;not null;uniqueIndex:uq_exam_question"`
	CategoryID types.CategoryID `gorm:"type:bigint;not null"`
	Order      int16            `gorm:"type:smallint;not null"`
	MaxScore   int16            `gorm:"type:smallint;not null"`
	Exam       Exam             `gorm:"foreignKey:ExamID"`
}

func (ExamQuestion) TableName() string {
	return constants.TABLE_EXAM_QUESTIONS
}
