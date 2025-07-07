package repositories

import (
	"gorm.io/gorm"

	"pariksha/common/pkg/types"
	"pariksha/paper/internal/models"
)

type PaperPermission struct {
	db *gorm.DB
}

func NewPaperPermission(db *gorm.DB) *PaperPermission {
	return &PaperPermission{db: db}
}

// getTx returns tx if not nil, otherwise returns r.db
func (r *PaperPermission) getTx(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}
	return tx
}

func (r *PaperPermission) Create(tx *gorm.DB, paperID types.PaperID, userID types.UserID) error {
	tx = r.getTx(tx)

	perm := models.PaperPermission{
		PaperID: paperID,
		UserID:  userID,
	}
	perm.SetWrite()

	return tx.Create(&perm).Error
}

func (r *PaperPermission) BulkDeleteByPaperIDs(tx *gorm.DB, paperIDs []types.PaperID) error {
	tx = r.getTx(tx)

	return tx.Where("paper_id IN ?", paperIDs).
		Delete(&models.PaperPermission{}).Error
}
