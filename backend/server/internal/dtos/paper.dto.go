package dtos

import (
	"encoding/json"
)

type UpdatePaperDto struct {
	Title           string `json:"title,omitempty"`
	DurationMinutes int32  `json:"duration_minutes,omitempty"`
}

type PaperResponse struct {
	ID              int64           `json:"id"`
	Title           string          `json:"title"`
	MaxScore        int             `json:"max_score"`
	DurationMinutes int             `json:"duration_minutes"`
	QuestionCounts  json.RawMessage `json:"question_counts"`
}

type UpdatePaperResponse struct {
	ID              int64           `json:"id"`
	Title           string          `json:"title"`
	MaxScore        int             `json:"max_score"`
	DurationMinutes int             `json:"duration_minutes"`
	QuestionCounts  json.RawMessage `json:"question_counts"`
}
