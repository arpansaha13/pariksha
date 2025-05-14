package dtos

import "encoding/json"

type ExamResultQuestionDTO struct {
	ID         int64           `json:"id"`
	Order      int32           `json:"order"`
	CategoryID int64           `json:"category_id"`
	Content    json.RawMessage `json:"content"`
	MaxScore   int32           `json:"max_score"`
}

type ExamResultAnswerDTO struct {
	Content      json.RawMessage `json:"content"`
	ScoreAwarded int32           `json:"score_awarded"`
	Comments     string          `json:"comments"`
}

type ExamResultItemDTO struct {
	Type     string                `json:"type"`
	Question ExamResultQuestionDTO `json:"question"`
	Answer   ExamResultAnswerDTO   `json:"answer"`
}
