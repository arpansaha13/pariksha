package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/models"
	"gorm.io/gorm"
)

func CreateQuestions(w http.ResponseWriter, r *http.Request) {
	var questionDtos []dtos.CreateQuestionDto
	if err := json.NewDecoder(r.Body).Decode(&questionDtos); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, questionDto := range questionDtos {
		// Validate the paperId
		var paper models.Paper
		if err := db.DB.Take(&paper, questionDto.PaperID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Paper not found", http.StatusNotFound)
				return
			} else {
				http.Error(w, "Failed to find paper", http.StatusInternalServerError)
				return
			}
		}

		// Validate the question type and unmarshal the question JSON
		var questionData json.RawMessage
		switch questionDto.Type {
		case constants.QUESTION_TYPE_MCQ:
			var mcq models.MCQQuestion
			if err := json.Unmarshal(questionDto.Question, &mcq); err != nil {
				http.Error(w, "Invalid MCQ question format", http.StatusBadRequest)
				return
			}
			questionData = questionDto.Question
		case constants.QUESTION_TYPE_SHORT, constants.QUESTION_TYPE_LONG:
			var general models.GeneralQuestion
			if err := json.Unmarshal(questionDto.Question, &general); err != nil {
				http.Error(w, "Invalid general question format", http.StatusBadRequest)
				return
			}
			questionData = questionDto.Question
		default:
			http.Error(w, "Invalid question type", http.StatusBadRequest)
			return
		}

		// Create the question
		question := models.Question{
			PaperID:       questionDto.PaperID,
			Question:      questionData,
			Category:      questionDto.Category,
			Type:          questionDto.Type,
			Tags:          questionDto.Tags,
			MaxScore:      questionDto.MaxScore,
			CorrectAnswer: questionDto.CorrectAnswer,
		}

		if err := db.DB.Create(&question).Error; err != nil {
			http.Error(w, "Failed to create question", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("Questions created successfully")
}
