package router

import (
	"github.com/gorilla/mux"

	"pariksha/server/internal/handlers"
	"pariksha/server/internal/middlewares"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// Apply CORS middleware to the entire router
	r.Use(mux.CORSMethodMiddleware(r))
	r.Use(middlewares.CorsMiddleware)

	authRouter := r.PathPrefix("/api/auth").Subrouter()

	protectedRouter := r.PathPrefix("/api").Subrouter()
	protectedRouter.Use(middlewares.AuthMiddleware)

	decryptIdRouter := protectedRouter.PathPrefix("").Subrouter()
	decryptIdRouter.Use(middlewares.DecryptIDMiddleware)

	// Auth Routes
	protectedRouter.HandleFunc("/check-auth", handlers.CheckAuth).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/logout", handlers.Logout).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/login", handlers.LoginWithPassword).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/login/otp", handlers.LoginWithOtp).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/signup", handlers.SignUp).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/verification/signup", handlers.VerifySignup).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/verification/login", handlers.VerifyLoginWithOtp).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/forgot-password", handlers.ForgotPassword).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/reset-password", handlers.ResetPassword).Methods("POST", "OPTIONS")

	// Paper Routes
	protectedRouter.HandleFunc("/papers", handlers.GetUserPapers).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/papers", handlers.CreatePaper).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/papers", handlers.DeletePapers).Methods("DELETE", "OPTIONS")
	decryptIdRouter.HandleFunc("/papers/{paperId}", handlers.GetPaper).Methods("GET", "OPTIONS")
	decryptIdRouter.HandleFunc("/papers/{paperId}", handlers.UpdatePaper).Methods("PATCH", "OPTIONS")
	decryptIdRouter.HandleFunc("/papers/{paperId}/questions", handlers.GetPaperQuestions).Methods("GET", "OPTIONS")
	decryptIdRouter.HandleFunc("/papers/{paperId}/questions", handlers.CreateQuestion).Methods("POST", "OPTIONS")
	decryptIdRouter.HandleFunc("/papers/{paperId}/permissions", handlers.GetPaperPermissions).Methods("GET", "OPTIONS")

	// Question Routes
	decryptIdRouter.HandleFunc("/questions/{questionId}", handlers.GetPaperQuestion).Methods("GET", "OPTIONS")
	decryptIdRouter.HandleFunc("/questions/{questionId}", handlers.UpdateQuestion).Methods("PATCH", "OPTIONS")
	decryptIdRouter.HandleFunc("/questions/{questionId}", handlers.DeleteQuestion).Methods("DELETE", "OPTIONS")
	decryptIdRouter.HandleFunc("/questions/{questionId}/test-cases", handlers.UpsertPaperTestCases).Methods("PUT", "OPTIONS")
	protectedRouter.HandleFunc("/categories/{category_id}/questions/reorder", handlers.ReorderQuestions).Methods("PATCH", "OPTIONS")

	// Question Category Routes
	decryptIdRouter.HandleFunc("/papers/{paperId}/categories", handlers.GetPaperCategories).Methods("GET", "OPTIONS")
	decryptIdRouter.HandleFunc("/papers/{paperId}/categories", handlers.CreateCategory).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/categories/{categoryId}", handlers.UpdateCategory).Methods("PATCH", "OPTIONS")
	protectedRouter.HandleFunc("/categories/{categoryId}", handlers.DeleteCategory).Methods("DELETE", "OPTIONS")
	decryptIdRouter.HandleFunc("/papers/{paperId}/categories/reorder", handlers.ReorderCategories).Methods("PATCH", "OPTIONS")

	// Exam Routes
	protectedRouter.HandleFunc("/exams", handlers.GetUserExams).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/exams", handlers.CreateExam).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/exams", handlers.DeleteExams).Methods("DELETE", "OPTIONS")
	decryptIdRouter.HandleFunc("/exams/{examId}", handlers.GetExam).Methods("GET", "OPTIONS")
	decryptIdRouter.HandleFunc("/exams/{examId}", handlers.UpdateExam).Methods("PATCH", "OPTIONS")
	decryptIdRouter.HandleFunc("/exams/{examId}/permission", handlers.GetExamPermission).Methods("GET", "OPTIONS")
	decryptIdRouter.HandleFunc("/exams/{examId}/start", handlers.StartExam).Methods("PATCH", "OPTIONS")
	decryptIdRouter.HandleFunc("/exams/{examId}/end", handlers.EndExam).Methods("PATCH", "OPTIONS")
	decryptIdRouter.HandleFunc("/exams/{examId}/questions", handlers.GetExamQuestions).Methods("GET", "OPTIONS")
	decryptIdRouter.HandleFunc("/exams/{examId}/categories", handlers.GetExamCategories).Methods("GET", "OPTIONS")
	decryptIdRouter.HandleFunc("/exams/questions/{questionId}", handlers.GetExamQuestion).Methods("GET", "OPTIONS")
	decryptIdRouter.HandleFunc("/exams/{examId}/results", handlers.GetExamResults).Methods("GET", "OPTIONS")

	// Exam Participant Routes
	decryptIdRouter.HandleFunc("/exams/{examId}/participants", handlers.GetExamParticipants).Methods("GET", "OPTIONS")
	decryptIdRouter.HandleFunc("/exams/{examId}/participants", handlers.AddExamParticipant).Methods("POST", "OPTIONS")
	decryptIdRouter.HandleFunc("/exams/{examId}/participants/{participantId}", handlers.RemoveExamParticipant).Methods("DELETE", "OPTIONS")
	decryptIdRouter.HandleFunc("/exams/{examId}/participants/current", handlers.GetExamParticipant).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/participants/{participantId}/evaluate", handlers.MarkParticipantAsEvaluated).Methods("PATCH", "OPTIONS")
	protectedRouter.HandleFunc("/participants/{participantId}", handlers.GetParticipantById).Methods("GET", "OPTIONS")

	// Answer Routes
	decryptIdRouter.HandleFunc("/exams/{examId}/answers", handlers.UpsertAnswer).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/participants/{participantId}/answers", handlers.GetParticipantAnswers).Methods("GET", "OPTIONS")
	decryptIdRouter.HandleFunc("/exams/{examId}/questions/{questionId}/answer", handlers.GetAnswerForExam).Methods("GET", "OPTIONS")

	// Answer Evaluation Routes
	decryptIdRouter.HandleFunc("/participants/{participantId}/questions/{questionId}/evaluation-data", handlers.GetAnswerEvaluationData).Methods("GET", "OPTIONS")
	decryptIdRouter.HandleFunc("/participants/{participantId}/questions/{questionId}/answer", handlers.GetAnswerForEvaluation).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/answers/{answerId}", handlers.UpdateAnswerForEvaluation).Methods("PATCH", "OPTIONS")

	// User Routes
	protectedRouter.HandleFunc("/users/me", handlers.GetAuthUser).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/users/me", handlers.UpdateAuthUser).Methods("PATCH", "OPTIONS")
	protectedRouter.HandleFunc("/users/{userId}", handlers.GetUser).Methods("GET", "OPTIONS")

	// Engine Routes
	protectedRouter.HandleFunc("/engine/run", handlers.RunCode).Methods("POST", "OPTIONS")

	// Boilerplate Routes
	decryptIdRouter.HandleFunc("/questions/{questionId}/languages/{languageId}/boilerplate", handlers.GetBoilerplate).Methods("GET", "OPTIONS")

	return r
}
