package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
	"github.com/smallestfisher/relaydeck/backend/internal/http/middleware"
	"github.com/smallestfisher/relaydeck/backend/internal/store"
)

type upstreamAccountRequest struct {
	Name                  string                        `json:"name"`
	Code                  string                        `json:"code"`
	PlatformKind          domain.UpstreamPlatformKind   `json:"platform_kind"`
	BaseURL               string                        `json:"base_url"`
	Enabled               bool                          `json:"enabled"`
	IncludeInRouting      bool                          `json:"include_in_routing"`
	Priority              int                           `json:"priority"`
	APIKey                string                        `json:"api_key"`
	AccountCredentialKind domain.UpstreamCredentialKind `json:"account_credential_kind"`
	AccountCredential     string                        `json:"account_credential"`
	AutoSyncModels        bool                          `json:"auto_sync_models"`
	AutoRefreshQuota      bool                          `json:"auto_refresh_quota"`
	AutoCheckin           bool                          `json:"auto_checkin"`
	Note                  string                        `json:"note"`
}

type upstreamAccountResponse struct {
	ID                    string                        `json:"id"`
	Name                  string                        `json:"name"`
	Code                  string                        `json:"code"`
	PlatformKind          domain.UpstreamPlatformKind   `json:"platform_kind"`
	BaseURL               string                        `json:"base_url"`
	Enabled               bool                          `json:"enabled"`
	IncludeInRouting      bool                          `json:"include_in_routing"`
	Priority              int                           `json:"priority"`
	APIKeyPrefix          string                        `json:"api_key_prefix"`
	HasAPICredential      bool                          `json:"has_api_credential"`
	AccountCredentialKind domain.UpstreamCredentialKind `json:"account_credential_kind"`
	HasAccountCredential  bool                          `json:"has_account_credential"`
	AutoSyncModels        bool                          `json:"auto_sync_models"`
	AutoRefreshQuota      bool                          `json:"auto_refresh_quota"`
	AutoCheckin           bool                          `json:"auto_checkin"`
	Note                  string                        `json:"note"`
	Status                domain.UpstreamAccountStatus  `json:"status"`
	CreatedAt             string                        `json:"created_at"`
	UpdatedAt             string                        `json:"updated_at"`
}

type batchRequest struct {
	IDs []string `json:"ids"`
}

type actionResult struct {
	ID            string                        `json:"id"`
	Status        string                        `json:"status"`
	Message       string                        `json:"message,omitempty"`
	AccountStatus *domain.UpstreamAccountStatus `json:"account_status,omitempty"`
}

