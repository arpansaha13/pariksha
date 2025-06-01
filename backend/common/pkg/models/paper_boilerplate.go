package models

import (
	"pariksha/common/pkg/constants"
)

type Boilerplate struct {
	ID         int64    `gorm:"primaryKey;type:bigint"`
	QuestionID int64    `gorm:"type:bigint;not null;uniqueIndex:idx_question_language;constraint:OnDelete:SET NULL"`
	LanguageID int16    `gorm:"type:smallint;not null;uniqueIndex:idx_question_language"`
	Code       string   `gorm:"type:text;not null"`
	Question   Question `gorm:"foreignKey:QuestionID"`
	Language   Language `gorm:"foreignKey:LanguageID"`
}

func (Boilerplate) TableName() string {
	return constants.TABLE_BOILERPLATES
}
