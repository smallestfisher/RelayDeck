package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestAdminSummaryRequiresSession(t *testing.T) {
	handler := New(config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/summary", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAdminSummaryRouteIsMountedForLoggedInUser(t *testing.T) {
	handler := New(config.Config{BootstrapOwnerEmail: "owner@example.com", BootstrapOwnerPassword: "change-me"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/admin/auth/login", strings.NewReader(`{"email":"owner@example.com","password":"change-me"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()

	handler.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d: %s", loginRec.Code, loginRec.Body.String())
	}
	cookies := loginRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected login to set a session cookie")
	}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/summary", nil)
	req.AddCookie(cookies[0])
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpstreamAdminRouteIsMountedForLoggedInUser(t *testing.T) {
	handler := New(config.Config{BootstrapOwnerEmail: "owner@example.com", BootstrapOwnerPassword: "change-me", UpstreamSecretKey: "0123456789abcdef0123456789abcdef"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/admin/auth/login", strings.NewReader(`{"email":"owner@example.com","password":"change-me"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()

	handler.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d: %s", loginRec.Code, loginRec.Body.String())
	}
	cookies := loginRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected login to set a session cookie")
	}
	req := httptest.NewRequest(http.MethodGet, "/api/admin/upstreams/accounts", nil)
	req.AddCookie(cookies[0])
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected upstream store unavailable without DATABASE_URL, got %d: %s", rec.Code, rec.Body.String())
	}
}
