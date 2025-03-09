package models

type QuestionCategory struct {
	ID      int    `gorm:"primaryKey"`
	PaperID int    `gorm:"not null"`
	Name    string `gorm:"type:varchar(255);not null"`
	Order   int    `gorm:"not null"`
	Paper   Paper  `gorm:"foreignKey:PaperID"`
}

func (QuestionCategory) TableName() string {
	return "question_categories"
}
