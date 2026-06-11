package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/domain"
	"github.com/smallestfisher/relaydeck/backend/internal/store"
)

func TestAdminLoginSetsSessionCookie(t *testing.T) {
	store := newTestAdminStore(t)
	sessions := auth.NewMemorySessionStore(fixedAdminNow)
	handler := New(store, sessions, fixedAdminNow)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/auth/login", strings.NewReader(`{"email":"owner@example.com","password":"change-me"}`))
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	cookie := rec.Result().Cookies()
	if len(cookie) != 1 {
		t.Fatalf("expected one cookie, got %d", len(cookie))
	}
	if cookie[0].Name != "relaydeck_session" {
		t.Fatalf("expected relaydeck_session cookie, got %s", cookie[0].Name)
	}
	if !cookie[0].HttpOnly {
		t.Fatal("expected HttpOnly session cookie")
	}
	if _, ok := sessions.Get(cookie[0].Value); !ok {
		t.Fatal("expected session to be stored")
	}
	if !strings.Contains(rec.Body.String(), "owner@example.com") {
		t.Fatalf("expected user payload in response, got %s", rec.Body.String())
	}
}

func TestAdminMeRequiresSession(t *testing.T) {
	store := newTestAdminStore(t)
	sessions := auth.NewMemorySessionStore(fixedAdminNow)
	handler := New(store, sessions, fixedAdminNow)

	user, _ := store.UserByEmail("owner@example.com")
	token, err := auth.NewSessionToken()
	if err != nil {
		t.Fatalf("expected token, got %v", err)
	}
	sessions.Create(auth.Session{
		Token:      token,
		UserID:     user.ID,
		Email:      user.Email,
		Role:       string(user.Role),
		IssuedAt:   fixedAdminNow(),
		ExpiresAt:  fixedAdminNow().Add(30 * time.Minute),
		LastSeenAt: fixedAdminNow(),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "relaydeck_session", Value: token})

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "owner@example.com") {
		t.Fatalf("expected current user payload, got %s", rec.Body.String())
	}
}

func TestAdminLogoutClearsSession(t *testing.T) {
	store := newTestAdminStore(t)
	sessions := auth.NewMemorySessionStore(fixedAdminNow)
	handler := New(store, sessions, fixedAdminNow)

	user, _ := store.UserByEmail("owner@example.com")
	token, err := auth.NewSessionToken()
	if err != nil {
		t.Fatalf("expected token, got %v", err)
	}
	sessions.Create(auth.Session{
		Token:      token,
		UserID:     user.ID,
		Email:      user.Email,
		Role:       string(user.Role),
		IssuedAt:   fixedAdminNow(),
		ExpiresAt:  fixedAdminNow().Add(30 * time.Minute),
		LastSeenAt: fixedAdminNow(),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "relaydeck_session", Value: token})

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
	if _, ok := sessions.Get(token); ok {
		t.Fatal("expected session to be removed")
	}
	cleared := rec.Result().Cookies()
	if len(cleared) == 0 || cleared[0].MaxAge >= 0 {
		t.Fatal("expected cleared session cookie")
	}
}

func newTestAdminStore(t *testing.T) *testAdminStore {
	t.Helper()
	hash, err := auth.HashPassword("change-me")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return &testAdminStore{users: []domain.User{{
		ID:           "user_owner",
		Email:        "owner@example.com",
		PasswordHash: hash,
		Role:         domain.UserRoleOwner,
		Status:       domain.UserStatusActive,
		CreatedAt:    fixedAdminNow(),
		UpdatedAt:    fixedAdminNow(),
	}}}
}

func fixedAdminNow() time.Time {
	return time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
}

type testAdminStore struct {
	users     []domain.User
	apiKeys   []domain.APIKey
	models    []domain.Model
	sites     []domain.UpstreamSite
	upstreams store.UpstreamAccountStore
}

func (s *testAdminStore) APIKeys() []domain.APIKey {
	return append([]domain.APIKey(nil), s.apiKeys...)
}
func (s *testAdminStore) Models() []domain.Model {
	return append([]domain.Model(nil), s.models...)
}
func (s *testAdminStore) Sites() []domain.UpstreamSite {
	return append([]domain.UpstreamSite(nil), s.sites...)
}
func (s *testAdminStore) Mappings() []domain.SiteModel { return nil }
func (s *testAdminStore) Users() []domain.User         { return append([]domain.User(nil), s.users...) }
func (s *testAdminStore) Upstreams() store.UpstreamAccountStore {
	return s.upstreams
}
func (s *testAdminStore) UserByEmail(email string) (domain.User, bool) {
	for _, user := range s.users {
		if user.Email == email {
			return user, true
		}
	}
	return domain.User{}, false
}
func (s *testAdminStore) UserByID(id string) (domain.User, bool) {
	for _, user := range s.users {
		if user.ID == id {
			return user, true
		}
	}
	return domain.User{}, false
}
