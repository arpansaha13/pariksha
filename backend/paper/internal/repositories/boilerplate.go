package repositories

import (
	"pariksha/common/pkg/models"

	"gorm.io/gorm"
)

type Boilerplate struct {
	db *gorm.DB
}

func NewBoilerplate(db *gorm.DB) *Boilerplate {
	return &Boilerplate{db: db}
}

// getTx returns tx if not nil, otherwise returns r.db
func (r *Boilerplate) getTx(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}
	return tx
}

// GetByQuestionAndLanguage fetches boilerplate code for a question and language
func (r *Boilerplate) GetByQuestionAndLanguage(questionHash string, languageID int32) (*models.Boilerplate, error) {
	tx := r.getTx(nil)
	var boilerplate models.Boilerplate
	err := tx.Joins("INNER JOIN questions ON questions.id = boilerplates.question_id").
		Where("questions.hash = ? AND boilerplates.language_id = ?", questionHash, languageID).
		Take(&boilerplate).Error
	return &boilerplate, err
}
