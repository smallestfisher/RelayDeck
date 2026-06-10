package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/smallestfisher/relaydeck/backend/internal/store"
)

func TestSummaryReturnsGatewayConfigurationCounts(t *testing.T) {
	handler := New(store.NewMemoryStore())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/summary", nil)

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
