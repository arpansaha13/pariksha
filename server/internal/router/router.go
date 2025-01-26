package router

import (
	"github.com/gorilla/mux"

	"github.com/arpansaha13/pariksha/internal/handlers"
	"github.com/arpansaha13/pariksha/internal/middlewares"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	authRouter := r.PathPrefix("/auth").Subrouter()
	protectedRouter := r.PathPrefix("/").Subrouter()

	protectedRouter.Use(middlewares.AuthMiddleware)

	// Auth Routes
	authRouter.HandleFunc("/login", handlers.Login).Methods("POST")
	authRouter.HandleFunc("/signup", handlers.SignUp).Methods("POST")
	authRouter.HandleFunc("/verification/{hash}", handlers.Verification).Methods("POST")

	// Paper Routes
	protectedRouter.HandleFunc("/papers", handlers.CreatePaper).Methods("POST")

	return r
}
