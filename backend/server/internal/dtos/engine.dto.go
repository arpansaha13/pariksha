package dtos

type TestCaseDto struct {
	Inputs []string `json:"inputs" validate:"required,min=1"`
}

type RunCodeRequestDto struct {
	Code        string        `json:"code" validate:"required"`
	Environment string        `json:"environment" validate:"required"`
	TestCases   []TestCaseDto `json:"testCases" validate:"required,min=1"`
}

type RunCodeResponseDto struct {
	Stdout        string `json:"stdout"`
	Stderr        string `json:"stderr"`
	Error         string `json:"error,omitempty"`
	ExitCode      int32  `json:"exit_code"`
	ExecutionTime int64  `json:"execution_time"`
}
