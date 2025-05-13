package models

import (
	"database/sql"

	"pariksha/common/pkg/constants"
)

type ExamParticipant struct {
	ID               int64 `gorm:"primaryKey;type:bigint"`
	ExamID           int64 `gorm:"type:bigint"`
	UserID           int64 `gorm:"type:bigint"`
	ScoreAwarded     int32
	Status           int16 `gorm:"type:smallint;default:1;not null"` // 1 = INVITED
	StartedAt        sql.NullTime
	EndedAt          sql.NullTime
	ScheduledEndTime sql.NullTime
	Exam             Exam     `gorm:"foreignKey:ExamID"`
	Answers          []Answer `gorm:"foreignKey:ExamParticipantID"`
}

func (ExamParticipant) TableName() string {
	return constants.TABLE_EXAM_PARTICIPANTS
}
