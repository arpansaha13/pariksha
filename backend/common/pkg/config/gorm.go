package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ______________________GORM CONNECTION STRING______________________

type GormDsn interface {
	Get() (string, error)
}

type GormDsnImpl struct {
	Addr     string
	User     string
	Password string
	Dbname   string
	Sslmode  string
}

func (gd *GormDsnImpl) missingFields() []string {
	missing := []string{}
	fieldMap := map[string]string{
		"Addr":     gd.Addr,
		"User":     gd.User,
		"Password": gd.Password,
		"Dbname":   gd.Dbname,
		"Sslmode":  gd.Sslmode,
	}

	for k, v := range fieldMap {
		if strings.TrimSpace(v) == "" {
			missing = append(missing, k)
		}
	}
	return missing
}

func (gd *GormDsnImpl) Get() (string, error) {
	missing := gd.missingFields()
	if len(missing) > 0 {
		return "", fmt.Errorf("missing required fields: %v", strings.Join(missing, ", "))
	}

	addr := strings.Split(gd.Addr, ":")
	if len(addr) != 2 {
		return "", fmt.Errorf("addr must be in format host:port")
	}
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		addr[0], gd.User, gd.Password, gd.Dbname, addr[1], gd.Sslmode,
	), nil
}

// ___________________________GORM CONFIG____________________________

func GetTestEnvGormConfig() *gorm.Config {
	return &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
}

func GetDevEnvGormConfig() *gorm.Config {
	return &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
			},
		),
	}
}

func GetDefaultGormConfig() *gorm.Config {
	return &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	}

}
