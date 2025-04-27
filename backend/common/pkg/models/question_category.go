package models

import "database/sql"

type QuestionCategory struct {
	ID      int64         `gorm:"primaryKey;type:bigint"`
	PaperID sql.NullInt64 `gorm:"type:bigint"`
	Name    string        `gorm:"type:varchar(255);not null"`
	Order   int16         `gorm:"type:smallint;not null"`
	// Indicates if the category is used in an exam and is locked for editing
	Locked bool  `gorm:"not null;default:false"`
	Paper  Paper `gorm:"foreignKey:PaperID"`
}

func (QuestionCategory) TableName() string {
	return "question_categories"
}
