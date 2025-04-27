package dtos

import (
	"encoding/json"
)

type UpsertAnswerDto struct {
	Answer     json.RawMessage `json:"answer" validate:"required"`
	QuestionID int64           `json:"question_id" validate:"required"`
}

type AnswerResponse struct {
	ID                int64           `json:"id"`
	ExamParticipantID int64           `json:"exam_participant_id"`
	QuestionID        int64           `json:"question_id"`
	Answer            json.RawMessage `json:"answer"`
	ScoreAwarded      int             `json:"score_awarded"`
	Comments          string          `json:"comments"`
}

type PartialAnswerResponse struct {
	ID     int64           `json:"id"`
	Answer json.RawMessage `json:"answer"`
}

type UpdateAnswerForEvaluationDTO struct {
	AnswerID  int64   `json:"answer_id" validate:"required"`
	NewScore  *int    `json:"new_score"`
	Evaluated *bool   `json:"evaluated"`
	Comments  *string `json:"comments"`
}
