package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/router"
	"github.com/arpansaha13/pariksha/internal/utils"
)

func main() {
	loadEnv()
	validateEnv()

	db.Init()

	// Ensure the database connection is closed on application exit
	sqlDb, _ := db.DB.DB()
	defer sqlDb.Close()

	r := router.SetupRouter()

	// Remove hostname in production
	// https://stackoverflow.com/questions/55201561/golang-run-on-windows-without-deal-with-the-firewall/65393403#65393403
	port := utils.GetEnvWithDefault("API_PORT", constants.DEFAULT_API_PORT)
	addr := "localhost:" + port
	fmt.Printf("Server starting on %s\n", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func loadEnv() {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func validateEnv() {
	requiredEnvVars := []string{
		"GO_ENV",
		"DB_HOST",
		"DB_USER",
		"DB_PASS",
		"DB_NAME",
		"JWT_SECRET_KEY",
		"SMTP_NAME",
		"SMTP_USER",
		"SMTP_FROM",
		"SMTP_PASSWORD",
		"SMTP_HOST",
		"SMTP_PORT",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s is not set", envVar)
		}
	}
}
