package validate

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
)

func TestCases(testCases []models.TestCaseContent, inputDefinitionsLength int) error {
	if len(testCases) > int(constants.MAX_CODING_TEST_CASES_COUNT) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("Number of test cases cannot be more than %d", constants.MAX_CODING_TEST_CASES_COUNT))
	}

	for _, testCase := range testCases {
		if err := validateTestCase(testCase, inputDefinitionsLength); err != nil {
			return err
		}
	}

	return nil
}

func validateTestCase(testCase models.TestCaseContent, inputDefinitionsLength int) error {
	if len(testCase.Inputs) != inputDefinitionsLength {
		return status.Error(codes.InvalidArgument, "number of inputs in test case must match number of input definitions")
	}

	// Check that no input is empty
	for _, input := range testCase.Inputs {
		if strings.TrimSpace(input) == "" {
			return status.Error(codes.InvalidArgument, "test case input cannot be empty")
		}
	}

	if strings.TrimSpace(testCase.Output) == "" {
		return status.Error(codes.InvalidArgument, "test case output cannot be empty")
	}
	if testCase.Explanation != nil && strings.TrimSpace(*testCase.Explanation) == "" {
		testCase.Explanation = nil
	}

	return nil
}
