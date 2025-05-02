package models

import (
	"encoding/json"
	"errors"
)

type QuestionCount struct {
	MCQ   int `json:"mcq"`
	Short int `json:"short"`
	Long  int `json:"long"`
}

type Paper struct {
	ID              int64           `gorm:"primaryKey;type:bigint"`
	Title           string          `gorm:"type:varchar(255);not null;default:'Untitled Paper'"`
	MaxScore        int             `gorm:"default:0"`
	DurationMinutes int             `gorm:"not null"`
	QuestionCounts  json.RawMessage `gorm:"type:json;default:'{\"mcq\":0,\"short\":0,\"long\":0}'"`
	CreatedBy       int64           `gorm:"type:bigint;not null"`

	Questions  []Question         `gorm:"foreignKey:PaperID"`
	Categories []QuestionCategory `gorm:"foreignKey:PaperID"`
}

func (p *Paper) GetQuestionCounts() (QuestionCount, error) {
	var counts QuestionCount
	if err := json.Unmarshal(p.QuestionCounts, &counts); err != nil {
		return counts, errors.New("failed to parse question counts")
	}
	return counts, nil
}

func (Paper) TableName() string {
	return "papers"
}
