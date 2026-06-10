package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/store"
)

func TestSummaryReturnsGatewayConfigurationCounts(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	sessions := auth.NewMemorySessionStore(fixedAdminNow)
	handler := New(memoryStore, sessions, fixedAdminNow)
	user, _ := memoryStore.UserByEmail(store.DefaultOwnerEmail)
	sessions.Create(auth.Session{
		Token:     "summary-session",
		UserID:    user.ID,
		Email:     user.Email,
		Role:      string(user.Role),
		IssuedAt:  fixedAdminNow(),
		ExpiresAt: fixedAdminNow().AddDate(0, 0, 1),
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/summary", nil)
	req.AddCookie(&http.Cookie{Name: "relaydeck_session", Value: "summary-session"})

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	for _, expected := range []string{`"sites":1`, `"models":2`, `"api_keys":1`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected %s in body, got %s", expected, body)
		}
	}
}

func TestSummaryRejectsMissingSession(t *testing.T) {
	handler := New(store.NewMemoryStore(), auth.NewMemorySessionStore(fixedAdminNow), fixedAdminNow)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/summary", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
