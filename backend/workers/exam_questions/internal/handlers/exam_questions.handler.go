package handlers

import (
	"encoding/json"
	"log"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/types"
	"pariksha/workers/exam_questions/internal/config/db"
)

func PrepareExamQuestions(body []byte) {
	var payload types.ExamQueuePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Default().Printf("Failed to unmarshal payload: %v", err)
		return
	}

	// Start a transaction in papers DB
	papersTx := db.Papers.Begin()
	if papersTx.Error != nil {
		log.Default().Printf("Failed to start papers transaction: %v", papersTx.Error)
		return
	}

	// Get all questions for the paper
	var questions []models.Question
	if err := papersTx.Where("paper_id = ?", payload.PaperID).Find(&questions).Error; err != nil {
		papersTx.Rollback()
		log.Default().Printf("Failed to fetch questions: %v", err)
		return
	}

	// Get all categories for the paper
	var categories []models.QuestionCategory
	if err := papersTx.Where("paper_id = ?", payload.PaperID).Find(&categories).Error; err != nil {
		papersTx.Rollback()
		log.Default().Printf("Failed to fetch categories: %v", err)
		return
	}

	// Lock all questions
	if err := papersTx.Model(&models.Question{}).Where("paper_id = ?", payload.PaperID).Update("locked", true).Error; err != nil {
		papersTx.Rollback()
		log.Default().Printf("Failed to lock questions: %v", err)
		return
	}

	// Lock all categories
	if err := papersTx.Model(&models.QuestionCategory{}).Where("paper_id = ?", payload.PaperID).Update("locked", true).Error; err != nil {
		papersTx.Rollback()
		log.Default().Printf("Failed to lock categories: %v", err)
		return
	}

	// Start a transaction in exams DB
	examsTx := db.Exams.Begin()
	if examsTx.Error != nil {
		papersTx.Rollback()
		log.Default().Printf("Failed to start exams transaction: %v", examsTx.Error)
		return
	}

	// Create exam questions
	for _, q := range questions {
		examQuestion := models.ExamQuestion{
			ExamID:     payload.ExamID,
			QuestionID: q.ID,
		}
		if err := examsTx.Create(&examQuestion).Error; err != nil {
			examsTx.Rollback()
			papersTx.Rollback()
			log.Default().Printf("Failed to create exam question: %v", err)
			return
		}
	}

	// Create exam categories
	for _, c := range categories {
		examCategory := models.ExamCategory{
			ExamID:     payload.ExamID,
			CategoryID: c.ID,
		}
		if err := examsTx.Create(&examCategory).Error; err != nil {
			examsTx.Rollback()
			papersTx.Rollback()
			log.Default().Printf("Failed to create exam category: %v", err)
			return
		}
	}

	// Commit both transactions
	if err := examsTx.Commit().Error; err != nil {
		examsTx.Rollback()
		papersTx.Rollback()
		log.Default().Printf("Failed to commit exams transaction: %v", err)
		return
	}

	if err := papersTx.Commit().Error; err != nil {
		log.Default().Printf("Failed to commit papers transaction: %v", err)
		return
	}

	log.Default().Printf("Successfully prepared exam questions for exam %d from paper %d", payload.ExamID, payload.PaperID)
}
