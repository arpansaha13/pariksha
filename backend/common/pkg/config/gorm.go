package config

import (
	"log"
	"os"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"pariksha/common/pkg/constants"
)

// GormLogger returns a GORM logger config based on environment
func GormLogger(env string) *gorm.Config {
	config := &gorm.Config{}

	switch env {
	case constants.GO_ENV_TEST:
		config.Logger = logger.Default.LogMode(logger.Silent)
	case constants.GO_ENV_DEV, constants.GO_ENV_DOCKER_DEV:
		config.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
			},
		)
	default:
		config.Logger = logger.Default.LogMode(logger.Error)
	}

	return config
}
