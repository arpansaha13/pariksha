package models

import "database/sql"

type ExamParticipant struct {
	ID           int `gorm:"primaryKey"`
	ExamID       int
	UserID       int
	ScoreAwarded int
	StartedAt    sql.NullTime
	EndedAt      sql.NullTime
	Exam         Exam     `gorm:"foreignKey:ExamID"`
	User         User     `gorm:"foreignKey:UserID"`
	Answers      []Answer `gorm:"foreignKey:ExamParticipantID"`
}

func (ExamParticipant) TableName() string {
	return "exam_participants"
}
