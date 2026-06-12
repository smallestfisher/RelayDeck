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
	if stored.Code == "" {
		t.Fatal("expected generated internal site code")
	}
	var created struct {
		Status domain.UpstreamAccountStatus `json:"status"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created account: %v", err)
	}
	if created.Status.APIStatus != domain.UpstreamAPIStatusUnknown || created.Status.AccountStatus != domain.AccountCredentialStatusActionRequired {
		t.Fatalf("expected new account to start unknown/action_required, got %+v", created.Status)
	}
	if created.Status.ModelCount != 0 || created.Status.LatencyMS != 0 || created.Status.BalanceUnit != "" {
		t.Fatalf("expected new account metrics to be unset, got %+v", created.Status)
	}
	if !created.Status.LastAPICheckedAt.IsZero() || !created.Status.LastModelSyncedAt.IsZero() {
		t.Fatalf("expected new account timestamps to be unset, got %+v", created.Status)
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

func TestListUpstreamAccountsSupportsLimitOffsetAndTotal(t *testing.T) {
	handler, sessions, adminStore, fakeStore := newUpstreamTestHandlerParts(t)
	cookie := createTestSession(t, sessions, adminStore)
	for _, name := range []string{"Account 1", "Account 2", "Account 3"} {
		_, err := fakeStore.CreateUpstreamAccount(domain.UpstreamAccount{
			Name:             name,
			Code:             strings.ToLower(strings.ReplaceAll(name, " ", "-")),
			PlatformKind:     domain.PlatformKindNewAPI,
			BaseURL:          "https://new-api.example.com",
			Enabled:          true,
			IncludeInRouting: true,
			APIKeyEncrypted:  mustEncryptTestSecret(t, "sk-test-secret"),
			APIKeyPrefix:     "sk-test",
		})
		if err != nil {
			t.Fatalf("create account: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/upstreams/accounts?limit=2&offset=1", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Items  []upstreamAccountResponse `json:"items"`
		Total  int                       `json:"total"`
		Limit  int                       `json:"limit"`
		Offset int                       `json:"offset"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Total != 3 || payload.Limit != 2 || payload.Offset != 1 {
		t.Fatalf("unexpected pagination metadata: %+v", payload)
	}
	if len(payload.Items) != 2 {
		t.Fatalf("expected two paginated accounts, got %d", len(payload.Items))
	}
}

