package upstream

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
		switch r.URL.Path {
		case "/api/usage/token":
			if r.Header.Get("Authorization") != "Bearer sk-test" {
				t.Fatalf("unexpected authorization: %q", r.Header.Get("Authorization"))
			}
			writeTestJSON(t, w, map[string]any{"data": map[string]any{"remain_quota": 500000}})
		case "/api/status":
			writeTestJSON(t, w, map[string]any{"quota_per_unit": 500000})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	adapter := NewNewAPIAccountAdapter(server.Client(), time.Second)
	result, err := adapter.RefreshQuota(context.Background(), accountForAdapter(server.URL, domain.PlatformKindNewAPI), "sk-test", "")
	if err != nil {
		t.Fatalf("refresh quota: %v", err)
	}
	if result.Status.APIStatus != domain.UpstreamAPIStatusHealthy || result.BalanceAmount != 1 || result.BalanceUnit != "usd" {
		t.Fatalf("unexpected quota result: %+v", result)
	}
}

func TestNewAPIAdapterAccountCredentialUsesAccessTokenUserHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user/self" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "token-123" {
			t.Fatalf("unexpected authorization: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("New-Api-User") != "42" {
			t.Fatalf("unexpected new-api user: %q", r.Header.Get("New-Api-User"))
		}
		writeTestJSON(t, w, map[string]any{"success": true, "data": map[string]any{"id": 42}})
	}))
	defer server.Close()

	adapter := NewNewAPIAccountAdapter(server.Client(), time.Second)
	account := accountForAdapter(server.URL, domain.PlatformKindNewAPI)
	account.AccountCredentialKind = domain.CredentialKindNewAPIAccessToken
	result, err := adapter.TestAccountCredential(context.Background(), account, `{"access_token":"token-123","user_id":42}`)
	if err != nil {
		t.Fatalf("test account credential: %v", err)
	}
	if result.Status.AccountStatus != domain.AccountCredentialStatusValid {
		t.Fatalf("unexpected account status: %+v", result.Status)
	}
}

func TestNewAPIAdapterAccountCredentialSuccessFalseExpiresCredential(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user/self" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		writeTestJSON(t, w, map[string]any{"success": false, "message": "access token invalid"})
	}))
	defer server.Close()

	adapter := NewNewAPIAccountAdapter(server.Client(), time.Second)
	account := accountForAdapter(server.URL, domain.PlatformKindNewAPI)
	account.AccountCredentialKind = domain.CredentialKindNewAPIAccessToken
	result, err := adapter.TestAccountCredential(context.Background(), account, `{"access_token":"token-123","user_id":"42"}`)
	if err != nil {
		t.Fatalf("test account credential: %v", err)
	}
	if result.Status.AccountStatus != domain.AccountCredentialStatusExpired || result.Status.LastErrorClass != domain.UpstreamErrorCredentialExpired {
		t.Fatalf("unexpected account status: %+v", result.Status)
	}
}

func TestNewAPIAdapterRefreshQuotaUsesUserSelfQuotaMinusUsedQuota(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/user/self":
			if r.Header.Get("Authorization") != "token-123" || r.Header.Get("New-Api-User") != "42" {
				t.Fatalf("unexpected headers: authorization=%q user=%q", r.Header.Get("Authorization"), r.Header.Get("New-Api-User"))
			}
			writeTestJSON(t, w, map[string]any{"success": true, "data": map[string]any{"quota": 600000, "used_quota": 100000}})
		case "/api/status":
			writeTestJSON(t, w, map[string]any{"quota_per_unit": 500000})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	adapter := NewNewAPIAccountAdapter(server.Client(), time.Second)
	account := accountForAdapter(server.URL, domain.PlatformKindNewAPI)
	account.AccountCredentialKind = domain.CredentialKindNewAPIAccessToken
	result, err := adapter.RefreshQuota(context.Background(), account, "sk-unused", `{"access_token":"token-123","user_id":42}`)
	if err != nil {
		t.Fatalf("refresh quota: %v", err)
	}
	if result.BalanceAmount != 1 || result.BalanceUnit != "usd" {
		t.Fatalf("unexpected quota result: %+v", result)
	}
}

func TestNewAPIAdapterCheckinClassifiesTurnstileAndAuthBodies(t *testing.T) {
	tests := []struct {
		name              string
		body              map[string]any
		wantStatus        domain.UpstreamCheckinStatus
		wantAccountStatus domain.AccountCredentialStatus
		wantErrorClass    domain.UpstreamErrorClass
	}{
		{
			name:           "turnstile",
			body:           map[string]any{"success": false, "message": "turnstile required"},
			wantStatus:     domain.CheckinStatusActionRequired,
			wantErrorClass: domain.UpstreamErrorActionRequired,
		},
		{
			name:              "auth",
			body:              map[string]any{"success": false, "message": "access token invalid"},
			wantStatus:        domain.CheckinStatusActionRequired,
			wantAccountStatus: domain.AccountCredentialStatusExpired,
			wantErrorClass:    domain.UpstreamErrorCredentialExpired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/user/checkin" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				writeTestJSON(t, w, tt.body)
			}))
			defer server.Close()

			adapter := NewNewAPIAccountAdapter(server.Client(), time.Second)
			account := accountForAdapter(server.URL, domain.PlatformKindNewAPI)
			account.AccountCredentialKind = domain.CredentialKindNewAPIAccessToken
			result, err := adapter.Checkin(context.Background(), account, `{"access_token":"token-123","user_id":42}`)
			if err != nil {
				t.Fatalf("checkin: %v", err)
			}
			if result.Status != tt.wantStatus || result.AccountStatus != tt.wantAccountStatus || result.LastErrorClass != tt.wantErrorClass {
				t.Fatalf("unexpected checkin result: %+v", result)
			}
		})
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
	result, err := adapter.TestAccountCredential(context.Background(), accountForAdapter(server.URL, domain.PlatformKindNewAPI), "session=bad")
	if err != nil {
		t.Fatalf("test account credential: %v", err)
	}
	status := result.Status
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

