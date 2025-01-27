package models

import (
	"encoding/json"
	"errors"

	"github.com/arpansaha13/pariksha/internal/constants"
)

type MCQQuestion struct {
	Statement string   `json:"statement"`
	Options   []string `json:"options"`
}

type GeneralQuestion struct {
	Statement string `json:"statement"`
}

type Question struct {
	ID            int             `gorm:"primaryKey"`
	Question      json.RawMessage `gorm:"type:json;not null"`
	Category      string          `gorm:"type:varchar(255)"`
	Type          string          `gorm:"type:varchar(20);not null;check:type IN ('MCQ', 'SHORT', 'LONG')"`
	Tags          []string        `gorm:"type:varchar[];default:'{}'"`
	PaperID       int
	MaxScore      int
	CorrectAnswer string   `gorm:"type:text"`
	Paper         Paper    `gorm:"foreignKey:PaperID"`
	Answers       []Answer `gorm:"foreignKey:QuestionID"`
}

// Unmarshal the raw JSON data into the appropriate struct based on the Type field
func (q *Question) GetQuestion() (interface{}, error) {
	switch q.Type {
	case constants.QUESTION_TYPE_MCQ:
		var mcq MCQQuestion
		if err := json.Unmarshal(q.Question, &mcq); err != nil {
			return nil, err
		}
		return mcq, nil
	case constants.QUESTION_TYPE_SHORT, constants.QUESTION_TYPE_LONG:
		var general GeneralQuestion
		if err := json.Unmarshal(q.Question, &general); err != nil {
			return nil, err
		}
		return general, nil
	default:
		return nil, errors.New("invalid question type")
	}
}

func (Question) TableName() string {
	return "questions"
}
