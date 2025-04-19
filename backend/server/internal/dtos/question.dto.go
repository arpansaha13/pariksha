package dtos

import "encoding/json"

type QuestionCategoryResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

type QuestionResponse struct {
	ID            int64           `json:"id"`
	Question      json.RawMessage `json:"question"`
	CategoryID    int64           `json:"category_id"`
	Type          string          `json:"type"`
	Tags          json.RawMessage `json:"tags"`
	PaperID       int64           `json:"paper_id"`
	MaxScore      int             `json:"max_score"`
	CorrectAnswer string          `json:"correct_answer"`
}

type QuestionMinimalResponse struct {
	ID         int64           `json:"id"`
	CategoryID int64           `json:"category_id"`
	PaperID    int64           `json:"paper_id"`
	Order      int             `json:"order"`
	Question   json.RawMessage `json:"question"`
}

type CreateQuestionDto struct {
	Question      json.RawMessage `json:"question" validate:"required"`
	CategoryID    int64           `json:"category_id" validate:"required"`
	Type          string          `json:"type" validate:"required"`
	Tags          json.RawMessage `json:"tags,omitempty" validate:"required"`
	MaxScore      int             `json:"max_score" validate:"required"`
	CorrectAnswer string          `json:"correct_answer,omitempty"`
}

type UpdateQuestionDto struct {
	Type          string          `json:"type,omitempty"`
	Question      json.RawMessage `json:"question,omitempty"`
	CategoryID    int64           `json:"category_id,omitempty"`
	MaxScore      int             `json:"max_score,omitempty"`
	Tags          json.RawMessage `json:"tags,omitempty"`
	CorrectAnswer string          `json:"correct_answer,omitempty"`
}

type ReorderQuestionsDto struct {
	Questions []int64 `json:"questions" validate:"required,min=1"`
}
