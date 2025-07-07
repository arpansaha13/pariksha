package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/types"
	"pariksha/workers/exam/internal/config/db"
	"pariksha/workers/exam/internal/interservice"
)

func PrepareExamQuestions(ctx context.Context, task *asynq.Task) error {
	var payload structs.PrepareQuestionsPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Default().Printf("Failed to unmarshal payload: %v", err)
		return err
	}

	questions, err := interservice.GetPaperQuestionsMeta(payload.PaperHash)
	if err != nil {
		log.Default().Printf("failed to fetch paper questions meta: %v", err)
		return err
	}

	categories, err := interservice.GetPaperCategoriesMeta(payload.PaperHash)
	if err != nil {
		log.Default().Printf("failed to fetch paper categories meta: %v", err)
		return err
	}

	// Gather question IDs and increment their exam indegree
	questionIDs := make([]types.QuestionID, len(questions))
	for i, q := range questions {
		questionIDs[i] = types.QuestionID(q.Id)
	}
	if err := interservice.IncQuestionExamIndegreeByIds(questionIDs); err != nil {
		log.Default().Printf("failed to increment question exam indegree: %v", err)
		return err
	}

	return db.Exams.Transaction(func(tx *gorm.DB) error {
		var totalMaxScore int32
		// Create exam questions
		for _, q := range questions {
			examQuestion := models.ExamQuestion{
				ExamID:     payload.ExamID,
				QuestionID: types.QuestionID(q.Id),
				CategoryID: types.CategoryID(q.CategoryId),
				Order:      int16(q.Order),
				MaxScore:   int16(q.MaxScore),
			}
			totalMaxScore += int32(q.MaxScore) // Accumulate total max_score

			if err := tx.Create(&examQuestion).Error; err != nil {
				log.Default().Printf("Failed to create exam question: %v", err)
				return err
			}
		}

		// Create exam categories
		for _, c := range categories {
			examCategory := models.ExamCategory{
				ExamID:     payload.ExamID,
				CategoryID: types.CategoryID(c.Id),
				Order:      int16(c.Order),
			}
			if err := tx.Create(&examCategory).Error; err != nil {
				log.Default().Printf("Failed to create exam category: %v", err)
				return err
			}
		}

		// Update exam's max_score
		if err := tx.Model(&models.Exam{}).Where("id = ?", payload.ExamID).Update("max_score", totalMaxScore).Error; err != nil {
			log.Default().Printf("Failed to update exam max score: %v", err)
			return err
		}

		return nil
	})
}
