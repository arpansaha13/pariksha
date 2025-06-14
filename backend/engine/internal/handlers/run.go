package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/engine/internal/config/db"
	"pariksha/engine/internal/config/env"
	engineConstants "pariksha/engine/internal/constants"
	"pariksha/engine/internal/templates"
)

type EngineServer struct {
	proto.UnimplementedEngineServer

	dockerClient *client.Client
}

// EnvironmentConfig contains all configuration needed for an execution environment
type EnvironmentConfig struct {
	Image        string   // Docker image to use
	FileExt      string   // File extension for the source code
	CommandName  string   // Command to execute the code
	CommandArgs  []string // Arguments for the command (excluding the script path)
	MountTarget  string   // Where to mount the script file in container
	TemplateFunc templates.TemplateFunc
}

// Environment configurations
var envConfigs = map[string]EnvironmentConfig{
	constants.LangNode: {
		Image:        engineConstants.NodeImage,
		FileExt:      ".js",
		CommandName:  "node",
		CommandArgs:  nil,
		MountTarget:  "/code/solution.js",
		TemplateFunc: templates.GenerateNodeScript,
	},
}

// NewEngineServer creates a new instance of EngineServer
func NewEngineServer() (*EngineServer, error) {
	// Ensure DOCKER_API_VERSION = 1.46
	// Error response from daemon: client version 1.48 is too new. Maximum supported API version is 1.46
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create Docker client: %v", err)
	}
	return &EngineServer{dockerClient: cli}, nil
}

type TestResult struct {
	Inputs         []string `json:"inputs"`
	Output         string   `json:"output"`
	ExpectedOutput string   `json:"expectedOutput"`
	Match          bool     `json:"match"`
	Error          string   `json:"error,omitempty"`
	ExecutionTime  int64    `json:"executionTime"`
}

func (s *EngineServer) RunCode(ctx context.Context, req *proto.RunCodeRequest) (*proto.RunCodeResponse, error) {
	// Get environment config
	envConfig, ok := envConfigs[req.Environment]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported environment: %s", req.Environment)
	}

	// Fetch input definitions
	// Will be used for parsing inputs
	inputDefinitions, err := fetchInputDefinitions(db.Papers, req.QuestionHash)
	if err != nil {
		return nil, err
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(engineConstants.ExecutionTimeout)*time.Second)
	defer cancel()

	// Check if image exists locally
	_, _, err = s.dockerClient.ImageInspectWithRaw(execCtx, envConfig.Image)
	if err != nil {
		// Image not found locally, pull it
		reader, err := s.dockerClient.ImagePull(execCtx, envConfig.Image, image.PullOptions{})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to pull docker image: %v", err)
		}
		defer reader.Close()

		// Wait for the image pull to complete
		if _, err := io.Copy(io.Discard, reader); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to pull docker image: %v", err)
		}
	}

	// Parse and validate all test cases. The string inputs need to be parsed
	// to their respective data types before sending as inputs to templates
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

	// Convert test cases to JSON
	testCasesJSON, err := json.Marshal(parsedTestCases)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal test cases: %v", err)
	}

	// Generate script using environment-specific template
	script, err := envConfig.TemplateFunc(req.Code, string(testCasesJSON))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate script: %v", err)
	}

	// Create temporary file for the script
	tmpDir, err := os.MkdirTemp("", "code-*")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use environment-specific file extension
	scriptPath := filepath.Join(tmpDir, "solution"+envConfig.FileExt)
	const scriptPerm os.FileMode = 0644
	if err := os.WriteFile(scriptPath, []byte(script), scriptPerm); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to write script file: %v", err)
	}

	// Prepare command with environment-specific values
	cmd := append([]string{envConfig.CommandName}, envConfig.CommandArgs...)
	cmd = append(cmd, envConfig.MountTarget)

	// Create container
	containerConfig := &container.Config{
		Image: envConfig.Image,
		Cmd:   cmd,
	}

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:   engineConstants.ContainerMemoryLimit,
			NanoCPUs: engineConstants.ContainerCPULimit,
		},
		SecurityOpt: []string{"no-new-privileges"},
		NetworkMode: "none",
		Mounts: []mount.Mount{
			{
				Type: mount.TypeBind,
				// Mount source has to be a host path, not a path inside engineService container
				Source:   filepath.Join(env.HOST_TMP_MOUNT_PATH, scriptPath),
				Target:   envConfig.MountTarget,
				ReadOnly: true,
			},
		},
	}

	resp, err := s.dockerClient.ContainerCreate(execCtx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create container: %v", err)
	}

	// Clean up container after execution
	defer func() {
		removeOptions := container.RemoveOptions{Force: true}
		if err := s.dockerClient.ContainerRemove(context.Background(), resp.ID, removeOptions); err != nil {
			fmt.Printf("failed to remove container: %v\n", err)
		}
	}()

	_, err = s.startContainerAndWait(&execCtx, resp.ID)
	if err != nil {
		return nil, err
	}

	// Get container logs
	logs, err := s.dockerClient.ContainerLogs(execCtx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get container logs: %v", err)
	}
	defer logs.Close()

	stdout, stderr := splitLogs(logs)

	// Check for compilation errors first
	if strings.Contains(stderr, "SyntaxError") {
		return &proto.RunCodeResponse{
			Compilation: &proto.CompilationResult{
				Success: false,
				Stderr:  &stderr,
			},
			Results: nil,
		}, nil
	}

	// Extract results from stdout
	results, err := extractResults(stdout)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to extract results: %v", err)
	}

	// Note: Make sure all streams have the start and end sequences
	stdoutParts := ExtractBetween(stdout, templates.TEST_CASE_START+"\n", templates.TEST_CASE_END)

	testCaseResults := prepareTestCaseResults(results, stdoutParts, len(req.TestCases))

	return &proto.RunCodeResponse{
		Compilation: &proto.CompilationResult{
			Success: true,
		},
		Results: testCaseResults,
	}, nil
}

