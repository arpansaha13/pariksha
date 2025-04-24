package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/models"
	"pariksha/common/pkg/types"
	"pariksha/workers/exam/internal/config/db"
)

// AutoEndExam handles the automatic ending of an exam after its duration expires
func AutoEndExam(ctx context.Context, task *asynq.Task) error {
	var payload types.AutoEndExamPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Default().Printf("Failed to unmarshal auto-end payload: %v", err)
		return err
	}

	err := db.Exams.Transaction(func(tx *gorm.DB) error {
		// Fetch participant and check if already ended
		var participant models.ExamParticipant
		if err := tx.Where("id = ? AND exam_id = ?", payload.ParticipantID, payload.ExamID).Take(&participant).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Default().Printf("Participant not found for auto-end: exam=%d, participant=%d", payload.ExamID, payload.ParticipantID)
				return nil
			}
			return err
		}

		// If already ended, skip
		if participant.Status == constants.PARTICIPANT_STATUS_ENDED {
			log.Default().Printf("Participant already ended: exam=%d, participant=%d", payload.ExamID, payload.ParticipantID)
			return nil
		}

		// Update participant status
		participant.Status = constants.PARTICIPANT_STATUS_ENDED
		participant.EndedAt = sql.NullTime{Time: time.Now(), Valid: true}

		// Update participant counts in exam
		var exam models.Exam
		if err := tx.Take(&exam, payload.ExamID).Error; err != nil {
			return err
		}

		counts, err := exam.GetParticipantCounts()
		if err != nil {
			return err
		}

		counts.Started--
		counts.Ended++
		countsJSON, err := json.Marshal(counts)
		if err != nil {
			return err
		}

		exam.ParticipantCounts = countsJSON

		// Save changes
		if err := tx.Save(&participant).Error; err != nil {
			return err
		}
		if err := tx.Save(&exam).Error; err != nil {
			return err
		}

		log.Default().Printf("Auto-ended exam for participant: exam=%d, participant=%d", payload.ExamID, payload.ParticipantID)
		return nil
	})

	if err != nil {
		log.Default().Printf("Failed to auto-end exam: %v", err)
		return err
	}

	return nil
}
