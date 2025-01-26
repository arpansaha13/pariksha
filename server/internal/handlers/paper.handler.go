package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/dtos"
	"github.com/arpansaha13/pariksha/internal/middlewares"
	"github.com/arpansaha13/pariksha/internal/models"
)

func CreatePaper(w http.ResponseWriter, r *http.Request) {
	var paperDto dtos.PaperDto
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
			Type:    constants.PAPER_TYPE_OWNER,
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