func TestSub2APIAdapterRefreshesTokenForProfileAndReturnsRotation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/refresh":
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode refresh body: %v", err)
			}
			if payload["refresh_token"] != "rt_old" {
				t.Fatalf("unexpected refresh token: %+v", payload)
			}
			writeTestJSON(t, w, map[string]any{"code": 0, "message": "success", "data": map[string]any{"access_token": "jwt-new", "refresh_token": "rt_new"}})
		case "/api/v1/user/profile":
			if r.Header.Get("Authorization") != "Bearer jwt-new" {
				t.Fatalf("unexpected authorization: %q", r.Header.Get("Authorization"))
			}
			writeTestJSON(t, w, map[string]any{"id": 1, "email": "user@example.com"})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	adapter := NewSub2APIAccountAdapter(server.Client(), time.Second)
	account := accountForAdapter(server.URL, domain.PlatformKindSub2API)
	account.AccountCredentialKind = domain.CredentialKindSub2APIRefreshToken
	result, err := adapter.TestAccountCredential(context.Background(), account, `{"refresh_token":"rt_old"}`)
	if err != nil {
		t.Fatalf("test account credential: %v", err)
	}
	if result.Status.AccountStatus != domain.AccountCredentialStatusValid {
		t.Fatalf("unexpected account status: %+v", result.Status)
	}
	if result.CredentialUpdate == nil || result.CredentialUpdate.Kind != domain.CredentialKindSub2APIRefreshToken || !strings.Contains(result.CredentialUpdate.Plaintext, "rt_new") {
		t.Fatalf("expected rotated credential update, got %+v", result.CredentialUpdate)
	}
}

func TestSub2APIAdapterRefreshQuotaUsesPlatformQuotasWithBearer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/refresh":
			writeTestJSON(t, w, map[string]any{"code": 0, "data": map[string]any{"access_token": "jwt-new", "refresh_token": "rt_old"}})
		case "/api/v1/user/platform-quotas":
			if r.Header.Get("Authorization") != "Bearer jwt-new" {
				t.Fatalf("unexpected authorization: %q", r.Header.Get("Authorization"))
			}
			writeTestJSON(t, w, map[string]any{
				"code": 0,
				"data": map[string]any{
					"platform_quotas": []map[string]any{
						{"platform": "openai", "monthly_limit_usd": 100, "monthly_usage_usd": 30},
						{"platform": "anthropic", "monthly_limit_usd": 50, "monthly_usage_usd": 20},
					},
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	adapter := NewSub2APIAccountAdapter(server.Client(), time.Second)
	account := accountForAdapter(server.URL, domain.PlatformKindSub2API)
	account.AccountCredentialKind = domain.CredentialKindSub2APIRefreshToken
	result, err := adapter.RefreshQuota(context.Background(), account, "sk-unused", `{"refresh_token":"rt_old"}`)
	if err != nil {
		t.Fatalf("refresh quota: %v", err)
	}
	if result.BalanceAmount != 100 || result.BalanceUnit != "usd" {
		t.Fatalf("unexpected quota result: %+v", result)
	}
}

func TestSub2APIAdapterInvalidRefreshTokenExpiresCredential(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/refresh" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusUnauthorized)
		writeTestJSON(t, w, map[string]any{"message": "refresh token expired"})
	}))
	defer server.Close()

	adapter := NewSub2APIAccountAdapter(server.Client(), time.Second)
	account := accountForAdapter(server.URL, domain.PlatformKindSub2API)
	account.AccountCredentialKind = domain.CredentialKindSub2APIRefreshToken
	result, err := adapter.TestAccountCredential(context.Background(), account, `{"refresh_token":"rt_old"}`)
	if err != nil {
		t.Fatalf("test account credential: %v", err)
	}
	if result.Status.AccountStatus != domain.AccountCredentialStatusExpired || result.Status.LastErrorClass != domain.UpstreamErrorCredentialExpired {
		t.Fatalf("unexpected account status: %+v", result.Status)
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
	result, err := adapter.TestAccountCredential(context.Background(), accountForAdapter(server.URL, domain.PlatformKindSub2API), "jwt-token")
	if err != nil {
		t.Fatalf("test account credential: %v", err)
	}
	status := result.Status
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
