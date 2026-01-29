package repository

import (
	"gorm.io/gorm"

	"pariksha/common/pkg/types"
	"pariksha/paper/internal/domain"
)

type PaperQuestion struct {
	db *gorm.DB
}

func NewPaperQuestion(db *gorm.DB) *PaperQuestion {
	return &PaperQuestion{db: db}
}

// getTx returns tx if not nil, otherwise returns r.db
func (r *PaperQuestion) getTx(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}
	return tx
}

func (r *PaperQuestion) Create(tx *gorm.DB, paperQuest *domain.PaperQuestion) error {
	tx = r.getTx(tx)
	return tx.Create(paperQuest).Error
}

func (r *PaperQuestion) DeleteByID(tx *gorm.DB, questionID types.QuestionID) error {
	tx = r.getTx(tx)
	return tx.Where("question_id = ?", questionID).
		Delete(&domain.PaperQuestion{}).Error
}

func (r *PaperQuestion) BulkDeleteByPaperIDs(tx *gorm.DB, paperIDs []types.PaperID) error {
	tx = r.getTx(tx)
	return tx.Where("paper_id IN ?", paperIDs).
		Delete(&domain.PaperQuestion{}).Error
}

func (r *PaperQuestion) BulkDeleteByPaperIDAndCategoryID(tx *gorm.DB, paperID types.PaperID, categoryID types.CategoryID) error {
	tx = r.getTx(tx)
	return tx.Where("paper_id = ? AND category_id = ?", paperID, categoryID).
		Delete(&domain.PaperQuestion{}).Error
}

func (r *PaperQuestion) GetAllByPaperHash(tx *gorm.DB, paperHash string) ([]domain.PaperQuestion, error) {
	tx = r.getTx(tx)

	var paperQuests []domain.PaperQuestion
	err := tx.
		Joins("JOIN papers ON papers.id = paper_questions.paper_id").
		Where("papers.hash = ?", paperHash).
		Find(&paperQuests).Error

	if err != nil {
		return nil, err
	}
	return paperQuests, nil
}

func (r *PaperQuestion) GetByPaperHashAndQuestionID(tx *gorm.DB, paperHash string, questionID types.QuestionID) (*domain.PaperQuestion, error) {
	tx = r.getTx(tx)

	var paperQuest domain.PaperQuestion
	err := tx.
		Joins("JOIN papers ON papers.id = paper_questions.paper_id").
		Where("papers.hash = ? AND paper_questions.question_id = ?", paperHash, questionID).
		Take(&paperQuest).Error

	if err != nil {
		return nil, err
	}
	return &paperQuest, nil
}

// GetMaxQuestionOrder gets max order for questions in a category
func (r *PaperQuestion) GetMaxQuestionOrder(tx *gorm.DB, categoryID int64) (int16, error) {
	tx = r.getTx(tx)
	var maxOrder struct{ MaxOrder int16 }
	err := tx.Model(&domain.PaperQuestion{}).
		Where("category_id = ?", categoryID).
		Select("COALESCE(MAX(\"order\"), 0) as max_order").
		Scan(&maxOrder).Error
	return maxOrder.MaxOrder, err
}

// ValidateCategoryQuestions checks if questions belong to category
func (r *PaperQuestion) ValidateCategoryQuestions(tx *gorm.DB, categoryID int64, questionIDs []types.QuestionID) (int64, error) {
	tx = r.getTx(tx)

	var count int64
	err := tx.Model(&domain.PaperQuestion{}).
		Where("category_id = ? AND question_id IN ?", categoryID, questionIDs).
		Count(&count).Error
	return count, err
}

// UpdateOrder updates question order
func (r *PaperQuestion) UpdateOrder(tx *gorm.DB, questionID types.QuestionID, order int16) error {
	tx = r.getTx(tx)
	return tx.Model(&domain.PaperQuestion{}).
		Where("question_id = ?", questionID).
		Update("order", order).Error
}

func (r *PaperQuestion) Save(tx *gorm.DB, paperQuest *domain.PaperQuestion) error {
	tx = r.getTx(tx)
	return tx.Save(paperQuest).Error
}

// GetAllByCategoryID gets all questions in a category
func (r *PaperQuestion) GetAllByCategoryID(tx *gorm.DB, categoryID types.CategoryID) ([]domain.PaperQuestion, error) {
	tx = r.getTx(tx)
	var paperQuests []domain.PaperQuestion
	err := tx.Where("category_id = ?", categoryID).Find(&paperQuests).Error
	return paperQuests, err
}
