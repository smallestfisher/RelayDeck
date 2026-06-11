package store

import "github.com/smallestfisher/relaydeck/backend/internal/domain"

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
