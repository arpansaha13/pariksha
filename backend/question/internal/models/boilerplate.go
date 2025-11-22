package models

import (
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
)

type Boilerplate struct {
	ID         types.BoilerplateID `gorm:"primaryKey;autoIncrement;type:bigint"`
	QuestionID types.QuestionID    `gorm:"type:bigint;not null;uniqueIndex:idx_question_language;constraint:OnDelete:SET NULL"`
	LanguageID types.LanguageID    `gorm:"type:smallint;not null;uniqueIndex:idx_question_language"`
	Code       string              `gorm:"type:text;not null"`
	Question   Question            `gorm:"foreignKey:QuestionID"`
	Language   Language            `gorm:"foreignKey:LanguageID"`
}

func (Boilerplate) TableName() string {
	return constants.TABLE_BOILERPLATES
}
