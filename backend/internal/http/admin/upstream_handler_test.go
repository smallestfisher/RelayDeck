package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/domain"
	"github.com/smallestfisher/relaydeck/backend/internal/secretbox"
	"github.com/smallestfisher/relaydeck/backend/internal/store"
	"github.com/smallestfisher/relaydeck/backend/internal/upstream"
)

func TestUpstreamAccountsRequireAdminSession(t *testing.T) {
	handler := newUpstreamTestHandler(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/upstreams/accounts", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateListAndUpdateUpstreamAccountRedactsSecrets(t *testing.T) {
	handler, sessions, adminStore, _ := newUpstreamTestHandlerParts(t)
	cookie := createTestSession(t, sessions, adminStore)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/accounts", strings.NewReader(`{
		"name":"New API Main",
		"code":"newapi-main",
		"platform_kind":"new_api",
		"base_url":"https://new-api.example.com",
		"enabled":true,
		"include_in_routing":true,
		"priority":10,
		"api_key":"sk-test-secret",
		"account_credential_kind":"cookie",
		"account_credential":"session=secret",
		"auto_sync_models":true,
		"auto_refresh_quota":true,
		"note":"primary"
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(cookie)
	createRec := httptest.NewRecorder()

	handler.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", createRec.Code, createRec.Body.String())
	}
	for _, leaked := range []string{"sk-test-secret", "session=secret"} {
		if strings.Contains(createRec.Body.String(), leaked) {
			t.Fatalf("response leaked secret %q: %s", leaked, createRec.Body.String())
		}
	}
	createdID := jsonField(t, createRec.Body.Bytes(), "id")
	stored, ok := adminStore.upstreams.UpstreamAccount(createdID)
	if !ok {
		t.Fatal("expected persisted account")
	}
	if stored.APIKeyEncrypted == "sk-test-secret" || stored.AccountCredentialEncrypted == "session=secret" {
		t.Fatalf("expected encrypted secrets, got %+v", stored)
	}
	if stored.APIKeyPrefix != "sk-test" {
		t.Fatalf("expected api key prefix, got %q", stored.APIKeyPrefix)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/admin/upstreams/accounts", nil)
	listReq.AddCookie(cookie)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d: %s", listRec.Code, listRec.Body.String())
	}
	if strings.Contains(listRec.Body.String(), "sk-test-secret") || strings.Contains(listRec.Body.String(), "session=secret") {
		t.Fatalf("list leaked secrets: %s", listRec.Body.String())
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/admin/upstreams/accounts/"+createdID, strings.NewReader(`{
		"name":"Renamed",
		"code":"newapi-main",
		"platform_kind":"new_api",
		"base_url":"https://new-api.example.com",
		"enabled":false,
		"include_in_routing":false,
		"priority":20,
		"account_credential_kind":"none",
		"note":"paused"
	}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.AddCookie(cookie)
	updateRec := httptest.NewRecorder()
	handler.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected update 200, got %d: %s", updateRec.Code, updateRec.Body.String())
	}
	updated, _ := adminStore.upstreams.UpstreamAccount(createdID)
	if updated.Name != "Renamed" || updated.Enabled {
		t.Fatalf("unexpected updated account: %+v", updated)
	}
	if updated.APIKeyEncrypted != stored.APIKeyEncrypted {
		t.Fatal("expected omitted api_key to preserve existing encrypted secret")
	}
}

func TestUpstreamActionEndpointsUpdateStatusModelsAndEvents(t *testing.T) {
	handler, sessions, adminStore, fakeStore := newUpstreamTestHandlerParts(t)
	cookie := createTestSession(t, sessions, adminStore)
	account := createStoredUpstreamAccount(t, fakeStore)

	actionReq := httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/accounts/"+account.ID+"/sync-models", nil)
	actionReq.AddCookie(cookie)
	actionRec := httptest.NewRecorder()
	handler.ServeHTTP(actionRec, actionReq)
	if actionRec.Code != http.StatusOK {
		t.Fatalf("expected action 200, got %d: %s", actionRec.Code, actionRec.Body.String())
	}
	status, ok := adminStore.upstreams.UpstreamAccountStatus(account.ID)
	if !ok || status.APIStatus != domain.UpstreamAPIStatusHealthy || status.ModelCount != 1 {
		t.Fatalf("unexpected status: %+v", status)
	}
	models := adminStore.upstreams.UpstreamModels(account.ID)
	if len(models) != 1 || models[0].NormalizedModelName != "gpt-4o-mini" {
		t.Fatalf("unexpected models: %+v", models)
	}
	events := adminStore.upstreams.UpstreamAccountEvents(account.ID, 10)
	if len(events) != 1 || events[0].Operation != "sync_models" || events[0].Status != "success" {
		t.Fatalf("expected sync event, got %+v", events)
	}
}

