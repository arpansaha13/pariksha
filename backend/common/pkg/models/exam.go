package models

import (
	"encoding/json"
	"errors"
	"time"
)

type ParticipantCount struct {
	Unattended int `json:"unattended"`
	Invited    int `json:"invited"`
	Started    int `json:"started"`
	Ended      int `json:"ended"`
}

type Exam struct {
	ID                 int64 `gorm:"primaryKey"`
	Title              string
	StartsAt           time.Time `gorm:"column:starts_at;not null"`
	EndsAt             time.Time `gorm:"column:ends_at;not null"`
	CreatedBy          int64
	Type               string `gorm:"type:varchar(16);default:LINK"`
	MaxCandidatesCount int32  `gorm:"not null"`
	DurationMinutes    int32  `gorm:"not null"`
	PaperID            int64
	ParticipantCounts  json.RawMessage   `gorm:"type:json;default:'{\"unattended\":0,\"invited\":0,\"started\":0,\"ended\":0}'"`
	Participants       []ExamParticipant `gorm:"foreignKey:ExamID"`
}

func (Exam) TableName() string {
	return "exams"
}

func (e *Exam) GetParticipantCounts() (ParticipantCount, error) {
	var counts ParticipantCount
	if err := json.Unmarshal(e.ParticipantCounts, &counts); err != nil {
		return counts, errors.New("failed to parse participant counts")
	}
	return counts, nil
}
