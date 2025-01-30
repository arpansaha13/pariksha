package dtos

import "encoding/json"

type QuestionResponse struct {
	ID            int             `json:"id"`
	Question      json.RawMessage `json:"question"`
	Category      string          `json:"category"`
	Type          string          `json:"type"`
	Tags          json.RawMessage `json:"tags"`
	PaperID       int             `json:"paper_id"`
	MaxScore      int             `json:"max_score"`
	CorrectAnswer string          `json:"correct_answer"`
}

type CreateQuestionDto struct {
	PaperID       int             `json:"paper_id"`
	Question      json.RawMessage `json:"question"`
	Category      string          `json:"category,omitempty"`
	Type          string          `json:"type"`
	Tags          json.RawMessage `json:"tags,omitempty"`
	MaxScore      int             `json:"max_score"`
	CorrectAnswer string          `json:"correct_answer,omitempty"`
}

type UpdateQuestionDto struct {
	Question      json.RawMessage `json:"question,omitempty"`
	Category      string          `json:"category,omitempty"`
	MaxScore      int             `json:"max_score,omitempty"`
	Tags          json.RawMessage `json:"tags,omitempty"`
	CorrectAnswer string          `json:"correct_answer,omitempty"`
}
