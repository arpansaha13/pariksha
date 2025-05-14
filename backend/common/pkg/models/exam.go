package models

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
)

type ParticipantCount struct {
	Unattended int16 `json:"unattended"`
	Invited    int16 `json:"invited"`
	Started    int16 `json:"started"`
	Ended      int16 `json:"ended"`
}

type Exam struct {
	ID                 int64 `gorm:"primaryKey;type:bigint"`
	Title              string
	StartsAt           time.Time       `gorm:"column:starts_at;not null"`
	EndsAt             time.Time       `gorm:"column:ends_at;not null"`
	CreatedBy          int64           `gorm:"type:bigint"`
	Type               string          `gorm:"type:varchar(16);default:LINK"`
	MaxCandidatesCount int32           `gorm:"not null"`
	MaxScore           int32           `gorm:"type:integer;default:0"`
	DurationMinutes    int16           `gorm:"type:smallint;not null;check:duration_minutes >= 0 AND duration_minutes <= 1440"`
	PaperID            int64           `gorm:"type:bigint"`
	ParticipantCounts  json.RawMessage `gorm:"type:json;default:'{\"unattended\":0,\"invited\":0,\"started\":0,\"ended\":0}'"`
	DeletedAt          gorm.DeletedAt  `gorm:"index"`

	Participants []ExamParticipant `gorm:"foreignKey:ExamID"`
}

func (Exam) TableName() string {
	return constants.TABLE_EXAMS
}

func (e *Exam) GetParticipantCounts() (ParticipantCount, error) {
	var counts ParticipantCount
	if err := json.Unmarshal(e.ParticipantCounts, &counts); err != nil {
		return counts, errors.New("failed to parse participant counts")
	}
	return counts, nil
}
