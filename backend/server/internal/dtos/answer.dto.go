package dtos

import (
	"encoding/json"
)

type UpsertAnswerDto struct {
	Answer     *json.RawMessage `json:"answer"`
	QuestionID int64            `json:"question_id" validate:"required"`
}

type AnswerResponseDto struct {
	ID                int64           `json:"id"`
	ExamParticipantID int64           `json:"exam_participant_id"`
	QuestionID        int64           `json:"question_id"`
	Answer            json.RawMessage `json:"answer"`
	ScoreAwarded      int             `json:"score_awarded"`
	Comments          string          `json:"comments"`
}

type AnswerMinimalResponseDto struct {
	ID         int64           `json:"id"`
	Answer     json.RawMessage `json:"answer"`
	QuestionID int64           `json:"question_id"`
}

type UpdateAnswerForEvaluationDto struct {
	AnswerID  int64   `json:"answer_id" validate:"required"`
	NewScore  *int    `json:"new_score"`
	Evaluated *bool   `json:"evaluated"`
	Comments  *string `json:"comments"`
}

type GetAnswerForEvaluationDto struct {
	ParticipantID int64 `json:"participantId" validate:"required"`
	QuestionID    int64 `json:"questionId" validate:"required"`
}
