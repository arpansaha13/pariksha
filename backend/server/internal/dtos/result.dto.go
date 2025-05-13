package dtos

import "encoding/json"

type ExamResultQuestionDTO struct {
	ID         int64           `json:"id"`
	Order      int32           `json:"order"`
	CategoryID int64           `json:"categoryId"`
	Type       string          `json:"type"`
	Content    json.RawMessage `json:"content"`
	MaxScore   int32           `json:"maxScore"`
}

type ExamResultAnswerDTO struct {
	Content      json.RawMessage `json:"content"`
	ScoreAwarded int32           `json:"scoreAwarded"`
	Comments     string          `json:"comments"`
}

type ExamResultItemDTO struct {
	Question ExamResultQuestionDTO `json:"question"`
	Answer   ExamResultAnswerDTO   `json:"answer"`
}
