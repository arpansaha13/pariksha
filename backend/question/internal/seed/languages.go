package seed

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"pariksha/common/pkg/constants"
	"pariksha/question/internal/models"
)

func Languages(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.Language{}) {
		return fmt.Errorf("table %s not found", constants.TABLE_LANGUAGES)
	}

	languages := []models.Language{
		{
			Slug:      "node",
			Name:      "JavaScript (Node)",
			Extension: "js",
			Version:   "22",
		},
	}

	for _, lang := range languages {
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(&lang)
	}

	return nil
}
