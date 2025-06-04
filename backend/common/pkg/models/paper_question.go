package models

import (
	"database/sql"
	"encoding/json"
	"errors"

	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/structs"
)

type Question struct {
	ID            int64            `gorm:"primaryKey;type:bigint"`
	CategoryID    int64            `gorm:"type:bigint;not null"`
	Question      json.RawMessage  `gorm:"type:jsonb;not null"`
	Order         int16            `gorm:"type:smallint;not null"`
	Type          int16            `gorm:"type:smallint;not null;check:type > 0 AND type <= 3"`
	Tags          *json.RawMessage `gorm:"type:jsonb;default:null"`
	PaperID       sql.NullInt64    `gorm:"type:bigint"`
	MaxScore      int16            `gorm:"type:smallint;not null;check:max_score >= 0 AND max_score <= 1000"`
	CorrectAnswer sql.NullString   `gorm:"type:text"`
	Locked        bool             `gorm:"not null;default:false"`
	DeletedAt     gorm.DeletedAt   `gorm:"index"`

	Paper    Paper            `gorm:"foreignKey:PaperID"`
	Category QuestionCategory `gorm:"foreignKey:CategoryID"`
}

// Unmarshal the raw JSON data into the appropriate struct based on the Type field
func (q *Question) GetQuestion() (any, error) {
	switch q.Type {
	case constants.QUESTION_TYPE_MCQ:
		var mcq structs.MCQQuestion
		if err := json.Unmarshal(q.Question, &mcq); err != nil {
			return nil, err
		}
		return mcq, nil
	case constants.QUESTION_TYPE_SUBJECTIVE:
		var subjective structs.SubjectiveQuestion
		if err := json.Unmarshal(q.Question, &subjective); err != nil {
			return nil, err
		}
		return subjective, nil
	case constants.QUESTION_TYPE_CODING:
		var coding structs.CodingQuestion
		if err := json.Unmarshal(q.Question, &coding); err != nil {
			return nil, err
		}
		return coding, nil
	default:
		return nil, errors.New("invalid question type")
	}
}

func (Question) TableName() string {
	return constants.TABLE_QUESTIONS
}
