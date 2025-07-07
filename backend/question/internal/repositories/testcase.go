package repositories

import (
	"database/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"pariksha/common/pkg/types"
	"pariksha/common/pkg/utils"
	"pariksha/question/internal/models"
)

type TestCase struct {
	db *gorm.DB
}

func NewTestcase(db *gorm.DB) *TestCase {
	return &TestCase{db: db}
}

func (r *TestCase) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return utils.TransactionHandler(r.db, fc, opts...)
}

// getTx returns tx if not nil, otherwise returns r.db
func (r *TestCase) getTx(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return r.db
	}
	return tx
}

// GetVisibleByQuestionID retrieves all non-hidden test cases for a given question ID
func (r *TestCase) GetVisibleByQuestionID(questionID int64) ([]*models.TestCase, error) {
	var testCases []*models.TestCase

	err := r.db.
		Where("question_id = ? AND hidden = FALSE", questionID).
		Order("\"order\" ASC").
		Find(&testCases).Error

	if err != nil {
		return nil, err
	}

	return testCases, nil
}

// GetAllByQuestionID retrieves all test cases for a given question ID, including hidden ones
func (r *TestCase) GetAllByQuestionID(tx *gorm.DB, questionID types.QuestionID) ([]models.TestCase, error) {
	tx = r.getTx(tx)

	var testCases []models.TestCase
	err := r.db.
		Where("question_id = ?", questionID).
		Order("\"order\" ASC").
		Find(&testCases).Error

	if err != nil {
		return nil, err
	}

	return testCases, nil
}

// GetUnscopedTestCasesByQuestionId returns test cases including soft-deleted ones with row-level locking
func (r *TestCase) GetUnscopedTestCasesByQuestionId(tx *gorm.DB, questionID types.QuestionID) ([]models.TestCase, error) {
	tx = r.getTx(tx)

	var existingTestCases []models.TestCase
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Unscoped(). // Include soft-deleted records
		Where("question_id = ?", questionID).
		Order("\"order\" ASC").
		Find(&existingTestCases).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch existing test cases")
	}

	return existingTestCases, nil
}

// UpdateUnscopedTestCase updates existing test case (includes reviving soft-deleted ones)
func (r *TestCase) UpdateUnscopedTestCase(tx *gorm.DB, tc models.TestCase) error {
	tx = r.getTx(tx)

	return tx.Unscoped().Model(&models.TestCase{}).
		Where("id = ?", tc.ID).
		Updates(map[string]any{
			"content":    tc.Content,
			"data_hash":  tc.DataHash,
			"hidden":     tc.Hidden,
			"deleted_at": nil,
		}).Error
}

func (r *TestCase) DeleteByIds(tx *gorm.DB, ids []types.TestCaseID) error {
	tx = r.getTx(tx)

	return tx.Delete(&models.TestCase{}, ids).Error
}

func (r *TestCase) Create(tx *gorm.DB, testCases *[]models.TestCase) error {
	tx = r.getTx(tx)

	return tx.Create(testCases).Error
}
