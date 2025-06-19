package constants

const (
	HEADER_SESSION_KEY string = "session-key"
	HEADER_CSRF_TOKEN  string = "csrf-token"
	HEADER_EXPIRES_AT  string = "expires-at"
)

const (
	// API Token for communication between exam service and paper service
	X_EXAM_API_TOKEN string = "x-exam-api-token"

	// API Token for communication between engine service and paper service
	X_ENGINE_API_TOKEN string = "x-engine-api-token"
)
