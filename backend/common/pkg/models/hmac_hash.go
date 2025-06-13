package models

import (
	"pariksha/common/pkg/constants"
	"pariksha/common/pkg/types"
)

type ExamHash struct {
	ID   types.ExamID `gorm:"primaryKey;type:bigint"`
	Hash string       `gorm:"type:varchar(64);uniqueIndex;not null"`
}

type QuestionHash struct {
	ID   types.QuestionID `gorm:"primaryKey;type:bigint"`
	Hash string           `gorm:"type:varchar(64);uniqueIndex;not null"`
}

type PaperHash struct {
	ID   types.PaperID `gorm:"primaryKey;type:bigint"`
	Hash string        `gorm:"type:varchar(64);uniqueIndex;not null"`
}

func (ExamHash) TableName() string {
	return constants.TABLE_EXAM_HASHES
}

func (QuestionHash) TableName() string {
	return constants.TABLE_QUESTION_HASHES
}

func (PaperHash) TableName() string {
	return constants.TABLE_PAPER_HASHES
}
