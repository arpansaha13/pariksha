package models

import "pariksha/common/pkg/constants"

type ExamPermission struct {
	ExamID      int64 `gorm:"primaryKey;type:bigint;not null"`
	UserID      int64 `gorm:"primaryKey;type:bigint;not null"`
	Permissions int16 `gorm:"type:smallint;not null"`
}

func (ExamPermission) TableName() string {
	return constants.TABLE_EXAM_PERMISSIONS
}

const (
	examReadBitShift        = 3 // Left-shift amount for the READ bit
	examWriteBitShift       = 2 // Left-shift amount for the WRITE bit
	examParticipateBitShift = 1 // Left-shift amount for the PARTICIPATE bit
	examEvaluateBitShift    = 0 // Left-shift amount for the EVALUATE bit
)

// CanRead checks if the READ bit is set
func (p *ExamPermission) CanRead() bool {
	return (p.Permissions & (1 << examReadBitShift)) != 0
}

// CanWrite checks if the WRITE bit is set
func (p *ExamPermission) CanWrite() bool {
	return (p.Permissions & (1 << examWriteBitShift)) != 0
}

// CanParticipate checks if the PARTICIPATE bit is set
func (p *ExamPermission) CanParticipate() bool {
	return (p.Permissions & (1 << examParticipateBitShift)) != 0
}

// CanEvaluate checks if the EVALUATE bit is set
func (p *ExamPermission) CanEvaluate() bool {
	return (p.Permissions & (1 << examEvaluateBitShift)) != 0
}

// SetRead sets the READ bit
func (p *ExamPermission) SetRead() {
	p.Permissions |= (1 << examReadBitShift)
}

// SetWrite sets the WRITE and READ bits
func (p *ExamPermission) SetWrite() {
	p.Permissions |= (1 << examWriteBitShift)
	p.Permissions |= (1 << examReadBitShift)
}

// SetParticipate sets the PARTICIPATE and READ bits
func (p *ExamPermission) SetParticipate() {
	p.Permissions |= (1 << examParticipateBitShift)
	p.Permissions |= (1 << examReadBitShift)
}

// SetEvaluate sets the EVALUATE and READ bits
func (p *ExamPermission) SetEvaluate() {
	p.Permissions |= (1 << examEvaluateBitShift)
	p.Permissions |= (1 << examReadBitShift)
}

// UnsetRead unsets READ, WRITE and PARTICIPATE bits
func (p *ExamPermission) UnsetRead() {
	p.Permissions &^= (1 << examReadBitShift)
	p.Permissions &^= (1 << examWriteBitShift)
	p.Permissions &^= (1 << examParticipateBitShift)
}

// UnsetWrite unsets WRITE bit
func (p *ExamPermission) UnsetWrite() {
	p.Permissions &^= (1 << examWriteBitShift)
}

// UnsetParticipate unsets PARTICIPANT bit
func (p *ExamPermission) UnsetParticipate() {
	p.Permissions &^= (1 << examParticipateBitShift)
}

// UnsetEvaluate unsets EVALUATE bit
func (p *ExamPermission) UnsetEvaluate() {
	p.Permissions &^= (1 << examEvaluateBitShift)
}
