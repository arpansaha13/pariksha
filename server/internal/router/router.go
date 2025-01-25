package router

import (
	"github.com/gorilla/mux"

	"github.com/arpansaha13/pariksha/internal/handlers"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// Auth Routes
	r.HandleFunc("/auth/login", handlers.Login).Methods("POST")
	r.HandleFunc("/auth/signup", handlers.SignUp).Methods("POST")
	r.HandleFunc("/auth/verification/{hash}", handlers.Verification).Methods("POST")

	return r
}
