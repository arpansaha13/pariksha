package dtos

import "encoding/json"

type QuestionCategoryResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

type QuestionResponse struct {
	ID            int                       `json:"id"`
	Question      json.RawMessage           `json:"question"`
	Category      *QuestionCategoryResponse `json:"category"`
	Type          string                    `json:"type"`
	Tags          json.RawMessage           `json:"tags"`
	PaperID       int                       `json:"paper_id"`
	MaxScore      int                       `json:"max_score"`
	CorrectAnswer string                    `json:"correct_answer"`
}

type QuestionMinimalResponse struct {
	ID         int             `json:"id"`
	CategoryID int             `json:"category_id"`
	PaperID    int             `json:"paper_id"`
	Question   json.RawMessage `json:"question"`
}

type CreateQuestionDto struct {
	Question      json.RawMessage `json:"question" validate:"required"`
	CategoryID    int             `json:"category_id" validate:"required"`
	Type          string          `json:"type" validate:"required"`
	Tags          json.RawMessage `json:"tags,omitempty" validate:"required"`
	MaxScore      int             `json:"max_score" validate:"required"`
	CorrectAnswer string          `json:"correct_answer,omitempty"`
}

type UpdateQuestionDto struct {
	Type          string          `json:"type,omitempty"`
	Question      json.RawMessage `json:"question,omitempty"`
	CategoryID    int             `json:"category_id,omitempty"`
	MaxScore      int             `json:"max_score,omitempty"`
	Tags          json.RawMessage `json:"tags,omitempty"`
	CorrectAnswer string          `json:"correct_answer,omitempty"`
}
