package repositories

import (
	"database/sql"
	"encoding/json"

	"gorm.io/gorm"

	"pariksha/common/pkg/models"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/structs"
	"pariksha/common/pkg/types"
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

// GetPaperQuestions fetches all questions for a paper
func (r *Question) GetPaperQuestions(tx *gorm.DB, paperHash string) ([]models.Question, error) {
	tx = r.getTx(tx)
	var questions []models.Question
	err := tx.Select("questions.id, question, category_id, paper_id, \"order\", questions.hash").
		Joins("INNER JOIN papers ON papers.id = questions.paper_id").
		Where("papers.hash = ?", paperHash).
		Order("\"order\" ASC").
		Find(&questions).Error
	return questions, err
}

// GetTestCasesForQuestion fetches test cases for a coding question
func (r *Question) GetTestCasesForQuestion(tx *gorm.DB, questionID types.QuestionID) ([]models.TestCase, error) {
	tx = r.getTx(tx)
	var testCases []models.TestCase
	err := tx.Where("question_id = ?", questionID).Find(&testCases).Error
	return testCases, err
}

// GetQuestionByHash fetches a question by its hash
func (r *Question) GetQuestionByHash(tx *gorm.DB, hash string) (*models.Question, error) {
	tx = r.getTx(tx)
	var question models.Question
	err := tx.Where("hash = ?", hash).Take(&question).Error
	return &question, err
}

// GetMaxQuestionOrder gets max order for questions in a category
func (r *Question) GetMaxQuestionOrder(tx *gorm.DB, categoryID int64) (int16, error) {
	tx = r.getTx(tx)
	var maxOrder struct{ MaxOrder int16 }
	err := tx.Model(&models.Question{}).
		Where("category_id = ?", categoryID).
		Select("COALESCE(MAX(\"order\"), 0) as max_order").
		Scan(&maxOrder).Error
	return maxOrder.MaxOrder, err
}

// CreateQuestion creates a new question
func (r *Question) CreateQuestion(tx *gorm.DB, question *models.Question) error {
	tx = r.getTx(tx)
	return tx.Create(question).Error
}

// UpdateQuestionHash updates question hash
func (r *Question) UpdateQuestionHash(tx *gorm.DB, questionID types.QuestionID, hash string) error {
	tx = r.getTx(tx)
	return tx.Model(&models.Question{}).Where("id = ?", questionID).Update("hash", hash).Error
}

// UnlinkFromPaper unlinks question from paper
func (r *Question) UnlinkFromPaper(tx *gorm.DB, questionID types.QuestionID) error {
	tx = r.getTx(tx)
	return tx.Model(&models.Question{}).
		Where("id = ?", questionID).
		Update("paper_id", sql.NullInt64{}).Error
}

// Delete permanently deletes a question
func (r *Question) Delete(tx *gorm.DB, questionID types.QuestionID) error {
	tx = r.getTx(tx)
	return tx.Delete(&models.Question{}, questionID).Error
}

// ValidateCategoryQuestions checks if questions belong to category
func (r *Question) ValidateCategoryQuestions(tx *gorm.DB, categoryID int64, questionHashes []string) (int64, error) {
	tx = r.getTx(tx)

	var count int64
	err := tx.Model(&models.Question{}).
		Where("category_id = ? AND hash IN ?", categoryID, questionHashes).
		Count(&count).Error
	return count, err
}

// UpdateOrder updates question order
func (r *Question) UpdateOrder(tx *gorm.DB, questionHash string, order int16) error {
	tx = r.getTx(tx)

	return tx.Model(&models.Question{}).
		Where("hash = ?", questionHash).
		Update("order", order).Error
}

// GetAllInCategory gets all questions in a category
func (r *Question) GetAllInCategory(tx *gorm.DB, categoryID types.CategoryID) ([]models.Question, error) {
	tx = r.getTx(tx)

	var questions []models.Question
	err := tx.Where("category_id = ?", categoryID).Find(&questions).Error
	return questions, err
}

// DeleteNonLocked deletes non-locked questions in category
func (r *Question) DeleteNonLocked(tx *gorm.DB, categoryID types.CategoryID) error {
	tx = r.getTx(tx)

	return tx.Where("category_id = ? AND locked = ?", categoryID, false).
		Delete(&models.Question{}).Error
}

// UnlinkLockedInCategoryFromPaper unlinks locked questions from paper
func (r *Question) UnlinkLockedInCategoryFromPaper(tx *gorm.DB, categoryID types.CategoryID) error {
	tx = r.getTx(tx)

	return tx.Model(&models.Question{}).
		Where("category_id = ? AND locked = true", categoryID).
		Update("paper_id", nil).Error
}

// UpdateCategory updates all questions to point to a new category
func (r *Question) UpdateCategory(tx *gorm.DB, oldCategoryID, newCategoryID types.CategoryID) error {
	tx = r.getTx(tx)

	return tx.Model(&models.Question{}).
		Where("category_id = ?", oldCategoryID).
		Update("category_id", newCategoryID).Error
}

// GetNonHiddenTestCases fetches only non-hidden test cases for a question
func (r *Question) GetNonHiddenTestCases(tx *gorm.DB, questionID types.QuestionID) ([]models.TestCase, error) {
	tx = r.getTx(tx)

	var testCases []models.TestCase
	err := tx.Where("question_id = ? AND hidden = ?", questionID, false).Find(&testCases).Error
	return testCases, err
}

// GetExamQuestionByHash fetches minimal question data for exam taking
func (r *Question) GetExamQuestionByHash(tx *gorm.DB, hash string) (*models.Question, error) {
	tx = r.getTx(tx)

	var question models.Question
	err := tx.Select("id, question, type").
		Where("hash = ?", hash).Take(&question).Error
	return &question, err
}

// GetQuestionHashes fetches questions by IDs and returns them
func (r *Question) GetQuestionHashes(tx *gorm.DB, questionIDs []int64) ([]models.Question, error) {
	tx = r.getTx(tx)

	var questions []models.Question
	err := tx.Select("id", "hash").
		Where("id IN ?", questionIDs).
		Find(&questions).Error
	return questions, err
}

// GetQuestionsByHashes fetches questions by their hashes
func (r *Question) GetQuestionsByHashes(tx *gorm.DB, questionHashes []string) ([]models.Question, error) {
	tx = r.getTx(tx)

	var questions []models.Question
	err := tx.Select("id", "hash").
		Where("hash IN ?", questionHashes).
		Find(&questions).Error
	return questions, err
}

// GetQuestionHashes fetches questions by IDs and returns them
func (r *Question) GetByIds(tx *gorm.DB, questionIDs []int64) ([]models.Question, error) {
	tx = r.getTx(tx)

	var questions []models.Question
	err := tx.
		Where("id IN ?", questionIDs).
		Find(&questions).Error
	return questions, err
}

// GetInputDefinitionsLength retrieves the length of InputDefinitions array
func (r *Question) GetInputDefinitionsLength(tx *gorm.DB, questionID types.QuestionID) (int, error) {
	tx = r.getTx(tx)

	var length int
	err := tx.Raw(`
			SELECT jsonb_array_length(
					(Question->>'input_definitions')::jsonb
			) FROM questions WHERE id = ?
    `, questionID).Scan(&length).Error
	return length, err
}

// GetInputDefinitionsByHash fetches input definitions for a coding question by hash.
func (r *Question) GetInputDefinitionsByHash(tx *gorm.DB, questionHash string) ([]structs.InputDefinition, error) {
	tx = r.getTx(tx)

	type QueryResult struct {
		InputDefs []byte             `gorm:"column:input_defs"`
		Type      proto.QuestionType `gorm:"column:type"`
	}
	var queryRes QueryResult
	if err := tx.Model(&models.Question{}).
		Select("type, question->>'input_definitions' as input_defs").
		Where("hash = ?", questionHash).
		Take(&queryRes).Error; err != nil {
		return nil, err
	}

	// Only allow coding questions
	if queryRes.Type != proto.QuestionType_CODING {
		return nil, sql.ErrNoRows
	}

	var inputDefs []structs.InputDefinition
	if err := json.Unmarshal(queryRes.InputDefs, &inputDefs); err != nil {
		return nil, err
	}
	return inputDefs, nil
}
