package db

import (
	"fmt"

	"gorm.io/gorm"

	"pariksha/common/pkg/config"
	"pariksha/common/pkg/models"
)

var (
	DB            *gorm.DB
	dbInitializer config.DBInitializer = &config.PostgresInitializer{}
)

func Init(gormDsn config.GormDsn, gormConfig *gorm.Config, migrator config.DBMigrator) error {
	var err error
	DB, err = dbInitializer.Init(gormDsn, gormConfig, migrator)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	return nil
}

func Close() {
	if DB != nil {
		sqlDb, _ := DB.DB()
		sqlDb.Close()
	}
}

// ____________________DB MIGRATOR IMPLEMENTATIONS___________________

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
