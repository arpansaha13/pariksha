package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/arpansaha13/pariksha/internal/db"
)

func main() {
	loadEnv()
	validateEnv()

	db.Init()

	// Ensure the database connection is closed on application exit
	defer db.DB.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		fmt.Fprintf(w, "Hello, World!")
	})

	// Remove hostname in production
	// https://stackoverflow.com/questions/55201561/golang-run-on-windows-without-deal-with-the-firewall/65393403#65393403
	addr := "localhost:4000"
	fmt.Printf("Server starting on %s\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
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
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s is not set", envVar)
		}
	}
}
