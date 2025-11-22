package models

import (
	"database/sql/driver"
	"encoding/json"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
)

type Answer struct {
	ID                types.AnswerID      `gorm:"primaryKey;autoIncrement;type:bigint"`
	ExamParticipantID types.ParticipantID `gorm:"type:bigint"`
	QuestionID        types.QuestionID    `gorm:"type:bigint"`
	Answer            *json.RawMessage    `gorm:"type:jsonb"`
	ScoreAwarded      int16               `gorm:"type:smallint;default:0;not null"`
	Evaluated         bool                `gorm:"default:false;not null"`
	ExamParticipant   ExamParticipant     `gorm:"foreignKey:ExamParticipantID"`
}

type MCQAnswer struct {
	OptionIndex *int `json:"optionIndex"`
}

type SubjectiveAnswer struct {
	Text string `json:"text"`
}

type CodingAnswer struct {
	Code string `json:"code"`
}

func (a *MCQAnswer) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *SubjectiveAnswer) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (Answer) TableName() string {
	return constants.TABLE_ANSWERS
}
