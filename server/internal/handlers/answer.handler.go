package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arpansaha13/pariksha/internal/config/validate"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/middlewares"
	"github.com/arpansaha13/pariksha/internal/models"
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

func CreateAnswers(w http.ResponseWriter, r *http.Request) {
	var answerDTOs []dtos.AnswerDTO
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

	var exam models.Exam
	if err := db.DB.Take(&exam, examID).Error; err != nil {
		http.Error(w, "Exam not found", http.StatusNotFound)
		return
	}

	var examParticipant models.ExamParticipant
	if err := db.DB.Where("exam_id = ? AND user_id = ?", examID, userID).Take(&examParticipant).Error; err != nil {
		http.Error(w, "Exam participant not found", http.StatusNotFound)
		return
	}

	totalCount := len(answerDTOs)
	skippedCount := 0

	for _, answerDTO := range answerDTOs {
		if answerDTO.SubmittedAt.After(exam.EndsAt) {
			skippedCount++
			continue
		}

		answer := models.Answer{
			ExamParticipantID: examParticipant.ID,
			QuestionID:        answerDTO.QuestionID,
			Answer:            sql.NullString{String: answerDTO.Answer, Valid: true},
		}

		if err := db.DB.Create(&answer).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	response := map[string]int{
		"totalCount":   totalCount,
		"skippedCount": skippedCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
