package models

type ExamCategory struct {
	ID         int64 `gorm:"primaryKey"`
	ExamID     int64 `gorm:"not null"`
	CategoryID int64 `gorm:"not null"`
	Exam       Exam  `gorm:"foreignKey:ExamID"`
}

func (ExamCategory) TableName() string {
	return "exam_categories"
}
