package dtos

import "time"

type UpsertAnswerDto struct {
	Answer      string    `json:"answer" validate:"required"`
	SubmittedAt time.Time `json:"submitted_at" validate:"required"`
	QuestionID  int64     `json:"question_id" validate:"required"`
}

type AnswerResponse struct {
	ID                int64
	ExamParticipantID int64
	QuestionID        int64
	Answer            string
	ScoreAwarded      int
	Comments          string
}

type UpdateAnswerForEvaluationDTO struct {
	AnswerID  int64   `json:"answer_id" validate:"required"`
	NewScore  *int    `json:"new_score"`
	Evaluated *bool   `json:"evaluated"`
	Comments  *string `json:"comments"`
}
