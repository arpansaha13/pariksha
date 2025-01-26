package repositories

import (
	"sync"

	"gorm.io/gorm"

	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/models"
)

type PaperOwnershipRepository interface {
	Create(paperOwnership *models.PaperOwnership) error
}

type paperOwnershipRepository struct {
	db *gorm.DB
}

var (
	paperOwnershipRepoInstance *paperOwnershipRepository
	paperOwnershipOnce         sync.Once
)

func GetPaperOwnershipRepository() PaperOwnershipRepository {
	paperOwnershipOnce.Do(func() {
		paperOwnershipRepoInstance = &paperOwnershipRepository{db: db.DB}
	})

	return paperOwnershipRepoInstance
}

func (r *paperOwnershipRepository) Create(paperOwnership *models.PaperOwnership) error {
	return r.db.Create(paperOwnership).Error
}
