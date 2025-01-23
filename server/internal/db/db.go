package db

import (
	"fmt"
	"log"
	"os"

	"database/sql"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *sql.DB

/* Make sure env variables are loaded before initializing db */
func Init() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	DB, _ = db.DB()

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
