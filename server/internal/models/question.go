package models

type Question struct {
	ID            int    `gorm:"primaryKey"`
	Question      string `gorm:"type:text"`
	Category      string
	Type          string   `gorm:"type:varchar(20)"`
	Tags          []string `gorm:"type:text[];default:'{}'"`
	PaperID       int
	MaxScore      int
	CorrectAnswer string   `gorm:"type:text"`
	Paper         Paper    `gorm:"foreignKey:PaperID"`
	Answers       []Answer `gorm:"foreignKey:QuestionID"`
}

func (Question) TableName() string {
	return "questions"
}
