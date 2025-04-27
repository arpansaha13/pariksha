package models

type ExamQuestion struct {
	ID         int64 `gorm:"primaryKey;type:bigint"`
	ExamID     int64 `gorm:"type:bigint;not null"`
	QuestionID int64 `gorm:"type:bigint;not null"`
	CategoryID int64 `gorm:"type:bigint;not null"`
	Order      int16 `gorm:"type:smallint;not null"`
	Exam       Exam  `gorm:"foreignKey:ExamID"`
}

func (ExamQuestion) TableName() string {
	return "exam_questions"
}
