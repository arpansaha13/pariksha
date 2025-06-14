package models

import (
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
)

type ExamQuestion struct {
	ID         types.QuestionID   `gorm:"primaryKey;type:bigint"`
	ExamID     types.ExamID       `gorm:"type:bigint;not null"`
	Type       proto.QuestionType `gorm:"type:smallint;not null;check:type > 0 AND type <= 3"`
	QuestionID types.QuestionID   `gorm:"type:bigint;not null"`
	CategoryID types.CategoryID   `gorm:"type:bigint;not null"`
	Order      int16              `gorm:"type:smallint;not null"`
	MaxScore   int16              `gorm:"type:smallint;not null"`
	Exam       Exam               `gorm:"foreignKey:ExamID"`
}

func (ExamQuestion) TableName() string {
	return constants.TABLE_EXAM_QUESTIONS
}
