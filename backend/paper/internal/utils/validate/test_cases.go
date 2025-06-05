package validate

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/models"
)

func TestCase(testCase *models.TestCaseContent, inputDefinitionsLength int) error {
	if len(testCase.Inputs) != inputDefinitionsLength {
		return status.Error(codes.InvalidArgument, "number of inputs in test case must match number of input definitions")
	}

	// Check that no input is empty
	for _, input := range testCase.Inputs {
		if strings.TrimSpace(input) == "" {
			return status.Error(codes.InvalidArgument, "test case input cannot be empty")
		}
	}

	// Check that output is not empty
	if strings.TrimSpace(testCase.Output) == "" {
		return status.Error(codes.InvalidArgument, "test case output cannot be empty")
	}

	return nil
}
