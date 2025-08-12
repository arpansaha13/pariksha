package services

import (
	"encoding/json"
	"fmt"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/engine/internal/interservice"
	"pariksha/engine/internal/runner"
)

type Engine struct {
	questionIntSvc *interservice.Question
	runners        map[string]runner.Runner
}

func NewEngine(
	questionIntSvc *interservice.Question,
	nodeRunner *runner.Node,
) *Engine {
	return &Engine{
		questionIntSvc: questionIntSvc,
		runners: map[string]runner.Runner{
			constants.LangNode: nodeRunner,
		},
	}
}

func (s *Engine) Run(req *proto.RunCodeRequest) (*proto.RunCodeResponse, error) {
	rnr, ok := s.runners[req.Environment]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported environment: %s", req.Environment)
	}

	// Fetch input definitions. Will be used for parsing inputs.
	inputDefinitions, err := s.questionIntSvc.GetInputDefinitions(req.QuestionHash)
	if err != nil {
		return nil, err
	}

	// Parse and validate all test cases. The string inputs need to be parsed
	// to their respective data types before sending as inputs to templates.
	parsedTestCases := make([]map[string]any, len(req.TestCases))
	for i, testCase := range req.TestCases {
		parsedInputs, err := parseTestCase(testCase.Inputs, inputDefinitions)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid test case %d: %v", i, err)
		}

		parsedTestCases[i] = map[string]any{
			"inputs":         parsedInputs,
			"expectedOutput": testCase.ExpectedOutput,
		}
	}

	return rnr.Run(&runner.RunnerArg{
		Code:            req.Code,
		ParsedTestCases: parsedTestCases,
		TestCasesCount:  int16(len(req.TestCases)),
	})
}

// parseTestCase parses all inputs in a test case according to their input definitions
func parseTestCase(testCaseInputs []string, inputDefs []*proto.InputDefinition) ([]any, error) {
	if len(testCaseInputs) != len(inputDefs) {
		return nil, status.Errorf(codes.InvalidArgument, "input count mismatch: expected %d, got %d", len(inputDefs), len(testCaseInputs))
	}

	parsedInputs := make([]any, len(testCaseInputs))
	for i, input := range testCaseInputs {
		parsed, err := parseValue(input, inputDefs[i].Type, inputDefs[i].Items)
		if err != nil {
			return nil, fmt.Errorf("failed to parse input %s: %v", inputDefs[i].VariableName, err)
		}
		parsedInputs[i] = parsed
	}
	return parsedInputs, nil
}

// parseValue converts a string input to its appropriate type based on the parameter type
func parseValue(input string, paramType proto.ParameterType, items []*proto.ParameterItem) (any, error) {
	switch paramType {
	case proto.ParameterType_NUMBER:
		return strconv.ParseFloat(input, 64)
	case proto.ParameterType_STRING:
		return input, nil
	case proto.ParameterType_BOOLEAN:
		return strconv.ParseBool(input)
	case proto.ParameterType_ARRAY:
		if items == nil {
			return nil, status.Errorf(codes.InvalidArgument, "items definition missing for array type")
		}
		var rawArray []any
		if err := json.Unmarshal([]byte(input), &rawArray); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid array format: %v", err)
		}

		// Parse each array element based on the item type
		itemType := items[0].Type // Assuming first item defines the array element type
		parsedArray := make([]any, len(rawArray))
		for i, item := range rawArray {
			// Convert item to string since all inputs are strings
			itemStr := fmt.Sprintf("%v", item)
			parsed, err := parseValue(itemStr, itemType, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to parse array item %d: %v", i, err)
			}
			parsedArray[i] = parsed
		}
		return parsedArray, nil
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported parameter type: %d", paramType)
	}
}
