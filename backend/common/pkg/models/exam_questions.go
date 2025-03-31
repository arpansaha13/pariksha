package models

type ExamQuestion struct {
	ID         int  `gorm:"primaryKey"`
	ExamID     int  `gorm:"not null"`
	QuestionID int  `gorm:"not null"`
	Exam       Exam `gorm:"foreignKey:ExamID"`
}

func (ExamQuestion) TableName() string {
	return "exam_questions"
}
