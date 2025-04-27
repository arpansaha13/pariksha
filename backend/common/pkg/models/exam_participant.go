package models

import "database/sql"

type ExamParticipant struct {
	ID               int64 `gorm:"primaryKey;type:bigint"`
	ExamID           int64 `gorm:"type:bigint"`
	UserID           int64 `gorm:"type:bigint"`
	ScoreAwarded     int
	Status           int `gorm:"default:1;not null"` // 1 = INVITED
	StartedAt        sql.NullTime
	EndedAt          sql.NullTime
	ScheduledEndTime sql.NullTime
	Exam             Exam     `gorm:"foreignKey:ExamID"`
	Answers          []Answer `gorm:"foreignKey:ExamParticipantID"`
}

func (ExamParticipant) TableName() string {
	return "exam_participants"
}
