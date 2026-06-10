package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/store"
	"github.com/smallestfisher/relaydeck/backend/internal/upstream"
)

func TestModelsRequiresAuthAndReturnsCanonicalModels(t *testing.T) {
	handler := New(store.NewMemoryStore(), upstream.NewClientWithHTTPClient(successHTTPClient(t)), fixedNow)

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
	handler := New(store.NewMemoryStore(), client, fixedNow)

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
	handler := New(store.NewMemoryStore(), upstream.NewClientWithHTTPClient(successHTTPClient(t)), fixedNow)
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
