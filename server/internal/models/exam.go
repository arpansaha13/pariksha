package models

import (
	"time"
)

type Exam struct {
	ID                 int `gorm:"primaryKey"`
	Title              string
	StartsAt           time.Time
	EndsAt             time.Time
	CreatedBy          int
	Type               string `gorm:"type:varchar(10)"`
	MaxCandidatesCount int    `gorm:"not null"`
	PaperID            int
	User               User              `gorm:"foreignKey:CreatedBy"`
	Paper              Paper             `gorm:"foreignKey:PaperID"`
	Participants       []ExamParticipant `gorm:"foreignKey:ExamID"`
}

func (Exam) TableName() string {
	return "exams"
}
