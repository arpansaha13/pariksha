package main

import (
	"fmt"
	"log"
	"net/http"

	"pariksha/server/internal/config/env"
	"pariksha/server/internal/router"
)

func main() {
	r := router.SetupRouter()

	port := env.API_PORT
	fmt.Printf("Server starting on localhost:%s\n", port)

	addr := ":" + port
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
