package validate

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/structs"
)

// CodingQuestionData validates coding question data
func CodingQuestionData(coding *structs.CodingQuestion) error {
	if strings.TrimSpace(coding.Title) == "" {
		return status.Error(codes.InvalidArgument, "question title cannot be empty")
	}
	if strings.TrimSpace(coding.Statement) == "" {
		return status.Error(codes.InvalidArgument, "question statement cannot be empty")
	}
	if len(coding.InputDefinitions) == 0 {
		return status.Error(codes.InvalidArgument, "coding question must have input definitions")
	}
	if len(coding.InputDefinitions) > int(constants.MAX_CODING_INPUTS_COUNT) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("Number of inputs cannot be more than %d", constants.MAX_CODING_INPUTS_COUNT))
	}
	if len(coding.TestCases) > int(constants.MAX_CODING_TEST_CASES_COUNT) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("Number of test cases cannot be more than %d", constants.MAX_CODING_TEST_CASES_COUNT))
	}

	if err := validateOutputDefinition(coding.OutputDefinition); err != nil {
		return err
	}

	for _, def := range coding.InputDefinitions {
		if err := validateInputDefinition(def); err != nil {
			return err
		}
	}

	for _, testCase := range coding.TestCases {
		if err := validateTestCase(testCase, len(coding.InputDefinitions)); err != nil {
			return err
		}
	}

	return nil
}

func validateInputDefinition(def structs.InputDefinition) error {
	if strings.TrimSpace(def.VariableName) == "" {
		return status.Error(codes.InvalidArgument, "variable name is required for input definition")
	}

	switch def.Type {
	case constants.PARAMETER_TYPE_ARRAY:
		if def.Items == nil || len(*def.Items) != 1 {
			return status.Error(codes.InvalidArgument, "array input definition must have exactly one item")
		}
		if (*def.Items)[0].PropertyName != nil {
			return status.Error(codes.InvalidArgument, "array input definition cannot have a property name")
		}
		// Validate item type is primitive
		switch (*def.Items)[0].Type {
		case constants.PARAMETER_TYPE_NUMBER, constants.PARAMETER_TYPE_STRING, constants.PARAMETER_TYPE_BOOLEAN:
			// Valid primitive type
		default:
			return status.Error(codes.InvalidArgument, "array items must have primitive types")
		}
	case constants.PARAMETER_TYPE_NUMBER, constants.PARAMETER_TYPE_STRING, constants.PARAMETER_TYPE_BOOLEAN:
		if def.Items != nil {
			return status.Error(codes.InvalidArgument, "primitive input definition cannot have items")
		}
	default:
		return status.Error(codes.InvalidArgument, "invalid input definition type")
	}

	return nil
}

func validateOutputDefinition(def structs.OutputDefinition) error {
	switch def.Type {
	case constants.PARAMETER_TYPE_ARRAY:
		if def.Items == nil || len(*def.Items) != 1 {
			return status.Error(codes.InvalidArgument, "array output definition must have exactly one item")
		}
		// Validate item type is primitive
		switch (*def.Items)[0].Type {
		case constants.PARAMETER_TYPE_NUMBER, constants.PARAMETER_TYPE_STRING, constants.PARAMETER_TYPE_BOOLEAN:
			// Valid primitive type
		default:
			return status.Error(codes.InvalidArgument, "array output items must have primitive types")
		}
	case constants.PARAMETER_TYPE_NUMBER, constants.PARAMETER_TYPE_STRING, constants.PARAMETER_TYPE_BOOLEAN:
		if def.Items != nil {
			return status.Error(codes.InvalidArgument, "primitive output definition cannot have items")
		}
	default:
		return status.Error(codes.InvalidArgument, "invalid output definition type")
	}

	return nil
}

func validateTestCase(testCase structs.CodingQuestionTestCase, inputDefinitionsLength int) error {
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
