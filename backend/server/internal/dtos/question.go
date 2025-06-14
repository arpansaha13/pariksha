package dtos

import (
	"encoding/json"
	"pariksha/common/pkg/proto"
)

type PaperTestCaseDto struct {
	Inputs      []string `json:"inputs"`
	Output      string   `json:"output"`
	Explanation *string  `json:"explanation"`
	Hidden      bool     `json:"hidden"`
	Order       int32    `json:"order"`
}

type QuestionResponseDto struct {
	ID            string              `json:"id"`
	Question      json.RawMessage     `json:"question"`
	CategoryID    int64               `json:"category_id"`
	Type          proto.QuestionType  `json:"type"`
	Tags          json.RawMessage     `json:"tags"`
	PaperID       string              `json:"paper_id"`
	MaxScore      int32               `json:"max_score"`
	TestCases     *[]PaperTestCaseDto `json:"test_cases"`
	CorrectAnswer string              `json:"correct_answer"`
}

type QuestionMinimalResponseDto struct {
	ID         string          `json:"id"`
	CategoryID int64           `json:"category_id"`
	PaperID    string          `json:"paper_id"`
	Order      int32           `json:"order"`
	Question   json.RawMessage `json:"question"`
}

type QuestionCategoryResponseDto struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Order int32  `json:"order"`
}

type CreateQuestionDto struct {
	Question      json.RawMessage    `json:"question" validate:"required"`
	CategoryID    int64              `json:"category_id" validate:"required"`
	Type          proto.QuestionType `json:"type" validate:"required"`
	Tags          json.RawMessage    `json:"tags,omitempty" validate:"required"`
	MaxScore      int32              `json:"max_score" validate:"required"`
	CorrectAnswer string             `json:"correct_answer,omitempty"`
}

type UpdateQuestionDto struct {
	Type          proto.QuestionType `json:"type,omitempty"`
	Question      json.RawMessage    `json:"question,omitempty"`
	MaxScore      int32              `json:"max_score,omitempty"`
	Tags          json.RawMessage    `json:"tags,omitempty"`
	CorrectAnswer string             `json:"correct_answer,omitempty"`
}

type ReorderQuestionsDto struct {
	Questions []string `json:"questions" validate:"required,min=1"`
}

type CreateQuestionResponseDto struct {
	ID string `json:"id"`
}

type UpdateQuestionResponseDto struct {
	ID string `json:"id"`
}

type UpsertTestCaseDto struct {
	Inputs      []string `json:"inputs" validate:"required"`
	Output      string   `json:"output" validate:"required"`
	Explanation *string  `json:"explanation,omitempty"`
	Hidden      bool     `json:"hidden"`
}

type UpsertTestCasesDto struct {
	TestCases []UpsertTestCaseDto `json:"test_cases" validate:"required,min=1"`
}