func TestBatchUpstreamActionReturnsPerAccountResults(t *testing.T) {
	handler, sessions, adminStore, fakeStore := newUpstreamTestHandlerParts(t)
	cookie := createTestSession(t, sessions, adminStore)
	account := createStoredUpstreamAccount(t, fakeStore)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/accounts/batch/test-api", strings.NewReader(`{"ids":["`+account.ID+`","missing"]}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected batch 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"id":"`+account.ID+`"`) || !strings.Contains(rec.Body.String(), `"id":"missing"`) {
		t.Fatalf("expected per-account results, got %s", rec.Body.String())
	}
}

func newUpstreamTestHandler(t *testing.T) http.Handler {
	t.Helper()
	handler, _, _, _ := newUpstreamTestHandlerParts(t)
	return handler
}

func newUpstreamTestHandlerParts(t *testing.T) (http.Handler, auth.SessionStore, *testAdminStore, *fakeUpstreamStore) {
	t.Helper()
	adminStore := newTestAdminStore(t)
	fakeStore := newFakeUpstreamStore()
	adminStore.upstreams = fakeStore
	sessions := auth.NewMemorySessionStore(fixedAdminNow)
	box, err := secretbox.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("secretbox: %v", err)
	}
	registry := upstream.NewAccountAdapterRegistry(map[domain.UpstreamPlatformKind]upstream.AccountAdapter{
		domain.PlatformKindNewAPI: fakeAccountAdapter{},
	})
	return NewWithDependencies(adminStore, sessions, fixedAdminNow, box, registry), sessions, adminStore, fakeStore
}

func createTestSession(t *testing.T, sessions auth.SessionStore, adminStore *testAdminStore) *http.Cookie {
	t.Helper()
	user, _ := adminStore.UserByEmail("owner@example.com")
	sessions.Create(auth.Session{
		Token:     "upstream-session",
		UserID:    user.ID,
		Email:     user.Email,
		Role:      string(user.Role),
		IssuedAt:  fixedAdminNow(),
		ExpiresAt: fixedAdminNow().Add(time.Hour),
	})
	return &http.Cookie{Name: "relaydeck_session", Value: "upstream-session"}
}

