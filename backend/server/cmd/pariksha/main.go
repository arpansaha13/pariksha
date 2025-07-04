package main

import (
	"fmt"
	"log"
	"net/http"

	"pariksha/server/internal/config/env"
	"pariksha/server/internal/interservice"
	"pariksha/server/internal/router"
)

func main() {
	r := router.SetupRouter()

	port := env.API_PORT
	fmt.Printf("Server starting on localhost:%s\n", port)

	defer interservice.CloseAuthConn()

	addr := ":" + port
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
