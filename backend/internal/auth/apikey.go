package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"slices"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

var (
	ErrInvalidKey    = errors.New("invalid api key")
	ErrInactiveKey   = errors.New("api key is not active")
	ErrExpiredKey    = errors.New("api key is expired")
	ErrInactiveOwner = errors.New("api key owner is disabled")
	ErrScopeDenied   = errors.New("api key scope does not allow this endpoint")
	ErrModelDenied   = errors.New("api key does not allow this model")
)

func HashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func VerifyGatewayKey(secret string, key domain.APIKey, req domain.GatewayRequest, now time.Time) (domain.GatewayPrincipal, error) {
	if HashSecret(secret) != key.Hash {
		return domain.GatewayPrincipal{}, ErrInvalidKey
	}
	if key.Status != domain.APIKeyStatusActive {
		return domain.GatewayPrincipal{}, ErrInactiveKey
	}
	if !key.ExpiresAt.IsZero() && !now.Before(key.ExpiresAt) {
		return domain.GatewayPrincipal{}, ErrExpiredKey
	}
	if !key.OwnerIsActive {
		return domain.GatewayPrincipal{}, ErrInactiveOwner
	}
	if !hasScope(key.Scopes, scopeForEndpoint(req.Endpoint)) {
		return domain.GatewayPrincipal{}, ErrScopeDenied
	}
	if len(key.AllowedModels) > 0 && !slices.Contains(key.AllowedModels, req.Model) {
		return domain.GatewayPrincipal{}, ErrModelDenied
	}

	return domain.GatewayPrincipal{
		APIKeyID: key.ID,
		UserID:   key.UserID,
		Scopes:   key.Scopes,
	}, nil
}

func scopeForEndpoint(endpoint domain.EndpointType) domain.Scope {
	switch endpoint {
	case domain.EndpointResponses:
		return domain.ScopeResponses
	case domain.EndpointEmbeddings:
		return domain.ScopeEmbeddings
	default:
		return domain.ScopeChatCompletions
	}
}

func hasScope(scopes []domain.Scope, target domain.Scope) bool {
	return slices.Contains(scopes, target) || slices.Contains(scopes, domain.ScopeAdmin)
}
