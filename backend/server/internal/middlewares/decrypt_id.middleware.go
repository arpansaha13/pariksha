package middlewares

import (
	"context"
	"net/http"

	"pariksha/server/internal/utils"

	"github.com/gorilla/mux"
)

type decryptedIDKey string

const (
	DecryptedExamID  decryptedIDKey = "decrypted_exam_id"
	DecryptedPaperID decryptedIDKey = "decrypted_paper_id"
)

// DecryptIDMiddleware decrypts examId and paperId from URL parameters
//
// Note: This middleware won't decrypt IDs in the request body
func DecryptIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ctx := r.Context()

		if examID, exists := vars["examId"]; exists {
			decryptedID, err := utils.DecryptID(examID)
			if err != nil {
				http.Error(w, "Invalid exam ID", http.StatusNotFound)
				return
			}
			ctx = context.WithValue(ctx, DecryptedExamID, decryptedID)
		}

		if paperID, exists := vars["paperId"]; exists {
			decryptedID, err := utils.DecryptID(paperID)
			if err != nil {
				http.Error(w, "Invalid paper ID", http.StatusNotFound)
				return
			}
			ctx = context.WithValue(ctx, DecryptedPaperID, decryptedID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
