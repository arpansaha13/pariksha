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

	// Auth Routes
	protectedRouter.HandleFunc("/check-auth", handlers.CheckAuth).Methods("GET", "OPTIONS")
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
	protectedRouter.HandleFunc("/papers/{paperId}", handlers.GetPaper).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/papers/{paperId}", handlers.UpdatePaper).Methods("PATCH", "OPTIONS")
	protectedRouter.HandleFunc("/papers/{paperId}/check", handlers.CheckPaperAccess).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/papers/{paperId}/questions", handlers.GetPaperQuestions).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/papers/{paperId}/questions", handlers.CreateQuestion).Methods("POST", "OPTIONS")

	// Question Routes
	protectedRouter.HandleFunc("/questions/{questionId}", handlers.GetQuestion).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/questions/{questionId}", handlers.UpdateQuestion).Methods("PATCH", "OPTIONS")
	protectedRouter.HandleFunc("/questions/{questionId}", handlers.DeleteQuestion).Methods("DELETE", "OPTIONS")
	protectedRouter.HandleFunc("/categories/{category_id}/questions/reorder", handlers.ReorderQuestions).Methods("PATCH", "OPTIONS")

	// Question Category Routes
	protectedRouter.HandleFunc("/papers/{paperId}/categories", handlers.GetPaperCategories).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/papers/{paperId}/categories", handlers.CreateCategory).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/categories/{categoryId}", handlers.UpdateCategory).Methods("PATCH", "OPTIONS")
	protectedRouter.HandleFunc("/categories/{categoryId}", handlers.DeleteCategory).Methods("DELETE", "OPTIONS")
	protectedRouter.HandleFunc("/papers/{paperId}/categories/reorder", handlers.ReorderCategories).Methods("PATCH", "OPTIONS")

	// Exam Routes
	protectedRouter.HandleFunc("/exams", handlers.GetUserExams).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/exams", handlers.CreateExam).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/exams/{examId}", handlers.GetExam).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/exams/{examId}", handlers.UpdateExam).Methods("PATCH", "OPTIONS")
	protectedRouter.HandleFunc("/exams/{examId}/permission", handlers.GetExamPermission).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/exams/{examId}/start", handlers.StartExam).Methods("PATCH", "OPTIONS")
	protectedRouter.HandleFunc("/exams/{examId}/end", handlers.EndExam).Methods("PATCH", "OPTIONS")
	protectedRouter.HandleFunc("/exams/{examId}/questions", handlers.GetExamQuestions).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/exams/questions/{questionId}", handlers.GetExamQuestion).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/exams/{examId}/categories", handlers.GetExamCategories).Methods("GET", "OPTIONS")

	// Exam Participant Routes
	protectedRouter.HandleFunc("/exams/{examId}/participants", handlers.GetExamParticipants).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/exams/{examId}/participants", handlers.AddExamParticipant).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/exams/{examId}/participants/{participantId}", handlers.RemoveExamParticipant).Methods("DELETE", "OPTIONS")
	protectedRouter.HandleFunc("/exams/{examId}/participants/current", handlers.GetExamParticipant).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/participants/{participantId}/evaluate", handlers.MarkAsEvaluated).Methods("PATCH", "OPTIONS")

	// Answer Routes
	protectedRouter.HandleFunc("/exams/{examId}/answers", handlers.UpsertAnswer).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/exams/{examId}/participants/{participantId}/answers", handlers.GetParticipantAnswers).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/exams/{examId}/questions/{questionId}/answer", handlers.GetAnswer).Methods("GET", "OPTIONS")

	// Answer Evaluation Routes
	protectedRouter.HandleFunc("/answers", handlers.UpdateAnswerForEvaluation).Methods("PATCH", "OPTIONS")

	// User Routes
	protectedRouter.HandleFunc("/users/me", handlers.GetAuthUser).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/users/me", handlers.UpdateAuthUser).Methods("PATCH", "OPTIONS")
	protectedRouter.HandleFunc("/users/{userId}", handlers.GetUser).Methods("GET", "OPTIONS")

	return r
}
