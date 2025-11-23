package middlewares

import (
	"context"
	"encoding/json"
	"net/http"

	"pariksha/common/pkg/proto"
	"pariksha/gateway/internal/config/env"
	"pariksha/gateway/internal/interservice"
)

type userContextKey string

const UserIDKey userContextKey = "user_id"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuthCheck := r.URL.Path == "/api/check-auth"

		sessionCookie, err := r.Cookie(env.SESSION_COOKIE_NAME)

		if err != nil {
			if isAuthCheck {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]bool{"valid": false})
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		csrfToken := r.Header.Get("X-CSRFToken")
		response, err := interservice.Authenticate(&proto.AuthenticateRequest{
			SessionKey: sessionCookie.Value,
			CsrfToken:  csrfToken,
		})

		if err != nil {
			if isAuthCheck {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]bool{"valid": false})
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if isAuthCheck {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"valid": true})
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, response.UserId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
