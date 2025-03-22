package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/server/internal/config/db"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
)

func questionToResponse(question models.Question) dtos.QuestionResponse {
	category := dtos.QuestionCategoryResponse{
		ID:    question.Category.ID,
		Name:  question.Category.Name,
		Order: question.Category.Order,
	}

	return dtos.QuestionResponse{
		ID:            question.ID,
		Question:      question.Question,
		Category:      &category,
		Type:          question.Type,
		Tags:          question.Tags,
		PaperID:       question.PaperID,
		MaxScore:      question.MaxScore,
		CorrectAnswer: question.CorrectAnswer.String,
	}
}

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

	// Fetch only needed fields from questions
	var questions []models.Question
	if err := db.DB.
		Select("id, category_id, paper_id, question").
		Where("paper_id = ?", paperID).
		Find(&questions).Error; err != nil {
		http.Error(w, "Failed to retrieve questions", http.StatusInternalServerError)
		return
	}

	response := make([]dtos.QuestionMinimalResponse, len(questions))
	for i, question := range questions {
		response[i] = dtos.QuestionMinimalResponse{
			ID:         question.ID,
			CategoryID: question.CategoryID,
			PaperID:    question.PaperID,
			Question:   question.Question,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetQuestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	questionID := vars["id"]

	var question models.Question
	if err := db.DB.Preload("Category").Take(&question, questionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Question not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to find question", http.StatusInternalServerError)
		}
		return
	}

	// Convert to response DTO
	response := questionToResponse(question)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreateQuestion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paperID, err := strconv.Atoi(vars["id"])

	if err != nil {
		http.Error(w, "Invalid request url", http.StatusBadRequest)
		return
	}

	var questionDto dtos.CreateQuestionDto
	if err := json.NewDecoder(r.Body).Decode(&questionDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(questionDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the paperId
	var paper models.Paper
	if err := db.DB.Take(&paper, paperID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Paper not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to find paper", http.StatusInternalServerError)
		return
	}

	questionCounts, err := paper.GetQuestionCounts()
	if err != nil {
		http.Error(w, "Failed to parse question counts", http.StatusInternalServerError)
		return
	}

	var response dtos.QuestionResponse

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		questionData, err := unmarshalQuestion(questionDto)
		if err != nil {
			return err
		}

		if err := validateQuestion(questionData, questionDto.Type); err != nil {
			return err
		}

		// Validate category exists
		var category models.QuestionCategory
		if err := tx.Where("id = ? AND paper_id = ?", questionDto.CategoryID, paperID).
			Take(&category).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("category not found")
			}
			return err
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
			PaperID:       paperID,
			Question:      questionDto.Question,
			CategoryID:    questionDto.CategoryID,
			Type:          questionDto.Type,
			Tags:          questionDto.Tags,
			MaxScore:      questionDto.MaxScore,
			CorrectAnswer: sql.NullString{String: questionDto.CorrectAnswer, Valid: questionDto.CorrectAnswer != ""},
		}

		if err := tx.Create(&question).Error; err != nil {
			return err
		}

		if err := tx.Preload("Category").Take(&question, question.ID).Error; err != nil {
			return err
		}

		response = questionToResponse(question)

		// Update the paper's max score
		paper.MaxScore += question.MaxScore

		// Update the paper with the new question counts
		paper.QuestionCounts, err = json.Marshal(questionCounts)
		if err != nil {
			return err
		}

		if err := tx.Save(&paper).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	var updateDto dtos.UpdateQuestionDto
	if err := json.NewDecoder(r.Body).Decode(&updateDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	questionID := vars["id"]

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var question models.Question
		if err := tx.Take(&question, questionID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Question not found", http.StatusNotFound)
				return err
			}
			return err
		}

		// Validate the paperId
		var paper models.Paper
		if err := tx.Take(&paper, question.PaperID).Error; err != nil {
			return err
		}

		questionCounts, err := paper.GetQuestionCounts()
		if err != nil {
			return err
		}

		isUpdated := false

		// Handle type update first as it affects question validation
		if updateDto.Type != "" && updateDto.Type != question.Type {
			// Must provide new question data when changing type
			if updateDto.Question == nil {
				http.Error(w, "Question data required when changing question type", http.StatusBadRequest)
				return errors.New("question data required when changing type")
			}

			// Validate new type
			switch updateDto.Type {
			case constants.QUESTION_TYPE_MCQ, constants.QUESTION_TYPE_SHORT, constants.QUESTION_TYPE_LONG:
				// Update question counts
				switch question.Type {
				case constants.QUESTION_TYPE_MCQ:
					questionCounts.MCQ--
				case constants.QUESTION_TYPE_SHORT:
					questionCounts.Short--
				case constants.QUESTION_TYPE_LONG:
					questionCounts.Long--
				}

				switch updateDto.Type {
				case constants.QUESTION_TYPE_MCQ:
					questionCounts.MCQ++
				case constants.QUESTION_TYPE_SHORT:
					questionCounts.Short++
				case constants.QUESTION_TYPE_LONG:
					questionCounts.Long++
				}

				question.Type = updateDto.Type
				isUpdated = true
			default:
				http.Error(w, "Invalid question type", http.StatusBadRequest)
				return errors.New("invalid question type")
			}
		}

		// Update question data
		if updateDto.Question != nil {
			// Unmarshal and validate based on current type (which might have just been updated)
			questionData, err := unmarshalQuestion(dtos.CreateQuestionDto{
				Question: updateDto.Question,
				Type:     question.Type,
			})
			if err != nil {
				return err
			}

			if err := validateQuestion(questionData, question.Type); err != nil {
				return err
			}

			question.Question = updateDto.Question
			isUpdated = true
		}

		// Update the fields based on the provided data
		if updateDto.CategoryID != 0 {
			var category models.QuestionCategory
			if err := tx.Where("id = ? AND paper_id = ?", updateDto.CategoryID, question.PaperID).
				Take(&category).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					http.Error(w, "Category not found", http.StatusBadRequest)
					return err
				}
				return err
			}
			question.CategoryID = updateDto.CategoryID
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

		isPaperUpdated := false

		if updateDto.MaxScore != 0 && updateDto.MaxScore != question.MaxScore {
			paper.MaxScore = paper.MaxScore - question.MaxScore + updateDto.MaxScore
			question.MaxScore = updateDto.MaxScore

			isPaperUpdated = true
			isUpdated = true
		}

		if isUpdated {
			// Update question counts if type was changed
			if updateDto.Type != "" && updateDto.Type != question.Type {
				paper.QuestionCounts, err = json.Marshal(questionCounts)
				if err != nil {
					return err
				}
				isPaperUpdated = true
			}

			if err := tx.Save(&question).Error; err != nil {
				return err
			}
		}

		if isPaperUpdated {
			if err := tx.Save(&paper).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		http.Error(w, "Failed to update question", http.StatusInternalServerError)
		return
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

	w.WriteHeader(http.StatusNoContent)
}

