package constants

// Container resource limits
const (
	// Memory limits (in bytes)
	ContainerMemoryLimit int64 = 128 * 1024 * 1024 // 128MB

	// CPU limits (in nano CPUs)
	ContainerCPULimit int64 = 100000000 // 0.1 CPU

	// Execution timeout (in seconds)
	ExecutionTimeout int64 = 10
)
