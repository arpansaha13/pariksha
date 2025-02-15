package constants

const StatusInvalidToken = 498

const (
	GO_ENV_DEV        string = "development"
	GO_ENV_PROD       string = "production"
	GO_ENV_DOCKER_DEV string = "docker-development"
)

const (
	VERIFICATION_OTP_LENGTH  int = 6
	VERIFICATION_HASH_LENGTH int = 10
	CSRF_TOKEN_LENGTH        int = 32
)

const (
	PAPER_OWNERSHIP_TYPE_OWNER  string = "OWNER"
	PAPER_OWNERSHIP_TYPE_SHARED string = "SHARED"
)

const (
	QUESTION_TYPE_MCQ   string = "MCQ"
	QUESTION_TYPE_SHORT string = "SHORT"
	QUESTION_TYPE_LONG  string = "LONG"
)

const (
	EXAM_TYPE_OPEN   string = "OPEN"
	EXAM_TYPE_INVITE string = "INVITE"
)

const (
	PARTICIPANT_STATUS_UNATTENDED = 0
	PARTICIPANT_STATUS_INVITED    = 1
	PARTICIPANT_STATUS_STARTED    = 2
	PARTICIPANT_STATUS_ENDED      = 3
	PARTICIPANT_STATUS_EVALUATED  = 4
)

const (
	OTP_PURPOSE_SIGNUP          = 1
	OTP_PURPOSE_LOGIN           = 2
	OTP_PURPOSE_FORGOT_PASSWORD = 3
)
