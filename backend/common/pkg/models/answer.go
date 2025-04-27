package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

type Answer struct {
	ID                int64 `gorm:"primaryKey"`
	ExamParticipantID int64
	QuestionID        int64
	Answer            json.RawMessage `gorm:"type:json"`
	ScoreAwarded      int             `gorm:"default:0;not null"`
	Comments          sql.NullString  `gorm:"type:text"`
	Evaluated         bool            `gorm:"default:false;not null"`
	ExamParticipant   ExamParticipant `gorm:"foreignKey:ExamParticipantID"`
}

type MCQAnswer struct {
	OptionIndex int `json:"optionIndex"`
}

type GeneralAnswer struct {
	Text string `json:"text"`
}

func (a *MCQAnswer) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *GeneralAnswer) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (Answer) TableName() string {
	return "answers"
}
