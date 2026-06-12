package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/domain"
	"github.com/smallestfisher/relaydeck/backend/internal/store"
)

type UpstreamStore struct {
	db *sql.DB
}

func NewUpstreamStore(db *sql.DB) *UpstreamStore {
	return &UpstreamStore{db: db}
}

func (s *UpstreamStore) EnsureSchema(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, upstreamSchemaSQL)
	return err
}

func (s *UpstreamStore) ListUpstreamAccounts() []domain.UpstreamAccount {
	rows, err := s.db.QueryContext(context.Background(), upstreamAccountSelectSQL+` ORDER BY priority DESC, created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	accounts := []domain.UpstreamAccount{}
	for rows.Next() {
		account, err := scanUpstreamAccount(rows)
		if err != nil {
			return nil
		}
		accounts = append(accounts, account)
	}
	return accounts
}

func (s *UpstreamStore) ListUpstreamAccountsPage(limit int, offset int) ([]domain.UpstreamAccount, int) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	if err := s.db.QueryRowContext(context.Background(), `SELECT count(*) FROM upstream_accounts`).Scan(&total); err != nil {
		return nil, 0
	}

	rows, err := s.db.QueryContext(context.Background(), upstreamAccountSelectSQL+` ORDER BY priority DESC, created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, total
	}
	defer rows.Close()

	accounts := []domain.UpstreamAccount{}
	for rows.Next() {
		account, err := scanUpstreamAccount(rows)
		if err != nil {
			return nil, total
		}
		accounts = append(accounts, account)
	}
	return accounts, total
}

