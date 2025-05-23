package dtos

type QuestionCountDto struct {
	MCQ        int32 `json:"mcq"`
	Subjective int32 `json:"subjective"`
}

type PaperResponseDto struct {
	ID              string           `json:"id"`
	Title           string           `json:"title"`
	MaxScore        int32            `json:"max_score"`
	DurationMinutes int32            `json:"duration_minutes"`
	QuestionCounts  QuestionCountDto `json:"question_counts"`
	CreatedBy       int64            `json:"created_by"`
}

type UpdatePaperDto struct {
	Title           string `json:"title,omitempty"`
	DurationMinutes int32  `json:"duration_minutes,omitempty"`
}

type PaperPermissionsDto struct {
	CanRead  bool `json:"can_read"`
	CanWrite bool `json:"can_write"`
}

type DeletePaperDto struct {
	PaperIDs []string `json:"paper_ids" validate:"required,min=1"`
}
