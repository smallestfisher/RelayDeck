package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

func TestSummaryReturnsGatewayConfigurationCounts(t *testing.T) {
	adminStore := newTestAdminStore(t)
	adminStore.apiKeys = []domain.APIKey{{ID: "key_1"}}
	adminStore.models = []domain.Model{{ID: "model_1"}, {ID: "model_2"}}
	adminStore.sites = []domain.UpstreamSite{{ID: "site_1"}}
	sessions := auth.NewMemorySessionStore(fixedAdminNow)
	handler := New(adminStore, sessions, fixedAdminNow)
	user, _ := adminStore.UserByEmail("owner@example.com")
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
	handler := New(newTestAdminStore(t), auth.NewMemorySessionStore(fixedAdminNow), fixedAdminNow)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/summary", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
