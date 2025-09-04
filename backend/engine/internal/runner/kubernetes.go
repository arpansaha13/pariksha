package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/utils/ptr"
	engineConstants "pariksha/engine/internal/constants"
	"pariksha/engine/internal/templates"
)

type Kubernetes struct {
	clientset *kubernetes.Clientset
	namespace string
}

func NewKubernetes() (*Kubernetes, error) {
	// Load in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to load in-cluster config: %v", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create kubernetes clientset: %v", err)
	}

	// Get namespace from environment or use default
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespace = "pariksha"
	}

	return &Kubernetes{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

func (r *Kubernetes) Run(args *RunnerArg) (*proto.RunCodeResponse, error) {
	envConfig, ok := envConfigs[constants.LangNode]
	if !ok {
		return nil, status.Errorf(codes.Internal, "could not find node env config")
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

	// Create a unique job name
	jobName := fmt.Sprintf("code-execution-%d", time.Now().UnixNano())

	// Create ConfigMap for the script
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: r.namespace,
		},
		Data: map[string]string{
			"solution" + envConfig.FileExt: script,
		},
	}

	_, err = r.clientset.CoreV1().ConfigMaps(r.namespace).Create(context.Background(), configMap, metav1.CreateOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create configmap: %v", err)
	}

	// Clean up ConfigMap after execution
	defer func() {
		r.clientset.CoreV1().ConfigMaps(r.namespace).Delete(context.Background(), jobName, metav1.DeleteOptions{})
	}()

	job := r.createJob(jobName, envConfig)

	_, err = r.clientset.BatchV1().Jobs(r.namespace).Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create job: %v", err)
	}

	// Clean up Job after execution
	defer func() {
		r.clientset.BatchV1().Jobs(r.namespace).Delete(context.Background(), jobName, metav1.DeleteOptions{})
	}()

	err = r.waitForJobCompletion(jobName)
	if err != nil {
		return nil, err
	}

	stdout, stderr, err := r.getJobLogs(jobName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get job logs: %v", err)
	}

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

	// Extract stdout parts between test case markers
	stdoutParts := r.extractBetween(stdout, templates.TEST_CASE_START+"\n", templates.TEST_CASE_END)

	testCaseResults := r.prepareTestCaseResults(results, stdoutParts, args.TestCasesCount)

	return &proto.RunCodeResponse{
		Compilation: &proto.CompilationResult{
			Success: true,
		},
		Results: testCaseResults,
	}, nil
}

func (r *Kubernetes) createJob(jobName string, envConfig environmentConfig) *batchv1.Job {
	// Prepare command with environment-specific values
	cmd := append([]string{envConfig.CommandName}, envConfig.CommandArgs...)
	cmd = append(cmd, envConfig.MountTarget)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: r.namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            ptr.Int32(0),
			TTLSecondsAfterFinished: ptr.Int32(10),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "engine-image-pull-secret",
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "code-executor",
							Image:   envConfig.Image,
							Command: cmd,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("128Mi"),
									corev1.ResourceCPU:    resource.MustParse("100m"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("64Mi"),
									corev1.ResourceCPU:    resource.MustParse("50m"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "code-volume",
									MountPath: "/code",
									ReadOnly:  true,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								RunAsNonRoot: ptr.Bool(true),
								RunAsUser:    ptr.Int64(1000),
								RunAsGroup:   ptr.Int64(1000),
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "code-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: jobName,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *Kubernetes) waitForJobCompletion(jobName string) error {
	timeout := time.Duration(engineConstants.ExecutionTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return status.Errorf(codes.DeadlineExceeded, "job execution timed out")
		default:
			job, err := r.clientset.BatchV1().Jobs(r.namespace).Get(ctx, jobName, metav1.GetOptions{})
			if err != nil {
				return status.Errorf(codes.Internal, "failed to get job status: %v", err)
			}

			if job.Status.Succeeded > 0 {
				return nil
			}

			if job.Status.Failed > 0 {
				return status.Errorf(codes.Internal, "job execution failed")
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (r *Kubernetes) getJobLogs(jobName string) (string, string, error) {
	// Get the pod name for the job
	pods, err := r.clientset.CoreV1().Pods(r.namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return "", "", err
	}

	if len(pods.Items) == 0 {
		return "", "", fmt.Errorf("no pods found for job %s", jobName)
	}

	podName := pods.Items[0].Name

	// Get logs from the pod
	req := r.clientset.CoreV1().Pods(r.namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: "code-executor",
	})

	logs, err := req.Stream(context.Background())
	if err != nil {
		return "", "", err
	}
	defer logs.Close()

	// Split logs into stdout and stderr
	stdout, stderr := r.splitLogs(logs)

	return stdout, stderr, nil
}

// splitLogs separates combined Kubernetes logs into stdout and stderr.
func (r *Kubernetes) splitLogs(logs io.Reader) (string, string) {
	var stdout, stderr strings.Builder
	scanner := bufio.NewScanner(logs)

	for scanner.Scan() {
		line := scanner.Text()
		// Kubernetes logs don't have the Docker-style headers, so we assume all output is stdout
		// unless it contains error indicators
		if strings.Contains(strings.ToLower(line), "error") || strings.Contains(line, "SyntaxError") {
			stderr.WriteString(line + "\n")
		} else {
			stdout.WriteString(line + "\n")
		}
	}

	return stdout.String(), stderr.String()
}

func (r *Kubernetes) extractResults(stdout string) ([]testResult, error) {
	start := strings.Index(stdout, templates.RESULTS_START)
	end := strings.Index(stdout, templates.RESULTS_END)
	if start == -1 || end == -1 {
		return nil, nil
	}

	startOffset := len(templates.RESULTS_START) + 1 // Add 1 for the \n character
	jsonStr := stdout[start+startOffset : end]

	var results []testResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(jsonStr)), &results); err != nil {
		return nil, err
	}
	return results, nil
}

// extractBetween takes a full string, start sequence, end sequence, and returns an array of extracted substrings in between start sequence and end sequence.
func (r *Kubernetes) extractBetween(text, startSeq, endSeq string) []string {
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

func (r *Kubernetes) prepareTestCaseResults(results []testResult, stdoutParts []string, resultsCount int16) []*proto.TestCaseResult {
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
