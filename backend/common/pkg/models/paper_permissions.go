package models

import (
	"gorm.io/gorm"

	"pariksha/common/pkg/constants"
)

type PaperPermissions struct {
	PaperID     int64          `gorm:"primaryKey;type:bigint;not null"`
	UserID      int64          `gorm:"primaryKey;type:bigint;not null"`
	Permissions int16          `gorm:"type:smallint;not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (PaperPermissions) TableName() string {
	return constants.TABLE_PAPER_PERMISSIONS
}

const (
	readBitShift  = 1 // Left-shift amount for the READ bit
	writeBitShift = 0 // Left-shift amount for the WRITE bit
)

// CanRead checks if the first bit of the `permissions` is set to 1.
func (p *PaperPermissions) CanRead() bool {
	return (p.Permissions & (1 << readBitShift)) != 0
}

// CanWrite checks if the second bit of the `permissions` is set to 1.
func (p *PaperPermissions) CanWrite() bool {
	return (p.Permissions & (1 << writeBitShift)) != 0
}

// SetRead sets the READ bit in the `permissions`.
func (p *PaperPermissions) SetRead() {
	p.Permissions |= (1 << readBitShift)
}

// SetWrite sets the READ and WRITE bits in the `permissions`.
func (p *PaperPermissions) SetWrite() {
	p.Permissions |= (1 << writeBitShift)
	p.Permissions |= (1 << readBitShift)
}

// UnsetRead unsets the READ and WRITE bits in the `permissions`.
func (p *PaperPermissions) UnsetRead() {
	p.Permissions &^= (1 << readBitShift)
	p.Permissions &^= (1 << writeBitShift)
}

// UnsetWrite unsets the WRITE bit in the `permissions`.
func (p *PaperPermissions) UnsetWrite() {
	p.Permissions &^= (1 << writeBitShift)
}

// &^ is the bit-clear operator in Go.
// It clears (sets to 0) the bits in the left-hand operand where the corresponding bits in the right-hand operand are 1.
// For example, if permissions = 0110 and (1 << readBitShift) = 0010,
// then permissions &^ (1 << readBitShift) results in 0100, effectively unsetting the READ bit.
// https://stackoverflow.com/questions/34459450/what-is-the-operator-in-golang
