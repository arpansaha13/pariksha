package models

type ExamCategory struct {
	ID         int  `gorm:"primaryKey"`
	ExamID     int  `gorm:"not null"`
	CategoryID int  `gorm:"not null"`
	Exam       Exam `gorm:"foreignKey:ExamID"`
}

func (ExamCategory) TableName() string {
	return "exam_categories"
}
