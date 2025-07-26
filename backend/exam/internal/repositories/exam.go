package repositories

import (
	"database/sql"

	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
)

type Exam struct {
	db *gorm.DB
}

func NewExam(db *gorm.DB) *Exam {
	return &Exam{db: db}
}

// getTx returns tx if not nil, otherwise returns r.db
func (r *Exam) getTx(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}
	return tx
}

func (r *Exam) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return utils.TransactionHandler(r.db, fc, opts...)
}

// GetByUserID gets all exams created by or participated in by a user.
func (r *Exam) GetByUserID(tx *gorm.DB, userID types.UserID) ([]models.Exam, error) {
	tx = r.getTx(tx)
	var exams []models.Exam
	err := tx.Where("created_by = ?", userID).
		Or("id IN (?)", tx.Model(&models.ExamParticipant{}).
			Select("exam_id").
			Where("user_id = ?", userID)).
		Find(&exams).Error
	return exams, err
}

func (r *Exam) GetByHash(tx *gorm.DB, examHash string) (*models.Exam, error) {
	tx = r.getTx(tx)

	var exam models.Exam
	if err := tx.Where("hash = ?", examHash).Take(&exam).Error; err != nil {
		return nil, err
	}

	return &exam, nil
}

// Create creates a new exam.
func (r *Exam) Create(tx *gorm.DB, exam *models.Exam) error {
	tx = r.getTx(tx)
	return tx.Create(exam).Error
}

// UpdateHash updates an exam's hash.
func (r *Exam) UpdateHash(tx *gorm.DB, exam *models.Exam) error {
	tx = r.getTx(tx)
	return tx.Model(exam).Update("hash", exam.Hash).Error
}

// Save saves all exam fields.
func (r *Exam) Save(tx *gorm.DB, exam *models.Exam) error {
	tx = r.getTx(tx)
	return tx.Save(exam).Error
}

// DeleteByHashes deletes exams by their hashes and returns their IDs.
func (r *Exam) DeleteByHashes(tx *gorm.DB, hashes []string) ([]types.ExamID, error) {
	tx = r.getTx(tx)
	var examIDs []types.ExamID
	err := tx.Model(&models.Exam{}).
		Where("hash IN ?", hashes).
		Pluck("id", &examIDs).Error
	if err != nil {
		return nil, err
	}
	err = tx.Where("id IN ?", examIDs).Delete(&models.Exam{}).Error
	return examIDs, err
}
