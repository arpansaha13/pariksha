package models

import "database/sql"

type Answer struct {
	ID                int `gorm:"primaryKey"`
	ExamParticipantID int
	QuestionID        int
	Answer            sql.NullString  `gorm:"type:text"`
	ScoreAwarded      int             `gorm:"default:0;not null"`
	Comments          sql.NullString  `gorm:"type:text"`
	ExamParticipant   ExamParticipant `gorm:"foreignKey:ExamParticipantID"`
	Question          Question        `gorm:"foreignKey:QuestionID"`
}

func (Answer) TableName() string {
	return "answers"
}
