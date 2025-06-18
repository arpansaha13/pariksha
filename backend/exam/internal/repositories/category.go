package repositories

import (
	"database/sql"

	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/utils"
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

func (r *Category) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return utils.TransactionHandler(r.db, fc, opts...)
}

// GetExamCategories gets all categories for an exam by exam hash.
func (r *Category) GetExamCategories(tx *gorm.DB, examHash string) ([]models.ExamCategory, error) {
	tx = r.getTx(tx)
	var categories []models.ExamCategory
	err := tx.Model(&models.ExamCategory{}).
		Select("exam_categories.category_id", "exam_categories.order").
		Joins("JOIN exams ON exams.id = exam_categories.exam_id").
		Where("exams.hash = ?", examHash).
		Find(&categories).Error
	return categories, err
}
