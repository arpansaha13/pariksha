package runner

import (
	"pariksha/common/pkg/proto"
	"pariksha/engine/internal/templates"
)

type Runner interface {
	Run(args *RunnerArg) (*proto.RunCodeResponse, error)
}

type RunnerArg struct {
	Code            string
	TestCasesCount  int16
	ParsedTestCases []map[string]any
}

type testResult struct {
	Inputs         []string `json:"inputs"`
	Output         string   `json:"output"`
	ExpectedOutput string   `json:"expectedOutput"`
	Match          bool     `json:"match"`
	Error          string   `json:"error,omitempty"`
	ExecutionTime  int64    `json:"executionTime"`
}

// environmentConfig contains all configuration needed for an execution environment
type environmentConfig struct {
	Image        string   // Docker image to use
	FileExt      string   // File extension for the source code
	CommandName  string   // Command to execute the code
	CommandArgs  []string // Arguments for the command (excluding the script path)
	MountTarget  string   // Where to mount the script file in container
	TemplateFunc templates.TemplateFunc
}
