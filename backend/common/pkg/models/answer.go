package models

import "database/sql"

type Answer struct {
	ID                int `gorm:"primaryKey"`
	ExamParticipantID int
	QuestionID        int
	Answer            sql.NullString  `gorm:"type:text"`
	ScoreAwarded      int             `gorm:"default:0;not null"`
	Comments          sql.NullString  `gorm:"type:text"`
	Evaluated         bool            `gorm:"default:false;not null"`
	ExamParticipant   ExamParticipant `gorm:"foreignKey:ExamParticipantID"`
}

func (Answer) TableName() string {
	return "answers"
}
