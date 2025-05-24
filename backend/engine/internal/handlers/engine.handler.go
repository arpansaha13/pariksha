package handlers

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/proto"
	engineConstants "pariksha/engine/internal/constants"
)

type EngineServer struct {
	proto.UnimplementedEngineServer
	dockerClient *client.Client
}

var mapEnvToImage = map[string]string{
	engineConstants.EnvNode: engineConstants.NodeImage,
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
	dockerImage, ok := mapEnvToImage[req.Environment]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported environment: %s", req.Environment)
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(engineConstants.ExecutionTimeout)*time.Second)
	defer cancel()

	// Check if image exists locally
	// type *client.Client has no field or method ImageInspect
	_, _, err := s.dockerClient.ImageInspectWithRaw(execCtx, dockerImage)
	if err != nil {
		// Image not found locally, pull it
		reader, err := s.dockerClient.ImagePull(execCtx, dockerImage, image.PullOptions{})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to pull docker image: %v", err)
		}
		defer reader.Close()

		// Wait for the image pull to complete
		if _, err := io.Copy(io.Discard, reader); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to pull docker image: %v", err)
		}
	}

	// Create container
	containerConfig := &container.Config{
		Image: dockerImage,
		Cmd:   []string{"node", "-e", req.Code},
	}

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:   engineConstants.ContainerMemoryLimit,
			NanoCPUs: engineConstants.ContainerCPULimit,
		},
		SecurityOpt: []string{"no-new-privileges"},
		NetworkMode: "none", // Disable network access
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
