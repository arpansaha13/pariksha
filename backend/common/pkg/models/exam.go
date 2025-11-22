package models

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/proto"
	"pariksha/common/pkg/types"
)

type ParticipantCount struct {
	Unattended int16 `json:"unattended"`
	Invited    int16 `json:"invited"`
	Started    int16 `json:"started"`
	Ended      int16 `json:"ended"`
}

type Exam struct {
	ID                 types.ExamID `gorm:"primaryKey;autoIncrement;type:bigint"`
	Title              string
	StartsAt           time.Time       `gorm:"column:starts_at;not null"`
	EndsAt             time.Time       `gorm:"column:ends_at;not null"`
	CreatedBy          types.UserID    `gorm:"type:bigint"`
	Type               proto.ExamType  `gorm:"type:smallint;default:1"`
	MaxCandidatesCount int32           `gorm:"not null"`
	MaxScore           int32           `gorm:"type:integer;default:0"`
	DurationMinutes    int16           `gorm:"type:smallint;not null;check:duration_minutes >= 0 AND duration_minutes <= 1440"`
	PaperHash          string          `gorm:"type:varchar(64)"`
	ParticipantCounts  json.RawMessage `gorm:"type:jsonb;default:'{\"unattended\":0,\"invited\":0,\"started\":0,\"ended\":0}'"`
	Hash               string          `gorm:"type:varchar(64);uniqueIndex;not null"`
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