func (h *Handler) mountUpstreamRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/admin/upstreams/accounts", h.requireAdmin(http.HandlerFunc(h.handleUpstreamAccounts)))
	mux.Handle("POST /api/admin/upstreams/accounts", h.requireAdmin(http.HandlerFunc(h.handleCreateUpstreamAccount)))
	mux.Handle("POST /api/admin/upstreams/test", h.requireAdmin(http.HandlerFunc(h.handleDraftUpstreamAPITest)))
	mux.Handle("GET /api/admin/upstreams/accounts/{id}", h.requireAdmin(http.HandlerFunc(h.handleUpstreamAccountByID)))
	mux.Handle("PUT /api/admin/upstreams/accounts/{id}", h.requireAdmin(http.HandlerFunc(h.handleUpdateUpstreamAccount)))
	mux.Handle("DELETE /api/admin/upstreams/accounts/{id}", h.requireAdmin(http.HandlerFunc(h.handleDeleteUpstreamAccount)))
	mux.Handle("POST /api/admin/upstreams/accounts/{id}/test-api", h.requireAdmin(http.HandlerFunc(h.handleUpstreamAction)))
	mux.Handle("POST /api/admin/upstreams/accounts/{id}/test-account", h.requireAdmin(http.HandlerFunc(h.handleUpstreamAction)))
	mux.Handle("POST /api/admin/upstreams/accounts/{id}/sync-models", h.requireAdmin(http.HandlerFunc(h.handleUpstreamAction)))
	mux.Handle("POST /api/admin/upstreams/accounts/{id}/refresh-quota", h.requireAdmin(http.HandlerFunc(h.handleUpstreamAction)))
	mux.Handle("POST /api/admin/upstreams/accounts/{id}/checkin", h.requireAdmin(http.HandlerFunc(h.handleUpstreamAction)))
	mux.Handle("POST /api/admin/upstreams/accounts/{id}/refresh-all", h.requireAdmin(http.HandlerFunc(h.handleUpstreamAction)))
	mux.Handle("POST /api/admin/upstreams/accounts/{id}/test-call", h.requireAdmin(http.HandlerFunc(h.handleTestCall)))
	mux.Handle("GET /api/admin/upstreams/accounts/{id}/models", h.requireAdmin(http.HandlerFunc(h.handleUpstreamModels)))
	mux.Handle("GET /api/admin/upstreams/accounts/{id}/events", h.requireAdmin(http.HandlerFunc(h.handleUpstreamEvents)))
	mux.Handle("POST /api/admin/upstreams/accounts/batch/test-api", h.requireAdmin(http.HandlerFunc(h.handleBatchUpstreamAction)))
	mux.Handle("POST /api/admin/upstreams/accounts/batch/sync-models", h.requireAdmin(http.HandlerFunc(h.handleBatchUpstreamAction)))
	mux.Handle("POST /api/admin/upstreams/accounts/batch/refresh-quota", h.requireAdmin(http.HandlerFunc(h.handleBatchUpstreamAction)))
	mux.Handle("POST /api/admin/upstreams/accounts/batch/checkin", h.requireAdmin(http.HandlerFunc(h.handleBatchUpstreamAction)))
	mux.Handle("POST /api/admin/upstreams/accounts/batch/refresh-all", h.requireAdmin(http.HandlerFunc(h.handleBatchUpstreamAction)))
}

func (h *Handler) requireAdmin(next http.Handler) http.Handler {
	return middleware.RequireAdminSession(next, h.sessions, h.store, h.now)
}

func (h *Handler) upstreamStore(w http.ResponseWriter) (store.UpstreamAccountStore, bool) {
	upstreams := h.store.Upstreams()
	if upstreams == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "upstream_store_unavailable"})
		return nil, false
	}
	return upstreams, true
}

func (h *Handler) handleUpstreamAccounts(w http.ResponseWriter, r *http.Request) {
	upstreams, ok := h.upstreamStore(w)
	if !ok {
		return
	}
	pagination := parseUpstreamAccountPagination(r)
	accounts, total := listUpstreamAccounts(upstreams, pagination.Limit, pagination.Offset)
	items := make([]upstreamAccountResponse, 0, len(accounts))
	for _, account := range accounts {
		status, _ := upstreams.UpstreamAccountStatus(account.ID)
		items = append(items, toUpstreamAccountResponse(account, status))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"total":  total,
		"limit":  pagination.Limit,
		"offset": pagination.Offset,
	})
}

func (h *Handler) handleCreateUpstreamAccount(w http.ResponseWriter, r *http.Request) {
	upstreams, ok := h.upstreamStore(w)
	if !ok {
		return
	}
	var payload upstreamAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
		return
	}
	if err := validateUpstreamAccountRequest(payload, true); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	account, err := h.accountFromRequest(payload, domain.UpstreamAccount{}, true)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "secret_encryption_failed"})
		return
	}
	created, err := upstreams.CreateUpstreamAccount(account)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "create_failed"})
		return
	}
	status := defaultStatusForAccount(created)
	_ = upstreams.UpsertUpstreamAccountStatus(status)
	writeJSON(w, http.StatusCreated, toUpstreamAccountResponse(created, status))
}

func (h *Handler) handleUpstreamAccountByID(w http.ResponseWriter, r *http.Request) {
	upstreams, ok := h.upstreamStore(w)
	if !ok {
		return
	}
	account, found := upstreams.UpstreamAccount(r.PathValue("id"))
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
		return
	}
	status, _ := upstreams.UpstreamAccountStatus(account.ID)
	writeJSON(w, http.StatusOK, toUpstreamAccountResponse(account, status))
}

