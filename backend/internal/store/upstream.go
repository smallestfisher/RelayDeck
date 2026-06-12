package store

import (
	"errors"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

var ErrNotFound = errors.New("not found")

// UpstreamAccountFilter describes server-side filtering and pagination for the
// account list. Empty string fields mean "no constraint" for that dimension.
type UpstreamAccountFilter struct {
	Query         string
	PlatformKind  string
	APIStatus     string
	AccountStatus string
	LatencyBand   string // "", "low", "medium", "high", "unknown"
	Limit         int
	Offset        int
}

// UpstreamAccountMetrics aggregates account counts across the whole filtered
// set, independent of pagination. Total matches the filtered row count.
type UpstreamAccountMetrics struct {
	Total   int `json:"total"`
	Healthy int `json:"healthy"`
	Warning int `json:"warning"`
	Manual  int `json:"manual"`
}

type UpstreamAccountStore interface {
	ListUpstreamAccounts() []domain.UpstreamAccount
	UpstreamAccount(id string) (domain.UpstreamAccount, bool)
	CreateUpstreamAccount(account domain.UpstreamAccount) (domain.UpstreamAccount, error)
	UpdateUpstreamAccount(account domain.UpstreamAccount) (domain.UpstreamAccount, error)
	DeleteUpstreamAccount(id string) error

	UpstreamAccountStatus(accountID string) (domain.UpstreamAccountStatus, bool)
	UpsertUpstreamAccountStatus(status domain.UpstreamAccountStatus) error

	UpstreamModels(accountID string) []domain.UpstreamSyncedModel
	ReplaceUpstreamModels(accountID string, models []domain.UpstreamSyncedModel) error

	UpstreamAccountEvents(accountID string, limit int) []domain.UpstreamAccountEvent
	AppendUpstreamAccountEvent(event domain.UpstreamAccountEvent) error
}
