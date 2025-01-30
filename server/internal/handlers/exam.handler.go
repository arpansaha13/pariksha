package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/middlewares"
	"github.com/arpansaha13/pariksha/internal/models"
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
	var participantsDto []dtos.AddExamParticipantDto
	if err := json.NewDecoder(r.Body).Decode(&participantsDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

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
			ExamID: participantDto.ExamID,
			UserID: userID,
		}
		if err := db.DB.Create(&participant).Error; err != nil {
			http.Error(w, "Failed to add participant", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}
