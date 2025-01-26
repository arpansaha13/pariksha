package repositories

import (
	"sync"

	"gorm.io/gorm"

	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/models"
)

type SessionRepository interface {
	Create(session *models.Session) error
	FindByKey(key string) (*models.Session, error)
}

type sessionRepository struct {
	db *gorm.DB
}

var (
	sessionRepoInstance *sessionRepository
	sessionOnce         sync.Once
)

func GetSessionRepository() SessionRepository {
	sessionOnce.Do(func() {
		sessionRepoInstance = &sessionRepository{db: db.DB}
	})

	return sessionRepoInstance
}

func (r *sessionRepository) Create(session *models.Session) error {
	return r.db.Create(session).Error
}

func (r *sessionRepository) FindByKey(key string) (*models.Session, error) {
	var session models.Session
	if err := r.db.Where("key = ?", key).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}
