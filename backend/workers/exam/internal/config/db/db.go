package db

import (
	"fmt"

	"gorm.io/gorm"

	"pariksha/common/pkg/config"
)

var (
	Exams         *gorm.DB
	dbInitializer config.DBInitializer = &config.PostgresInitializer{}
)

func Init(gormDsn config.GormDsn, gormConfig *gorm.Config) error {
	var err error
	Exams, err = dbInitializer.Init(gormDsn, gormConfig, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	return nil
}

func Close() {
	if Exams != nil {
		sqlDb, _ := Exams.DB()
		sqlDb.Close()
	}
}
