package models

import "database/sql"

type Answer struct {
	ID                int64 `gorm:"primaryKey"`
	ExamParticipantID int64
	QuestionID        int64
	Answer            sql.NullString  `gorm:"type:text"`
	ScoreAwarded      int             `gorm:"default:0;not null"`
	Comments          sql.NullString  `gorm:"type:text"`
	Evaluated         bool            `gorm:"default:false;not null"`
	ExamParticipant   ExamParticipant `gorm:"foreignKey:ExamParticipantID"`
}

func (Answer) TableName() string {
	return "answers"
}
