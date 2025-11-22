package models

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
	"pariksha/question/internal/structs"
)

type Question struct {
	ID            types.QuestionID   `gorm:"primaryKey;autoIncrement;type:bigint"`
	Question      json.RawMessage    `gorm:"type:jsonb;not null"`
	Type          proto.QuestionType `gorm:"type:smallint;not null;check:type > 0 AND type <= 3"`
	Hash          string             `gorm:"type:varchar(64);uniqueIndex;not null"`
	PaperIndegree int32              `gorm:"not null"`
	ExamIndegree  int32              `gorm:"not null"`
	DeletedAt     gorm.DeletedAt     `gorm:"index"`
}

// Unmarshal the raw JSON data into the appropriate struct based on the Type field
func (q *Question) GetQuestion() (any, error) {
	switch q.Type {
	case proto.QuestionType_MCQ:
		var mcq structs.MCQQuestion
		if err := json.Unmarshal(q.Question, &mcq); err != nil {
			return nil, err
		}
		return mcq, nil
	case proto.QuestionType_SUBJECTIVE:
		var subjective structs.SubjectiveQuestion
		if err := json.Unmarshal(q.Question, &subjective); err != nil {
			return nil, err
		}
		return subjective, nil
	case proto.QuestionType_CODING:
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
