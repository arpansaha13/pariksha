package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/arpansaha13/common/pkg/constants"
	"github.com/arpansaha13/common/pkg/models"
	"github.com/arpansaha13/pariksha/internal/config/db"
	"github.com/arpansaha13/pariksha/internal/config/validate"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/middlewares"
)

func GetParticipantAnswers(w http.ResponseWriter, r *http.Request) {
	participantID := mux.Vars(r)["participantId"]

	var answers []models.Answer
	if err := db.DB.Where("exam_participant_id = ?", participantID).Find(&answers).Error; err != nil {
		http.Error(w, "Answers not found", http.StatusNotFound)
		return
	}

	if len(answers) == 0 {
		http.Error(w, "Answers not found", http.StatusNotFound)
		return
	}

	var response []dtos.AnswerResponse
	for _, answer := range answers {
		response = append(response, dtos.AnswerResponse{
			ID:                answer.ID,
			ExamParticipantID: answer.ExamParticipantID,
			QuestionID:        answer.QuestionID,
			Answer:            answer.Answer.String,
			Comments:          answer.Comments.String,
			ScoreAwarded:      answer.ScoreAwarded,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpsertAnswers(w http.ResponseWriter, r *http.Request) {
	var answerDTOs []dtos.UpsertAnswerDto
	if err := json.NewDecoder(r.Body).Decode(&answerDTOs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, answerDTO := range answerDTOs {
		if err := validate.Do.Struct(answerDTO); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int)
	examID := mux.Vars(r)["examId"]

	var examParticipant models.ExamParticipant
	if err := db.DB.Where("exam_id = ? AND user_id = ?", examID, userID).Take(&examParticipant).Error; err != nil {
		http.Error(w, "Exam participant not found", http.StatusNotFound)
		return
	}

	if examParticipant.Status != constants.PARTICIPANT_STATUS_STARTED {
		http.Error(w, "Exam has not started", http.StatusBadRequest)
		return
	}

	totalCount := len(answerDTOs)
	skippedCount := 0

	for _, answerDTO := range answerDTOs {
		if answerDTO.SubmittedAt.After(examParticipant.ScheduledEndTime.Time) {
			skippedCount++
			continue
		}

		var answer models.Answer
		if err := db.DB.Where("exam_participant_id = ? AND question_id = ?", examParticipant.ID, answerDTO.QuestionID).Take(&answer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create a new answer if it does not exist
				answer = models.Answer{
					ExamParticipantID: examParticipant.ID,
					QuestionID:        answerDTO.QuestionID,
					Answer:            sql.NullString{String: answerDTO.Answer, Valid: true},
				}
				if err := db.DB.Create(&answer).Error; err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// Update the existing answer
			answer.Answer = sql.NullString{String: answerDTO.Answer, Valid: true}
			if err := db.DB.Save(&answer).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	response := map[string]int{
		"totalCount":   totalCount,
		"skippedCount": skippedCount,
	}

	w.Header().Set("Content-Type", "application/json")
	if skippedCount == totalCount {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	json.NewEncoder(w).Encode(response)
}

func UpdateAnswerForEvaluation(w http.ResponseWriter, r *http.Request) {
	var updateDTO dtos.UpdateAnswerForEvaluationDTO
	if err := json.NewDecoder(r.Body).Decode(&updateDTO); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(updateDTO); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var answer models.Answer
		if err := tx.Take(&answer, updateDTO.AnswerID).Error; err != nil {
			http.Error(w, "Answer not found", http.StatusNotFound)
			return err
		}

		var question models.Question
		if err := tx.Take(&question, answer.QuestionID).Error; err != nil {
			http.Error(w, "Question not found", http.StatusNotFound)
			return err
		}

		if updateDTO.NewScore != nil && *updateDTO.NewScore > question.MaxScore {
			http.Error(w, "New score exceeds max score for the question", http.StatusBadRequest)
			return errors.New("new score exceeds max score for the question")
		}

		isUpdated := false

		if updateDTO.NewScore != nil {
			var examParticipant models.ExamParticipant
			if err := tx.Take(&examParticipant, answer.ExamParticipantID).Error; err != nil {
				http.Error(w, "Exam participant not found", http.StatusNotFound)
				return err
			}

			examParticipant.ScoreAwarded = examParticipant.ScoreAwarded - answer.ScoreAwarded + *updateDTO.NewScore
			answer.ScoreAwarded = *updateDTO.NewScore

			if err := tx.Save(&examParticipant).Error; err != nil {
				http.Error(w, "Failed to update exam participant", http.StatusInternalServerError)
				return err
			}

			isUpdated = true
		}

		if updateDTO.Evaluated != nil {
			answer.Evaluated = *updateDTO.Evaluated
			isUpdated = true
		}

		if updateDTO.Comments != nil {
			answer.Comments = sql.NullString{String: *updateDTO.Comments, Valid: true}
			isUpdated = true
		}

		if isUpdated {
			if err := tx.Save(&answer).Error; err != nil {
				http.Error(w, "Failed to update answer", http.StatusInternalServerError)
				return err
			}
		}

		return nil
	})

	if err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
}

func MarkAsEvaluated(w http.ResponseWriter, r *http.Request) {
	participantID := mux.Vars(r)["participantId"]

	var examParticipant models.ExamParticipant
	if err := db.DB.Take(&examParticipant, participantID).Error; err != nil {
		http.Error(w, "Exam participant not found", http.StatusNotFound)
		return
	}

	if examParticipant.Status != constants.PARTICIPANT_STATUS_ENDED {
		http.Error(w, "Evaluation can only start if the exam has ended", http.StatusBadRequest)
		return
	}

	var unevaluatedCount int64
	if err := db.DB.Model(&models.Answer{}).Where("exam_participant_id = ? AND evaluated = ?", participantID, false).Count(&unevaluatedCount).Error; err != nil {
		http.Error(w, "Failed to count unevaluated answers", http.StatusInternalServerError)
		return
	}

	if unevaluatedCount == 0 {
		examParticipant.Status = constants.PARTICIPANT_STATUS_EVALUATED
		if err := db.DB.Save(&examParticipant).Error; err != nil {
			http.Error(w, "Failed to update exam participant status", http.StatusInternalServerError)
			return
		}
	}

	response := map[string]int64{
		"unevaluatedCount": unevaluatedCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
