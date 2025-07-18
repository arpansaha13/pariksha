package repositories

import (
	"gorm.io/gorm"

	"pariksha/question/internal/models"
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
func (r *Category) Create(tx *gorm.DB, category *models.Category) error {
	tx = r.getTx(tx)
	return tx.Create(category).Error
}

// GetByIDs retrieves categories by their IDs
func (r *Category) GetByIDs(tx *gorm.DB, ids []int64) ([]models.Category, error) {
	if len(ids) == 0 {
		return []models.Category{}, nil
	}

	tx = r.getTx(tx)
	var categories []models.Category
	err := tx.Where("id IN ?", ids).Find(&categories).Error
	return categories, err
}

func (r *Category) UpdateName(tx *gorm.DB, id int64, name string) error {
	tx = r.getTx(tx)

	result := tx.Model(&models.Category{}).
		Where("id = ?", id).
		Update("name", name)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
