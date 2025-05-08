package dtos

import (
	"encoding/json"
)

type UpdatePaperDto struct {
	Title           string `json:"title,omitempty"`
	DurationMinutes int32  `json:"duration_minutes,omitempty"`
}

type PaperResponseDto struct {
	ID              int64           `json:"id"`
	Title           string          `json:"title"`
	MaxScore        int             `json:"max_score"`
	DurationMinutes int             `json:"duration_minutes"`
	QuestionCounts  json.RawMessage `json:"question_counts"`
	CreatedBy       int64           `json:"created_by"`
}

type UpdatePaperResponseDto struct {
	ID              int64           `json:"id"`
	Title           string          `json:"title"`
	MaxScore        int             `json:"max_score"`
	DurationMinutes int             `json:"duration_minutes"`
	QuestionCounts  json.RawMessage `json:"question_counts"`
	CreatedBy       int64           `json:"created_by"`
}

type PaperPermissionsDto struct {
	CanRead  bool `json:"can_read"`
	CanWrite bool `json:"can_write"`
}
