package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arpansaha13/pariksha/internal/config/db"
	"github.com/arpansaha13/pariksha/internal/config/env"
	"github.com/arpansaha13/pariksha/internal/router"
)

func main() {
	r := router.SetupRouter()

	port := env.API_PORT
	fmt.Printf("Server starting on localhost:%s\n", port)

	addr := ":" + port
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}

	defer closeConnections()
}

// Ensure the connections are closed on application exit
func closeConnections() {
	sqlDb, _ := db.DB.DB()
	sqlDb.Close()
}
