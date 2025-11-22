package models

import (
	"database/sql"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
)

type ExamParticipant struct {
	ID               types.ParticipantID `gorm:"primaryKey;autoIncrement;type:bigint"`
	ExamID           types.ExamID        `gorm:"type:bigint"`
	UserID           types.UserID        `gorm:"type:bigint"`
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
