package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/structs"
	"pariksha/workers/exam/internal/config/db"
)

func PostDeleteExamsCleanup(ctx context.Context, task *asynq.Task) error {
	var payload structs.DeleteExamsPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Default().Printf("Failed to unmarshal delete payload: %v", err)
		return err
	}

	err := db.Exams.Transaction(func(tx *gorm.DB) error {
		// Delete exam questions
		if err := tx.Where("exam_id IN ?", payload.ExamIDs).
			Delete(&models.ExamQuestion{}).Error; err != nil {
			return err
		}

		// Delete exam categories
		if err := tx.Where("exam_id IN ?", payload.ExamIDs).
			Delete(&models.ExamCategory{}).Error; err != nil {
			return err
		}

		// Get participant IDs first
		var participantIDs []int64
		if err := tx.Model(&models.ExamParticipant{}).
			Where("exam_id IN ?", payload.ExamIDs).
			Pluck("id", &participantIDs).Error; err != nil {
			return err
		}

		// Delete answers for these participants
		if len(participantIDs) > 0 {
			if err := tx.Where("exam_participant_id IN ?", participantIDs).
				Delete(&models.Answer{}).Error; err != nil {
				return err
			}
		}

		// Delete exam participants
		if err := tx.Where("exam_id IN ?", payload.ExamIDs).
			Delete(&models.ExamParticipant{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Default().Printf("Failed to delete exam data: %v", err)
		return err
	}

	log.Default().Printf("Successfully deleted data for exams: %v", payload.ExamIDs)
	return nil
}
