package models

import (
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
)

type PaperCategory struct {
	PaperID    types.PaperID    `gorm:"primaryKey;type:bigint;not null;constraint:OnDelete:CASCADE;references:id"`
	CategoryID types.CategoryID `gorm:"primaryKey;type:bigint;not null;constraint:OnDelete:CASCADE;references:id"`
	Order      int16            `gorm:"type:smallint;not null"`
}

func (PaperCategory) TableName() string {
	return constants.TABLE_PAPER_CATEGORIES
}
