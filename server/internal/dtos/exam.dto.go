package dtos

import (
	"time"
)

type CreateExamDto struct {
	Title              string    `json:"title" validate:"required"`
	StartsAt           time.Time `json:"starts_at" validate:"required"`
	EndsAt             time.Time `json:"ends_at" validate:"required"`
	MaxCandidatesCount int       `json:"max_candidates_count" validate:"required"`
	Type               string    `json:"type" validate:"required"`
	PaperID            int       `json:"paper_id" validate:"required"`
}

type UpdateExamDto struct {
	Title              string    `json:"title"`
	StartsAt           time.Time `json:"starts_at"`
	EndsAt             time.Time `json:"ends_at"`
	MaxCandidatesCount int       `json:"max_candidates_count"`
	Type               string    `json:"type"`
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

type ExamParticipantResponse struct {
	ID           int       `json:"id"`
	UserID       int       `json:"userId"`
	FirstName    string    `json:"firstName,omitempty"`
	LastName     string    `json:"lastName,omitempty"`
	Email        string    `json:"email"`
	Status       int       `json:"status"`
	ScoreAwarded int       `json:"scoreAwarded"`
	StartedAt    time.Time `json:"startedAt,omitempty"`
	EndedAt      time.Time `json:"endedAt,omitempty"`
}

type AddExamParticipantDto struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type AddExamParticipantResponse struct {
	AddedCount     int    `json:"added_count"`
	OmittedCount   int    `json:"omitted_count"`
	MaxLimitReason string `json:"max_limit_reason,omitempty"`
}
