package repositories

import (
	"database/sql"

	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
)

// PermissionFlags represents the four types of permissions for an exam.
type PermissionFlags struct {
	Read        bool
	Write       bool
	Participate bool
	Evaluate    bool
}

type Permission struct {
	db *gorm.DB
}

func NewPermission(db *gorm.DB) *Permission {
	return &Permission{db: db}
}

func (r *Permission) getTx(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}
	return tx
}

func (r *Permission) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return utils.TransactionHandler(r.db, fc, opts...)
}

func (r *Permission) Create(tx *gorm.DB, examID types.ExamID, userID types.UserID, flags *PermissionFlags) error {
	tx = r.getTx(tx)
	permission := models.ExamPermission{
		ExamID: examID,
		UserID: userID,
	}
	if flags.Read {
		permission.SetRead()
	}
	if flags.Write {
		permission.SetWrite()
	}
	if flags.Participate {
		permission.SetParticipate()
	}
	if flags.Evaluate {
		permission.SetEvaluate()
	}
	return tx.Create(&permission).Error
}

func (r *Permission) DeleteByExamIDs(tx *gorm.DB, examIDs []types.ExamID) error {
	tx = r.getTx(tx)
	return tx.Where("exam_id IN ?", examIDs).Delete(&models.ExamPermission{}).Error
}

func (r *Permission) GetByExamHashesAndUserId(tx *gorm.DB, examHashes []string, userID types.UserID) ([]models.ExamPermission, error) {
	tx = r.getTx(tx)

	var permissions []models.ExamPermission
	err := tx.Joins("INNER JOIN exams ON exams.id = permissions.exam_id").
		Where("exams.hash IN ? AND permissions.user_id = ?", examHashes, userID).
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *Permission) GetByExamHashAndUserId(tx *gorm.DB, examHash string, userID types.UserID) (*models.ExamPermission, error) {
	tx = r.getTx(tx)

	permissions, err := r.GetByExamHashesAndUserId(tx, []string{examHash}, userID)

	if err != nil {
		return nil, err
	}
	if len(permissions) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &permissions[0], nil
}
