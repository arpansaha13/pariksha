package dtos

import (
	"encoding/json"
	"time"
)

type CreateExamDto struct {
	Title           string    `json:"title" validate:"required"`
	StartsAt        time.Time `json:"starts_at" validate:"required"`
	EndsAt          time.Time `json:"ends_at" validate:"required"`
	Type            string    `json:"type"`
	PaperID         int64     `json:"paper_id" validate:"required"`
	DurationMinutes int32     `json:"duration_minutes" validate:"required,gt=0"`
}

type UpdateExamDto struct {
	Title           string    `json:"title"`
	StartsAt        time.Time `json:"starts_at"`
	EndsAt          time.Time `json:"ends_at"`
	Type            string    `json:"type"`
	DurationMinutes *int32    `json:"duration_minutes" validate:"omitempty,gt=0"`
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
	DurationMinutes    int32     `json:"duration_minutes"`
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
	AccessType        string `json:"access_type"` // "OWNER" or "PARTICIPANT"
	ParticipantStatus *int   `json:"participant_status,omitempty"`
}

type ExamQuestionMinimalResponse struct {
	QuestionID int64 `json:"id"`
	CategoryID int64 `json:"category_id"`
	Order      int   `json:"order"`
}

type ExamCategoriesResponse struct {
	CategoryID int64  `json:"id"`
	Name       string `json:"name"`
	Order      int32  `json:"order"`
}

type ExamQuestionResponse struct {
	ID         int64           `json:"id"`
	Question   json.RawMessage `json:"question"`
	CategoryID int64           `json:"category_id"`
	Type       string          `json:"type"`
	MaxScore   int             `json:"max_score"`
}

type CheckExamParticipantResponse struct {
	ParticipantStatus int `json:"participant_status"`
}

type GetExamParticipantResponse struct {
	ID               int64     `json:"id"`
	StartedAt        time.Time `json:"started_at,omitempty"`
	ScheduledEndTime time.Time `json:"scheduled_end_time,omitempty"`
}
