package repositories

import (
	"gorm.io/gorm"

	"pariksha/common/pkg/types"
	"pariksha/paper/internal/models"
)

type PaperCategory struct {
	db *gorm.DB
}

func NewPaperCategory(db *gorm.DB) *PaperCategory {
	return &PaperCategory{db: db}
}

// getTx returns tx if not nil, otherwise returns r.db
func (r *PaperCategory) getTx(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}
	return tx
}

func (r *PaperCategory) Create(tx *gorm.DB, paperCat *models.PaperCategory) error {
	tx = r.getTx(tx)
	return tx.Create(paperCat).Error
}

func (r *PaperCategory) DeleteByID(tx *gorm.DB, paperID types.PaperID, categoryID types.CategoryID) error {
	tx = r.getTx(tx)
	return tx.Where("paper_id = ? AND category_id = ?", paperID, categoryID).
		Delete(&models.PaperCategory{}).Error
}

func (r *PaperCategory) BulkDeleteByPaperIDs(tx *gorm.DB, paperIDs []types.PaperID) error {
	tx = r.getTx(tx)

	return tx.Where("paper_id IN ?", paperIDs).
		Delete(&models.PaperCategory{}).Error
}

func (r *PaperCategory) GetAllByPaperHash(tx *gorm.DB, paperHash string) ([]models.PaperCategory, error) {
	tx = r.getTx(tx)

	var paperCategories []models.PaperCategory
	err := tx.
		Joins("JOIN papers ON papers.id = paper_categories.paper_id").
		Where("papers.hash = ?", paperHash).
		Find(&paperCategories).Error

	if err != nil {
		return nil, err
	}
	return paperCategories, nil
}

func (r *PaperCategory) GetByPaperHashAndCategoryID(tx *gorm.DB, paperHash string, categoryID types.CategoryID) (*models.PaperCategory, error) {
	tx = r.getTx(tx)

	var paperCategory models.PaperCategory
	err := tx.
		Joins("JOIN papers ON papers.id = paper_categories.paper_id").
		Where("papers.hash = ? AND category_id = ?", paperHash, categoryID).
		Take(&paperCategory).Error

	if err != nil {
		return nil, err
	}
	return &paperCategory, nil
}

// GetCategoriesCountByPaperHash counts the categories in a paper by paper hash
func (r *PaperCategory) GetCategoriesCountByPaperHash(tx *gorm.DB, paperHash string, categoryIDs []int64) (int64, error) {
	var count int64
	err := tx.Model(&models.PaperCategory{}).
		Joins("JOIN papers ON papers.id = paper_categories.paper_id").
		Where("papers.hash = ? AND category_id IN ?", paperHash, categoryIDs).
		Count(&count).Error
	return count, err
}

// UpdateOrder updates category order
func (r *PaperCategory) UpdateOrder(tx *gorm.DB, categoryID int64, order int16) error {
	return tx.Model(&models.PaperCategory{}).
		Where("id = ?", categoryID).
		Update("order", order).Error
}

// GetMaxOrder gets max order for categories in paper
func (r *PaperCategory) GetMaxOrder(tx *gorm.DB, paperID types.PaperID) (int16, error) {
	var maxOrder struct{ MaxOrder int16 }
	err := tx.Model(&models.PaperCategory{}).
		Where("paper_id = ?", paperID).
		Select("COALESCE(MAX(\"order\"), 0) as max_order").
		Scan(&maxOrder).Error
	return maxOrder.MaxOrder, err
}

// GetCountByPaperId gets total categories count for paper
func (r *PaperCategory) GetCountByPaperId(tx *gorm.DB, paperID types.PaperID) (int64, error) {
	var count int64
	err := tx.Model(&models.PaperCategory{}).
		Where("paper_id = ?", paperID).
		Count(&count).Error
	return count, err
}

// GetCountByPaperHash gets total categories count for the paper
func (r *PaperCategory) GetCountByPaperHash(tx *gorm.DB, paperHash string) (int64, error) {
	var count int64
	err := tx.Model(&models.PaperCategory{}).
		Joins("JOIN papers ON papers.id = paper_categories.paper_id").
		Where("papers.hash = ?", paperHash).
		Count(&count).Error
	return count, err
}
