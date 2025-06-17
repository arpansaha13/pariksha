package repositories

import (
	"database/sql"

	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/types"
)

type Category struct {
	db *gorm.DB
}

func NewCategory(db *gorm.DB) *Category {
	return &Category{db: db}
}

// getTx returns tx if not nil, otherwise returns r.db
func (r *Category) getTx(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}
	return tx
}

// Create creates a new category
func (r *Category) Create(tx *gorm.DB, category *models.QuestionCategory) error {
	tx = r.getTx(tx)
	return tx.Create(category).Error
}

// Delete delete a category by its ID
func (r *Category) Delete(tx *gorm.DB, categoryID types.CategoryID) error {
	tx = r.getTx(tx)
	return tx.Model(&models.QuestionCategory{}).
		Delete("id = ?", categoryID).Error
}

// GetAllByPaperHash fetches all categories for a paper
func (r *Category) GetAllByPaperHash(paperHash string) ([]models.QuestionCategory, error) {
	var categories []models.QuestionCategory
	err := r.db.Joins("INNER JOIN papers ON papers.id = categories.paper_id").
		Where("papers.hash = ?", paperHash).
		Order("\"order\" ASC").
		Find(&categories).Error
	return categories, err
}

// ValidatePaperCategories checks if categories belong to paper
func (r *Category) ValidatePaperCategories(tx *gorm.DB, paperID types.PaperID, categoryIDs []int64) (int64, error) {
	var count int64
	err := tx.Model(&models.QuestionCategory{}).
		Where("paper_id = ? AND id IN ?", paperID, categoryIDs).
		Count(&count).Error
	return count, err
}

// UpdateOrder updates category order
func (r *Category) UpdateOrder(tx *gorm.DB, categoryID int64, order int16) error {
	return tx.Model(&models.QuestionCategory{}).
		Where("id = ?", categoryID).
		Update("order", order).Error
}

// GetMaxOrder gets max order for categories in paper
func (r *Category) GetMaxOrder(tx *gorm.DB, paperID types.PaperID) (int16, error) {
	var maxOrder struct{ MaxOrder int16 }
	err := tx.Model(&models.QuestionCategory{}).
		Where("paper_id = ?", paperID).
		Select("COALESCE(MAX(\"order\"), 0) as max_order").
		Scan(&maxOrder).Error
	return maxOrder.MaxOrder, err
}

// GetByID fetches category by ID
func (r *Category) GetByID(tx *gorm.DB, categoryID int64) (*models.QuestionCategory, error) {
	var category models.QuestionCategory
	err := tx.Where("id = ?", categoryID).Take(&category).Error
	return &category, err
}

// UpdateName updates category name
func (r *Category) UpdateName(tx *gorm.DB, categoryID int64, name string) error {
	return tx.Model(&models.QuestionCategory{}).
		Where("id = ?", categoryID).
		Update("name", name).Error
}

// UnlinkFromPaper unlinks category from paper
func (r *Category) UnlinkFromPaper(tx *gorm.DB, categoryID types.CategoryID) error {
	return tx.Model(&models.QuestionCategory{}).
		Where("id = ?", categoryID).
		Update("paper_id", sql.NullInt64{}).Error
}

// GetCountByPaperId gets total categories count for paper
func (r *Category) GetCountByPaperId(tx *gorm.DB, paperID sql.NullInt64) (int64, error) {
	var count int64
	err := tx.Model(&models.QuestionCategory{}).
		Where("paper_id = ?", paperID).
		Count(&count).Error
	return count, err
}

// GetByIds fetches multiple categories by their IDs
func (r *Category) GetByIds(tx *gorm.DB, categoryIDs []int64) ([]models.QuestionCategory, error) {
	tx = r.getTx(tx)
	var categories []models.QuestionCategory
	err := tx.Where("id IN ?", categoryIDs).Find(&categories).Error
	return categories, err
}
