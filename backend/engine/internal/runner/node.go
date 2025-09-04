package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	engineConstants "pariksha/engine/internal/constants"
	"pariksha/engine/internal/templates"
)

type Node struct {
	dockerClient *client.Client
}

// Kept for running tests (for now)
func NewNode(dockerClient *client.Client) *Node {
	return &Node{
		dockerClient: dockerClient,
	}
}

func (r Node) Run(args *RunnerArg) (*proto.RunCodeResponse, error) {
	envConfig, ok := envConfigs[constants.LangNode]
	if !ok {
		return nil, status.Errorf(codes.Internal, "could not find node env config")
	}

	// Check if image exists locally
	_, _, err := r.dockerClient.ImageInspectWithRaw(context.Background(), envConfig.Image)
	if err != nil {
		// Image not found locally, pull it
		reader, err := r.dockerClient.ImagePull(context.Background(), envConfig.Image, image.PullOptions{})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to pull docker image: %v", err)
		}
		defer reader.Close()

		// Wait for the image pull to complete
		if _, err := io.Copy(io.Discard, reader); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to pull docker image: %v", err)
		}
	}

	// Convert test cases to JSON
	testCasesJSON, err := json.Marshal(args.ParsedTestCases)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal test cases: %v", err)
	}

	// Generate script using environment-specific template
	script, err := envConfig.TemplateFunc(args.Code, string(testCasesJSON))
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
				Source:   filepath.Join("", scriptPath),
				Target:   envConfig.MountTarget,
				ReadOnly: true,
			},
		},
	}

	resp, err := r.dockerClient.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create container: %v", err)
	}

	// Clean up container after execution
	defer func() {
		removeOptions := container.RemoveOptions{Force: true}
		if err := r.dockerClient.ContainerRemove(context.Background(), resp.ID, removeOptions); err != nil {
			fmt.Printf("failed to remove container: %v\n", err)
		}
	}()

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(context.Background(), time.Duration(engineConstants.ExecutionTimeout)*time.Second)
	defer cancel()

	_, err = r.startContainerAndWait(execCtx, resp.ID)
	if err != nil {
		return nil, err
	}

	// Get container logs
	logs, err := r.dockerClient.ContainerLogs(context.Background(), resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get container logs: %v", err)
	}
	defer logs.Close()

	stdout, stderr := r.splitLogs(logs)

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
	results, err := r.extractResults(stdout)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to extract results: %v", err)
	}

	// Note: Make sure all streams have the start and end sequences
	stdoutParts := r.extractBetween(stdout, templates.TEST_CASE_START+"\n", templates.TEST_CASE_END)

	testCaseResults := r.prepareTestCaseResults(results, stdoutParts, args.TestCasesCount)

	return &proto.RunCodeResponse{
		Compilation: &proto.CompilationResult{
			Success: true,
		},
		Results: testCaseResults,
	}, nil
}

func (r *Node) startContainerAndWait(execCtx context.Context, containerID string) (*int64, error) {
	// Start container
	if err := r.dockerClient.ContainerStart(execCtx, containerID, container.StartOptions{}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start container: %v", err)
	}

	// Wait for container to finish
	statusCh, errCh := r.dockerClient.ContainerWait(execCtx, containerID, container.WaitConditionNotRunning)
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
func (r *Node) splitLogs(logs io.Reader) (string, string) {
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

func (r *Node) extractResults(stdout string) ([]testResult, error) {
	start := strings.Index(stdout, templates.RESULTS_START)
	end := strings.Index(stdout, templates.RESULTS_END)
	if start == -1 || end == -1 {
		return nil, nil
	}

	startOffset := len(templates.RESULTS_START) + 1 // Add 1 for the \n character
	jsonStr := stdout[start+startOffset : end]

	// Note: parsed inputs are sent to the templates.
	// Make sure the templates stringify the inputs before adding to results
	var results []testResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(jsonStr)), &results); err != nil {
		return nil, err
	}
	return results, nil
}

// extractBetween takes a full string, start sequence, end sequence, and returns an array of extracted substrings in between start sequence and end sequence.
//
// extractBetween scans the string character by character, detecting start and end sequences.
func (r *Node) extractBetween(text, startSeq, endSeq string) []string {
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

func (r *Node) prepareTestCaseResults(results []testResult, stdoutParts []string, resultsCount int16) []*proto.TestCaseResult {
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
