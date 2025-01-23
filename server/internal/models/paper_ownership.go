package models

type PaperOwnership struct {
	ID      int `gorm:"primaryKey"`
	UserID  int
	PaperID int
	Path    string
	Type    string `gorm:"type:varchar(10)"`
	User    User   `gorm:"foreignKey:UserID"`
	Paper   Paper  `gorm:"foreignKey:PaperID"`
}

func (PaperOwnership) TableName() string {
	return "paper_ownerships"
}