func (h *Handler) handleUpdateUpstreamAccount(w http.ResponseWriter, r *http.Request) {
	upstreams, ok := h.upstreamStore(w)
	if !ok {
		return
	}
	existing, found := upstreams.UpstreamAccount(r.PathValue("id"))
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
		return
	}
	var payload upstreamAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
		return
	}
	if err := validateUpstreamAccountRequest(payload, false); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	account, err := h.accountFromRequest(payload, existing, false)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "secret_encryption_failed"})
		return
	}
	updated, err := upstreams.UpdateUpstreamAccount(account)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "update_failed"})
		return
	}
	status, _ := upstreams.UpstreamAccountStatus(updated.ID)
	writeJSON(w, http.StatusOK, toUpstreamAccountResponse(updated, status))
}

func (h *Handler) handleDeleteUpstreamAccount(w http.ResponseWriter, r *http.Request) {
	upstreams, ok := h.upstreamStore(w)
	if !ok {
		return
	}
	if err := upstreams.DeleteUpstreamAccount(r.PathValue("id")); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "delete_failed"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleDraftUpstreamAPITest(w http.ResponseWriter, r *http.Request) {
	var payload upstreamAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
		return
	}
	if err := validateDraftUpstreamAPITestRequest(payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	adapter, ok := h.upstreams.For(payload.PlatformKind)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "platform_kind_unsupported"})
		return
	}
	account := domain.UpstreamAccount{
		ID:           "draft",
		PlatformKind: payload.PlatformKind,
		BaseURL:      strings.TrimRight(strings.TrimSpace(payload.BaseURL), "/"),
		Enabled:      true,
	}
	status, err := adapter.TestAPI(r.Context(), account, strings.TrimSpace(payload.APIKey))
	result := actionResult{ID: "draft", Status: "success"}
	if err != nil {
		result.Status = "failed"
		result.Message = err.Error()
	} else if status.LastErrorMessage != "" {
		result.Message = status.LastErrorMessage
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) handleUpstreamModels(w http.ResponseWriter, r *http.Request) {
	upstreams, ok := h.upstreamStore(w)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": upstreams.UpstreamModels(r.PathValue("id"))})
}

func (h *Handler) handleUpstreamEvents(w http.ResponseWriter, r *http.Request) {
	upstreams, ok := h.upstreamStore(w)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": upstreams.UpstreamAccountEvents(r.PathValue("id"), 100)})
}

func (h *Handler) handleUpstreamAction(w http.ResponseWriter, r *http.Request) {
	result := h.runUpstreamAction(r.Context(), r.PathValue("id"), actionFromPath(r.URL.Path))
	status := http.StatusOK
	if result.Status == "not_found" {
		status = http.StatusNotFound
	}
	writeJSON(w, status, result)
}

