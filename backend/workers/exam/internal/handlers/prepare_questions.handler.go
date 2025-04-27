package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/structs"
	"pariksha/workers/exam/internal/config/db"
)

func PrepareExamQuestions(ctx context.Context, task *asynq.Task) error {
	var payload structs.PrepareQuestionsPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Default().Printf("Failed to unmarshal payload: %v", err)
		return err
	}

	// Start a transaction in papers DB
	papersTx := db.Papers.Begin()
	if papersTx.Error != nil {
		log.Default().Printf("Failed to start papers transaction: %v", papersTx.Error)
		return papersTx.Error
	}

	// Get all questions for the paper
	var questions []models.Question
	err := papersTx.Select("id", "category_id", "order").Where("paper_id = ?", payload.PaperID).Find(&questions).Error
	if err != nil {
		papersTx.Rollback()
		log.Default().Printf("Failed to fetch questions: %v", err)
		return err
	}

	// Get all categories for the paper
	var categories []models.QuestionCategory
	err = papersTx.Select("id").Where("paper_id = ?", payload.PaperID).Find(&categories).Error
	if err != nil {
		papersTx.Rollback()
		log.Default().Printf("Failed to fetch categories: %v", err)
		return err
	}

	// Lock all questions
	err = papersTx.Model(&models.Question{}).Where("paper_id = ?", payload.PaperID).Update("locked", true).Error
	if err != nil {
		papersTx.Rollback()
		log.Default().Printf("Failed to lock questions: %v", err)
		return err
	}

	// Lock all categories
	err = papersTx.Model(&models.QuestionCategory{}).Where("paper_id = ?", payload.PaperID).Update("locked", true).Error
	if err != nil {
		papersTx.Rollback()
		log.Default().Printf("Failed to lock categories: %v", err)
		return err
	}

	// Start a transaction in exams DB
	examsTx := db.Exams.Begin()
	if examsTx.Error != nil {
		papersTx.Rollback()
		log.Default().Printf("Failed to start exams transaction: %v", examsTx.Error)
		return examsTx.Error
	}

	// Create exam questions
	for _, q := range questions {
		examQuestion := models.ExamQuestion{
			ExamID:     payload.ExamID,
			QuestionID: q.ID,
			CategoryID: q.CategoryID,
			Order:      q.Order,
		}
		if err := examsTx.Create(&examQuestion).Error; err != nil {
			examsTx.Rollback()
			papersTx.Rollback()
			log.Default().Printf("Failed to create exam question: %v", err)
			return err
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
			return err
		}
	}

	// Commit both transactions
	if err := examsTx.Commit().Error; err != nil {
		examsTx.Rollback()
		papersTx.Rollback()
		log.Default().Printf("Failed to commit exams transaction: %v", err)
		return err
	}

	if err := papersTx.Commit().Error; err != nil {
		log.Default().Printf("Failed to commit papers transaction: %v", err)
		return err
	}

	log.Default().Printf("Successfully prepared exam questions for exam %d from paper %d", payload.ExamID, payload.PaperID)
	return nil
}
