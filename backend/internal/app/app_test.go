package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smallestfisher/relaydeck/backend/internal/config"
)

func TestHealthzReturnsOK(t *testing.T) {
	handler := New(config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok\n" {
		t.Fatalf("expected ok body, got %q", rec.Body.String())
	}
}

func TestAdminSummaryRouteIsMounted(t *testing.T) {
	handler := New(config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/summary", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
