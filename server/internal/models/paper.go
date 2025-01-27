package models

type Paper struct {
	ID             int            `gorm:"primaryKey"`
	Title          string         `gorm:"type:varchar(255);not null"`
	MaxScore       int            `gorm:"default:0"`
	Questions      []Question     `gorm:"foreignKey:PaperID"`
	PaperOwnership PaperOwnership `gorm:"foreignKey:PaperID"`
}

func (Paper) TableName() string {
	return "papers"
}
