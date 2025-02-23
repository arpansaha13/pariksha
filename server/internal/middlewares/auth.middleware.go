package middlewares

import (
	"context"
	"net/http"

	"github.com/arpansaha13/common/pkg/proto"
	"github.com/arpansaha13/pariksha/internal/config/env"
	"github.com/arpansaha13/pariksha/internal/services"
)

type userContextKey string

const UserIDKey userContextKey = "user_id"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie(env.SESSION_COOKIE_NAME)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		csrfToken := r.Header.Get("X-CSRFToken")
		authService := services.GetAuthService()
		response, err := authService.Client().Authenticate(context.Background(), &proto.AuthenticateRequest{
			SessionKey: sessionCookie.Value,
			CsrfToken:  csrfToken,
		})

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, int(response.UserId))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
