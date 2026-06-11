package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/domain"
	"github.com/smallestfisher/relaydeck/backend/internal/upstream"
)

func TestModelsRequiresAuthAndReturnsCanonicalModels(t *testing.T) {
	handler := New(newTestGatewayStore(), upstream.NewClientWithHTTPClient(successHTTPClient(t)), fixedNow)

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/v1/models", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status 401, got %d", unauthorized.Code)
	}

	authorized := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer rd_live_dev_test_secret")
	handler.ServeHTTP(authorized, req)

	if authorized.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", authorized.Code, authorized.Body.String())
	}
	if !strings.Contains(authorized.Body.String(), `"object":"list"`) || !strings.Contains(authorized.Body.String(), `"id":"gpt-4o-mini"`) {
		t.Fatalf("unexpected models response: %s", authorized.Body.String())
	}
}

func TestChatCompletionsProxiesToMappedUpstream(t *testing.T) {
	client := upstream.NewClientWithHTTPClient(roundTripClient(func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read upstream request: %v", err)
		}
		if !strings.Contains(string(body), `"model":"upstream-gpt-4o-mini"`) {
			t.Fatalf("expected mapped upstream model, got %s", string(body))
		}
		if req.Header.Get("Authorization") != "Bearer upstream-dev-secret" {
			t.Fatalf("unexpected upstream authorization: %q", req.Header.Get("Authorization"))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"id":"chatcmpl_test","object":"chat.completion"}`)),
		}, nil
	}))
	handler := New(newTestGatewayStore(), client, fixedNow)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}`))
	req.Header.Set("Authorization", "Bearer rd_live_dev_test_secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "chatcmpl_test") {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestResponsesReturnsCompatibilityErrorWhenUnsupported(t *testing.T) {
	handler := New(newTestGatewayStore(), upstream.NewClientWithHTTPClient(successHTTPClient(t)), fixedNow)
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{"model":"gpt-4o-mini","input":"hi"}`))
	req.Header.Set("Authorization", "Bearer rd_live_dev_test_secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "unsupported_endpoint") {
		t.Fatalf("expected compatibility error, got %s", rec.Body.String())
	}
}

func fixedNow() time.Time {
	return time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
}

type testGatewayStore struct {
	apiKeys  []domain.APIKey
	models   []domain.Model
	sites    []domain.UpstreamSite
	mappings []domain.SiteModel
}

func newTestGatewayStore() *testGatewayStore {
	now := fixedNow()
	return &testGatewayStore{
		apiKeys: []domain.APIKey{
			{
				ID:              "key_dev",
				UserID:          "user_admin",
				Prefix:          "rd_live_dev",
				Hash:            auth.HashSecret("rd_live_dev_test_secret"),
				Status:          domain.APIKeyStatusActive,
				Scopes:          []domain.Scope{domain.ScopeChatCompletions, domain.ScopeResponses},
				AllowedModels:   []string{"gpt-4o-mini", "gpt-4o"},
				ExpiresAt:       now.AddDate(10, 0, 0),
				OwnerIsActive:   true,
				RPM:             120,
				TPM:             60000,
				AllowedCIDRs:    []string{"0.0.0.0/0"},
				MonthlyQuotaTPM: 1000000,
			},
		},
		models: []domain.Model{
			{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Capabilities: []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming}},
			{ID: "gpt-4o", Name: "GPT-4o", Capabilities: []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming, domain.CapabilityVision}},
		},
		sites: []domain.UpstreamSite{
			{
				ID:           "upstream_dev",
				Name:         "Dev Upstream",
				BaseURL:      "https://upstream.example",
				Credential:   "upstream-dev-secret",
				Enabled:      true,
				Weight:       80,
				HealthScore:  95,
				SuccessRate:  99,
				LatencyMS:    120,
				Circuit:      domain.CircuitClosed,
				Capabilities: []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming},
			},
		},
		mappings: []domain.SiteModel{
			{
				SiteID:        "upstream_dev",
				Model:         "gpt-4o-mini",
				UpstreamModel: "upstream-gpt-4o-mini",
				EndpointTypes: []domain.EndpointType{domain.EndpointChatCompletions},
				Capabilities:  []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming},
			},
		},
	}
}

func (s *testGatewayStore) APIKeys() []domain.APIKey {
	return append([]domain.APIKey(nil), s.apiKeys...)
}

func (s *testGatewayStore) Models() []domain.Model {
	return append([]domain.Model(nil), s.models...)
}

func (s *testGatewayStore) Sites() []domain.UpstreamSite {
	return append([]domain.UpstreamSite(nil), s.sites...)
}

func (s *testGatewayStore) Mappings() []domain.SiteModel {
	return append([]domain.SiteModel(nil), s.mappings...)
}

func successHTTPClient(t *testing.T) *http.Client {
	t.Helper()
	return roundTripClient(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	})
}

func roundTripClient(fn func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{Transport: roundTripFunc(fn)}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
