package middlewares

import (
	"context"
	"net/http"

	"github.com/arpansaha13/pariksha/internal/constants"
	"github.com/arpansaha13/pariksha/internal/db"
	"github.com/arpansaha13/pariksha/internal/models"
	"github.com/arpansaha13/pariksha/internal/utils"
)

type sessionContextKey string

const SessionKey sessionContextKey = "session"

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionCookieName := utils.GetEnvWithDefault("SESSION_COOKIE_NAME", constants.DEFAULT_SESSION_COOKIE_NAME)
		sessionCookie, err := r.Cookie(sessionCookieName)
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
