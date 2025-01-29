package models

import (
	"time"
)

type User struct {
	ID        int    `gorm:"primaryKey"`
	Username  string `gorm:"type:varchar(255);not null;unique"`
	Email     string `gorm:"type:varchar(255);not null;unique"`
	Password  string `gorm:"type:varchar(255)"`
	FirstName string `gorm:"column:first_name;type:varchar(255)"`
	LastName  string `gorm:"column:last_name;type:varchar(255)"`
	IsGuest   bool   `gorm:"default:false;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Exams     []Exam `gorm:"foreignKey:CreatedBy"`
}

func (User) TableName() string {
	return "users"
}
