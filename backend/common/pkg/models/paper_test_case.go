package models

import (
	"encoding/json"

	"pariksha/common/pkg/constants"
)

type TestCaseContent struct {
	Inputs      []string `json:"inputs"`
	Output      string   `json:"output"`
	Explanation *string  `json:"explanation,omitempty"`
}

type TestCase struct {
	ID         int64           `gorm:"primaryKey;type:bigint"`
	QuestionID int64           `gorm:"type:bigint;not null"`
	Content    json.RawMessage `gorm:"type:jsonb;not null"`

	// Hidden test cases won't be shown as examples
	Hidden bool `gorm:"column:hidden;default:false;not null"`

	Question Question `gorm:"foreignKey:QuestionID"`
}

func (TestCase) TableName() string {
	return constants.TABLE_TEST_CASES
}
