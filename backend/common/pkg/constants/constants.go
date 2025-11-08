package constants

const StatusInvalidToken = 498

const (
	GO_ENV_DEV  string = "development"
	GO_ENV_PROD string = "production"
	GO_ENV_TEST string = "test"
)

const (
	CSRF_TOKEN_LENGTH int16 = 32
)

const (
	EXAM_ACCESS_TYPE_LINK   int16 = 1
	EXAM_ACCESS_TYPE_INVITE int16 = 2
)

const (
	PARTICIPANT_STATUS_UNATTENDED int16 = 0
	PARTICIPANT_STATUS_INVITED    int16 = 1
	PARTICIPANT_STATUS_STARTED    int16 = 2
	PARTICIPANT_STATUS_ENDED      int16 = 3
	PARTICIPANT_STATUS_EVALUATED  int16 = 4
)

const (
	EXAM_QUEUE_NAME                   string = "exam_queue"
	EXAM_QUEUE_TASK_PREPARE_QUESTIONS string = "exam_queue:prepare_questions"
	EXAM_QUEUE_TASK_AUTO_END          string = "exam_queue:auto_end"
	EXAM_QUEUE_TASK_DELETE_EXAMS      string = "exam_queue:delete_exams"
)

const (
	MAX_EXAM_DURATION_MINUTES int16 = 1440 // 24 hours
)
