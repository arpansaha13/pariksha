package repositories

import (
	"encoding/json"
	"errors"

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/models"
)

type QuestionRepository interface {
	UnmarshalQuestion(questionDto dtos.CreateQuestionDto) (json.RawMessage, error)
}

type questionRepository struct{}

func GetQuestionRepository() QuestionRepository {
	return &questionRepository{}
}

func (r *questionRepository) UnmarshalQuestion(questionDto dtos.CreateQuestionDto) (json.RawMessage, error) {
	var questionData json.RawMessage

	switch questionDto.Type {
	case constants.QUESTION_TYPE_MCQ:
		var mcq models.MCQQuestion
		if err := json.Unmarshal(questionDto.Question, &mcq); err != nil {
			return nil, errors.New("invalid MCQ question format")
		}
		questionData = questionDto.Question
	case constants.QUESTION_TYPE_SHORT, constants.QUESTION_TYPE_LONG:
		var general models.GeneralQuestion
		if err := json.Unmarshal(questionDto.Question, &general); err != nil {
			return nil, errors.New("invalid general question format")
		}
		questionData = questionDto.Question
	default:
		return nil, errors.New("invalid question type")
	}

	return questionData, nil
}
