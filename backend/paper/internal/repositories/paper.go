package repositories

import (
	"database/sql"

	"gorm.io/gorm"

	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/paper/internal/models"
)

type Paper struct {
	db *gorm.DB
}

func NewPaper(db *gorm.DB) *Paper {
	return &Paper{db: db}
}

// getTx returns tx if not nil, otherwise returns r.db
func (r *Paper) getTx(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}
	return tx
}

func (r *Paper) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return utils.TransactionHandler(r.db, fc, opts...)
}

// GetAllByUserId fetches all papers accessible to a user
func (r *Paper) GetAllByUserId(tx *gorm.DB, userID types.UserID) ([]models.Paper, error) {
	tx = r.getTx(tx)
	var papers []models.Paper
	err := tx.
		Select("papers.id, papers.hash, papers.title, papers.duration_minutes, papers.created_by").
		Joins("INNER JOIN permissions ON permissions.paper_id = papers.id").
		Where("permissions.user_id = ?", userID).
		Find(&papers).Error
	return papers, err
}

// Create creates a new paper and its associated records
func (r *Paper) Create(tx *gorm.DB, paper *models.Paper, userID types.UserID) error {
	tx = r.getTx(tx)
	return tx.Create(paper).Error
}

// UpdateHash updates the hash of a paper
func (r *Paper) UpdateHash(tx *gorm.DB, paper *models.Paper, hash string) error {
	tx = r.getTx(tx)
	return tx.Model(paper).Update("hash", hash).Error
}

// GetByID fetches a paper by its id
func (r *Paper) GetByID(tx *gorm.DB, paperId types.PaperID) (*models.Paper, error) {
	tx = r.getTx(tx)
	var paper models.Paper
	err := tx.Where("id = ?", paperId).Take(&paper).Error
	return &paper, err
}

// GetByHash fetches a paper by its hash
func (r *Paper) GetByHash(tx *gorm.DB, hash string) (*models.Paper, error) {
	tx = r.getTx(tx)
	var paper models.Paper
	err := tx.Where("hash = ?", hash).Take(&paper).Error
	return &paper, err
}

// Update updates the paper's details
func (r *Paper) Update(tx *gorm.DB, paper *models.Paper) error {
	tx = r.getTx(tx)
	return tx.Save(paper).Error
}

// GetIDsByHashes retrieves paper IDs from their hashes
func (r *Paper) GetIDsByHashes(tx *gorm.DB, hashes []string) ([]types.PaperID, error) {
	tx = r.getTx(tx)
	var paperIDs []types.PaperID
	err := tx.Model(&models.Paper{}).
		Where("hash IN ?", hashes).
		Pluck("id", &paperIDs).Error
	return paperIDs, err
}

// BulkDelete deletes papers by their IDs
func (r *Paper) BulkDelete(tx *gorm.DB, paperIDs []types.PaperID) error {
	tx = r.getTx(tx)

	return tx.Where("id IN ?", paperIDs).Delete(&models.Paper{}).Error
}
