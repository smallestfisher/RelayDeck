package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/domain"
	"github.com/smallestfisher/relaydeck/backend/internal/http/middleware"
)

type authPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type currentUserPayload struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	var payload authPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
		return
	}
	user, ok := h.store.UserByEmail(payload.Email)
	if !ok || user.Status != domain.UserStatusActive || !auth.VerifyPassword(user.PasswordHash, payload.Password) {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid_credentials"})
		return
	}
	token, err := auth.NewSessionToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "session_unavailable"})
		return
	}
	session := auth.Session{
		Token:      token,
		UserID:     user.ID,
		Email:      user.Email,
		Role:       string(user.Role),
		IssuedAt:   h.now(),
		ExpiresAt:  h.now().Add(24 * time.Hour),
		LastSeenAt: h.now(),
	}
	h.sessions.Create(session)
	http.SetCookie(w, &http.Cookie{
		Name:     "relaydeck_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	})
	writeJSON(w, http.StatusOK, map[string]any{"user": toCurrentUser(user)})
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	user, ok := h.currentUser(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": toCurrentUser(user)})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("relaydeck_session")
	if err == nil && cookie.Value != "" {
		h.sessions.Delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "relaydeck_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) currentUser(r *http.Request) (domain.User, bool) {
	user, ok := middleware.AdminUserFromContext(r.Context())
	if ok {
		return user, true
	}
	cookie, err := r.Cookie("relaydeck_session")
	if err != nil || cookie.Value == "" {
		return domain.User{}, false
	}
	session, ok := h.sessions.Get(cookie.Value)
	if !ok {
		return domain.User{}, false
	}
	user, ok = h.store.UserByID(session.UserID)
	if !ok || user.Status != domain.UserStatusActive {
		return domain.User{}, false
	}
	return user, true
}

func toCurrentUser(user domain.User) currentUserPayload {
	return currentUserPayload{ID: user.ID, Email: user.Email, Role: string(user.Role), Status: string(user.Status)}
}