func (h *Handler) handleBatchUpstreamAction(w http.ResponseWriter, r *http.Request) {
	var payload batchRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
		return
	}
	results := make([]actionResult, 0, len(payload.IDs))
	action := actionFromPath(r.URL.Path)
	for _, id := range payload.IDs {
		results = append(results, h.runUpstreamAction(r.Context(), id, action))
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *Handler) runUpstreamAction(ctx context.Context, id string, action string) actionResult {
	upstreams := h.store.Upstreams()
	if upstreams == nil {
		return actionResult{ID: id, Status: "failed", Message: "upstream store unavailable"}
	}
	account, ok := upstreams.UpstreamAccount(id)
	if !ok {
		return actionResult{ID: id, Status: "not_found", Message: "account not found"}
	}
	adapter, ok := h.upstreams.For(account.PlatformKind)
	if !ok {
		return actionResult{ID: id, Status: "failed", Message: "unsupported platform"}
	}
	apiKey, accountCredential, err := h.decryptAccountSecrets(account)
	if err != nil {
		return actionResult{ID: id, Status: "failed", Message: "decrypt secret failed"}
	}
	switch action {
	case "test-api":
		status, err := adapter.TestAPI(ctx, account, apiKey)
		return h.storeActionStatus(upstreams, account, action, status, nil, err)
	case "test-account":
		result, err := adapter.TestAccountCredential(ctx, account, accountCredential)
		if err == nil {
			if updated, updateErr := h.applyCredentialUpdate(upstreams, account, result.CredentialUpdate); updateErr != nil {
				return actionResult{ID: id, Status: "failed", Message: updateErr.Error()}
			} else {
				account = updated
			}
		}
		return h.storeActionStatus(upstreams, account, action, result.Status, nil, err)
	case "sync-models":
		result, status, err := adapter.SyncModels(ctx, account, apiKey)
		return h.storeActionStatus(upstreams, account, action, status, result.Models, err)
	case "refresh-quota":
		result, err := adapter.RefreshQuota(ctx, account, apiKey, accountCredential)
		if err == nil {
			if updated, updateErr := h.applyCredentialUpdate(upstreams, account, result.CredentialUpdate); updateErr != nil {
				return actionResult{ID: id, Status: "failed", Message: updateErr.Error()}
			} else {
				account = updated
			}
		}
		return h.storeActionStatus(upstreams, account, action, result.Status, nil, err)
	case "checkin":
		result, err := adapter.Checkin(ctx, account, accountCredential)
		if err == nil {
			if updated, updateErr := h.applyCredentialUpdate(upstreams, account, result.CredentialUpdate); updateErr != nil {
				return actionResult{ID: id, Status: "failed", Message: updateErr.Error()}
			} else {
				account = updated
			}
		}
		status := domain.UpstreamAccountStatus{UpstreamAccountID: account.ID, CheckinStatus: result.Status, UpdatedAt: h.now()}
		if result.Status == domain.CheckinStatusChecked {
			status.LastCheckinAt = h.now()
			if account.HasAccountCredential() {
				status.AccountStatus = domain.AccountCredentialStatusValid
			}
		}
		if result.AccountStatus != "" {
			status.AccountStatus = result.AccountStatus
		}
		status.LastErrorClass = result.LastErrorClass
		status.LastErrorMessage = firstNonEmpty(result.LastErrorMessage, result.Message)
		status.ActionRequiredReason = result.ActionRequiredReason
		actionResult := h.storeActionStatus(upstreams, account, action, status, nil, err)
		if actionResult.Message == "" {
			actionResult.Message = result.Message
		}
		return actionResult
	case "refresh-all":
		// Test account credential + refresh quota
		finalStatus := domain.UpstreamAccountStatus{UpstreamAccountID: account.ID, UpdatedAt: h.now()}
		var models []domain.UpstreamSyncedModel

		if accountCredential != "" {
			testResult, err := adapter.TestAccountCredential(ctx, account, accountCredential)
			if err == nil {
				if updated, updateErr := h.applyCredentialUpdate(upstreams, account, testResult.CredentialUpdate); updateErr == nil {
					account = updated
				}
				finalStatus.AccountStatus = testResult.Status.AccountStatus
				finalStatus.CheckinStatus = testResult.Status.CheckinStatus
				finalStatus.LastAccountCheckedAt = testResult.Status.LastAccountCheckedAt
				finalStatus.BalanceAmount = testResult.Status.BalanceAmount
				finalStatus.BalanceUnit = testResult.Status.BalanceUnit
				if testResult.Status.LastErrorClass != "" {
					finalStatus.LastErrorClass = testResult.Status.LastErrorClass
					finalStatus.LastErrorMessage = testResult.Status.LastErrorMessage
				}
			}
		}

		// Sync models
		syncResult, syncStatus, syncErr := adapter.SyncModels(ctx, account, apiKey)
		if syncErr == nil {
			models = syncResult.Models
			finalStatus.APIStatus = syncStatus.APIStatus
			finalStatus.ModelCount = syncStatus.ModelCount
			finalStatus.LatencyMS = syncStatus.LatencyMS
			finalStatus.LastAPICheckedAt = syncStatus.LastAPICheckedAt
			finalStatus.LastModelSyncedAt = syncStatus.LastModelSyncedAt
			if syncStatus.LastErrorClass != "" {
				finalStatus.LastErrorClass = syncStatus.LastErrorClass
				finalStatus.LastErrorMessage = syncStatus.LastErrorMessage
			}
		}

		return h.storeActionStatus(upstreams, account, action, finalStatus, models, nil)
	default:
		return actionResult{ID: id, Status: "failed", Message: "unknown action"}
	}
}

