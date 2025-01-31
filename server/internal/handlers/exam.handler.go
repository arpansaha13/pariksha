package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/middlewares"
	"github.com/arpansaha13/pariksha/internal/models"
	"github.com/gorilla/mux"
)

func CreateExam(w http.ResponseWriter, r *http.Request) {
	var examDto dtos.CreateExamDto
	if err := json.NewDecoder(r.Body).Decode(&examDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the paperId
	var paper models.Paper
	if err := db.DB.Take(&paper, examDto.PaperID).Error; err != nil {
		http.Error(w, "Paper not found", http.StatusNotFound)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int)

	exam := models.Exam{
		Title:              examDto.Title,
		StartsAt:           examDto.StartsAt,
		EndsAt:             examDto.EndsAt,
		CreatedBy:          userID,
		Type:               examDto.Type,
		MaxCandidatesCount: examDto.MaxCandidatesCount,
		PaperID:            examDto.PaperID,
	}

	if err := db.DB.Create(&exam).Error; err != nil {
		http.Error(w, "Failed to create exam", http.StatusInternalServerError)
		return
	}

	response := dtos.ExamResponse{
		ID:                 exam.ID,
		Title:              exam.Title,
		StartsAt:           exam.StartsAt,
		EndsAt:             exam.EndsAt,
		CreatedBy:          exam.CreatedBy,
		Type:               exam.Type,
		MaxCandidatesCount: exam.MaxCandidatesCount,
		PaperID:            exam.PaperID,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func AddExamParticipants(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	examID, err := strconv.Atoi(params["examId"])
	if err != nil {
		http.Error(w, "Invalid exam ID", http.StatusBadRequest)
		return
	}

	var participantsDto []dtos.AddExamParticipantDto
	if err := json.NewDecoder(r.Body).Decode(&participantsDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var examParticipants []models.ExamParticipant

	for _, participantDto := range participantsDto {
		var userID int

		if participantDto.UserID != 0 {
			userID = participantDto.UserID
		} else if participantDto.Email != "" {
			// Create a guest user
			username := strings.Split(participantDto.Email, "@")[0]
			user := models.User{
				Email:    participantDto.Email,
				IsGuest:  true,
				Username: username,
			}

			if participantDto.FirstName != "" {
				user.FirstName = sql.NullString{String: participantDto.FirstName, Valid: true}
			}

			if participantDto.LastName != "" {
				user.FirstName = sql.NullString{String: participantDto.LastName, Valid: true}
			}

			if err := db.DB.Create(&user).Error; err != nil {
				http.Error(w, "Failed to create guest user", http.StatusInternalServerError)
				return
			}
			userID = user.ID
		} else {
			// Either user_id or email must be provided
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Add the participant to the exam
		participant := models.ExamParticipant{
			ExamID: examID,
			UserID: userID,
		}

		examParticipants = append(examParticipants, participant)
	}

	if err := db.DB.Create(&examParticipants).Error; err != nil {
		http.Error(w, "Failed to add participants", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func RemoveExamParticipant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	participantID := vars["participantId"]

	if err := db.DB.Delete(&models.ExamParticipant{}, participantID).Error; err != nil {
		http.Error(w, "Failed to remove participant", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func StartExam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	examID, err := strconv.Atoi(vars["examId"])
	if err != nil {
		http.Error(w, "Invalid exam ID", http.StatusBadRequest)
		return
	}

	participantID, err := strconv.Atoi(vars["participantId"])
	if err != nil {
		http.Error(w, "Invalid participant ID", http.StatusBadRequest)
		return
	}

	var exam models.Exam
	if err := db.DB.Preload("Paper").Take(&exam, examID).Error; err != nil {
		http.Error(w, "Exam not found", http.StatusNotFound)
		return
	}

	var participant models.ExamParticipant
	if err := db.DB.Take(&participant, participantID).Error; err != nil {
		http.Error(w, "Participant not found", http.StatusNotFound)
		return
	}

	if participant.Status != constants.PARTICIPANT_STATUS_INVITED {
		http.Error(w, "Participant has already started the exam", http.StatusBadRequest)
		return
	}

	if participant.ExamID != examID {
		http.Error(w, "Participant does not belong to this exam", http.StatusBadRequest)
		return
	}

	now := time.Now()
	endTime := now.Add(time.Duration(exam.Paper.DurationMinutes) * time.Minute)

	// Update participant status and times
	participant.Status = constants.PARTICIPANT_STATUS_STARTED
	participant.StartedAt = sql.NullTime{Time: now, Valid: true}
	participant.EndedAt = sql.NullTime{Time: endTime, Valid: true}

	if err := db.DB.Save(&participant).Error; err != nil {
		http.Error(w, "Failed to start exam", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
