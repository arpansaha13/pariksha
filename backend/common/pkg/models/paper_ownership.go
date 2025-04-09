package models

type PaperOwnership struct {
	ID      int64 `gorm:"primaryKey"`
	UserID  int64
	PaperID int64
	Path    string
	Type    string `gorm:"type:varchar(10);not null;check:type IN ('OWNER', 'SHARED')"`
}

func (PaperOwnership) TableName() string {
	return "paper_ownerships"
}