// fetchInputDefinitions retrieves the InputDefinitions array
// from the JSONB Question field for a given question ID.
func fetchInputDefinitions(db *gorm.DB, questionHash string) ([]structs.InputDefinition, error) {
	type QuestionFields struct {
		Type      proto.QuestionType `gorm:"column:type"`
		InputDefs []byte             `gorm:"column:input_defs"`
	}

	var fields QuestionFields
	if err := db.Model(&models.Question{}).
		Select("type, question->>'input_definitions' as input_defs").
		Joins("INNER JOIN question_hashes qh ON qh.id = questions.id").
		Where("qh.hash = ?", questionHash).
		Take(&fields).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch question: %v", err)
	}

	if fields.Type != proto.QuestionType_CODING {
		return nil, status.Errorf(codes.InvalidArgument, "question type must be coding, got type: %d", fields.Type)
	}

	var inputDefinitions []structs.InputDefinition
	if err := json.Unmarshal(fields.InputDefs, &inputDefinitions); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal input definitions: %v", err)
	}

	return inputDefinitions, nil
}

// parseTestCase parses all inputs in a test case according to their input definitions
func parseTestCase(testCaseInputs []string, inputDefs []structs.InputDefinition) ([]any, error) {
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
func parseValue(input string, paramType constants.ParameterType, items *[]structs.ParameterItem) (any, error) {
	switch paramType {
	case constants.PARAMETER_TYPE_NUMBER:
		return strconv.ParseFloat(input, 64)
	case constants.PARAMETER_TYPE_STRING:
		return input, nil
	case constants.PARAMETER_TYPE_BOOLEAN:
		return strconv.ParseBool(input)
	case constants.PARAMETER_TYPE_ARRAY:
		if items == nil {
			return nil, status.Errorf(codes.InvalidArgument, "items definition missing for array type")
		}
		var rawArray []any
		if err := json.Unmarshal([]byte(input), &rawArray); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid array format: %v", err)
		}

		// Parse each array element based on the item type
		itemType := (*items)[0].Type // Assuming first item defines the array element type
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

func (s *EngineServer) startContainerAndWait(execCtx *context.Context, containerID string) (*int64, error) {
	// Start container
	if err := s.dockerClient.ContainerStart(*execCtx, containerID, container.StartOptions{}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start container: %v", err)
	}

	// Wait for container to finish
	statusCh, errCh := s.dockerClient.ContainerWait(*execCtx, containerID, container.WaitConditionNotRunning)
	var statusCode int64
	select {
	case err := <-errCh:
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error waiting for container: %v", err)
		}
	case status := <-statusCh:
		statusCode = status.StatusCode
	}

	return &statusCode, nil
}

// splitLogs separates combined Docker logs into stdout and stderr.
func splitLogs(logs io.Reader) (string, string) {
	var stdout, stderr strings.Builder
	reader := bufio.NewReader(logs) // Use bufio.Reader to handler partial reads and incomplete headers
	headerBuf := make([]byte, 8)    // Temporary buffer for header storage

	for {
		// Read the 8-byte header
		_, err := io.ReadFull(reader, headerBuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Sprintf("error reading logs: %v", err)
		}

		// Extract header info
		header := headerBuf[0] // First byte determines stdout/stderr

		// Read the actual log message
		content, err := reader.ReadBytes('\n') // Read until newline
		if err != nil && err != io.EOF {
			return "", fmt.Sprintf("error reading log content: %v", err)
		}

		if header == 1 {
			stdout.Write(content)
		} else if header == 2 {
			stderr.Write(content)
		}
	}

	return stdout.String(), stderr.String()
}

func extractResults(stdout string) ([]TestResult, error) {
	start := strings.Index(stdout, templates.RESULTS_START)
	end := strings.Index(stdout, templates.RESULTS_END)
	if start == -1 || end == -1 {
		return nil, nil
	}

	startOffset := len(templates.RESULTS_START) + 1 // Add 1 for the \n character
	jsonStr := stdout[start+startOffset : end]

	// Note: parsed inputs are sent to the templates.
	// Make sure the templates stringify the inputs before adding to results
	var results []TestResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(jsonStr)), &results); err != nil {
		return nil, err
	}
	return results, nil
}

