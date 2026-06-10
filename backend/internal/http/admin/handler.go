package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/domain"
	"github.com/smallestfisher/relaydeck/backend/internal/http/middleware"
)

type Store interface {
	APIKeys() []domain.APIKey
	Models() []domain.Model
	Sites() []domain.UpstreamSite
	Users() []domain.User
	UserByEmail(email string) (domain.User, bool)
	UserByID(id string) (domain.User, bool)
}

type Handler struct {
	store    Store
	sessions auth.SessionStore
	now      func() time.Time
}

func New(store Store, sessions auth.SessionStore, now func() time.Time) http.Handler {
	if now == nil {
		now = time.Now
	}
	h := &Handler{store: store, sessions: sessions, now: now}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/admin/auth/login", h.handleLogin)
	mux.HandleFunc("POST /api/admin/auth/logout", h.handleLogout)
	mux.Handle("GET /api/admin/auth/me", middleware.RequireAdminSession(http.HandlerFunc(h.handleMe), sessions, store, now))
	mux.Handle("GET /api/admin/summary", middleware.RequireAdminSession(http.HandlerFunc(h.handleSummary), sessions, store, now))
	return mux
}

func (h *Handler) handleSummary(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]int{
		"sites":    len(h.store.Sites()),
		"models":   len(h.store.Models()),
		"api_keys": len(h.store.APIKeys()),
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
