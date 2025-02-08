package dtos

import "time"

type AnswerDTO struct {
	Answer      string    `json:"answer" validate:"required"`
	SubmittedAt time.Time `json:"submitted_at" validate:"required"`
	QuestionID  int       `json:"question_id" validate:"required"`
}

type AnswerResponse struct {
	ID                int
	ExamParticipantID int
	QuestionID        int
	Answer            string
	ScoreAwarded      int
	Comments          string
}

type UpdateAnswerForEvaluationDTO struct {
	AnswerID  int     `json:"answer_id" validate:"required"`
	NewScore  *int    `json:"new_score"`
	Evaluated *bool   `json:"evaluated"`
	Comments  *string `json:"comments"`
}