func (h *Handler) applyCredentialUpdate(upstreams store.UpstreamAccountStore, account domain.UpstreamAccount, update *domain.UpstreamCredentialUpdate) (domain.UpstreamAccount, error) {
	if update == nil {
		return account, nil
	}
	if update.Kind == "" || strings.TrimSpace(update.Plaintext) == "" {
		return domain.UpstreamAccount{}, errors.New("credential update is invalid")
	}
	encrypted, err := h.encryptSecret(update.Plaintext)
	if err != nil {
		return domain.UpstreamAccount{}, err
	}
	account.AccountCredentialKind = update.Kind
	account.AccountCredentialEncrypted = encrypted
	return upstreams.UpdateUpstreamAccount(account)
}

func (h *Handler) storeActionStatus(upstreams store.UpstreamAccountStore, account domain.UpstreamAccount, action string, status domain.UpstreamAccountStatus, models []domain.UpstreamSyncedModel, err error) actionResult {
	if existing, ok := upstreams.UpstreamAccountStatus(account.ID); ok {
		status = mergeActionStatus(action, existing, status)
	}
	status.UpstreamAccountID = account.ID
	if status.UpdatedAt.IsZero() {
		status.UpdatedAt = h.now()
	}
	eventStatus := "success"
	message := ""
	if err != nil {
		eventStatus = "failed"
		message = err.Error()
	}
	if len(models) > 0 {
		_ = upstreams.ReplaceUpstreamModels(account.ID, models)
	}
	_ = upstreams.UpsertUpstreamAccountStatus(status)
	_ = upstreams.AppendUpstreamAccountEvent(domain.UpstreamAccountEvent{
		UpstreamAccountID: account.ID,
		Operation:         strings.ReplaceAll(action, "-", "_"),
		Status:            eventStatus,
		ErrorClass:        status.LastErrorClass,
		Message:           firstNonEmpty(message, status.LastErrorMessage),
		LatencyMS:         status.LatencyMS,
	})
	snapshot := status
	return actionResult{ID: account.ID, Status: eventStatus, Message: message, AccountStatus: &snapshot}
}

func mergeActionStatus(action string, existing domain.UpstreamAccountStatus, status domain.UpstreamAccountStatus) domain.UpstreamAccountStatus {
	if status.APIStatus == "" || action == "test-account" || action == "checkin" {
		status.APIStatus = existing.APIStatus
		status.LastAPICheckedAt = existing.LastAPICheckedAt
	}
	if status.AccountStatus == "" || action == "test-api" || action == "sync-models" {
		status.AccountStatus = existing.AccountStatus
		status.LastAccountCheckedAt = existing.LastAccountCheckedAt
	}
	if status.CheckinStatus == "" || action == "test-api" || action == "sync-models" || action == "refresh-quota" {
		status.CheckinStatus = existing.CheckinStatus
		status.LastCheckinAt = existing.LastCheckinAt
	}
	if action != "test-api" && action != "sync-models" {
		status.ModelCount = existing.ModelCount
		status.LastModelSyncedAt = existing.LastModelSyncedAt
	}
	if status.LatencyMS == 0 {
		status.LatencyMS = existing.LatencyMS
	}
	if action != "refresh-quota" {
		status.BalanceAmount = existing.BalanceAmount
		status.BalanceUnit = existing.BalanceUnit
	}
	if status.UpdatedAt.IsZero() {
		status.UpdatedAt = existing.UpdatedAt
	}
	return status
}

