package constants

const StatusInvalidToken = 498

const (
	DEFAULT_API_PORT                 string = "4000"
	DEFAULT_OTP_EXPIRES_IN_MINUTES   string = "15"
	DEFAULT_SESSION_EXPIRES_IN_HOURS string = "24"
	DEFAULT_SESSION_COOKIE_NAME      string = "token"
)

const (
	VERIFICATION_OTP_LENGTH  int = 6
	VERIFICATION_HASH_LENGTH int = 10
)

const (
	PAPER_TYPE_OWNER  string = "OWNER"
	PAPER_TYPE_SHARED string = "SHARED"
)
