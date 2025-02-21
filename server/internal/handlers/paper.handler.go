package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/arpansaha13/common/pkg/constants"
	"github.com/arpansaha13/pariksha/internal/config/validate"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/middlewares"
	"github.com/arpansaha13/pariksha/internal/models"
)

func GetUserPapers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(int)

	var papers []models.Paper
	err := db.DB.Joins(
		"INNER JOIN paper_ownerships ON paper_ownerships.paper_id = papers.id AND paper_ownerships.user_id = ?",
		userID,
	).Find(&papers).Error
	if err != nil {
		http.Error(w, "Failed to retrieve papers", http.StatusInternalServerError)
		return
	}

	var response []dtos.PaperResponse
	for _, paper := range papers {
		response = append(response, dtos.PaperResponse{
			ID:             paper.ID,
			Title:          paper.Title,
			MaxScore:       paper.MaxScore,
			QuestionCounts: paper.QuestionCounts,
			PaperOwnership: dtos.PaperOwnershipResponse{
				ID:   paper.PaperOwnership.ID,
				Path: paper.PaperOwnership.Path,
				Type: paper.PaperOwnership.Type,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreatePaper(w http.ResponseWriter, r *http.Request) {
	var paperDto dtos.CreatePaperDto
	if err := json.NewDecoder(r.Body).Decode(&paperDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(paperDto)
	if errs != nil {
		http.Error(w, "Invald request body", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int)

	var paper models.Paper
	var paperOwnership models.PaperOwnership

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		paper = models.Paper{
			Title: paperDto.Title,
		}

		if err := tx.Create(&paper).Error; err != nil {
			return err
		}

		paperOwnership = models.PaperOwnership{
			UserID:  userID,
			PaperID: paper.ID,
			Type:    constants.PAPER_OWNERSHIP_TYPE_OWNER,
		}

		if err := tx.Create(&paperOwnership).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, "Failed to create paper", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dtos.PaperResponse{
		ID:             paper.ID,
		Title:          paper.Title,
		MaxScore:       paper.MaxScore,
		QuestionCounts: paper.QuestionCounts,
		PaperOwnership: dtos.PaperOwnershipResponse{
			ID:   paperOwnership.ID,
			Path: paperOwnership.Path,
			Type: paperOwnership.Type,
		},
	})
}

func UpdatePaper(w http.ResponseWriter, r *http.Request) {
	var paperDto dtos.UpdatePaperDto
	if err := json.NewDecoder(r.Body).Decode(&paperDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errs := validate.Do.Struct(paperDto)
	if errs != nil {
		http.Error(w, "Invald request body", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	paperID := vars["id"]

	var paper models.Paper
	if err := db.DB.Take(&paper, paperID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Paper not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to find paper", http.StatusInternalServerError)
		}
		return
	}

	isUpdated := false

	if paperDto.Title != "" {
		paper.Title = paperDto.Title
		isUpdated = true
	}

	if isUpdated {
		if err := db.DB.Save(&paper).Error; err != nil {
			http.Error(w, "Failed to update paper title", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
