package dtos

import (
	"encoding/json"
)

type UpdatePaperDto struct {
	Title string `json:"title"`
}

type PaperResponse struct {
	ID              int                        `json:"id"`
	Title           string                     `json:"title"`
	MaxScore        int                        `json:"max_score"`
	DurationMinutes int                        `json:"duration_minutes"`
	QuestionCounts  json.RawMessage            `json:"question_counts"`
	Categories      []QuestionCategoryResponse `json:"categories"`
	PaperOwnership  PaperOwnershipResponse     `json:"ownership"`
}

type PaperOwnershipResponse struct {
	ID   int    `json:"id"`
	Path string `json:"path"`
	Type string `json:"type"`
}

type UpdatePaperResponse struct {
	ID              int             `json:"id"`
	Title           string          `json:"title"`
	MaxScore        int             `json:"max_score"`
	DurationMinutes int             `json:"duration_minutes"`
	QuestionCounts  json.RawMessage `json:"question_counts"`
}
