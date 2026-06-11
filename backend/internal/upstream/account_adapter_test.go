package upstream

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

func TestNewAPIAdapterSyncModelsUsesAPIKeyModelList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Fatalf("unexpected authorization: %q", r.Header.Get("Authorization"))
		}
		writeTestJSON(t, w, map[string]any{
			"data": []map[string]any{
				{"id": "gpt-4o-mini"},
				{"id": "claude-3-5-sonnet"},
			},
		})
	}))
	defer server.Close()

	adapter := NewNewAPIAccountAdapter(server.Client(), time.Second)
	result, status, err := adapter.SyncModels(context.Background(), accountForAdapter(server.URL, domain.PlatformKindNewAPI), "sk-test")
	if err != nil {
		t.Fatalf("sync models: %v", err)
	}
	if status.APIStatus != domain.UpstreamAPIStatusHealthy || status.ModelCount != 2 {
		t.Fatalf("unexpected status: %+v", status)
	}
	if len(result.Models) != 2 || result.Models[0].NormalizedModelName != "gpt-4o-mini" {
		t.Fatalf("unexpected models: %+v", result.Models)
	}
}

func TestNewAPIAdapterRefreshQuotaUsesTokenUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/usage/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Fatalf("unexpected authorization: %q", r.Header.Get("Authorization"))
		}
		writeTestJSON(t, w, map[string]any{"data": map[string]any{"remain_quota": 42.5}})
	}))
	defer server.Close()

	adapter := NewNewAPIAccountAdapter(server.Client(), time.Second)
	result, err := adapter.RefreshQuota(context.Background(), accountForAdapter(server.URL, domain.PlatformKindNewAPI), "sk-test", "")
	if err != nil {
		t.Fatalf("refresh quota: %v", err)
	}
	if result.Status.APIStatus != domain.UpstreamAPIStatusHealthy || result.BalanceAmount != 42.5 {
		t.Fatalf("unexpected quota result: %+v", result)
	}
}

func TestNewAPIAdapterAccountCredentialFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user/self" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	adapter := NewNewAPIAccountAdapter(server.Client(), time.Second)
	status, err := adapter.TestAccountCredential(context.Background(), accountForAdapter(server.URL, domain.PlatformKindNewAPI), "session=bad")
	if err != nil {
		t.Fatalf("test account credential: %v", err)
	}
	if status.AccountStatus != domain.AccountCredentialStatusExpired || status.LastErrorClass != domain.UpstreamErrorCredentialExpired {
		t.Fatalf("unexpected account status: %+v", status)
	}
}

func TestNewAPIAdapterCheckinActionRequired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user/checkin" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusBadRequest)
		writeTestJSON(t, w, map[string]any{"message": "turnstile required"})
	}))
	defer server.Close()

	adapter := NewNewAPIAccountAdapter(server.Client(), time.Second)
	result, err := adapter.Checkin(context.Background(), accountForAdapter(server.URL, domain.PlatformKindNewAPI), "session=valid")
	if err != nil {
		t.Fatalf("checkin: %v", err)
	}
	if result.Status != domain.CheckinStatusActionRequired {
		t.Fatalf("expected action_required, got %+v", result)
	}
}

func TestSub2APIAdapterSyncModelsAndUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			if r.Header.Get("Authorization") != "Bearer sk-sub2api" {
				t.Fatalf("unexpected authorization: %q", r.Header.Get("Authorization"))
			}
			writeTestJSON(t, w, map[string]any{"data": []map[string]any{{"id": "claude-sonnet-4"}}})
		case "/v1/usage":
			if r.Header.Get("Authorization") != "Bearer sk-sub2api" {
				t.Fatalf("unexpected authorization: %q", r.Header.Get("Authorization"))
			}
			writeTestJSON(t, w, map[string]any{"remaining": 88})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	adapter := NewSub2APIAccountAdapter(server.Client(), time.Second)
	account := accountForAdapter(server.URL, domain.PlatformKindSub2API)
	models, status, err := adapter.SyncModels(context.Background(), account, "sk-sub2api")
	if err != nil {
		t.Fatalf("sync models: %v", err)
	}
	if len(models.Models) != 1 || status.ModelCount != 1 {
		t.Fatalf("unexpected model sync: %+v %+v", models, status)
	}
	quota, err := adapter.RefreshQuota(context.Background(), account, "sk-sub2api", "")
	if err != nil {
		t.Fatalf("refresh quota: %v", err)
	}
	if quota.BalanceAmount != 88 {
		t.Fatalf("unexpected quota: %+v", quota)
	}
}

func TestSub2APIAdapterProfileCredentialAndUnsupportedCheckin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/user/profile" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer jwt-token" {
			t.Fatalf("unexpected authorization: %q", r.Header.Get("Authorization"))
		}
		writeTestJSON(t, w, map[string]any{"id": 1, "email": "user@example.com"})
	}))
	defer server.Close()

	adapter := NewSub2APIAccountAdapter(server.Client(), time.Second)
	status, err := adapter.TestAccountCredential(context.Background(), accountForAdapter(server.URL, domain.PlatformKindSub2API), "jwt-token")
	if err != nil {
		t.Fatalf("test account credential: %v", err)
	}
	if status.AccountStatus != domain.AccountCredentialStatusValid {
		t.Fatalf("unexpected account status: %+v", status)
	}
	checkin, err := adapter.Checkin(context.Background(), accountForAdapter(server.URL, domain.PlatformKindSub2API), "jwt-token")
	if err != nil {
		t.Fatalf("checkin: %v", err)
	}
	if checkin.Status != domain.CheckinStatusUnsupported {
		t.Fatalf("expected unsupported checkin, got %+v", checkin)
	}
}

func accountForAdapter(baseURL string, kind domain.UpstreamPlatformKind) domain.UpstreamAccount {
	return domain.UpstreamAccount{
		ID:           "acct_test",
		Name:         "Test Account",
		PlatformKind: kind,
		BaseURL:      baseURL,
		Enabled:      true,
	}
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("write json: %v", err)
	}
}