func (s *UpstreamStore) FilterUpstreamAccounts(filter store.UpstreamAccountFilter) ([]domain.UpstreamAccount, store.UpstreamAccountMetrics) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// Build dynamic WHERE clause by joining account and status tables.
	// effective_api_status: if account is disabled, override to 'disabled'.
	// effective_account_status: when no status row exists and account has no credential, use 'not_configured'; with credential, 'action_required'.
	// effective_latency: prefer api_latency_ms when last_api_checked_at is set, else latency_ms when last_model_synced_at is set, else 0.
	baseQuery := `
WITH joined AS (
  SELECT a.*,
    CASE WHEN NOT a.enabled THEN 'disabled'
         WHEN COALESCE(s.api_status, 'unknown') = '' THEN 'unknown'
         ELSE COALESCE(s.api_status, 'unknown')
    END AS effective_api_status,
    CASE WHEN s.upstream_account_id IS NULL THEN
           CASE WHEN a.account_credential_kind != 'none' AND a.account_credential_enc != '' THEN 'action_required'
                ELSE 'not_configured'
           END
         WHEN COALESCE(s.account_status, 'not_configured') = '' THEN 'not_configured'
         ELSE COALESCE(s.account_status, 'not_configured')
    END AS effective_account_status,
    COALESCE(s.checkin_status, 'unsupported') AS effective_checkin_status,
    COALESCE(s.model_count, 0) AS s_model_count,
    COALESCE(s.latency_ms, 0) AS s_latency_ms,
    COALESCE(s.api_latency_ms, 0) AS s_api_latency_ms,
    CASE WHEN s.last_api_checked_at IS NOT NULL THEN COALESCE(s.api_latency_ms, 0)
         WHEN s.last_model_synced_at IS NOT NULL THEN COALESCE(s.latency_ms, 0)
         ELSE 0
    END AS effective_latency
  FROM upstream_accounts a
  LEFT JOIN upstream_account_status s ON s.upstream_account_id = a.id
)`

	var conditions []string
	var args []any
	argIdx := 1

	if filter.Query != "" {
		pattern := "%" + strings.ToLower(filter.Query) + "%"
		conditions = append(conditions, fmt.Sprintf(
			`(lower(name) LIKE $%d OR lower(base_url) LIKE $%d OR lower(note) LIKE $%d)`,
			argIdx, argIdx, argIdx))
		args = append(args, pattern)
		argIdx++
	}
	if filter.PlatformKind != "" {
		conditions = append(conditions, fmt.Sprintf(`platform_kind = $%d`, argIdx))
		args = append(args, filter.PlatformKind)
		argIdx++
	}
	if filter.APIStatus != "" {
		conditions = append(conditions, fmt.Sprintf(`effective_api_status = $%d`, argIdx))
		args = append(args, filter.APIStatus)
		argIdx++
	}
	if filter.AccountStatus != "" {
		conditions = append(conditions, fmt.Sprintf(`effective_account_status = $%d`, argIdx))
		args = append(args, filter.AccountStatus)
		argIdx++
	}
	if filter.LatencyBand != "" {
		switch filter.LatencyBand {
		case "unknown":
			conditions = append(conditions, `effective_latency <= 0`)
		case "low":
			conditions = append(conditions, `effective_latency > 0 AND effective_latency < 300`)
		case "medium":
			conditions = append(conditions, `effective_latency >= 300 AND effective_latency <= 1000`)
		case "high":
			conditions = append(conditions, `effective_latency > 1000`)
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Single query: get paginated rows + metrics via window functions.
	fullSQL := baseQuery + fmt.Sprintf(`
SELECT
  id::text, name, code, platform_kind, base_url, enabled, include_in_routing, priority,
  api_key_enc, api_key_prefix, account_credential_kind, account_credential_enc,
  auto_sync_models, auto_refresh_quota, auto_checkin, note, created_at, updated_at,
  count(*) OVER () AS filtered_total,
  count(*) FILTER (WHERE enabled AND effective_api_status = 'healthy') OVER () AS cnt_healthy,
  count(*) FILTER (WHERE effective_api_status = 'warning' OR effective_account_status = 'expired') OVER () AS cnt_warning,
  count(*) FILTER (WHERE effective_account_status = 'action_required') OVER () AS cnt_manual
FROM joined
%s
ORDER BY priority DESC, created_at DESC
LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := s.db.QueryContext(context.Background(), fullSQL, args...)
	if err != nil {
		return nil, store.UpstreamAccountMetrics{}
	}
	defer rows.Close()

	var accounts []domain.UpstreamAccount
	var metrics store.UpstreamAccountMetrics
	for rows.Next() {
		var account domain.UpstreamAccount
		var platformKind, credentialKind string
		var filteredTotal, cntHealthy, cntWarning, cntManual int
		if err := rows.Scan(
			&account.ID,
			&account.Name,
			&account.Code,
			&platformKind,
			&account.BaseURL,
			&account.Enabled,
			&account.IncludeInRouting,
			&account.Priority,
			&account.APIKeyEncrypted,
			&account.APIKeyPrefix,
			&credentialKind,
			&account.AccountCredentialEncrypted,
			&account.AutoSyncModels,
			&account.AutoRefreshQuota,
			&account.AutoCheckin,
			&account.Note,
			&account.CreatedAt,
			&account.UpdatedAt,
			&filteredTotal,
			&cntHealthy,
			&cntWarning,
			&cntManual,
		); err != nil {
			return nil, store.UpstreamAccountMetrics{}
		}
		account.PlatformKind = domain.UpstreamPlatformKind(platformKind)
		account.AccountCredentialKind = domain.UpstreamCredentialKind(credentialKind)
		accounts = append(accounts, account)
		metrics.Total = filteredTotal
		metrics.Healthy = cntHealthy
		metrics.Warning = cntWarning
		metrics.Manual = cntManual
	}
	return accounts, metrics
}

func (s *UpstreamStore) UpstreamAccount(id string) (domain.UpstreamAccount, bool) {
	account, err := scanUpstreamAccount(s.db.QueryRowContext(context.Background(), upstreamAccountSelectSQL+` WHERE id = $1::uuid`, id))
	return account, err == nil
}

func (s *UpstreamStore) CreateUpstreamAccount(account domain.UpstreamAccount) (domain.UpstreamAccount, error) {
	id, err := newUUID()
	if err != nil {
		return domain.UpstreamAccount{}, err
	}
	account.ID = id
	row := s.db.QueryRowContext(context.Background(), `
INSERT INTO upstream_accounts (
  id, name, code, platform_kind, base_url, enabled, include_in_routing, priority,
  api_key_enc, api_key_prefix, account_credential_kind, account_credential_enc,
  auto_sync_models, auto_refresh_quota, auto_checkin, note, created_at, updated_at
) VALUES (
  $1::uuid, $2, $3, $4, $5, $6, $7, $8,
  $9, $10, $11, $12, $13, $14, $15, $16, now(), now()
)
RETURNING id::text, name, code, platform_kind, base_url, enabled, include_in_routing, priority,
  api_key_enc, api_key_prefix, account_credential_kind, account_credential_enc,
  auto_sync_models, auto_refresh_quota, auto_checkin, note, created_at, updated_at`,
		account.ID,
		account.Name,
		account.Code,
		account.PlatformKind,
		account.BaseURL,
		account.Enabled,
		account.IncludeInRouting,
		account.Priority,
		account.APIKeyEncrypted,
		account.APIKeyPrefix,
		credentialKindOrNone(account.AccountCredentialKind),
		account.AccountCredentialEncrypted,
		account.AutoSyncModels,
		account.AutoRefreshQuota,
		account.AutoCheckin,
		account.Note,
	)
	return scanUpstreamAccount(row)
}

func (s *UpstreamStore) UpdateUpstreamAccount(account domain.UpstreamAccount) (domain.UpstreamAccount, error) {
	if account.ID == "" {
		return domain.UpstreamAccount{}, errors.New("upstream account id is required")
	}
	row := s.db.QueryRowContext(context.Background(), `
UPDATE upstream_accounts
SET name = $2,
    code = $3,
    platform_kind = $4,
    base_url = $5,
    enabled = $6,
    include_in_routing = $7,
    priority = $8,
    api_key_enc = $9,
    api_key_prefix = $10,
    account_credential_kind = $11,
    account_credential_enc = $12,
    auto_sync_models = $13,
    auto_refresh_quota = $14,
    auto_checkin = $15,
    note = $16,
    updated_at = now()
WHERE id = $1::uuid
RETURNING id::text, name, code, platform_kind, base_url, enabled, include_in_routing, priority,
  api_key_enc, api_key_prefix, account_credential_kind, account_credential_enc,
  auto_sync_models, auto_refresh_quota, auto_checkin, note, created_at, updated_at`,
		account.ID,
		account.Name,
		account.Code,
		account.PlatformKind,
		account.BaseURL,
		account.Enabled,
		account.IncludeInRouting,
		account.Priority,
		account.APIKeyEncrypted,
		account.APIKeyPrefix,
		credentialKindOrNone(account.AccountCredentialKind),
		account.AccountCredentialEncrypted,
		account.AutoSyncModels,
		account.AutoRefreshQuota,
		account.AutoCheckin,
		account.Note,
	)
	return scanUpstreamAccount(row)
}

func (s *UpstreamStore) DeleteUpstreamAccount(id string) error {
	_, err := s.db.ExecContext(context.Background(), `DELETE FROM upstream_accounts WHERE id = $1::uuid`, id)
	return err
}

func (s *UpstreamStore) UpstreamAccountStatus(accountID string) (domain.UpstreamAccountStatus, bool) {
	status, err := scanUpstreamAccountStatus(s.db.QueryRowContext(context.Background(), upstreamStatusSelectSQL+` WHERE upstream_account_id = $1::uuid`, accountID))
	return status, err == nil
}

func (s *UpstreamStore) UpsertUpstreamAccountStatus(status domain.UpstreamAccountStatus) error {
	_, err := s.db.ExecContext(context.Background(), `
INSERT INTO upstream_account_status (
  upstream_account_id, api_status, account_status, checkin_status, model_count, latency_ms, api_latency_ms,
  balance_amount, balance_unit, last_api_checked_at, last_account_checked_at, last_model_synced_at,
  last_checkin_at, last_error_class, last_error_message, action_required_reason, updated_at
) VALUES (
  $1::uuid, $2, $3, $4, $5, $6, $7,
  $8, $9, $10, $11, $12,
  $13, $14, $15, $16, now()
)
ON CONFLICT (upstream_account_id) DO UPDATE
SET api_status = EXCLUDED.api_status,
    account_status = EXCLUDED.account_status,
    checkin_status = EXCLUDED.checkin_status,
    model_count = EXCLUDED.model_count,
    latency_ms = EXCLUDED.latency_ms,
    api_latency_ms = EXCLUDED.api_latency_ms,
    balance_amount = EXCLUDED.balance_amount,
    balance_unit = EXCLUDED.balance_unit,
    last_api_checked_at = EXCLUDED.last_api_checked_at,
    last_account_checked_at = EXCLUDED.last_account_checked_at,
    last_model_synced_at = EXCLUDED.last_model_synced_at,
    last_checkin_at = EXCLUDED.last_checkin_at,
    last_error_class = EXCLUDED.last_error_class,
    last_error_message = EXCLUDED.last_error_message,
    action_required_reason = EXCLUDED.action_required_reason,
    updated_at = now()`,
		status.UpstreamAccountID,
		apiStatusOrUnknown(status.APIStatus),
		accountStatusOrNotConfigured(status.AccountStatus),
		checkinStatusOrUnsupported(status.CheckinStatus),
		status.ModelCount,
		status.LatencyMS,
		status.APILatencyMS,
		status.BalanceAmount,
		status.BalanceUnit,
		nullableTime(status.LastAPICheckedAt),
		nullableTime(status.LastAccountCheckedAt),
		nullableTime(status.LastModelSyncedAt),
		nullableTime(status.LastCheckinAt),
		status.LastErrorClass,
		status.LastErrorMessage,
		status.ActionRequiredReason,
	)
	return err
}

func (s *UpstreamStore) UpstreamModels(accountID string) []domain.UpstreamSyncedModel {
	rows, err := s.db.QueryContext(context.Background(), `
SELECT id::text, upstream_account_id::text, normalized_model_name, upstream_model_name, display_name,
       native_wire_protocol, array_to_json(supported_wire_protocols), array_to_json(capabilities), status, raw_metadata, last_synced_at
FROM upstream_synced_models
WHERE upstream_account_id = $1::uuid
ORDER BY normalized_model_name ASC`, accountID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	models := []domain.UpstreamSyncedModel{}
	for rows.Next() {
		model, err := scanUpstreamModel(rows)
		if err != nil {
			return nil
		}
		models = append(models, model)
	}
	return models
}

func (s *UpstreamStore) ReplaceUpstreamModels(accountID string, models []domain.UpstreamSyncedModel) error {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(context.Background(), `DELETE FROM upstream_synced_models WHERE upstream_account_id = $1::uuid`, accountID); err != nil {
		return err
	}
	for _, model := range models {
		id := model.ID
		if id == "" {
			id, err = newUUID()
			if err != nil {
				return err
			}
		}
		rawMetadata, err := json.Marshal(model.RawMetadata)
		if err != nil {
			return err
		}
		if model.Status == "" {
			model.Status = "active"
		}
		if _, err := tx.ExecContext(context.Background(), `
INSERT INTO upstream_synced_models (
  id, upstream_account_id, normalized_model_name, upstream_model_name, display_name,
  native_wire_protocol, supported_wire_protocols, capabilities, status, raw_metadata, last_synced_at
) VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, now())`,
			id,
			accountID,
			model.NormalizedModelName,
			model.UpstreamModelName,
			model.DisplayName,
			model.NativeWireProtocol,
			protocolsToStrings(model.SupportedWireProtocols),
			capabilitiesToStrings(model.Capabilities),
			model.Status,
			string(rawMetadata),
		); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(context.Background(), `
UPDATE upstream_account_status
SET model_count = $2, last_model_synced_at = now(), updated_at = now()
WHERE upstream_account_id = $1::uuid`, accountID, len(models)); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *UpstreamStore) UpstreamAccountEvents(accountID string, limit int) []domain.UpstreamAccountEvent {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.QueryContext(context.Background(), `
SELECT id::text, upstream_account_id::text, operation, status, error_class, message, latency_ms, metadata, created_at
FROM upstream_account_events
WHERE upstream_account_id = $1::uuid
ORDER BY created_at DESC
LIMIT $2`, accountID, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	events := []domain.UpstreamAccountEvent{}
	for rows.Next() {
		event, err := scanUpstreamEvent(rows)
		if err != nil {
			return nil
		}
		events = append(events, event)
	}
	return events
}

func (s *UpstreamStore) AppendUpstreamAccountEvent(event domain.UpstreamAccountEvent) error {
	id := event.ID
	var err error
	if id == "" {
		id, err = newUUID()
		if err != nil {
			return err
		}
	}
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(context.Background(), `
INSERT INTO upstream_account_events (
  id, upstream_account_id, operation, status, error_class, message, latency_ms, metadata, created_at
) VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8::jsonb, now())`,
		id,
		event.UpstreamAccountID,
		event.Operation,
		event.Status,
		event.ErrorClass,
		event.Message,
		event.LatencyMS,
		string(metadata),
	)
	return err
}

const upstreamAccountSelectSQL = `
SELECT id::text, name, code, platform_kind, base_url, enabled, include_in_routing, priority,
       api_key_enc, api_key_prefix, account_credential_kind, account_credential_enc,
       auto_sync_models, auto_refresh_quota, auto_checkin, note, created_at, updated_at
FROM upstream_accounts`

const upstreamStatusSelectSQL = `
SELECT upstream_account_id::text, api_status, account_status, checkin_status, model_count, latency_ms, api_latency_ms,
       balance_amount, balance_unit, last_api_checked_at, last_account_checked_at, last_model_synced_at,
       last_checkin_at, last_error_class, last_error_message, action_required_reason, updated_at
FROM upstream_account_status`

type scanner interface {
	Scan(dest ...any) error
}

func scanUpstreamAccount(row scanner) (domain.UpstreamAccount, error) {
	var account domain.UpstreamAccount
	var platformKind string
	var credentialKind string
	if err := row.Scan(
		&account.ID,
		&account.Name,
		&account.Code,
		&platformKind,
		&account.BaseURL,
		&account.Enabled,
		&account.IncludeInRouting,
		&account.Priority,
		&account.APIKeyEncrypted,
		&account.APIKeyPrefix,
		&credentialKind,
		&account.AccountCredentialEncrypted,
		&account.AutoSyncModels,
		&account.AutoRefreshQuota,
		&account.AutoCheckin,
		&account.Note,
		&account.CreatedAt,
		&account.UpdatedAt,
	); err != nil {
		return domain.UpstreamAccount{}, err
	}
	account.PlatformKind = domain.UpstreamPlatformKind(platformKind)
	account.AccountCredentialKind = domain.UpstreamCredentialKind(credentialKind)
	return account, nil
}

func scanUpstreamAccountStatus(row scanner) (domain.UpstreamAccountStatus, error) {
	var status domain.UpstreamAccountStatus
	var apiStatus string
	var accountStatus string
	var checkinStatus string
	var lastAPI sql.NullTime
	var lastAccount sql.NullTime
	var lastModel sql.NullTime
	var lastCheckin sql.NullTime
	var errorClass string
	if err := row.Scan(
		&status.UpstreamAccountID,
		&apiStatus,
		&accountStatus,
		&checkinStatus,
		&status.ModelCount,
		&status.LatencyMS,
		&status.APILatencyMS,
		&status.BalanceAmount,
		&status.BalanceUnit,
		&lastAPI,
		&lastAccount,
		&lastModel,
		&lastCheckin,
		&errorClass,
		&status.LastErrorMessage,
		&status.ActionRequiredReason,
		&status.UpdatedAt,
	); err != nil {
		return domain.UpstreamAccountStatus{}, err
	}
	status.APIStatus = domain.UpstreamAPIStatus(apiStatus)
	status.AccountStatus = domain.AccountCredentialStatus(accountStatus)
	status.CheckinStatus = domain.UpstreamCheckinStatus(checkinStatus)
	status.LastErrorClass = domain.UpstreamErrorClass(errorClass)
	status.LastAPICheckedAt = timeFromNull(lastAPI)
	status.LastAccountCheckedAt = timeFromNull(lastAccount)
	status.LastModelSyncedAt = timeFromNull(lastModel)
	status.LastCheckinAt = timeFromNull(lastCheckin)
	return status, nil
}

func scanUpstreamModel(row scanner) (domain.UpstreamSyncedModel, error) {
	var model domain.UpstreamSyncedModel
	var nativeProtocol string
	var protocolsJSON []byte
	var capabilitiesJSON []byte
	var rawMetadata []byte
	if err := row.Scan(
		&model.ID,
		&model.UpstreamAccountID,
		&model.NormalizedModelName,
		&model.UpstreamModelName,
		&model.DisplayName,
		&nativeProtocol,
		&protocolsJSON,
		&capabilitiesJSON,
		&model.Status,
		&rawMetadata,
		&model.LastSyncedAt,
	); err != nil {
		return domain.UpstreamSyncedModel{}, err
	}
	model.NativeWireProtocol = domain.Protocol(nativeProtocol)
	var protocols []string
	if len(protocolsJSON) > 0 {
		if err := json.Unmarshal(protocolsJSON, &protocols); err != nil {
			return domain.UpstreamSyncedModel{}, err
		}
	}
	var capabilities []string
	if len(capabilitiesJSON) > 0 {
		if err := json.Unmarshal(capabilitiesJSON, &capabilities); err != nil {
			return domain.UpstreamSyncedModel{}, err
		}
	}
	model.SupportedWireProtocols = stringsToProtocols(protocols)
	model.Capabilities = stringsToCapabilities(capabilities)
	if len(rawMetadata) > 0 {
		if err := json.Unmarshal(rawMetadata, &model.RawMetadata); err != nil {
			return domain.UpstreamSyncedModel{}, err
		}
	}
	return model, nil
}

func scanUpstreamEvent(row scanner) (domain.UpstreamAccountEvent, error) {
	var event domain.UpstreamAccountEvent
	var errorClass string
	var metadata []byte
	if err := row.Scan(
		&event.ID,
		&event.UpstreamAccountID,
		&event.Operation,
		&event.Status,
		&errorClass,
		&event.Message,
		&event.LatencyMS,
		&metadata,
		&event.CreatedAt,
	); err != nil {
		return domain.UpstreamAccountEvent{}, err
	}
	event.ErrorClass = domain.UpstreamErrorClass(errorClass)
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &event.Metadata); err != nil {
			return domain.UpstreamAccountEvent{}, err
		}
	}
	return event, nil
}

func credentialKindOrNone(kind domain.UpstreamCredentialKind) domain.UpstreamCredentialKind {
	if kind == "" {
		return domain.CredentialKindNone
	}
	return kind
}

func apiStatusOrUnknown(status domain.UpstreamAPIStatus) domain.UpstreamAPIStatus {
	if status == "" {
		return domain.UpstreamAPIStatusUnknown
	}
	return status
}

func accountStatusOrNotConfigured(status domain.AccountCredentialStatus) domain.AccountCredentialStatus {
	if status == "" {
		return domain.AccountCredentialStatusNotConfigured
	}
	return status
}

func checkinStatusOrUnsupported(status domain.UpstreamCheckinStatus) domain.UpstreamCheckinStatus {
	if status == "" {
		return domain.CheckinStatusUnsupported
	}
	return status
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func timeFromNull(value sql.NullTime) time.Time {
	if value.Valid {
		return value.Time
	}
	return time.Time{}
}

func protocolsToStrings(values []domain.Protocol) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, string(value))
	}
	return result
}

func stringsToProtocols(values []string) []domain.Protocol {
	result := make([]domain.Protocol, 0, len(values))
	for _, value := range values {
		result = append(result, domain.Protocol(value))
	}
	return result
}

func capabilitiesToStrings(values []domain.Capability) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, string(value))
	}
	return result
}

func stringsToCapabilities(values []string) []domain.Capability {
	result := make([]domain.Capability, 0, len(values))
	for _, value := range values {
		result = append(result, domain.Capability(value))
	}
	return result
}

const upstreamSchemaSQL = `
CREATE TABLE IF NOT EXISTS upstream_accounts (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  code TEXT NOT NULL UNIQUE,
  platform_kind TEXT NOT NULL,
  base_url TEXT NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT true,
  include_in_routing BOOLEAN NOT NULL DEFAULT true,
  priority INTEGER NOT NULL DEFAULT 0,
  api_key_enc TEXT NOT NULL,
  api_key_prefix TEXT NOT NULL DEFAULT '',
  account_credential_kind TEXT NOT NULL DEFAULT 'none',
  account_credential_enc TEXT NOT NULL DEFAULT '',
  auto_sync_models BOOLEAN NOT NULL DEFAULT true,
  auto_refresh_quota BOOLEAN NOT NULL DEFAULT false,
  auto_checkin BOOLEAN NOT NULL DEFAULT false,
  note TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS upstream_account_status (
  upstream_account_id UUID PRIMARY KEY REFERENCES upstream_accounts(id) ON DELETE CASCADE,
  api_status TEXT NOT NULL DEFAULT 'unknown',
  account_status TEXT NOT NULL DEFAULT 'not_configured',
  checkin_status TEXT NOT NULL DEFAULT 'unsupported',
  model_count INTEGER NOT NULL DEFAULT 0,
  latency_ms INTEGER NOT NULL DEFAULT 0,
  api_latency_ms INTEGER NOT NULL DEFAULT 0,
  balance_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
  balance_unit TEXT NOT NULL DEFAULT '',
  last_api_checked_at TIMESTAMPTZ,
  last_account_checked_at TIMESTAMPTZ,
  last_model_synced_at TIMESTAMPTZ,
  last_checkin_at TIMESTAMPTZ,
  last_error_class TEXT NOT NULL DEFAULT '',
  last_error_message TEXT NOT NULL DEFAULT '',
  action_required_reason TEXT NOT NULL DEFAULT '',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE upstream_account_status
  ADD COLUMN IF NOT EXISTS api_latency_ms INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS upstream_synced_models (
  id UUID PRIMARY KEY,
  upstream_account_id UUID NOT NULL REFERENCES upstream_accounts(id) ON DELETE CASCADE,
  normalized_model_name TEXT NOT NULL,
  upstream_model_name TEXT NOT NULL,
  display_name TEXT NOT NULL DEFAULT '',
  native_wire_protocol TEXT NOT NULL DEFAULT '',
  supported_wire_protocols TEXT[] NOT NULL DEFAULT '{}',
  capabilities TEXT[] NOT NULL DEFAULT '{}',
  status TEXT NOT NULL DEFAULT 'active',
  raw_metadata JSONB NOT NULL DEFAULT '{}',
  last_synced_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(upstream_account_id, normalized_model_name, upstream_model_name)
);

CREATE TABLE IF NOT EXISTS upstream_account_events (
  id UUID PRIMARY KEY,
  upstream_account_id UUID NOT NULL REFERENCES upstream_accounts(id) ON DELETE CASCADE,
  operation TEXT NOT NULL,
  status TEXT NOT NULL,
  error_class TEXT NOT NULL DEFAULT '',
  message TEXT NOT NULL DEFAULT '',
  latency_ms INTEGER NOT NULL DEFAULT 0,
  metadata JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS upstream_account_events_account_created_idx
  ON upstream_account_events(upstream_account_id, created_at DESC);
`
