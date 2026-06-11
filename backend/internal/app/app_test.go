package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/smallestfisher/relaydeck/backend/internal/config"
	"github.com/smallestfisher/relaydeck/backend/internal/store/postgres"
)

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}
	if !strings.Contains(cfg.DatabaseURL, "test") {
		t.Fatalf("DATABASE_URL must point to a test database, got %q", cfg.DatabaseURL)
	}
	cfg.DatabaseURL = isolatedDatabaseURL(t, cfg.DatabaseURL)
	return New(cfg)
}

func isolatedDatabaseURL(t *testing.T, databaseURL string) string {
	t.Helper()
	schema := testSchemaName(t)
	ctx := context.Background()
	db, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA %s`, schema)); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		cleanupDB, err := postgres.Open(context.Background(), databaseURL)
		if err != nil {
			t.Fatalf("open postgres for cleanup: %v", err)
		}
		defer cleanupDB.Close()
		if _, err := cleanupDB.ExecContext(context.Background(), fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, schema)); err != nil {
			t.Fatalf("drop test schema: %v", err)
		}
	})
	return databaseURLWithSearchPath(t, databaseURL, schema)
}

func databaseURLWithSearchPath(t *testing.T, databaseURL string, schema string) string {
	t.Helper()
	parsed, err := url.Parse(databaseURL)
	if err != nil || parsed.Scheme == "" {
		t.Fatalf("DATABASE_URL must be a postgres URL for isolated tests, got %q", databaseURL)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func testSchemaName(t *testing.T) string {
	t.Helper()
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		t.Fatalf("generate schema suffix: %v", err)
	}
	return "test_" + hex.EncodeToString(suffix[:])
}

func TestHealthzReturnsOK(t *testing.T) {
	handler := newTestHandler(t)

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
	handler := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/summary", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAdminSummaryRouteIsMountedForLoggedInUser(t *testing.T) {
	handler := newTestHandler(t)
	cookie := loginAndGetCookie(t, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/summary", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpstreamAdminRouteIsMountedForLoggedInUser(t *testing.T) {
	handler := newTestHandler(t)
	cookie := loginAndGetCookie(t, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/upstreams/accounts", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected upstream accounts to be served with DATABASE_URL set, got %d: %s", rec.Code, rec.Body.String())
	}
}

func loginAndGetCookie(t *testing.T, handler http.Handler) *http.Cookie {
	t.Helper()
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
	return cookies[0]
}
