package dtos

import "time"

type CreateExamDto struct {
	Title              string    `json:"title"`
	StartsAt           time.Time `json:"starts_at"`
	EndsAt             time.Time `json:"ends_at"`
	MaxCandidatesCount int       `json:"max_candidates_count"`
	Type               string    `json:"type"`
	PaperID            int       `json:"paper_id"`
}

type ExamResponse struct {
	ID                 int       `json:"id"`
	Title              string    `json:"title"`
	StartsAt           time.Time `json:"starts_at"`
	EndsAt             time.Time `json:"ends_at"`
	CreatedBy          int       `json:"created_by"`
	Type               string    `json:"type"`
	MaxCandidatesCount int       `json:"max_candidates_count"`
	PaperID            int       `json:"paper_id"`
}
