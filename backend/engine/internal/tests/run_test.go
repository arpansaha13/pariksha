package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
)

func TestRunCode(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		inputDefs    []proto.InputDefinition
		outputDef    structs.OutputDefinition
		testCases    []*proto.TestCase
		expectedCode codes.Code
		validate     func(t *testing.T, resp *proto.RunCodeResponse)
	}{
		{
			name: "Success - Add two numbers",
			code: `function solve(a, b) { return a + b; }`,
			inputDefs: []proto.InputDefinition{
				{VariableName: "a", Type: proto.ParameterType_NUMBER},
				{VariableName: "b", Type: proto.ParameterType_NUMBER},
			},
			outputDef: structs.OutputDefinition{
				Type: proto.ParameterType_NUMBER,
			},
			testCases: []*proto.TestCase{
				{
					Inputs:         []string{"1", "2"},
					ExpectedOutput: "3",
				},
				{
					Inputs:         []string{"-1", "5"},
					ExpectedOutput: "4",
				},
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.RunCodeResponse) {
				require.True(t, resp.Compilation.Success)
				require.Len(t, resp.Results, 2)

				for _, result := range resp.Results {
					assert.Equal(t, proto.ExecutionStatus_SUCCESS, result.Status)
				}
			},
		},
		{
			name: "Compilation Error - Invalid syntax",
			code: `function solve(a, b {`, // Missing closing parenthesis
			inputDefs: []proto.InputDefinition{
				{VariableName: "a", Type: proto.ParameterType_NUMBER},
				{VariableName: "b", Type: proto.ParameterType_NUMBER},
			},
			outputDef: structs.OutputDefinition{
				Type: proto.ParameterType_NUMBER,
			},
			testCases: []*proto.TestCase{
				{
					Inputs:         []string{"1", "2"},
					ExpectedOutput: "3",
				},
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.RunCodeResponse) {
				assert.False(t, resp.Compilation.Success)
				assert.NotNil(t, resp.Compilation.Stderr)
				assert.Contains(t, *resp.Compilation.Stderr, "SyntaxError")
			},
		},
		{
			name: "Runtime Error - Undeclared variable",
			code: `function solve(a, b) { return a / x; }`, // x is undeclared
			inputDefs: []proto.InputDefinition{
				{VariableName: "a", Type: proto.ParameterType_NUMBER},
				{VariableName: "b", Type: proto.ParameterType_NUMBER},
			},
			outputDef: structs.OutputDefinition{
				Type: proto.ParameterType_NUMBER,
			},
			testCases: []*proto.TestCase{
				{
					Inputs:         []string{"1", "0"},
					ExpectedOutput: "error",
				},
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.RunCodeResponse) {
				require.True(t, resp.Compilation.Success)
				require.Len(t, resp.Results, 1)
				assert.Equal(t, proto.ExecutionStatus_RUNTIME_ERROR, resp.Results[0].Status)
			},
		},
		{
			name: "Success - String concatenation",
			code: `function solve(first, last) { return first + " " + last; }`,
			inputDefs: []proto.InputDefinition{
				{VariableName: "first", Type: proto.ParameterType_STRING},
				{VariableName: "last", Type: proto.ParameterType_STRING},
			},
			outputDef: structs.OutputDefinition{
				Type: proto.ParameterType_STRING,
			},
			testCases: []*proto.TestCase{
				{
					Inputs:         []string{"John", "Doe"},
					ExpectedOutput: `"John Doe"`,
				},
				{
					Inputs:         []string{"Hello", "World"},
					ExpectedOutput: `"Hello World"`,
				},
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.RunCodeResponse) {
				require.True(t, resp.Compilation.Success)
				require.Len(t, resp.Results, 2)
				for _, result := range resp.Results {
					assert.Equal(t, proto.ExecutionStatus_SUCCESS, result.Status)
				}
			},
		},
		{
			name: "Success - Boolean operations",
			code: `function solve(a, b) { return a && b; }`,
			inputDefs: []proto.InputDefinition{
				{VariableName: "a", Type: proto.ParameterType_BOOLEAN},
				{VariableName: "b", Type: proto.ParameterType_BOOLEAN},
			},
			outputDef: structs.OutputDefinition{
				Type: proto.ParameterType_BOOLEAN,
			},
			testCases: []*proto.TestCase{
				{
					Inputs:         []string{"true", "true"},
					ExpectedOutput: "true",
				},
				{
					Inputs:         []string{"true", "false"},
					ExpectedOutput: "false",
				},
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.RunCodeResponse) {
				require.True(t, resp.Compilation.Success)
				require.Len(t, resp.Results, 2)
				for _, result := range resp.Results {
					assert.Equal(t, proto.ExecutionStatus_SUCCESS, result.Status)
				}
			},
		},
		{
			name: "Success - Array operations",
			code: `function solve(arr) { return arr.reduce((sum, n) => sum + n, 0); }`,
			inputDefs: []proto.InputDefinition{
				{
					VariableName: "arr",
					Type:         proto.ParameterType_ARRAY,
					Items: []*proto.ParameterItem{
						{Type: proto.ParameterType_NUMBER},
					},
				},
			},
			outputDef: structs.OutputDefinition{
				Type: proto.ParameterType_NUMBER,
			},
			testCases: []*proto.TestCase{
				{
					Inputs:         []string{"[1,2,3,4]"},
					ExpectedOutput: "10",
				},
				{
					Inputs:         []string{"[-1,5,10]"},
					ExpectedOutput: "14",
				},
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.RunCodeResponse) {
				require.True(t, resp.Compilation.Success)
				require.Len(t, resp.Results, 2)
				for _, result := range resp.Results {
					assert.Equal(t, proto.ExecutionStatus_SUCCESS, result.Status)
				}
			},
		},
		{
			name: "Success - Mixed types array",
			code: `function solve(arr) { return arr.join(""); }`,
			inputDefs: []proto.InputDefinition{
				{
					VariableName: "arr",
					Type:         proto.ParameterType_ARRAY,
					Items: []*proto.ParameterItem{
						{Type: proto.ParameterType_STRING},
					},
				},
			},
			outputDef: structs.OutputDefinition{
				Type: proto.ParameterType_STRING,
			},
			testCases: []*proto.TestCase{
				{
					Inputs:         []string{`["a","b","c"]`},
					ExpectedOutput: `"abc"`,
				},
				{
					Inputs:         []string{`["hello"," ","world"]`},
					ExpectedOutput: `"hello world"`,
				},
			},
			expectedCode: codes.OK,
			validate: func(t *testing.T, resp *proto.RunCodeResponse) {
				require.True(t, resp.Compilation.Success)
				require.Len(t, resp.Results, 2)
				for _, result := range resp.Results {
					assert.Equal(t, proto.ExecutionStatus_SUCCESS, result.Status)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupTest(t)

			// Mock input/output definitions
			mockInputDefinitions = tt.inputDefs

			// Run code
			req := &proto.RunCodeRequest{
				QuestionHash: mockQuestionHash,
				Code:         tt.code,
				TestCases:    tt.testCases,
				Environment:  constants.LangNode,
			}

			resp, err := server.RunCode(context.Background(), req)
			if tt.expectedCode != codes.OK {
				assert.Equal(t, tt.expectedCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			tt.validate(t, resp)
		})
	}
}
