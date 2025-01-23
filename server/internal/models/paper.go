package models

type Paper struct {
	ID        int `gorm:"primaryKey"`
	Title     string
	MaxScore  int
	Questions []Question `gorm:"foreignKey:PaperID"`
}

func (Paper) TableName() string {
	return "papers"
}
