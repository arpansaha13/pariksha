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

type ExamResponseDto struct {
	ID                 int64     `json:"id"`
	Title              string    `json:"title"`
	StartsAt           time.Time `json:"starts_at"`
	EndsAt             time.Time `json:"ends_at"`
	CreatedBy          int64     `json:"created_by"`
	Type               string    `json:"type"`
	MaxCandidatesCount int32     `json:"max_candidates_count"`
	PaperID            int64     `json:"paper_id"`
	DurationMinutes    int32     `json:"duration_minutes"`
	MaxScore           int32     `json:"max_score"`
}

type ExamParticipantResponseDto struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	Status           int32      `json:"status"`
	ScoreAwarded     int32      `json:"score_awarded"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	ScheduledEndTime *time.Time `json:"scheduled_end_time,omitempty"`

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

type AddExamParticipantResponseDto struct {
	ID           int64 `json:"id"`
	UserID       int64 `json:"user_id"`
	Status       int32 `json:"status"`
	ScoreAwarded int32 `json:"score_awarded"`
}

type ExamQuestionMinimalResponseDto struct {
	QuestionID int64 `json:"id"`
	CategoryID int64 `json:"category_id"`
	Order      int32 `json:"order"`
	MaxScore   int32 `json:"max_score"`
}

type ExamCategoriesResponseDto struct {
	CategoryID int64  `json:"id"`
	Name       string `json:"name"`
	Order      int32  `json:"order"`
}

type ExamQuestionResponseDto struct {
	ID       int64           `json:"id"`
	Question json.RawMessage `json:"question"`
	Type     string          `json:"type"`
}

type GetExamParticipantResponseDto struct {
	ID               int64     `json:"id"`
	StartedAt        time.Time `json:"started_at,omitempty"`
	ScheduledEndTime time.Time `json:"scheduled_end_time,omitempty"`
}

type ExamPermissionResponseDto struct {
	CanRead           bool `json:"can_read"`
	CanWrite          bool `json:"can_write"`
	CanParticipate    bool `json:"can_participate"`
	CanEvaluate       bool `json:"can_evaluate"`
	ParticipantStatus *int `json:"participant_status,omitempty"`
}

type DeleteExamsDto struct {
	ExamIds []int64 `json:"exam_ids" validate:"required,min=1"`
}
