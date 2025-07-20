package db

import (
	"gorm.io/gorm"

	"pariksha/paper/internal/models"
)

type AutoMigrator struct{}

func (m *AutoMigrator) Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Paper{},
		&models.PaperQuestion{},
		&models.PaperCategory{},
		&models.PaperPermission{},
	)
}
