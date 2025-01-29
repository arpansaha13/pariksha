package handlers

import (
	"encoding/json"
	"net/http"

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
