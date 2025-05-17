package dtos

import (
	"encoding/json"
)

type AnswerListQuestionDto struct {
	ID         int64           `json:"id"`
	Order      int32           `json:"order"`
	CategoryID int64           `json:"category_id"`
	Content    json.RawMessage `json:"content"`
	MaxScore   int32           `json:"max_score"`
}

type AnswerListAnswerDto struct {
	ID      int64           `json:"id"`
	Content json.RawMessage `json:"content"`
}

type AnswerListItemDto struct {
	Type     string                `json:"type"`
	Question AnswerListQuestionDto `json:"question"`
	Answer   *AnswerListAnswerDto  `json:"answer"`
}

type UpsertAnswerDto struct {
	Answer     *json.RawMessage `json:"answer"`
	QuestionID int64            `json:"question_id" validate:"required"`
}

type AnswerMinimalResponseDto struct {
	ID         int64           `json:"id"`
	Answer     json.RawMessage `json:"answer"`
	QuestionID int64           `json:"question_id"`
}
