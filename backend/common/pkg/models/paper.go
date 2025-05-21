package models

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
)

type QuestionCount struct {
	MCQ        int16 `json:"mcq"`
	Subjective int16 `json:"subjective"`
	Coding     int16 `json:"coding"`
}

// Sum of max_score of all questions in the paper may exceed int16 range.
// Use int32 for paper's max_score.

type Paper struct {
	ID              int64           `gorm:"primaryKey;type:bigint"`
	Title           string          `gorm:"type:varchar(255);not null;default:'Untitled Paper'"`
	MaxScore        int32           `gorm:"type:integer;default:0"`
	DurationMinutes int16           `gorm:"type:smallint;not null;check:duration_minutes >= 0 AND duration_minutes <= 1440"`
	QuestionCounts  json.RawMessage `gorm:"type:json;default:'{\"mcq\":0,\"subjective\":0,\"coding\":0}'"`
	CreatedBy       int64           `gorm:"type:bigint;not null"`
	DeletedAt       gorm.DeletedAt  `gorm:"index"`

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
	return constants.TABLE_PAPERS
}