func (h *Handler) accountFromRequest(payload upstreamAccountRequest, existing domain.UpstreamAccount, create bool) (domain.UpstreamAccount, error) {
	account := existing
	account.Name = strings.TrimSpace(payload.Name)
	account.Code = strings.TrimSpace(payload.Code)
	account.PlatformKind = payload.PlatformKind
	account.BaseURL = strings.TrimRight(strings.TrimSpace(payload.BaseURL), "/")
	account.Enabled = payload.Enabled
	account.IncludeInRouting = payload.IncludeInRouting
	account.Priority = payload.Priority
	account.AccountCredentialKind = credentialKindOrNone(payload.AccountCredentialKind)
	account.AutoSyncModels = payload.AutoSyncModels
	account.AutoRefreshQuota = payload.AutoRefreshQuota
	account.AutoCheckin = payload.AutoCheckin
	account.Note = payload.Note
	if payload.APIKey != "" {
		encrypted, err := h.encryptSecret(payload.APIKey)
		if err != nil {
			return domain.UpstreamAccount{}, err
		}
		account.APIKeyEncrypted = encrypted
		account.APIKeyPrefix = secretPrefix(payload.APIKey)
	}
	if create && account.APIKeyEncrypted == "" {
		return domain.UpstreamAccount{}, errors.New("api key is required")
	}
	if payload.AccountCredential != "" {
		encrypted, err := h.encryptSecret(payload.AccountCredential)
		if err != nil {
			return domain.UpstreamAccount{}, err
		}
		account.AccountCredentialEncrypted = encrypted
	}
	if account.AccountCredentialKind == domain.CredentialKindNone {
		account.AccountCredentialEncrypted = ""
	}
	return account, nil
}

func (h *Handler) encryptSecret(plaintext string) (string, error) {
	if h.secrets == nil {
		return "", errors.New("secretbox unavailable")
	}
	return h.secrets.Encrypt(plaintext)
}

func (h *Handler) decryptAccountSecrets(account domain.UpstreamAccount) (string, string, error) {
	if h.secrets == nil {
		return "", "", errors.New("secretbox unavailable")
	}
	apiKey := ""
	accountCredential := ""
	var err error
	if account.APIKeyEncrypted != "" {
		apiKey, err = h.secrets.Decrypt(account.APIKeyEncrypted)
		if err != nil {
			return "", "", err
		}
	}
	if account.AccountCredentialEncrypted != "" {
		accountCredential, err = h.secrets.Decrypt(account.AccountCredentialEncrypted)
		if err != nil {
			return "", "", err
		}
	}
	return apiKey, accountCredential, nil
}

func validateUpstreamAccountRequest(payload upstreamAccountRequest, create bool) error {
	if strings.TrimSpace(payload.Name) == "" {
		return errors.New("name_required")
	}
	if strings.TrimSpace(payload.Code) == "" {
		return errors.New("code_required")
	}
	if payload.PlatformKind != domain.PlatformKindNewAPI && payload.PlatformKind != domain.PlatformKindSub2API {
		return errors.New("platform_kind_invalid")
	}
	if strings.TrimSpace(payload.BaseURL) == "" {
		return errors.New("base_url_required")
	}
	if create && strings.TrimSpace(payload.APIKey) == "" {
		return errors.New("api_key_required")
	}
	return nil
}

func validateDraftUpstreamAPITestRequest(payload upstreamAccountRequest) error {
	if payload.PlatformKind != domain.PlatformKindNewAPI && payload.PlatformKind != domain.PlatformKindSub2API {
		return errors.New("platform_kind_invalid")
	}
	if strings.TrimSpace(payload.BaseURL) == "" {
		return errors.New("base_url_required")
	}
	if strings.TrimSpace(payload.APIKey) == "" {
		return errors.New("api_key_required")
	}
	return nil
}

type upstreamAccountPagination struct {
	Limit  int
	Offset int
}

