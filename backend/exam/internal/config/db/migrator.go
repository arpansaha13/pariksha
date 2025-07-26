package db

import (
	"gorm.io/gorm"

	"pariksha/common/pkg/models"
)

type AutoMigrator struct{}

func (m *AutoMigrator) Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Exam{},
		&models.Answer{},
		&models.ExamParticipant{},
		&models.ExamPermission{},
		&models.ExamQuestion{},
		&models.ExamCategory{},
	)
}
