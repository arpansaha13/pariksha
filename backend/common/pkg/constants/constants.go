package constants

const StatusInvalidToken = 498

const (
	GO_ENV_DEV        string = "development"
	GO_ENV_PROD       string = "production"
	GO_ENV_DOCKER_DEV string = "docker-development"
	GO_ENV_TEST       string = "test"
)

const (
	VERIFICATION_OTP_LENGTH int16 = 6
	CSRF_TOKEN_LENGTH       int16 = 32
)

const (
	EXAM_ACCESS_TYPE_LINK   string = "LINK"
	EXAM_ACCESS_TYPE_INVITE string = "INVITE"
)

const (
	PARTICIPANT_STATUS_UNATTENDED int16 = 0
	PARTICIPANT_STATUS_INVITED    int16 = 1
	PARTICIPANT_STATUS_STARTED    int16 = 2
	PARTICIPANT_STATUS_ENDED      int16 = 3
	PARTICIPANT_STATUS_EVALUATED  int16 = 4
)

const (
	OTP_PURPOSE_SIGNUP          int16 = 1
	OTP_PURPOSE_LOGIN           int16 = 2
	OTP_PURPOSE_FORGOT_PASSWORD int16 = 3
)

const (
	EXAM_QUEUE_NAME                   string = "exam_queue"
	EXAM_QUEUE_TASK_PREPARE_QUESTIONS string = "exam_queue:prepare_questions"
	EXAM_QUEUE_TASK_AUTO_END          string = "exam_queue:auto_end"
	EXAM_QUEUE_TASK_DELETE_EXAMS      string = "exam_queue:delete_exams"
)

const (
	MAIL_QUEUE_NAME                      string = "mail_queue"
	MAIL_QUEUE_TASK_SEND_VERIFICATION    string = "mail_queue:send_verification"
	MAIL_QUEUE_TASK_SEND_LOGIN_OTP       string = "mail_queue:send_login_otp"
	MAIL_QUEUE_TASK_SEND_FORGOT_PASSWORD string = "mail_queue:send_forgot_password"
	MAIL_QUEUE_TASK_SEND_RESET_PASSWORD  string = "mail_queue:send_reset_password"
)

const (
	MAX_EXAM_DURATION_MINUTES int16 = 1440 // 24 hours
)
