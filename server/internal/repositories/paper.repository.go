package repositories

import (
	"sync"

	"gorm.io/gorm"

	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/models"
)

type PaperRepository interface {
	Create(paper *models.Paper) error
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
