package db

import (
	"gorm.io/gorm"

	"pariksha/question/internal/models"
)

type AutoMigrator struct{}

func (m *AutoMigrator) Migrate(dbInst *gorm.DB) error {
	return dbInst.AutoMigrate(
		&models.Question{},
		&models.Category{},
		&models.Boilerplate{},
		&models.Language{},
		&models.TestCase{},
	)
}