func unmarshalQuestion(questionDto dtos.CreateQuestionDto) (interface{}, error) {
	switch questionDto.Type {
	case constants.QUESTION_TYPE_MCQ:
		var mcq models.MCQQuestion
		if err := json.Unmarshal(questionDto.Question, &mcq); err != nil {
			return nil, errors.New("invalid MCQ question format")
		}
		return mcq, nil

	case constants.QUESTION_TYPE_SHORT, constants.QUESTION_TYPE_LONG:
		var general models.GeneralQuestion
		if err := json.Unmarshal(questionDto.Question, &general); err != nil {
			return nil, errors.New("invalid general question format")
		}
		return general, nil

	default:
		return nil, errors.New("invalid question type")
	}
}

func validateQuestion(question interface{}, questionType string) error {
	switch questionType {
	case constants.QUESTION_TYPE_MCQ:
		mcq := question.(models.MCQQuestion)

		if strings.TrimSpace(mcq.Statement) == "" {
			return errors.New("question statement cannot be empty")
		}
		if len(mcq.Options) < 2 {
			return errors.New("MCQ questions must have at least 2 options")
		}
		if len(mcq.Options) > 5 {
			return errors.New("MCQ questions cannot have more than 5 options")
		}

	case constants.QUESTION_TYPE_SHORT, constants.QUESTION_TYPE_LONG:
		general := question.(models.GeneralQuestion)

		if strings.TrimSpace(general.Statement) == "" {
			return errors.New("question statement cannot be empty")
		}
	}

	return nil
}
