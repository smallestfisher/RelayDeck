package auth

import (
	"testing"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

func TestVerifyGatewayKey(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	secret := "rd_live_dev_test_secret"
	key := domain.APIKey{
		ID:              "key_1",
		UserID:          "user_1",
		Prefix:          "rd_live_dev",
		Hash:            HashSecret(secret),
		Status:          domain.APIKeyStatusActive,
		Scopes:          []domain.Scope{domain.ScopeChatCompletions, domain.ScopeResponses},
		AllowedModels:   []string{"gpt-4o-mini"},
		ExpiresAt:       now.Add(24 * time.Hour),
		OwnerIsActive:   true,
		RPM:             60,
		TPM:             10000,
		AllowedCIDRs:    []string{"0.0.0.0/0"},
		MonthlyQuotaTPM: 1000000,
	}

	tests := []struct {
		name    string
		secret  string
		key     domain.APIKey
		req     domain.GatewayRequest
		wantErr bool
	}{
		{
			name:   "accepts valid key",
			secret: secret,
			key:    key,
			req: domain.GatewayRequest{
				Endpoint: domain.EndpointChatCompletions,
				Model:    "gpt-4o-mini",
			},
		},
		{
			name:    "rejects invalid secret",
			secret:  "rd_live_wrong",
			key:     key,
			req:     domain.GatewayRequest{Endpoint: domain.EndpointChatCompletions, Model: "gpt-4o-mini"},
			wantErr: true,
		},
		{
			name:   "rejects inactive key",
			secret: secret,
			key: func() domain.APIKey {
				copy := key
				copy.Status = domain.APIKeyStatusRevoked
				return copy
			}(),
			req:     domain.GatewayRequest{Endpoint: domain.EndpointChatCompletions, Model: "gpt-4o-mini"},
			wantErr: true,
		},
		{
			name:   "rejects expired key",
			secret: secret,
			key: func() domain.APIKey {
				copy := key
				copy.ExpiresAt = now.Add(-time.Hour)
				return copy
			}(),
			req:     domain.GatewayRequest{Endpoint: domain.EndpointChatCompletions, Model: "gpt-4o-mini"},
			wantErr: true,
		},
		{
			name:    "rejects disallowed model",
			secret:  secret,
			key:     key,
			req:     domain.GatewayRequest{Endpoint: domain.EndpointChatCompletions, Model: "gpt-4o"},
			wantErr: true,
		},
		{
			name:    "rejects missing scope",
			secret:  secret,
			key:     key,
			req:     domain.GatewayRequest{Endpoint: domain.EndpointEmbeddings, Model: "gpt-4o-mini"},
			wantErr: true,
		},
		{
			name:   "rejects disabled owner",
			secret: secret,
			key: func() domain.APIKey {
				copy := key
				copy.OwnerIsActive = false
				return copy
			}(),
			req:     domain.GatewayRequest{Endpoint: domain.EndpointChatCompletions, Model: "gpt-4o-mini"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			principal, err := VerifyGatewayKey(tt.secret, tt.key, tt.req, now)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if principal.APIKeyID != tt.key.ID || principal.UserID != tt.key.UserID {
				t.Fatalf("unexpected principal: %+v", principal)
			}
		})
	}
}
