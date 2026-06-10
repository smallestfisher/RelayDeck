package store

import (
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

const DevAPIKeySecret = "rd_live_dev_test_secret"

const (
	DefaultOwnerEmail    = "owner@example.com"
	DefaultOwnerPassword = "change-me"
)

type MemoryStore struct {
	users    []domain.User
	apiKeys  []domain.APIKey
	models   []domain.Model
	sites    []domain.UpstreamSite
	mappings []domain.SiteModel
}

func mustHashPassword(password string) string {
	hash, err := auth.HashPassword(password)
	if err != nil {
		panic(err)
	}
	return hash
}

func NewMemoryStore() *MemoryStore {
	return NewMemoryStoreWithBootstrap(DefaultOwnerEmail, DefaultOwnerPassword)
}

func NewMemoryStoreWithBootstrap(ownerEmail string, ownerPassword string) *MemoryStore {
	if ownerEmail == "" {
		ownerEmail = DefaultOwnerEmail
	}
	if ownerPassword == "" {
		ownerPassword = DefaultOwnerPassword
	}
	now := time.Now()
	return &MemoryStore{
		users: []domain.User{
			{
				ID:           "user_admin",
				Email:        ownerEmail,
				PasswordHash: mustHashPassword(ownerPassword),
				Role:         domain.UserRoleOwner,
				Status:       domain.UserStatusActive,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		},
		apiKeys: []domain.APIKey{
			{
				ID:              "key_dev",
				UserID:          "user_admin",
				Prefix:          "rd_live_dev",
				Hash:            auth.HashSecret(DevAPIKeySecret),
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

func (s *MemoryStore) APIKeys() []domain.APIKey {
	return append([]domain.APIKey(nil), s.apiKeys...)
}

func (s *MemoryStore) Users() []domain.User {
	return append([]domain.User(nil), s.users...)
}

func (s *MemoryStore) UserByEmail(email string) (domain.User, bool) {
	for _, user := range s.users {
		if user.Email == email {
			return user, true
		}
	}
	return domain.User{}, false
}

func (s *MemoryStore) UserByID(id string) (domain.User, bool) {
	for _, user := range s.users {
		if user.ID == id {
			return user, true
		}
	}
	return domain.User{}, false
}

func (s *MemoryStore) Models() []domain.Model {
	return append([]domain.Model(nil), s.models...)
}

func (s *MemoryStore) Sites() []domain.UpstreamSite {
	return append([]domain.UpstreamSite(nil), s.sites...)
}

func (s *MemoryStore) Mappings() []domain.SiteModel {
	return append([]domain.SiteModel(nil), s.mappings...)
}
