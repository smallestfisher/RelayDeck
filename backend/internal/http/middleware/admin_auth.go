package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

type UserReader interface {
	UserByID(id string) (domain.User, bool)
}

type adminUserContextKey struct{}

func AdminUserFromContext(ctx context.Context) (domain.User, bool) {
	user, ok := ctx.Value(adminUserContextKey{}).(domain.User)
	return user, ok
}

func RequireAdminSession(next http.Handler, sessions auth.SessionStore, users UserReader, now func() time.Time) http.Handler {
	if now == nil {
		now = time.Now
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("relaydeck_session")
		if err != nil || cookie.Value == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		session, ok := sessions.Get(cookie.Value)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		user, ok := users.UserByID(session.UserID)
		if !ok || user.Status != domain.UserStatusActive {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		session.LastSeenAt = now()
		_ = sessions.Create(session)
		ctx := context.WithValue(r.Context(), adminUserContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
