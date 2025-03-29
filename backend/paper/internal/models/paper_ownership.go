package models

type PaperOwnership struct {
	ID      int `gorm:"primaryKey"`
	UserID  int
	PaperID int
	Path    string
	Type    string `gorm:"type:varchar(10);not null;check:type IN ('OWNER', 'SHARED')"`
}

func (PaperOwnership) TableName() string {
	return "paper_ownerships"
}
