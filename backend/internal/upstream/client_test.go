package upstream

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

func TestClientDoJSONSendsBearerTokenAndBody(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer upstream-secret" {
			t.Fatalf("unexpected authorization: %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["model"] != "upstream-gpt-4o-mini" {
			t.Fatalf("expected mapped model, got %#v", body["model"])
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"id":"chatcmpl_1","object":"chat.completion"}`)),
		}, nil
	})

	client := NewClientWithHTTPClient(&http.Client{Transport: transport, Timeout: 2 * time.Second})
	resp, err := client.DoJSON(context.Background(), domain.UpstreamSite{BaseURL: "https://upstream.example", Credential: "upstream-secret"}, "/v1/chat/completions", []byte(`{"model":"upstream-gpt-4o-mini"}`))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(string(resp.Body), "chatcmpl_1") {
		t.Fatalf("unexpected body: %s", string(resp.Body))
	}
}

func TestClientDoJSONNormalizesUpstreamError(t *testing.T) {
	transport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("upstream exploded")),
		}, nil
	})

	client := NewClientWithHTTPClient(&http.Client{Transport: transport, Timeout: 2 * time.Second})
	_, err := client.DoJSON(context.Background(), domain.UpstreamSite{BaseURL: "https://upstream.example", Credential: "upstream-secret"}, "/v1/chat/completions", []byte(`{"model":"gpt"}`))
	if err == nil {
		t.Fatal("expected upstream error")
	}
	upstreamErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if upstreamErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", upstreamErr.StatusCode)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
