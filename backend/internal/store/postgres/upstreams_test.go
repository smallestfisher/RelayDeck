package postgres

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

func TestUpstreamStorePersistsAccountStateModelsAndEvents(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}
	if !strings.Contains(databaseURL, "test") {
		t.Fatalf("DATABASE_URL must point to a test database, got %q", databaseURL)
	}

	ctx := context.Background()
	db, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(ctx, `
DROP TABLE IF EXISTS upstream_account_events;
DROP TABLE IF EXISTS upstream_synced_models;
DROP TABLE IF EXISTS upstream_account_status;
DROP TABLE IF EXISTS upstream_accounts;
`); err != nil {
		t.Fatalf("reset upstream tables: %v", err)
	}

	upstreams := NewUpstreamStore(db)
	if err := upstreams.EnsureSchema(ctx); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	runUpstreamStoreBehavior(t, upstreams)
}

type upstreamStoreContract interface {
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

func runUpstreamStoreBehavior(t *testing.T, upstreams upstreamStoreContract) {
	t.Helper()

	account := domain.UpstreamAccount{
		Name:                       "New API Account",
		Code:                       "newapi-main",
		PlatformKind:               domain.PlatformKindNewAPI,
		BaseURL:                    "https://new-api.example.com",
		Enabled:                    true,
		IncludeInRouting:           true,
		Priority:                   10,
		APIKeyEncrypted:            "encrypted-api-key",
		APIKeyPrefix:               "sk-live",
		AccountCredentialKind:      domain.CredentialKindCookie,
		AccountCredentialEncrypted: "encrypted-cookie",
		AutoSyncModels:             true,
		AutoRefreshQuota:           true,
		Note:                       "primary",
	}

	created, err := upstreams.CreateUpstreamAccount(account)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected generated account id")
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatal("expected timestamps")
	}

	listed := upstreams.ListUpstreamAccounts()
	if len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("expected created account in list, got %+v", listed)
	}

	status := domain.UpstreamAccountStatus{
		UpstreamAccountID: created.ID,
		APIStatus:         domain.UpstreamAPIStatusHealthy,
		AccountStatus:     domain.AccountCredentialStatusValid,
		CheckinStatus:     domain.CheckinStatusUnsupported,
		ModelCount:        1,
		LatencyMS:         88,
	}
	if err := upstreams.UpsertUpstreamAccountStatus(status); err != nil {
		t.Fatalf("upsert status: %v", err)
	}
	storedStatus, ok := upstreams.UpstreamAccountStatus(created.ID)
	if !ok {
		t.Fatal("expected account status")
	}
	if storedStatus.APIStatus != domain.UpstreamAPIStatusHealthy || storedStatus.LatencyMS != 88 {
		t.Fatalf("unexpected status: %+v", storedStatus)
	}

	models := []domain.UpstreamSyncedModel{{
		NormalizedModelName:    "gpt-4o-mini",
		UpstreamModelName:      "gpt-4o-mini",
		DisplayName:            "GPT-4o Mini",
		NativeWireProtocol:     domain.ProtocolOpenAIChat,
		SupportedWireProtocols: []domain.Protocol{domain.ProtocolOpenAIChat},
		Capabilities:           []domain.Capability{domain.CapabilityChat, domain.CapabilityStreaming},
		Status:                 "active",
		RawMetadata:            map[string]any{"source": "test"},
	}}
	if err := upstreams.ReplaceUpstreamModels(created.ID, models); err != nil {
		t.Fatalf("replace models: %v", err)
	}
	storedModels := upstreams.UpstreamModels(created.ID)
	if len(storedModels) != 1 {
		t.Fatalf("expected one model, got %+v", storedModels)
	}
	if storedModels[0].ID == "" || storedModels[0].UpstreamAccountID != created.ID {
		t.Fatalf("expected model id and account id, got %+v", storedModels[0])
	}

	if err := upstreams.AppendUpstreamAccountEvent(domain.UpstreamAccountEvent{
		UpstreamAccountID: created.ID,
		Operation:         "sync_models",
		Status:            "success",
		Message:           "synced",
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}
	events := upstreams.UpstreamAccountEvents(created.ID, 10)
	if len(events) != 1 || events[0].ID == "" || events[0].Operation != "sync_models" {
		t.Fatalf("expected event history, got %+v", events)
	}

	created.Name = "Renamed Account"
	created.Priority = 20
	updated, err := upstreams.UpdateUpstreamAccount(created)
	if err != nil {
		t.Fatalf("update account: %v", err)
	}
	if updated.Name != "Renamed Account" || updated.Priority != 20 {
		t.Fatalf("unexpected updated account: %+v", updated)
	}

	if err := upstreams.DeleteUpstreamAccount(created.ID); err != nil {
		t.Fatalf("delete account: %v", err)
	}
	if _, ok := upstreams.UpstreamAccount(created.ID); ok {
		t.Fatal("expected account to be deleted")
	}
	if models := upstreams.UpstreamModels(created.ID); len(models) != 0 {
		t.Fatalf("expected account models to be deleted, got %+v", models)
	}
}
