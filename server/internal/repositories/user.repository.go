package repositories

import (
	"sync"

	"gorm.io/gorm"

	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	FindOne(id int) (*models.User, error)
	FindOneByEmail(email string) (*models.User, error)
	Delete(id int) error
}

type userRepository struct {
	db *gorm.DB
}

var (
	userRepoInstance *userRepository
	userOnce         sync.Once
)

func GetUserRepository() UserRepository {
	userOnce.Do(func() {
		userRepoInstance = &userRepository{db: db.DB}
	})

	return userRepoInstance
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindOne(id int) (*models.User, error) {
	var user models.User
	result := r.db.First(&user, id)

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *userRepository) FindOneByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (r *userRepository) Delete(id int) error {
	return r.db.Delete(&models.User{}, id).Error
}