// ExtractBetween takes a full string, start sequence, end sequence, and returns an array of extracted substrings in between start sequence and end sequence.
//
// ExtractBetween scans the string character by character, detecting start and end sequences.
func ExtractBetween(text, startSeq, endSeq string) []string {
	var results []string
	var buffer strings.Builder
	inCaptureMode := false

	for i := 0; i < len(text); i++ {
		// Check if start sequence is forming
		if strings.HasPrefix(text[i:], startSeq) {
			inCaptureMode = true
			i += len(startSeq) - 1 // Skip past the start sequence
			continue
		}

		// If we are recording, add characters to buffer
		if inCaptureMode {
			// Check if end sequence is forming
			if strings.HasPrefix(text[i:], endSeq) {
				results = append(results, buffer.String()) // Save recorded segment
				buffer.Reset()                             // Clear buffer for next segment
				inCaptureMode = false
				i += len(endSeq) - 1 // Skip past the end sequence
				continue
			}
			buffer.WriteByte(text[i])
		}
	}

	return results
}

func prepareTestCaseResults(results []TestResult, stdoutParts []string, resultsCount int) []*proto.TestCaseResult {
	testCaseResults := make([]*proto.TestCaseResult, 0, resultsCount)
	for i, result := range results {
		status := proto.ExecutionStatus_RUNTIME_ERROR
		if result.Error != "" {
			status = proto.ExecutionStatus_RUNTIME_ERROR
		} else if result.Match {
			status = proto.ExecutionStatus_SUCCESS
		} else {
			status = proto.ExecutionStatus_WRONG_ANSWER
		}

		testCaseResults = append(testCaseResults, &proto.TestCaseResult{
			Status:         status,
			Inputs:         result.Inputs,
			Output:         result.Output,
			ExpectedOutput: result.ExpectedOutput,
			ExecutionTime:  result.ExecutionTime,
			Stdout:         stdoutParts[i],
			Error:          result.Error,
		})
	}

	return testCaseResults
}
