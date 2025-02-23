package middlewares

import (
	"net/http"

	"github.com/arpansaha13/common/pkg/models"
)

var unsafeMethods = map[string]bool{
	"POST":   true,
	"PUT":    true,
	"PATCH":  true,
	"DELETE": true,
}

func CsrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !unsafeMethods[r.Method] {
			next.ServeHTTP(w, r)
			return
		}

		session, ok := r.Context().Value(SessionKey).(*models.Session)
		if !ok || session == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		csrfToken := r.Header.Get("X-CSRFToken")
		if csrfToken == "" {
			http.Error(w, "CSRF token missing", http.StatusForbidden)
			return
		}

		if session.CsrfToken != csrfToken {
			http.Error(w, "CSRF token invalid", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
