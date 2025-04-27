package models

type ExamCategory struct {
	ID         int64 `gorm:"primaryKey;type:bigint"`
	ExamID     int64 `gorm:"type:bigint;not null"`
	CategoryID int64 `gorm:"type:bigint;not null"`
	Exam       Exam  `gorm:"foreignKey:ExamID"`
}

func (ExamCategory) TableName() string {
	return "exam_categories"
}
