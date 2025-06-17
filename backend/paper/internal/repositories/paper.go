package repositories

import (
	"database/sql"
	"encoding/json"

	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/types"
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

// GetAllByUserId fetches all papers accessible to a user
func (r *Paper) GetAllByUserId(tx *gorm.DB, userID types.UserID) ([]models.Paper, error) {
	tx = r.getTx(tx)
	var papers []models.Paper
	err := tx.
		Select("papers.id, papers.hash, papers.title, papers.max_score, papers.duration_minutes, papers.question_counts, papers.created_by").
		Joins("INNER JOIN permissions ON permissions.paper_id = papers.id").
		Where("permissions.user_id = ?", userID).
		Find(&papers).Error
	return papers, err
}

// Create creates a new paper and its associated records
func (r *Paper) Create(tx *gorm.DB, paper *models.Paper, userID types.UserID) error {
	tx = r.getTx(tx)
	if err := tx.Create(paper).Error; err != nil {
		return err
	}

	// Create default category
	defaultCategory := models.QuestionCategory{
		PaperID: sql.NullInt64{Int64: int64(paper.ID), Valid: true},
		Name:    "Category 1",
		Order:   1,
	}
	if err := tx.Create(&defaultCategory).Error; err != nil {
		return err
	}

	// Create permissions entry
	permissions := models.PaperPermission{
		PaperID: paper.ID,
		UserID:  userID,
	}
	permissions.SetWrite()

	return tx.Create(&permissions).Error
}

// UpdateHash updates the hash of a paper
func (r *Paper) UpdateHash(tx *gorm.DB, paper *models.Paper, hash string) error {
	tx = r.getTx(tx)
	return tx.Model(paper).Update("hash", hash).Error
}

// GetByHash fetches a paper by its hash
func (r *Paper) GetByHash(tx *gorm.DB, hash string) (*models.Paper, error) {
	tx = r.getTx(tx)
	var paper models.Paper
	err := tx.Where("hash = ?", hash).First(&paper).Error
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

// BulkDelete deletes papers and their related entities
func (r *Paper) BulkDelete(tx *gorm.DB, paperIDs []types.PaperID) error {
	tx = r.getTx(tx)
	// Delete papers
	if err := tx.Where("id IN ?", paperIDs).Delete(&models.Paper{}).Error; err != nil {
		return err
	}

	// Delete non-locked questions
	if err := tx.Where("paper_id IN ? AND locked = ?", paperIDs, false).
		Delete(&models.Question{}).Error; err != nil {
		return err
	}

	// Delete non-locked categories
	if err := tx.Where("paper_id IN ? AND locked = ?", paperIDs, false).
		Delete(&models.QuestionCategory{}).Error; err != nil {
		return err
	}

	// Delete permissions
	return tx.Where("paper_id IN ?", paperIDs).
		Delete(&models.PaperPermission{}).Error
}

// GetDetails fetches paper details by ID
func (r *Paper) GetDetails(tx *gorm.DB, paperID types.PaperID) (*models.Paper, error) {
	tx = r.getTx(tx)
	var paper models.Paper
	err := tx.Select("id, question_counts").Where("id = ?", paperID).Take(&paper).Error
	return &paper, err
}

// UpdateMaxScore updates the paper's max score
func (r *Paper) UpdateMaxScore(tx *gorm.DB, paperID types.PaperID, scoreDiff int16) error {
	tx = r.getTx(tx)
	return tx.Model(&models.Paper{}).
		Where("id = ?", paperID).
		UpdateColumn("max_score", gorm.Expr("max_score - ?", scoreDiff)).Error
}

// UpdateStats updates paper's statistics
func (r *Paper) UpdateStats(tx *gorm.DB, paper models.Paper, scoreDiff int32, newQuestionCounts json.RawMessage) error {
	tx = r.getTx(tx)
	return tx.Model(&paper).
		Updates(map[string]any{
			"max_score":       gorm.Expr("max_score + ?", scoreDiff),
			"question_counts": newQuestionCounts,
		}).Error
}

// UpdateQuestionCounts updates the paper's question counts
func (r *Paper) UpdateQuestionCounts(tx *gorm.DB, paperID types.PaperID, counts json.RawMessage) error {
	tx = r.getTx(tx)
	return tx.Model(&models.Paper{}).
		Where("id = ?", paperID).
		Update("question_counts", counts).Error
}
