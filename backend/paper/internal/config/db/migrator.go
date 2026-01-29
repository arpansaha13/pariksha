package db

import (
	"gorm.io/gorm"

	"pariksha/paper/internal/domain"
)

type AutoMigrator struct{}

func (m *AutoMigrator) Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.Paper{},
		&domain.PaperQuestion{},
		&domain.PaperCategory{},
		&domain.PaperPermission{},
	)
}
