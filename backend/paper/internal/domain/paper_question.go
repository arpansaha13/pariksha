package domain

import (
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
)

type PaperQuestion struct {
	PaperID    types.PaperID    `gorm:"primaryKey;type:bigint;not null;constraint:OnDelete:CASCADE;references:id"`
	QuestionID types.QuestionID `gorm:"primaryKey;type:bigint;not null;constraint:OnDelete:CASCADE;references:id"`
	CategoryID types.CategoryID `gorm:"type:bigint;not null"`
	Order      int16            `gorm:"type:smallint;not null"`
	MaxScore   int16            `gorm:"type:smallint;not null;check:max_score >= 0 AND max_score <= 1000"`
}

func (PaperQuestion) TableName() string {
	return constants.TABLE_PAPER_QUESTIONS
}
