package router

import (
	"github.com/gorilla/mux"

	"github.com/arpansaha13/pariksha/internal/controllers"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// Auth Routes
	r.HandleFunc("/login", controllers.Login).Methods("POST")
	r.HandleFunc("/signup", controllers.SignUp).Methods("POST")

	return r
}
