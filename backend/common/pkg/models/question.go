package models

import (
	"database/sql"
	"encoding/json"
	"errors"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/structs"
)

type Question struct {
	ID            int64            `gorm:"primaryKey;type:bigint"`
	CategoryID    int64            `gorm:"type:bigint;not null"`
	Question      json.RawMessage  `gorm:"type:json;not null"`
	Order         int16            `gorm:"type:smallint;not null"`
	Type          string           `gorm:"type:varchar(20);not null;check:type IN ('MCQ', 'SHORT', 'LONG')"`
	Tags          json.RawMessage  `gorm:"type:json;default:'[]'"`
	PaperID       sql.NullInt64    `gorm:"type:bigint"`
	MaxScore      int16            `gorm:"type:smallint;not null;check:max_score >= 0 AND max_score <= 1000"`
	CorrectAnswer sql.NullString   `gorm:"type:text"`
	Locked        bool             `gorm:"not null;default:false"`
	Paper         Paper            `gorm:"foreignKey:PaperID"`
	Category      QuestionCategory `gorm:"foreignKey:CategoryID"`
}

// Unmarshal the raw JSON data into the appropriate struct based on the Type field
func (q *Question) GetQuestion() (interface{}, error) {
	switch q.Type {
	case constants.QUESTION_TYPE_MCQ:
		var mcq structs.MCQQuestion
		if err := json.Unmarshal(q.Question, &mcq); err != nil {
			return nil, err
		}
		return mcq, nil
	case constants.QUESTION_TYPE_SHORT, constants.QUESTION_TYPE_LONG:
		var general structs.GeneralQuestion
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
