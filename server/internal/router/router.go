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
	authRouter.HandleFunc("/login/otp", handlers.LoginWithOtp).Methods("POST")
	authRouter.HandleFunc("/signup", handlers.SignUp).Methods("POST")
	authRouter.HandleFunc("/verification/signup", handlers.VerifySignup).Methods("POST")
	authRouter.HandleFunc("/verification/login", handlers.VerifyLogin).Methods("POST")

	// Paper Routes
	protectedRouter.HandleFunc("/papers", handlers.GetUserPapers).Methods("GET")
	protectedRouter.HandleFunc("/papers", handlers.CreatePaper).Methods("POST")
	protectedRouter.HandleFunc("/papers/{id}", handlers.UpdatePaper).Methods("PATCH")
	protectedRouter.HandleFunc("/papers/{id}/questions", handlers.GetPaperQuestions).Methods("GET")
	protectedRouter.HandleFunc("/papers/{id}/questions", handlers.CreatePaperQuestions).Methods("POST")

	// Question Routes
	protectedRouter.HandleFunc("/questions/{id}", handlers.UpdateQuestion).Methods("PATCH")
	protectedRouter.HandleFunc("/questions/{id}", handlers.DeleteQuestion).Methods("DELETE")

	// Exam Routes
	protectedRouter.HandleFunc("/exams", handlers.GetUserExams).Methods("GET")
	protectedRouter.HandleFunc("/exams", handlers.CreateExam).Methods("POST")
	protectedRouter.HandleFunc("/exams/{examId}", handlers.UpdateExam).Methods("PATCH")
	protectedRouter.HandleFunc("/exams/{examId}/participants", handlers.GetExamParticipants).Methods("GET")
	protectedRouter.HandleFunc("/exams/{examId}/participants", handlers.AddExamParticipants).Methods("POST")
	protectedRouter.HandleFunc("/exams/{examId}/participants/{participantId}", handlers.RemoveExamParticipant).Methods("DELETE")
	protectedRouter.HandleFunc("/exams/{examId}/start", handlers.StartExam).Methods("PATCH")

	// User Routes
	protectedRouter.HandleFunc("/user", handlers.GetUser).Methods("GET")
	protectedRouter.HandleFunc("/user", handlers.UpdateUser).Methods("PATCH")

	return r
}
