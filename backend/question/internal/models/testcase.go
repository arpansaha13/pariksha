package models

import (
	"encoding/json"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
)

type TestCase struct {
	ID         types.TestCaseID `gorm:"primaryKey;type:bigint"`
	QuestionID types.QuestionID `gorm:"type:bigint;not null;uniqueIndex:idx_test_case_order;constraint:OnDelete:CASCADE;references:id"`
	Order      int16            `gorm:"type:smallint;not null;uniqueIndex:idx_test_case_order"`
	Content    json.RawMessage  `gorm:"type:jsonb;not null"`
	DeletedAt  gorm.DeletedAt   `gorm:"index"`

	// SHA256 hash of content
	DataHash string `gorm:"type:varchar(64);not null"`

	// Hidden test cases won't be shown as examples
	Hidden bool `gorm:"column:hidden;default:false;not null"`
}

func (TestCase) TableName() string {
	return constants.TABLE_TEST_CASES
}

type TestCaseContent struct {
	Inputs      []string `json:"inputs"`
	Output      string   `json:"output"`
	Explanation *string  `json:"explanation,omitempty"`
}

func (tc *TestCaseContent) Validate(inputDefinitionsLength int) error {

	if len(tc.Inputs) != inputDefinitionsLength {
		return status.Error(codes.InvalidArgument, "number of inputs in test case must match number of input definitions")
	}

	// Check that no input is empty
	for _, input := range tc.Inputs {
		if strings.TrimSpace(input) == "" {
			return status.Error(codes.InvalidArgument, "test case input cannot be empty")
		}
	}

	// Check that output is not empty
	if strings.TrimSpace(tc.Output) == "" {
		return status.Error(codes.InvalidArgument, "test case output cannot be empty")
	}

	return nil
}
