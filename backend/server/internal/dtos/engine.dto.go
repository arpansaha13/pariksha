package dtos

import "pariksha/common/pkg/proto"

type TestCaseDto struct {
	Inputs         []string `json:"inputs" validate:"required,min=1"`
	ExpectedOutput string   `json:"expectedOutput" validate:"required"`
}

type RunCodeRequestDto struct {
	QuestionID  string        `json:"question_id" validate:"required"`
	Code        string        `json:"code" validate:"required"`
	Environment string        `json:"environment" validate:"required"`
	TestCases   []TestCaseDto `json:"test_cases" validate:"required,min=1"`
}

type CompilationResult struct {
	Success bool    `json:"success"`
	Stderr  *string `json:"stderr,omitempty"`
}

type TestCaseResult struct {
	Inputs         []string              `json:"inputs"`
	Output         string                `json:"output"`
	ExpectedOutput string                `json:"expected_output"`
	Status         proto.ExecutionStatus `json:"status"`
	Stdout         string                `json:"stdout"`
	Error          string                `json:"error"`
	ExecutionTime  int64                 `json:"execution_time"`
}

type RunCodeResponseDto struct {
	Compilation CompilationResult `json:"compilation"`
	Results     []TestCaseResult  `json:"results"`
}
