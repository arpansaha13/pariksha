package dtos

import (
	"time"
)

type CreateExamDto struct {
	Title    string    `json:"title" validate:"required"`
	StartsAt time.Time `json:"starts_at" validate:"required"`
	EndsAt   time.Time `json:"ends_at" validate:"required"`
	Type     string    `json:"type"`
	PaperID  int64     `json:"paper_id" validate:"required"`
}

type UpdateExamDto struct {
	Title    string    `json:"title"`
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
	Type     string    `json:"type"`
}

type ExamResponse struct {
	ID                 int64     `json:"id"`
	Title              string    `json:"title"`
	StartsAt           time.Time `json:"starts_at"`
	EndsAt             time.Time `json:"ends_at"`
	CreatedBy          int64     `json:"created_by"`
	Type               string    `json:"type"`
	MaxCandidatesCount int       `json:"max_candidates_count"`
	PaperID            int64     `json:"paper_id"`
}

type ExamParticipantResponse struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Status       int       `json:"status"`
	ScoreAwarded int       `json:"score_awarded"`
	StartedAt    time.Time `json:"started_at,omitempty"`
	EndedAt      time.Time `json:"ended_at,omitempty"`

	// From auth service
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
}

type AddExamParticipantDto struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type AddExamParticipantResponse struct {
	ID           int64 `json:"id"`
	UserID       int64 `json:"user_id"`
	Status       int   `json:"status"`
	ScoreAwarded int   `json:"score_awarded"`
}

type ExamAccessResponse struct {
	AccessType string `json:"access_type"` // "OWNER" or "PARTICIPANT"
}
