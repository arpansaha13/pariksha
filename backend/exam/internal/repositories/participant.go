package repositories

import (
	"database/sql"

	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
)

type Participant struct {
	db *gorm.DB
}

func NewParticipant(db *gorm.DB) *Participant {
	return &Participant{db: db}
}

// getTx returns tx if not nil, otherwise returns r.db
func (r *Participant) getTx(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}
	return tx
}

func (r *Participant) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return utils.TransactionHandler(r.db, fc, opts...)
}

// GetParticipantsByExamHash gets all participants for an exam
func (r *Participant) GetAllByExamHash(tx *gorm.DB, examHash string) ([]models.ExamParticipant, error) {
	tx = r.getTx(tx)
	var participants []models.ExamParticipant
	err := tx.Joins("JOIN exams ON exams.id = exam_participants.exam_id").
		Where("exams.hash = ?", examHash).
		Find(&participants).Error
	return participants, err
}

// Create creates a new exam participant
func (r *Participant) Create(tx *gorm.DB, participant *models.ExamParticipant) error {
	tx = r.getTx(tx)
	return tx.Create(participant).Error
}

// GetParticipantByID gets a participant by ID
func (r *Participant) GetByID(tx *gorm.DB, participantID types.ParticipantID) (*models.ExamParticipant, error) {
	tx = r.getTx(tx)
	var participant models.ExamParticipant
	err := tx.Take(&participant, participantID).Error
	if err != nil {
		return nil, err
	}
	return &participant, nil
}

// DeleteParticipant deletes a participant and their permissions
func (r *Participant) Delete(tx *gorm.DB, participant *models.ExamParticipant) error {
	tx = r.getTx(tx)
	if err := tx.Where("exam_id = ? AND user_id = ?", participant.ExamID, participant.UserID).
		Delete(&models.ExamPermission{}).Error; err != nil {
		return err
	}
	return tx.Delete(participant).Error
}

// GetByExamHashAndUserID fetches a participant by exam hash and user ID.
func (r *Participant) GetByExamHashAndUserID(tx *gorm.DB, examHash string, userID types.UserID) (*models.ExamParticipant, error) {
	tx = r.getTx(tx)
	var participant models.ExamParticipant
	err := tx.Joins("JOIN exams ON exams.id = exam_participants.exam_id").
		Where("exams.hash = ? AND exam_participants.user_id = ?", examHash, userID).
		Take(&participant).Error
	if err != nil {
		return nil, err
	}
	return &participant, nil
}

// UpdateStatus updates the status of a participant.
func (r *Participant) UpdateStatus(tx *gorm.DB, participant *models.ExamParticipant, status int16) error {
	tx = r.getTx(tx)
	participant.Status = status
	return tx.Save(participant).Error
}

// Save saves a participant.
func (r *Participant) Save(tx *gorm.DB, participant *models.ExamParticipant) error {
	tx = r.getTx(tx)
	return tx.Save(participant).Error
}

// GetByExamAndUser gets a participant by exam ID and user ID.
func (r *Participant) GetByExamAndUser(tx *gorm.DB, examID types.ExamID, userID types.UserID) (*models.ExamParticipant, error) {
	tx = r.getTx(tx)
	var participant models.ExamParticipant
	err := tx.Where("exam_id = ? AND user_id = ?", examID, userID).Take(&participant).Error
	if err != nil {
		return nil, err
	}
	return &participant, nil
}
