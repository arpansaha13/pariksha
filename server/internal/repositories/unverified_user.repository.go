package repositories

import (
	"sync"

	"gorm.io/gorm"

	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/models"
)

type UnverifiedUserRepository interface {
	Create(user *models.UnverifiedUser) error
	Update(user *models.UnverifiedUser) error
	FindOne(hash string) (*models.UnverifiedUser, error)
	FindOneByEmail(email string) (*models.UnverifiedUser, error)
	DeleteByPointer(*models.UnverifiedUser) error
}

type unverifiedUserRepository struct {
	db *gorm.DB
}

var (
	unverifiedUserRepoInstance *unverifiedUserRepository
	once                       sync.Once
)

func GetUnverifiedUserRepository() UnverifiedUserRepository {
	once.Do(func() {
		unverifiedUserRepoInstance = &unverifiedUserRepository{db: db.DB}
	})
	return unverifiedUserRepoInstance
}

func (r *unverifiedUserRepository) Create(user *models.UnverifiedUser) error {
	return r.db.Create(user).Error
}

func (r *unverifiedUserRepository) Update(user *models.UnverifiedUser) error {
	return r.db.Save(user).Error
}

func (r *unverifiedUserRepository) FindOne(hash string) (*models.UnverifiedUser, error) {
	var user models.UnverifiedUser
	result := r.db.Where("hash = ?", hash).First(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *unverifiedUserRepository) FindOneByEmail(email string) (*models.UnverifiedUser, error) {
	var user models.UnverifiedUser
	result := r.db.Where("email = ?", email).First(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *unverifiedUserRepository) DeleteByPointer(user *models.UnverifiedUser) error {
	return r.db.Delete(user).Error
}
