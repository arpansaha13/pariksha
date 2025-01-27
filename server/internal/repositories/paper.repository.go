package repositories

import (
	"sync"

	"gorm.io/gorm"

	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/models"
)

type PaperRepository interface {
	Create(paper *models.Paper) error
	FindByUserID(userID int) ([]models.Paper, error)
}

type paperRepository struct {
	db *gorm.DB
}

var (
	paperRepoInstance *paperRepository
	paperOnce         sync.Once
)

func GetPaperRepository() PaperRepository {
	paperOnce.Do(func() {
		paperRepoInstance = &paperRepository{db: db.DB}
	})

	return paperRepoInstance
}

func (r *paperRepository) Create(paper *models.Paper) error {
	return r.db.Create(paper).Error
}

func (r *paperRepository) FindByUserID(userID int) ([]models.Paper, error) {
	var papers []models.Paper
	err := r.db.Model(&models.Paper{}).Preload("PaperOwnership").Find(&papers).Where("paper_ownerships.user_id = ?", userID).Error
	return papers, err
}
