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
	ExamID    int    `json:"exam_id"`
	UserID    int    `json:"user_id,omitempty"`
	Email     string `json:"email,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}
