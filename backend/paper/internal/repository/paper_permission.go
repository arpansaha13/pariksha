package repository

import (
	"gorm.io/gorm"

	"pariksha/common/pkg/types"
	"pariksha/paper/internal/domain"
)

type PaperPermission struct {
	db *gorm.DB
}

var _ IPaperPermissionRepository = (*PaperPermission)(nil)

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

	perm := domain.PaperPermission{
		PaperID: paperID,
		UserID:  userID,
	}
	perm.SetWrite()

	return tx.Create(&perm).Error
}

func (r *PaperPermission) GetByPaperHashesAndUserId(tx *gorm.DB, paperHashes []string, userID types.UserID) ([]domain.PaperPermission, error) {
	tx = r.getTx(tx)

	var permissions []domain.PaperPermission
	if err := tx.Joins("INNER JOIN papers ON papers.id = permissions.paper_id").
		Where("papers.hash IN ? AND permissions.user_id = ?", paperHashes, userID).
		Find(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *PaperPermission) GetByPaperHashAndUserId(tx *gorm.DB, paperHash string, userID types.UserID) (*domain.PaperPermission, error) {
	permissions, err := r.GetByPaperHashesAndUserId(tx, []string{paperHash}, userID)
	if err != nil {
		return nil, err
	}
	if len(permissions) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &permissions[0], nil
}

func (r *PaperPermission) BulkDeleteByPaperIDs(tx *gorm.DB, paperIDs []types.PaperID) error {
	tx = r.getTx(tx)

	return tx.Where("paper_id IN ?", paperIDs).
		Delete(&domain.PaperPermission{}).Error
}
