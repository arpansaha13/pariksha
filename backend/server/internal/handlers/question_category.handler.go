package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/server/internal/config/db"
	"pariksha/server/internal/config/validate"
	"pariksha/server/internal/dtos"
)

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paperID := vars["paper_id"]

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Get current category count
		var count int64
		err := tx.Model(&models.QuestionCategory{}).
			Where("paper_id = ?", paperID).
			Count(&count).Error
		if err != nil {
			return err
		}

		// Get max order
		var maxOrder struct{ MaxOrder int }
		err = tx.Model(&models.QuestionCategory{}).
			Where("paper_id = ?", paperID).
			Select("COALESCE(MAX(\"order\"), 0) as max_order").
			Scan(&maxOrder).Error
		if err != nil {
			return err
		}

		category := models.QuestionCategory{
			PaperID: mustAtoi(paperID),
			Name:    fmt.Sprintf("Category %d", count+1),
			Order:   maxOrder.MaxOrder + 1,
		}

		if err := tx.Create(&category).Error; err != nil {
			return err
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dtos.QuestionCategoryResponse{
			ID:    category.ID,
			Name:  category.Name,
			Order: category.Order,
		})
		return nil
	})

	if err != nil {
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		return
	}
}

func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID := vars["id"]

	var categoryDto dtos.UpdateCategoryDto
	if err := json.NewDecoder(r.Body).Decode(&categoryDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(categoryDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var category models.QuestionCategory
	if err := db.DB.First(&category, categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to find category", http.StatusInternalServerError)
		}
		return
	}

	category.Name = categoryDto.Name
	if err := db.DB.Save(&category).Error; err != nil {
		http.Error(w, "Failed to update category", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func ReorderCategories(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paperID := vars["paper_id"]

	var reorderDto dtos.ReorderCategoryDto
	if err := json.NewDecoder(r.Body).Decode(&reorderDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Do.Struct(reorderDto); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Verify all categories belong to the paper
		var count int64
		err := tx.Model(&models.QuestionCategory{}).
			Where("paper_id = ? AND id IN ?", paperID, getIDs(reorderDto.Categories)).
			Count(&count).Error
		if err != nil {
			return err
		}

		if int(count) != len(reorderDto.Categories) {
			http.Error(w, "Invalid category IDs", http.StatusBadRequest)
			return nil
		}

		// Update orders
		for _, category := range reorderDto.Categories {
			if err := tx.Model(&models.QuestionCategory{}).
				Where("id = ?", category.ID).
				Update("order", category.Order).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		http.Error(w, "Failed to reorder categories", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetPaperCategories(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paperID := vars["paper_id"]

	var categories []models.QuestionCategory
	if err := db.DB.Where("paper_id = ?", paperID).Find(&categories).Error; err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}

	var response []dtos.QuestionCategoryResponse
	for _, category := range categories {
		response = append(response, dtos.QuestionCategoryResponse{
			ID:    category.ID,
			Name:  category.Name,
			Order: category.Order,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func mustAtoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func getIDs(categories []dtos.CategoryOrderDto) []int {
	ids := make([]int, len(categories))
	for i, c := range categories {
		ids[i] = c.ID
	}
	return ids
}
