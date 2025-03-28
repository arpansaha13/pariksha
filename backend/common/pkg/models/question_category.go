package models

import "database/sql"

type QuestionCategory struct {
	ID      int `gorm:"primaryKey"`
	PaperID sql.NullInt64
	Name    string `gorm:"type:varchar(255);not null"`
	Order   int    `gorm:"not null"`
	Locked  bool   `gorm:"not null;default:false"`
	Paper   Paper  `gorm:"foreignKey:PaperID"`
}

func (QuestionCategory) TableName() string {
	return "question_categories"
}
