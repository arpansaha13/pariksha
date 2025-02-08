package dtos

import "time"

type AnswerDTO struct {
	Answer      string    `json:"answer" validate:"required"`
	SubmittedAt time.Time `json:"submitted_at" validate:"required"`
	QuestionID  int       `json:"question_id" validate:"required"`
}