func createStoredUpstreamAccount(t *testing.T, upstreams *fakeUpstreamStore) domain.UpstreamAccount {
	t.Helper()
	account, err := upstreams.CreateUpstreamAccount(domain.UpstreamAccount{
		Name:                  "New API Main",
		Code:                  "newapi-main",
		PlatformKind:          domain.PlatformKindNewAPI,
		BaseURL:               "https://new-api.example.com",
		Enabled:               true,
		IncludeInRouting:      true,
		APIKeyEncrypted:       mustEncryptTestSecret(t, "sk-test-secret"),
		APIKeyPrefix:          "sk-test",
		AccountCredentialKind: domain.CredentialKindNone,
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	return account
}

func mustEncryptTestSecret(t *testing.T, plaintext string) string {
	t.Helper()
	box, err := secretbox.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("secretbox: %v", err)
	}
	encrypted, err := box.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return encrypted
}

func jsonField(t *testing.T, body []byte, field string) string {
	t.Helper()
	payload := map[string]any{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	value, _ := payload[field].(string)
	return value
}

type fakeAccountAdapter struct{}

func (fakeAccountAdapter) TestAPI(context.Context, domain.UpstreamAccount, string) (domain.UpstreamAccountStatus, error) {
	return domain.UpstreamAccountStatus{APIStatus: domain.UpstreamAPIStatusHealthy, CheckinStatus: domain.CheckinStatusUnsupported}, nil
}
func (fakeAccountAdapter) TestAccountCredential(context.Context, domain.UpstreamAccount, string) (domain.UpstreamAccountStatus, error) {
	return domain.UpstreamAccountStatus{AccountStatus: domain.AccountCredentialStatusValid}, nil
}
func (fakeAccountAdapter) SyncModels(context.Context, domain.UpstreamAccount, string) (domain.ModelSyncResult, domain.UpstreamAccountStatus, error) {
	models := []domain.UpstreamSyncedModel{{
		NormalizedModelName:    "gpt-4o-mini",
		UpstreamModelName:      "gpt-4o-mini",
		DisplayName:            "GPT-4o Mini",
		NativeWireProtocol:     domain.ProtocolOpenAIChat,
		SupportedWireProtocols: []domain.Protocol{domain.ProtocolOpenAIChat},
		Capabilities:           []domain.Capability{domain.CapabilityChat},
		Status:                 "active",
	}}
	return domain.ModelSyncResult{AccountID: "acct", CreatedModels: 1, UpdatedModels: 1, Models: models},
		domain.UpstreamAccountStatus{APIStatus: domain.UpstreamAPIStatusHealthy, ModelCount: 1, CheckinStatus: domain.CheckinStatusUnsupported}, nil
}
func (fakeAccountAdapter) RefreshQuota(context.Context, domain.UpstreamAccount, string, string) (domain.QuotaRefreshResult, error) {
	status := domain.UpstreamAccountStatus{APIStatus: domain.UpstreamAPIStatusHealthy, BalanceAmount: 12, BalanceUnit: "quota"}
	return domain.QuotaRefreshResult{Status: status, BalanceAmount: 12, BalanceUnit: "quota"}, nil
}
func (fakeAccountAdapter) Checkin(context.Context, domain.UpstreamAccount, string) (domain.CheckinResult, error) {
	return domain.CheckinResult{Status: domain.CheckinStatusUnsupported}, nil
}

type fakeUpstreamStore struct {
	accounts map[string]domain.UpstreamAccount
	statuses map[string]domain.UpstreamAccountStatus
	models   map[string][]domain.UpstreamSyncedModel
	events   map[string][]domain.UpstreamAccountEvent
	nextID   int
}

func newFakeUpstreamStore() *fakeUpstreamStore {
	return &fakeUpstreamStore{
		accounts: map[string]domain.UpstreamAccount{},
		statuses: map[string]domain.UpstreamAccountStatus{},
		models:   map[string][]domain.UpstreamSyncedModel{},
		events:   map[string][]domain.UpstreamAccountEvent{},
	}
}

func (s *fakeUpstreamStore) ListUpstreamAccounts() []domain.UpstreamAccount {
	accounts := make([]domain.UpstreamAccount, 0, len(s.accounts))
	for _, account := range s.accounts {
		accounts = append(accounts, account)
	}
	return accounts
}
func (s *fakeUpstreamStore) UpstreamAccount(id string) (domain.UpstreamAccount, bool) {
	account, ok := s.accounts[id]
	return account, ok
}
func (s *fakeUpstreamStore) CreateUpstreamAccount(account domain.UpstreamAccount) (domain.UpstreamAccount, error) {
	s.nextID++
	if account.ID == "" {
		account.ID = "acct_test_" + string(rune('0'+s.nextID))
	}
	account.CreatedAt = fixedAdminNow()
	account.UpdatedAt = fixedAdminNow()
	s.accounts[account.ID] = account
	return account, nil
}
func (s *fakeUpstreamStore) UpdateUpstreamAccount(account domain.UpstreamAccount) (domain.UpstreamAccount, error) {
	existing, ok := s.accounts[account.ID]
	if !ok {
		return domain.UpstreamAccount{}, store.ErrNotFound
	}
	account.CreatedAt = existing.CreatedAt
	account.UpdatedAt = fixedAdminNow()
	s.accounts[account.ID] = account
	return account, nil
}
func (s *fakeUpstreamStore) DeleteUpstreamAccount(id string) error {
	delete(s.accounts, id)
	delete(s.statuses, id)
	delete(s.models, id)
	delete(s.events, id)
	return nil
}
func (s *fakeUpstreamStore) UpstreamAccountStatus(accountID string) (domain.UpstreamAccountStatus, bool) {
	status, ok := s.statuses[accountID]
	return status, ok
}
func (s *fakeUpstreamStore) UpsertUpstreamAccountStatus(status domain.UpstreamAccountStatus) error {
	s.statuses[status.UpstreamAccountID] = status
	return nil
}
func (s *fakeUpstreamStore) UpstreamModels(accountID string) []domain.UpstreamSyncedModel {
	return append([]domain.UpstreamSyncedModel(nil), s.models[accountID]...)
}
func (s *fakeUpstreamStore) ReplaceUpstreamModels(accountID string, models []domain.UpstreamSyncedModel) error {
	for i := range models {
		models[i].ID = "model_test"
		models[i].UpstreamAccountID = accountID
	}
	s.models[accountID] = append([]domain.UpstreamSyncedModel(nil), models...)
	return nil
}
func (s *fakeUpstreamStore) UpstreamAccountEvents(accountID string, limit int) []domain.UpstreamAccountEvent {
	events := append([]domain.UpstreamAccountEvent(nil), s.events[accountID]...)
	if limit > 0 && len(events) > limit {
		return events[:limit]
	}
	return events
}
func (s *fakeUpstreamStore) AppendUpstreamAccountEvent(event domain.UpstreamAccountEvent) error {
	event.ID = "event_test"
	event.CreatedAt = fixedAdminNow()
	s.events[event.UpstreamAccountID] = append([]domain.UpstreamAccountEvent{event}, s.events[event.UpstreamAccountID]...)
	return nil
}
