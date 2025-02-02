package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/arpansaha13/pariksha/internal/config/validate"
	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/models"
)

func GetPaperQuestions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paperID := vars["id"]

	// Validate the paperId
	var paper models.Paper
	if err := db.DB.Take(&paper, paperID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Paper not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to find paper", http.StatusInternalServerError)
		}
		return
	}

	// Fetch questions for the paper
	var questions []models.Question
	if err := db.DB.Where("paper_id = ?", paperID).Find(&questions).Error; err != nil {
		http.Error(w, "Failed to retrieve questions", http.StatusInternalServerError)
		return
	}

	var response []dtos.QuestionResponse
	for _, question := range questions {
		response = append(response, dtos.QuestionResponse{
			ID:            question.ID,
			Question:      question.Question,
			Category:      question.Category.String,
			Type:          question.Type,
			Tags:          question.Tags,
			PaperID:       question.PaperID,
			MaxScore:      question.MaxScore,
			CorrectAnswer: question.CorrectAnswer.String,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreatePaperQuestions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paperID := vars["id"]

	var questionDtos []dtos.CreateQuestionDto
	if err := json.NewDecoder(r.Body).Decode(&questionDtos); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, questionDto := range questionDtos {
		errs := validate.Do.Struct(questionDto)
		if errs != nil {
			http.Error(w, "Invald request body", http.StatusBadRequest)
			return
		}
	}

	// Validate the paperId
	var paper models.Paper
	if err := db.DB.Take(&paper, paperID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Paper not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, "Failed to find paper", http.StatusInternalServerError)
			return
		}
	}

	questionCounts, err := paper.GetQuestionCounts()
	if err != nil {
		http.Error(w, "Failed to parse question counts", http.StatusInternalServerError)
		return
	}

	var questions []models.Question

	for _, questionDto := range questionDtos {
		// Unmarshal and validate the question JSON
		questionData, err := unmarshalQuestion(questionDto)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Increment the question count based on the question type
		switch questionDto.Type {
		case constants.QUESTION_TYPE_MCQ:
			questionCounts.MCQ++
		case constants.QUESTION_TYPE_SHORT:
			questionCounts.Short++
		case constants.QUESTION_TYPE_LONG:
			questionCounts.Long++
		}

		// Create the question
		question := models.Question{
			PaperID:       questionDto.PaperID,
			Question:      questionData,
			Category:      sql.NullString{String: questionDto.Category, Valid: true},
			Type:          questionDto.Type,
			Tags:          questionDto.Tags,
			MaxScore:      questionDto.MaxScore,
			CorrectAnswer: sql.NullString{String: questionDto.CorrectAnswer, Valid: true},
		}

		questions = append(questions, question)

		// Update the paper's max score
		paper.MaxScore += question.MaxScore
	}

	// Bulk insert the questions
	if err := db.DB.Create(&questions).Error; err != nil {
		http.Error(w, "Failed to create questions", http.StatusInternalServerError)
		return
	}

	// Update the paper with the new question counts
	paper.QuestionCounts, err = json.Marshal(questionCounts)
	if err != nil {
		http.Error(w, "Failed to marshal question counts", http.StatusInternalServerError)
		return
	}

	if err := db.DB.Save(&paper).Error; err != nil {
		http.Error(w, "Failed to update paper", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	var updateDto dtos.UpdateQuestionDto
	if err := json.NewDecoder(r.Body).Decode(&updateDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	questionID := vars["id"]

	var question models.Question
	if err := db.DB.Take(&question, questionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Question not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to find question", http.StatusInternalServerError)
		}
		return
	}

	// Validate the paperId
	var paper models.Paper
	if err := db.DB.Take(&paper, question.PaperID).Error; err != nil {
		http.Error(w, "Failed to find paper", http.StatusInternalServerError)
		return
	}

	isUpdated := false

	// Update the fields based on the provided data
	if updateDto.Question != nil {
		switch question.Type {
		case constants.QUESTION_TYPE_MCQ:
			var mcq models.MCQQuestion
			if err := json.Unmarshal(updateDto.Question, &mcq); err != nil {
				http.Error(w, "Invalid MCQ question format", http.StatusBadRequest)
				return
			}
			question.Question = updateDto.Question
		case constants.QUESTION_TYPE_SHORT, constants.QUESTION_TYPE_LONG:
			var general models.GeneralQuestion
			if err := json.Unmarshal(updateDto.Question, &general); err != nil {
				http.Error(w, "Invalid general question format", http.StatusBadRequest)
				return
			}
			question.Question = updateDto.Question
		default:
			http.Error(w, "Invalid question type", http.StatusBadRequest)
			return
		}
		isUpdated = true
	}

	if updateDto.Tags != nil {
		question.Tags = updateDto.Tags
		isUpdated = true
	}

	if updateDto.CorrectAnswer != "" {
		question.CorrectAnswer = sql.NullString{String: updateDto.CorrectAnswer, Valid: true}
		isUpdated = true
	}

	if updateDto.Category != "" {
		question.Category = sql.NullString{String: updateDto.Category, Valid: true}
		isUpdated = true
	}

	if updateDto.MaxScore != 0 && updateDto.MaxScore != question.MaxScore {
		paper.MaxScore = paper.MaxScore - question.MaxScore + updateDto.MaxScore
		question.MaxScore = updateDto.MaxScore

		if err := db.DB.Save(&paper).Error; err != nil {
			http.Error(w, "Failed to update paper", http.StatusInternalServerError)
			return
		}
		isUpdated = true
	}

	if isUpdated {
		if err := db.DB.Save(&question).Error; err != nil {
			http.Error(w, "Failed to update question", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	questionID := vars["id"]

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var question models.Question
		if err := tx.Take(&question, questionID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Question not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to find question", http.StatusInternalServerError)
			}
			return err
		}

		// Validate the paperId
		var paper models.Paper
		if err := tx.Take(&paper, question.PaperID).Error; err != nil {
			http.Error(w, "Failed to find paper", http.StatusInternalServerError)
			return err
		}

		questionCounts, err := paper.GetQuestionCounts()
		if err != nil {
			http.Error(w, "Failed to parse question counts", http.StatusInternalServerError)
			return err
		}

		// Decrement the question count based on the question type
		switch question.Type {
		case constants.QUESTION_TYPE_MCQ:
			questionCounts.MCQ--
		case constants.QUESTION_TYPE_SHORT:
			questionCounts.Short--
		case constants.QUESTION_TYPE_LONG:
			questionCounts.Long--
		}

		// Update the paper's max score
		paper.MaxScore -= question.MaxScore

		// Delete the question
		if err := tx.Delete(&question).Error; err != nil {
			http.Error(w, "Failed to delete question", http.StatusInternalServerError)
			return err
		}

		// Update the paper with the new question counts
		paper.QuestionCounts, err = json.Marshal(questionCounts)
		if err != nil {
			http.Error(w, "Failed to marshal question counts", http.StatusInternalServerError)
			return err
		}
		if err := tx.Save(&paper).Error; err != nil {
			http.Error(w, "Failed to update paper question counts", http.StatusInternalServerError)
			return err
		}

		return nil
	})

	if err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
}

func unmarshalQuestion(questionDto dtos.CreateQuestionDto) (json.RawMessage, error) {
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
