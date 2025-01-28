package dtos

import "encoding/json"

type CreateQuestionDto struct {
	PaperID       int             `json:"paper_id"`
	Question      json.RawMessage `json:"question"`
	Category      string          `json:"category,omitempty"`
	Type          string          `json:"type"`
	Tags          json.RawMessage `json:"tags,omitempty"`
	MaxScore      int             `json:"max_score"`
	CorrectAnswer string          `json:"correct_answer,omitempty"`
}
