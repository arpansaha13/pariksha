package middlewares

import (
	"context"
	"net/http"

	"github.com/arpansaha13/pariksha/internal/config/env"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/models"
)

type sessionContextKey string

const SessionKey sessionContextKey = "session"

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie(env.SESSION_COOKIE_NAME)
		if err != nil {
			ctx := context.WithValue(r.Context(), SessionKey, nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		var session models.Session
		err = db.DB.Where("key = ?", sessionCookie.Value).Take(&session).Error
		if err != nil {
			ctx := context.WithValue(r.Context(), SessionKey, nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ctx := context.WithValue(r.Context(), SessionKey, &session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
