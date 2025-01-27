package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/middlewares"
	"github.com/arpansaha13/pariksha/internal/models"
)

func CreatePaper(w http.ResponseWriter, r *http.Request) {
	var paperDto dtos.CreatePaperDto
	if err := json.NewDecoder(r.Body).Decode(&paperDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middlewares.UserIDKey).(int)

	var paper models.Paper

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		paper = models.Paper{
			Title: paperDto.Title,
		}

		if err := tx.Create(&paper).Error; err != nil {
			return err
		}

		paperOwnership := models.PaperOwnership{
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
	json.NewEncoder(w).Encode(paper)
}

func UpdatePaper(w http.ResponseWriter, r *http.Request) {
	var paperDto dtos.UpdatePaperDto
	if err := json.NewDecoder(r.Body).Decode(&paperDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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

	paper.Title = paperDto.Title
	if err := db.DB.Save(&paper).Error; err != nil {
		http.Error(w, "Failed to update paper title", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(paper)
}