type upstreamAccountPager interface {
	ListUpstreamAccountsPage(limit int, offset int) ([]domain.UpstreamAccount, int)
}

const (
	defaultUpstreamAccountLimit = 50
	maxUpstreamAccountLimit     = 200
)

func parseUpstreamAccountPagination(r *http.Request) upstreamAccountPagination {
	limit := defaultUpstreamAccountLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > maxUpstreamAccountLimit {
		limit = maxUpstreamAccountLimit
	}
	offset := 0
	if raw := r.URL.Query().Get("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			offset = parsed
		}
	}
	return upstreamAccountPagination{Limit: limit, Offset: offset}
}

func listUpstreamAccounts(upstreams store.UpstreamAccountStore, limit int, offset int) ([]domain.UpstreamAccount, int) {
	if pager, ok := upstreams.(upstreamAccountPager); ok {
		return pager.ListUpstreamAccountsPage(limit, offset)
	}
	accounts := upstreams.ListUpstreamAccounts()
	total := len(accounts)
	if offset >= total {
		return []domain.UpstreamAccount{}, total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return accounts[offset:end], total
}

func toUpstreamAccountResponse(account domain.UpstreamAccount, status domain.UpstreamAccountStatus) upstreamAccountResponse {
	if status.APIStatus == "" {
		status = defaultStatusForAccount(account)
	}
	return upstreamAccountResponse{
		ID:                    account.ID,
		Name:                  account.Name,
		Code:                  account.Code,
		PlatformKind:          account.PlatformKind,
		BaseURL:               account.BaseURL,
		Enabled:               account.Enabled,
		IncludeInRouting:      account.IncludeInRouting,
		Priority:              account.Priority,
		APIKeyPrefix:          account.APIKeyPrefix,
		HasAPICredential:      account.HasAPICredential(),
		AccountCredentialKind: credentialKindOrNone(account.AccountCredentialKind),
		HasAccountCredential:  account.HasAccountCredential(),
		AutoSyncModels:        account.AutoSyncModels,
		AutoRefreshQuota:      account.AutoRefreshQuota,
		AutoCheckin:           account.AutoCheckin,
		Note:                  account.Note,
		Status:                status,
		CreatedAt:             account.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:             account.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func defaultStatusForAccount(account domain.UpstreamAccount) domain.UpstreamAccountStatus {
	apiStatus := domain.UpstreamAPIStatusUnknown
	if !account.Enabled {
		apiStatus = domain.UpstreamAPIStatusDisabled
	}
	accountStatus := domain.AccountCredentialStatusNotConfigured
	if account.HasAccountCredential() {
		accountStatus = domain.AccountCredentialStatusValid
	}
	return domain.UpstreamAccountStatus{
		UpstreamAccountID: account.ID,
		APIStatus:         apiStatus,
		AccountStatus:     accountStatus,
		CheckinStatus:     domain.CheckinStatusUnsupported,
	}
}

func credentialKindOrNone(kind domain.UpstreamCredentialKind) domain.UpstreamCredentialKind {
	if kind == "" {
		return domain.CredentialKindNone
	}
	return kind
}

func actionFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func secretPrefix(secret string) string {
	secret = strings.TrimSpace(secret)
	if len(secret) <= 7 {
		return secret
	}
	return secret[:7]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

type testCallRequest struct {
	ModelName string `json:"model_name"`
	Protocol  string `json:"protocol"`
	Streaming bool   `json:"streaming"`
	Message   string `json:"message"`
}

func (h *Handler) handleTestCall(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req testCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	upstreams := h.store.Upstreams()
	if upstreams == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "upstream store unavailable"})
		return
	}
	account, ok := upstreams.UpstreamAccount(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "account not found"})
		return
	}

	adapter, ok := h.upstreams.For(account.PlatformKind)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unsupported platform"})
		return
	}

	apiKey, _, err := h.decryptAccountSecrets(account)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "decrypt secret failed"})
		return
	}

	result, err := adapter.TestCall(r.Context(), account, apiKey, req.ModelName, req.Protocol, req.Streaming, req.Message)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}