func TestDraftUpstreamAPITestDoesNotPersistAccount(t *testing.T) {
	handler, sessions, adminStore, fakeStore := newUpstreamTestHandlerParts(t)
	cookie := createTestSession(t, sessions, adminStore)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/test", strings.NewReader(`{
		"platform_kind":"new_api",
		"base_url":"https://new-api.example.com/",
		"api_key":"sk-draft-secret"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected draft test 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result actionResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Status != "success" {
		t.Fatalf("expected successful draft test, got %+v", result)
	}
	if len(fakeStore.accounts) != 0 {
		t.Fatalf("expected draft test not to persist accounts, got %+v", fakeStore.accounts)
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

func TestTestCallUpdatesAPIStatus(t *testing.T) {
	handler, sessions, adminStore, fakeStore := newUpstreamTestHandlerParts(t)
	cookie := createTestSession(t, sessions, adminStore)
	account := createStoredUpstreamAccount(t, fakeStore)
	if err := fakeStore.ReplaceUpstreamModels(account.ID, []domain.UpstreamSyncedModel{{
		NormalizedModelName:    "gpt-4o-mini",
		UpstreamModelName:      "gpt-4o-mini",
		NativeWireProtocol:     domain.ProtocolOpenAIChat,
		SupportedWireProtocols: []domain.Protocol{domain.ProtocolOpenAIChat, domain.ProtocolOpenAIResponses},
	}}); err != nil {
		t.Fatalf("seed models: %v", err)
	}
	if err := fakeStore.UpsertUpstreamAccountStatus(domain.UpstreamAccountStatus{
		UpstreamAccountID: account.ID,
		APIStatus:         domain.UpstreamAPIStatusUnknown,
		AccountStatus:     domain.AccountCredentialStatusActionRequired,
		CheckinStatus:     domain.CheckinStatusUnsupported,
		ModelCount:        3,
		LatencyMS:         188,
		LastModelSyncedAt: fixedAdminNow(),
		UpdatedAt:         fixedAdminNow(),
	}); err != nil {
		t.Fatalf("seed status: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/accounts/"+account.ID+"/test-call", strings.NewReader(`{"model_name":"gpt-4o-mini","protocol":"auto","message":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected test call 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var response testCallResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.OK || response.Protocol != "openai-chat" || response.HTTPStatus != http.StatusOK {
		t.Fatalf("unexpected test response: %+v", response)
	}
	status, ok := fakeStore.UpstreamAccountStatus(account.ID)
	if !ok || status.APIStatus != domain.UpstreamAPIStatusHealthy || status.APILatencyMS != 42 || status.LastAPICheckedAt.IsZero() {
		t.Fatalf("expected healthy API status, got %+v", status)
	}
	if status.LatencyMS != 188 || status.ModelCount != 3 || status.LastModelSyncedAt.IsZero() {
		t.Fatalf("expected test call to preserve site latency and model sync status, got %+v", status)
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

func TestUpstreamActionResponseIncludesAccountStatusSnapshot(t *testing.T) {
	handler, sessions, adminStore, fakeStore := newUpstreamTestHandlerParts(t)
	cookie := createTestSession(t, sessions, adminStore)
	account := createStoredUpstreamAccount(t, fakeStore)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/accounts/"+account.ID+"/test-account", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected action 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result actionResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode action response: %v", err)
	}
	if result.AccountStatus == nil || result.AccountStatus.AccountStatus != domain.AccountCredentialStatusValid {
		t.Fatalf("expected account status snapshot, got %+v", result)
	}

	batchReq := httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/accounts/batch/test-api", strings.NewReader(`{"ids":["`+account.ID+`"]}`))
	batchReq.Header.Set("Content-Type", "application/json")
	batchReq.AddCookie(cookie)
	batchRec := httptest.NewRecorder()
	handler.ServeHTTP(batchRec, batchReq)
	if batchRec.Code != http.StatusOK {
		t.Fatalf("expected batch 200, got %d: %s", batchRec.Code, batchRec.Body.String())
	}
	var batchPayload struct {
		Results []actionResult `json:"results"`
	}
	if err := json.Unmarshal(batchRec.Body.Bytes(), &batchPayload); err != nil {
		t.Fatalf("decode batch response: %v", err)
	}
	if len(batchPayload.Results) != 1 || batchPayload.Results[0].AccountStatus == nil || batchPayload.Results[0].AccountStatus.APIStatus != domain.UpstreamAPIStatusHealthy {
		t.Fatalf("expected batch account status snapshot, got %+v", batchPayload)
	}
}

func TestRefreshAllRoutesAreMounted(t *testing.T) {
	handler, sessions, adminStore, fakeStore := newUpstreamTestHandlerParts(t)
	cookie := createTestSession(t, sessions, adminStore)
	account := createStoredUpstreamAccount(t, fakeStore)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/accounts/"+account.ID+"/refresh-all", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected single refresh-all 200, got %d: %s", rec.Code, rec.Body.String())
	}

	batchReq := httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/accounts/batch/refresh-all", strings.NewReader(`{"ids":["`+account.ID+`"]}`))
	batchReq.Header.Set("Content-Type", "application/json")
	batchReq.AddCookie(cookie)
	batchRec := httptest.NewRecorder()
	handler.ServeHTTP(batchRec, batchReq)
	if batchRec.Code != http.StatusOK {
		t.Fatalf("expected batch refresh-all 200, got %d: %s", batchRec.Code, batchRec.Body.String())
	}
}

func TestRefreshAllActionPersistsQuotaAndModels(t *testing.T) {
	handler, sessions, adminStore, fakeStore := newUpstreamTestHandlerParts(t)
	cookie := createTestSession(t, sessions, adminStore)
	account := createStoredUpstreamAccount(t, fakeStore)
	if err := fakeStore.UpsertUpstreamAccountStatus(domain.UpstreamAccountStatus{
		UpstreamAccountID: account.ID,
		APIStatus:         domain.UpstreamAPIStatusHealthy,
		ModelCount:        7,
		BalanceAmount:     2,
		BalanceUnit:       "usd",
		UpdatedAt:         fixedAdminNow(),
	}); err != nil {
		t.Fatalf("seed status: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/accounts/"+account.ID+"/refresh-all", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected refresh-all 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result actionResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode action response: %v", err)
	}
	if result.AccountStatus == nil || result.AccountStatus.ModelCount != 1 || result.AccountStatus.BalanceAmount != 12 || result.AccountStatus.BalanceUnit != "quota" {
		t.Fatalf("expected fresh quota and model status in response, got %+v", result.AccountStatus)
	}
	status, ok := fakeStore.UpstreamAccountStatus(account.ID)
	if !ok {
		t.Fatalf("expected stored status")
	}
	if status.ModelCount != 1 || status.BalanceAmount != 12 || status.BalanceUnit != "quota" {
		t.Fatalf("expected fresh quota and model status in store, got %+v", status)
	}
	models := fakeStore.UpstreamModels(account.ID)
	if len(models) != 1 || models[0].NormalizedModelName != "gpt-4o-mini" {
		t.Fatalf("expected synced models, got %+v", models)
	}
}

func TestRefreshAllKeepsCredentialStatusSeparateFromQuotaAndUsesAPIResult(t *testing.T) {
	handler, sessions, adminStore, fakeStore := newUpstreamTestHandlerPartsWithAdapter(t, quotaHealthySyncFailedAdapter{})
	cookie := createTestSession(t, sessions, adminStore)
	account := createStoredUpstreamAccount(t, fakeStore)
	if err := fakeStore.UpsertUpstreamAccountStatus(domain.UpstreamAccountStatus{
		UpstreamAccountID: account.ID,
		APIStatus:         domain.UpstreamAPIStatusHealthy,
		AccountStatus:     domain.AccountCredentialStatusActionRequired,
		APILatencyMS:      999,
		UpdatedAt:         fixedAdminNow(),
	}); err != nil {
		t.Fatalf("seed status: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/accounts/"+account.ID+"/refresh-all", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected refresh-all 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result actionResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode action response: %v", err)
	}
	if result.AccountStatus == nil {
		t.Fatalf("expected status snapshot")
	}
	if result.AccountStatus.APIStatus != domain.UpstreamAPIStatusHealthy || result.AccountStatus.APILatencyMS != 999 {
		t.Fatalf("expected refresh-all to preserve API test status, got %+v", result.AccountStatus)
	}
	if result.AccountStatus.AccountStatus != domain.AccountCredentialStatusActionRequired {
		t.Fatalf("expected quota refresh not to overwrite credential status, got %+v", result.AccountStatus)
	}
	if result.AccountStatus.BalanceAmount != 12 || result.AccountStatus.BalanceUnit != "usd" {
		t.Fatalf("expected quota to be refreshed, got %+v", result.AccountStatus)
	}
}

func TestUpstreamActionPersistsRotatedAccountCredential(t *testing.T) {
	handler, sessions, adminStore, fakeStore := newUpstreamTestHandlerPartsWithAdapter(t, rotatingAccountAdapter{})
	cookie := createTestSession(t, sessions, adminStore)
	account := createStoredUpstreamAccount(t, fakeStore)
	account.AccountCredentialKind = domain.CredentialKindSub2APIRefreshToken
	account.AccountCredentialEncrypted = mustEncryptTestSecret(t, `{"refresh_token":"rt_old"}`)
	if _, err := fakeStore.UpdateUpstreamAccount(account); err != nil {
		t.Fatalf("update test account: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/upstreams/accounts/"+account.ID+"/test-account", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected action 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "rt_new") {
		t.Fatalf("response leaked rotated credential: %s", rec.Body.String())
	}
	updated, _ := fakeStore.UpstreamAccount(account.ID)
	if updated.AccountCredentialKind != domain.CredentialKindSub2APIRefreshToken {
		t.Fatalf("unexpected credential kind: %s", updated.AccountCredentialKind)
	}
	decrypted := mustDecryptTestSecret(t, updated.AccountCredentialEncrypted)
	if !strings.Contains(decrypted, "rt_new") {
		t.Fatalf("expected rotated credential to be persisted, got %q", decrypted)
	}
	for _, event := range fakeStore.UpstreamAccountEvents(account.ID, 10) {
		if strings.Contains(event.Message, "rt_new") {
			t.Fatalf("event leaked rotated credential: %+v", event)
		}
	}
}

func newUpstreamTestHandler(t *testing.T) http.Handler {
	t.Helper()
	handler, _, _, _ := newUpstreamTestHandlerParts(t)
	return handler
}

func newUpstreamTestHandlerParts(t *testing.T) (http.Handler, auth.SessionStore, *testAdminStore, *fakeUpstreamStore) {
	t.Helper()
	return newUpstreamTestHandlerPartsWithAdapter(t, fakeAccountAdapter{})
}

func newUpstreamTestHandlerPartsWithAdapter(t *testing.T, adapter upstream.AccountAdapter) (http.Handler, auth.SessionStore, *testAdminStore, *fakeUpstreamStore) {
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
		domain.PlatformKindNewAPI:  adapter,
		domain.PlatformKindSub2API: adapter,
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

func mustDecryptTestSecret(t *testing.T, ciphertext string) string {
	t.Helper()
	box, err := secretbox.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("secretbox: %v", err)
	}
	plaintext, err := box.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	return plaintext
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
func (fakeAccountAdapter) TestAccountCredential(context.Context, domain.UpstreamAccount, string) (domain.AccountCredentialTestResult, error) {
	return domain.AccountCredentialTestResult{Status: domain.UpstreamAccountStatus{AccountStatus: domain.AccountCredentialStatusValid}}, nil
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

func (fakeAccountAdapter) TestCall(context.Context, domain.UpstreamAccount, string, string, string, bool, string) (domain.UpstreamTestCallResult, error) {
	return domain.UpstreamTestCallResult{HTTPStatus: http.StatusOK, Protocol: "openai-chat", OK: true, LatencyMS: 42}, nil
}

type rotatingAccountAdapter struct {
	fakeAccountAdapter
}

func (rotatingAccountAdapter) TestAccountCredential(context.Context, domain.UpstreamAccount, string) (domain.AccountCredentialTestResult, error) {
	return domain.AccountCredentialTestResult{
		Status: domain.UpstreamAccountStatus{AccountStatus: domain.AccountCredentialStatusValid},
		CredentialUpdate: &domain.UpstreamCredentialUpdate{
			Kind:      domain.CredentialKindSub2APIRefreshToken,
			Plaintext: `{"refresh_token":"rt_new"}`,
		},
	}, nil
}

type quotaHealthySyncFailedAdapter struct {
	fakeAccountAdapter
}

func (quotaHealthySyncFailedAdapter) SyncModels(context.Context, domain.UpstreamAccount, string) (domain.ModelSyncResult, domain.UpstreamAccountStatus, error) {
	return domain.ModelSyncResult{AccountID: "acct"},
		domain.UpstreamAccountStatus{
			APIStatus:         domain.UpstreamAPIStatusFailed,
			AccountStatus:     domain.AccountCredentialStatusNotConfigured,
			CheckinStatus:     domain.CheckinStatusUnsupported,
			ModelCount:        0,
			LatencyMS:         321,
			LastAPICheckedAt:  fixedAdminNow(),
			LastErrorClass:    domain.UpstreamErrorAuthError,
			LastErrorMessage:  "api test failed",
			LastModelSyncedAt: fixedAdminNow(),
		}, nil
}

func (quotaHealthySyncFailedAdapter) RefreshQuota(context.Context, domain.UpstreamAccount, string, string) (domain.QuotaRefreshResult, error) {
	status := domain.UpstreamAccountStatus{
		APIStatus:            domain.UpstreamAPIStatusHealthy,
		AccountStatus:        domain.AccountCredentialStatusNotConfigured,
		BalanceAmount:        12,
		BalanceUnit:          "usd",
		LastAPICheckedAt:     fixedAdminNow(),
		LastAccountCheckedAt: fixedAdminNow(),
	}
	return domain.QuotaRefreshResult{Status: status, BalanceAmount: 12, BalanceUnit: "usd"}, nil
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
