package constants

// Container resource limits
const (
	// Memory limits (in bytes)
	ContainerMemoryLimit int64 = 128 * 1024 * 1024 // 128MB

	// CPU limits (in nano CPUs)
	ContainerCPULimit int64 = 100000000 // 0.1 CPU

	// Execution timeout (in seconds)
	ExecutionTimeout int64 = 20
)

// Kubernetes job resource limits (for code execution jobs)
const (
	// CPU request per job (in millicores)
	JobCPURequest string = "40m"

	// CPU limit per job (in millicores)
	JobCPULimit string = "80m"

	// Memory request per job (in MiB)
	JobMemoryRequest string = "48Mi"

	// Memory limit per job (in MiB)
	JobMemoryLimit string = "96Mi"
)
