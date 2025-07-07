package repositories

import (
	"database/sql"

	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/utils"
)

type Question struct {
	db *gorm.DB
}

func NewQuestion(db *gorm.DB) *Question {
	return &Question{db: db}
}

// getTx returns tx if not nil, otherwise returns r.db
func (r *Question) getTx(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}
	return tx
}

func (r *Question) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return utils.TransactionHandler(r.db, fc, opts...)
}

// GetExamQuestions gets all questions for an exam by exam hash.
func (r *Question) GetExamQuestions(tx *gorm.DB, examHash string) ([]models.ExamQuestion, error) {
	tx = r.getTx(tx)
	var questions []models.ExamQuestion
	err := tx.Model(&models.ExamQuestion{}).
		Select("exam_questions.question_id", "exam_questions.category_id", "exam_questions.order", "exam_questions.max_score").
		Joins("JOIN exams ON exams.id = exam_questions.exam_id").
		Where("exams.hash = ?", examHash).
		Find(&questions).Error
	return questions, err
}
