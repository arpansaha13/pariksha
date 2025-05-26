package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/proto"
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
	engineConstants.EnvNode: {
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
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}
	return &EngineServer{dockerClient: cli}, nil
}

func (s *EngineServer) RunCode(ctx context.Context, req *proto.RunCodeRequest) (*proto.RunCodeResponse, error) {
	// Get environment config
	envConfig, ok := envConfigs[req.Environment]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported environment: %s", req.Environment)
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(engineConstants.ExecutionTimeout)*time.Second)
	defer cancel()

	// Check if image exists locally
	_, _, err := s.dockerClient.ImageInspectWithRaw(execCtx, envConfig.Image)
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

	// Convert test cases to JSON
	testCasesJSON, err := json.Marshal(req.TestCases)
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
				Type:     mount.TypeBind,
				Source:   filepath.Join(engineConstants.HOST_TMP_MOUNT_PATH, scriptPath),
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

	// Start container
	startTime := time.Now()
	if err := s.dockerClient.ContainerStart(execCtx, resp.ID, container.StartOptions{}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start container: %v", err)
	}

	// Wait for container to finish
	statusCh, errCh := s.dockerClient.ContainerWait(execCtx, resp.ID, container.WaitConditionNotRunning)
	var statusCode int64
	select {
	case err := <-errCh:
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error waiting for container: %v", err)
		}
	case status := <-statusCh:
		statusCode = status.StatusCode
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
	executionTime := time.Since(startTime).Milliseconds()

	return &proto.RunCodeResponse{
		Stdout:        stdout,
		Stderr:        stderr,
		ExitCode:      int32(statusCode),
		ExecutionTime: executionTime,
	}, nil
}
