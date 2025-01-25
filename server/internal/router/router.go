package router

import (
	"github.com/gorilla/mux"

	"github.com/arpansaha13/pariksha/internal/controllers"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// Auth Routes
	r.HandleFunc("/auth/login", controllers.Login).Methods("POST")
	r.HandleFunc("/auth/signup", controllers.SignUp).Methods("POST")
	r.HandleFunc("/auth/verification/{hash}", controllers.Verification).Methods("POST")

	return r
}
