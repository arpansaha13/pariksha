package repository

import (
	"database/sql"

	"gorm.io/gorm"

	"pariksha/common/pkg/types"
	"pariksha/paper/internal/domain"
)

// IPaperRepository defines the interface for paper repository operations.
type IPaperRepository interface {
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error
	GetAllByUserId(tx *gorm.DB, userID types.UserID) ([]domain.Paper, error)
	Create(tx *gorm.DB, paper *domain.Paper, userID types.UserID) error
	UpdateHash(tx *gorm.DB, paper *domain.Paper, hash string) error
	GetByHash(tx *gorm.DB, hash string) (*domain.Paper, error)
	Update(tx *gorm.DB, paper *domain.Paper) error
	GetIDsByHashes(tx *gorm.DB, hashes []string) ([]types.PaperID, error)
	BulkDelete(tx *gorm.DB, paperIDs []types.PaperID) error
}

// IPaperCategoryRepository defines the interface for paper category operations.
type IPaperCategoryRepository interface {
	Create(tx *gorm.DB, paperCat *domain.PaperCategory) error
	DeleteByID(tx *gorm.DB, paperID types.PaperID, categoryID types.CategoryID) error
	BulkDeleteByPaperIDs(tx *gorm.DB, paperIDs []types.PaperID) error
	GetAllByPaperHash(tx *gorm.DB, paperHash string) ([]domain.PaperCategory, error)
	GetByPaperHashAndCategoryID(tx *gorm.DB, paperHash string, categoryID types.CategoryID) (*domain.PaperCategory, error)
	GetCountByPaperHash(tx *gorm.DB, paperHash string) (int64, error)
	UpdateOrder(tx *gorm.DB, categoryID int64, order int16) error
	GetMaxOrder(tx *gorm.DB, paperID types.PaperID) (int16, error)
	GetCountByPaperId(tx *gorm.DB, paperID types.PaperID) (int64, error)
}

// IPaperPermissionRepository defines the interface for paper permission operations.
type IPaperPermissionRepository interface {
	Create(tx *gorm.DB, paperID types.PaperID, userID types.UserID) error
	GetByPaperHashesAndUserId(tx *gorm.DB, paperHashes []string, userID types.UserID) ([]domain.PaperPermission, error)
	GetByPaperHashAndUserId(tx *gorm.DB, paperHash string, userID types.UserID) (*domain.PaperPermission, error)
	BulkDeleteByPaperIDs(tx *gorm.DB, paperIDs []types.PaperID) error
}

// IPaperQuestionRepository defines the interface for paper question operations.
type IPaperQuestionRepository interface {
	Create(tx *gorm.DB, paperQuest *domain.PaperQuestion) error
	DeleteByID(tx *gorm.DB, questionID types.QuestionID) error
	BulkDeleteByPaperIDs(tx *gorm.DB, paperIDs []types.PaperID) error
	BulkDeleteByPaperIDAndCategoryID(tx *gorm.DB, paperID types.PaperID, categoryID types.CategoryID) error
	GetAllByPaperHash(tx *gorm.DB, paperHash string) ([]domain.PaperQuestion, error)
	GetByPaperHashAndQuestionID(tx *gorm.DB, paperHash string, questionID types.QuestionID) (*domain.PaperQuestion, error)
	GetMaxQuestionOrder(tx *gorm.DB, categoryID int64) (int16, error)
	ValidateCategoryQuestions(tx *gorm.DB, categoryID int64, questionIDs []types.QuestionID) (int64, error)
	UpdateOrder(tx *gorm.DB, questionID types.QuestionID, order int16) error
	Save(tx *gorm.DB, paperQuest *domain.PaperQuestion) error
	GetAllByCategoryID(tx *gorm.DB, categoryID types.CategoryID) ([]domain.PaperQuestion, error)
}
