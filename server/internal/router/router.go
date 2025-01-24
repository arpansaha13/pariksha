package router

import (
	"github.com/gorilla/mux"

	"github.com/arpansaha13/pariksha/internal/controllers"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", controllers.GetUsers).Methods("GET")

	return r
}
