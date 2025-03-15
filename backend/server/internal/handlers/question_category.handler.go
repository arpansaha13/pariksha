package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
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
			Where("paper_id = ? AND id IN ?", paperID, reorderDto.Categories).
			Count(&count).Error
		if err != nil {
			return err
		}

		if int(count) != len(reorderDto.Categories) {
			http.Error(w, "Invalid category IDs", http.StatusBadRequest)
			return nil
		}

		// Update orders based on array position
		for i, categoryID := range reorderDto.Categories {
			if err := tx.Model(&models.QuestionCategory{}).
				Where("id = ?", categoryID).
				Update("order", i+1).Error; err != nil {
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

func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID := vars["id"]

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Get category to verify it exists
		var category models.QuestionCategory
		if err := tx.Take(&category, categoryID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Category not found", http.StatusNotFound)
				return err
			}
			return err
		}

		// Get paper to update question counts
		var paper models.Paper
		if err := tx.First(&paper, category.PaperID).Error; err != nil {
			return err
		}

		// Get questions in this category
		var questions []models.Question
		if err := tx.Where("category_id = ?", categoryID).Find(&questions).Error; err != nil {
			return err
		}

		// Update paper's question counts
		var questionCounts models.QuestionCount
		if err := json.Unmarshal(paper.QuestionCounts, &questionCounts); err != nil {
			return err
		}

		// Decrement counts based on questions being deleted
		for _, q := range questions {
			switch q.Type {
			case constants.QUESTION_TYPE_MCQ:
				questionCounts.MCQ--
			case constants.QUESTION_TYPE_SHORT:
				questionCounts.Short--
			case constants.QUESTION_TYPE_LONG:
				questionCounts.Long--
			}
			paper.MaxScore -= q.MaxScore
		}

		// Update paper with new counts
		newCounts, err := json.Marshal(questionCounts)
		if err != nil {
			return err
		}
		paper.QuestionCounts = newCounts

		if err := tx.Save(&paper).Error; err != nil {
			return err
		}

		// Delete all questions in this category
		if err := tx.Where("category_id = ?", categoryID).Delete(&models.Question{}).Error; err != nil {
			return err
		}

		// Delete the category
		if err := tx.Delete(&category).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
