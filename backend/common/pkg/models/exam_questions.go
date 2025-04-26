package models

type ExamQuestion struct {
	ID         int64 `gorm:"primaryKey"`
	ExamID     int64 `gorm:"not null"`
	QuestionID int64 `gorm:"not null"`
	CategoryID int64 `gorm:"not null"`
	Order      int   `gorm:"not null"`
	Exam       Exam  `gorm:"foreignKey:ExamID"`
}

func (ExamQuestion) TableName() string {
	return "exam_questions"
}
