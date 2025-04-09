package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int64          `gorm:"primaryKey"`
	Username  string         `gorm:"type:varchar(255);not null;unique"`
	Email     string         `gorm:"type:varchar(255);not null;unique"`
	Password  sql.NullString `gorm:"type:varchar(255)"`
	FirstName sql.NullString `gorm:"column:first_name;type:varchar(255)"`
	LastName  sql.NullString `gorm:"column:last_name;type:varchar(255)"`
	Verified  bool           `gorm:"default:false;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Exams     []Exam `gorm:"foreignKey:CreatedBy"`
}

func (User) TableName() string {
	return "users"
}
