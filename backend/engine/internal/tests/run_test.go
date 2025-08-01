package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/test"
)

// Mock data for testing
var (
	defaultMetadata = metadata.New(map[string]string{
		"user_id": "1",
	})
)

func TestRunCode(t *testing.T) {
	type SetupReturn struct {
		InputDefs []*proto.InputDefinition
	}

	testCases := []test.TestCase[*proto.RunCodeRequest, *proto.RunCodeResponse, *SetupReturn]{
		{
			Name:     "Success - Add two numbers",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					InputDefs: []*proto.InputDefinition{
						{VariableName: "a", Type: proto.ParameterType_NUMBER},
						{VariableName: "b", Type: proto.ParameterType_NUMBER},
					},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.RunCodeRequest {
				return &proto.RunCodeRequest{
					QuestionHash: "test_add_numbers",
					Code:         `function solve(a, b) { return a + b; }`,
					TestCases: []*proto.TestCase{
						{
							Inputs:         []string{"1", "2"},
							ExpectedOutput: "3",
						},
						{
							Inputs:         []string{"-1", "5"},
							ExpectedOutput: "4",
						},
					},
					Environment: constants.LangNode,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.RunCodeResponse, setupData *SetupReturn) {
				require.True(t, resp.Compilation.Success)
				require.Len(t, resp.Results, 2)

				for _, result := range resp.Results {
					assert.Equal(t, proto.ExecutionStatus_SUCCESS, result.Status)
				}
			},
		},
		{
			Name:     "Compilation Error - Invalid syntax",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					InputDefs: []*proto.InputDefinition{
						{VariableName: "a", Type: proto.ParameterType_NUMBER},
						{VariableName: "b", Type: proto.ParameterType_NUMBER},
					},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.RunCodeRequest {
				return &proto.RunCodeRequest{
					QuestionHash: "test_add_numbers",
					Code:         `function solve(a, b {`, // Missing closing parenthesis
					TestCases: []*proto.TestCase{
						{
							Inputs:         []string{"1", "2"},
							ExpectedOutput: "3",
						},
					},
					Environment: constants.LangNode,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.RunCodeResponse, setupData *SetupReturn) {
				assert.False(t, resp.Compilation.Success)
				assert.NotNil(t, resp.Compilation.Stderr)
				assert.Contains(t, *resp.Compilation.Stderr, "SyntaxError")
			},
		},
		{
			Name:     "Runtime Error - Undeclared variable",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					InputDefs: []*proto.InputDefinition{
						{VariableName: "a", Type: proto.ParameterType_NUMBER},
						{VariableName: "b", Type: proto.ParameterType_NUMBER},
					},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.RunCodeRequest {
				return &proto.RunCodeRequest{
					QuestionHash: "test_add_numbers",
					Code:         `function solve(a, b) { return a / x; }`, // x is undeclared
					TestCases: []*proto.TestCase{
						{
							Inputs:         []string{"1", "0"},
							ExpectedOutput: "error",
						},
					},
					Environment: constants.LangNode,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.RunCodeResponse, setupData *SetupReturn) {
				require.True(t, resp.Compilation.Success)
				require.Len(t, resp.Results, 1)
				assert.Equal(t, proto.ExecutionStatus_RUNTIME_ERROR, resp.Results[0].Status)
			},
		},
		{
			Name:     "Success - String concatenation",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					InputDefs: []*proto.InputDefinition{
						{VariableName: "first", Type: proto.ParameterType_STRING},
						{VariableName: "last", Type: proto.ParameterType_STRING},
					},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.RunCodeRequest {
				return &proto.RunCodeRequest{
					QuestionHash: "test_string_concat",
					Code:         `function solve(first, last) { return first + " " + last; }`,
					TestCases: []*proto.TestCase{
						{
							Inputs:         []string{"John", "Doe"},
							ExpectedOutput: `"John Doe"`,
						},
						{
							Inputs:         []string{"Hello", "World"},
							ExpectedOutput: `"Hello World"`,
						},
					},
					Environment: constants.LangNode,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.RunCodeResponse, setupData *SetupReturn) {
				require.True(t, resp.Compilation.Success)
				require.Len(t, resp.Results, 2)
				for _, result := range resp.Results {
					assert.Equal(t, proto.ExecutionStatus_SUCCESS, result.Status)
				}
			},
		},
		{
			Name:     "Success - Boolean operations",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					InputDefs: []*proto.InputDefinition{
						{VariableName: "a", Type: proto.ParameterType_BOOLEAN},
						{VariableName: "b", Type: proto.ParameterType_BOOLEAN},
					},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.RunCodeRequest {
				return &proto.RunCodeRequest{
					QuestionHash: "test_boolean_ops",
					Code:         `function solve(a, b) { return a && b; }`,
					TestCases: []*proto.TestCase{
						{
							Inputs:         []string{"true", "true"},
							ExpectedOutput: "true",
						},
						{
							Inputs:         []string{"true", "false"},
							ExpectedOutput: "false",
						},
					},
					Environment: constants.LangNode,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.RunCodeResponse, setupData *SetupReturn) {
				require.True(t, resp.Compilation.Success)
				require.Len(t, resp.Results, 2)
				for _, result := range resp.Results {
					assert.Equal(t, proto.ExecutionStatus_SUCCESS, result.Status)
				}
			},
		},
		{
			Name:     "Success - Array operations",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					InputDefs: []*proto.InputDefinition{
						{
							VariableName: "arr",
							Type:         proto.ParameterType_ARRAY,
							Items: []*proto.ParameterItem{
								{Type: proto.ParameterType_NUMBER},
							},
						},
					},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.RunCodeRequest {
				return &proto.RunCodeRequest{
					QuestionHash: "test_array_numbers",
					Code:         `function solve(arr) { return arr.reduce((sum, n) => sum + n, 0); }`,
					TestCases: []*proto.TestCase{
						{
							Inputs:         []string{"[1,2,3,4]"},
							ExpectedOutput: "10",
						},
						{
							Inputs:         []string{"[-1,5,10]"},
							ExpectedOutput: "14",
						},
					},
					Environment: constants.LangNode,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.RunCodeResponse, setupData *SetupReturn) {
				require.True(t, resp.Compilation.Success)
				require.Len(t, resp.Results, 2)
				for _, result := range resp.Results {
					assert.Equal(t, proto.ExecutionStatus_SUCCESS, result.Status)
				}
			},
		},
		{
			Name:     "Success - Mixed types array",
			Metadata: defaultMetadata,
			Setup: func(t *testing.T) *SetupReturn {
				return &SetupReturn{
					InputDefs: []*proto.InputDefinition{
						{
							VariableName: "arr",
							Type:         proto.ParameterType_ARRAY,
							Items: []*proto.ParameterItem{
								{Type: proto.ParameterType_STRING},
							},
						},
					},
				}
			},
			GetRequest: func(setupData *SetupReturn) *proto.RunCodeRequest {
				return &proto.RunCodeRequest{
					QuestionHash: "test_array_strings",
					Code:         `function solve(arr) { return arr.join(""); }`,
					TestCases: []*proto.TestCase{
						{
							Inputs:         []string{`["a","b","c"]`},
							ExpectedOutput: `"abc"`,
						},
						{
							Inputs:         []string{`["hello"," ","world"]`},
							ExpectedOutput: `"hello world"`,
						},
					},
					Environment: constants.LangNode,
				}
			},
			ExpectedCode: codes.OK,
			Validate: func(t *testing.T, resp *proto.RunCodeResponse, setupData *SetupReturn) {
				require.True(t, resp.Compilation.Success)
				require.Len(t, resp.Results, 2)
				for _, result := range resp.Results {
					assert.Equal(t, proto.ExecutionStatus_SUCCESS, result.Status)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			test.Runner(t, tc, client.RunCode)
		})
	}
}
