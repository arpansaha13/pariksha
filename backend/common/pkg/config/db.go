package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ___________________________DB MIGRATOR____________________________

// DBMigrator interface for handling database migrations
type DBMigrator interface {
	Migrate(db *gorm.DB) error
}

// __________________________DB INITIALIZER__________________________

// DBInitializer provides a common interface for database initialization
type DBInitializer interface {
	Init(dsnConfig GormDsn, gormConfig *gorm.Config, migrator DBMigrator) (*gorm.DB, error)
}

// PostgresInitializer implements DBInitializer for PostgreSQL
type PostgresInitializer struct{}

// Init initializes a PostgreSQL database connection
func (p *PostgresInitializer) Init(dsnConfig GormDsn, gormConfig *gorm.Config, migrator DBMigrator) (*gorm.DB, error) {
	dsn, err := dsnConfig.Get()
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	if migrator != nil {
		if err := migrator.Migrate(db); err != nil {
			return nil, err
		}
	}

	return db, nil
}
