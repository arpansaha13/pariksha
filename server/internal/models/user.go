package models

import (
	"time"
)

type User struct {
	ID        int `gorm:"primaryKey"`
	Username  string
	Email     string
	Password  string
	FirstName string `gorm:"column:first_name"`
	LastName  string `gorm:"column:last_name"`
	Role      string `gorm:"type:varchar(10)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Exams     []Exam `gorm:"foreignKey:CreatedBy"`
}

func (User) TableName() string {
	return "users"
}
