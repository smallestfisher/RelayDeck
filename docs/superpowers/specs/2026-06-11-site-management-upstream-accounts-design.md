# RelayDeck Site Management Upstream Accounts Design

Date: 2026-06-11

## Context

The site management page from `Design_image/5.PNG` is the first full page-level implementation for RelayDeck's aggregation management console. In this product, a "site" means an upstream ordinary user account that RelayDeck can call through, not an upstream administrator account.

The first supported upstream platform kinds are:

- `new_api`
- `sub2api`

Direct provider accounts such as OpenAI, Anthropic, Gemini, and Azure OpenAI are intentionally out of scope for this page milestone.

Reference directories are used only to inspect behavior:

- `Reference/new-api`: user-facing relay, token, billing, check-in, model, and session endpoints.
- `Reference/sub2api`: user-facing API key, model, usage, quota, JWT, and gateway endpoints.

## Goals

- Implement the site management page as a complete usable interaction, not a minimal mock.
- Let administrators add, edit, test, disable, delete, batch-check, and inspect upstream ordinary user accounts.
- Use API keys for model API calls, model discovery, health probes, and runtime forwarding.
- Use optional account credentials for account-only state such as balance, quota, profile, check-in status, and user dashboard data.
- Keep credential capability visible in the UI so missing account credentials are not treated as upstream failures.
- Persist upstream accounts, synced model data, status snapshots, and operation history in the backend.
- Keep the frontend naming close to the prototype: "site management"; keep backend naming precise: upstream accounts.

## Non-Goals

- No upstream administrator API integration.
- No direct provider account support in this page milestone.
- No 2FA or TOTP support.
- No automated bypass for CAPTCHA, Turnstile, QR confirmation, or other human verification.
- No billing settlement engine or downstream user quota changes.
- No attempt to manage upstream platform users, channels, or internal admin settings.

## Credential Model

Each upstream account has two independent credential layers.

### API credential

This is required. It is the API key RelayDeck uses for:

- `GET /v1/models` model listing where supported;
- lightweight model call probes;
- runtime request forwarding;
- token or key-level usage endpoints where the platform exposes them.

The API key is stored encrypted. The UI only shows a masked prefix.

### Account credential

This is optional. It is used only for endpoints that require a logged-in user session rather than an API key:

- account profile;
- account-level balance or platform quota;
- check-in status;
- check-in action;
- user dashboard usage;
- user-owned API key list, when exposed by the upstream platform.

Supported first-version account credential kinds:

- cookie or raw session header;
- access token;
- refresh token;
- combined JSON credentials for platform-specific auth headers.

If account credentials are missing, account-only fields should display `Not configured`, not `Failed`.

### Human verification

If an upstream action requires CAPTCHA, Turnstile, QR confirmation, email code, or another challenge, RelayDeck records the operation as `action_required`. The UI should show an actionable state and leave completion to an administrator. RelayDeck must not attempt to bypass the challenge.

## Platform Behavior

### new-api

Observed reference behavior:

- API key endpoints:
  - `GET /v1/models`
  - `POST /v1/chat/completions`
  - `POST /v1/messages`
  - `POST /v1/responses`
  - `GET /api/usage/token`
  - `GET /v1/dashboard/billing/subscription`
  - `GET /v1/dashboard/billing/usage`
- Session-backed user endpoints:
  - `GET /api/user/self`
  - `GET /api/user/models`
  - `GET /api/user/checkin`
  - `POST /api/user/checkin`
  - `GET /api/token`

`POST /api/user/checkin` may require Turnstile depending on deployment settings. When that happens, RelayDeck returns `action_required`.

### sub2api

Observed reference behavior:

- API key endpoints:
  - `GET /v1/models`
  - `GET /v1/usage`
  - `POST /v1/messages`
  - `POST /v1/responses`
  - `POST /v1/chat/completions`
  - `GET /v1beta/models`
- JWT-backed user endpoints:
  - `GET /api/v1/user/profile`
  - `GET /api/v1/user/platform-quotas`
  - `GET /api/v1/keys`
  - `GET /api/v1/usage/dashboard/stats`
  - `GET /api/v1/usage/dashboard/trend`
  - `GET /api/v1/usage/dashboard/models`

Sub2API check-in is not assumed in this milestone. If a deployment exposes an equivalent user-side check-in endpoint later, the adapter can add it without changing the page contract.

## Page UX

The page keeps the prototype structure:

- header with title and primary `Add site` action;
- four metric cards;
- filter/search toolbar;
- selectable data table;
- pagination;
- right-side drawer for add/edit;
- icon actions per row.

### Metric cards

Cards should derive from backend account data:

- total upstream accounts;
- healthy API-call accounts;
- accounts with warnings;
- accounts needing manual action or account credential setup.

### Filters

Filters:

- search by name, base URL, note, or model;
- platform kind: all, new-api, sub2api;
- API call status;
- account credential status;
- check-in status;
- latency band.

### Table columns

Columns:

- selection checkbox;
- site name, code, base URL, note;
- platform kind;
- API call status;
- account credential status;
- synced model count;
- average latency;
- balance or quota summary;
- check-in status;
- last checked time;
- row actions.

Status values:

- API call status: `unknown`, `healthy`, `warning`, `failed`, `disabled`.
- account credential status: `not_configured`, `valid`, `expired`, `failed`, `action_required`.
- check-in status: `unsupported`, `not_configured`, `checked`, `unchecked`, `failed`, `action_required`.

### Row actions

Actions:

- view health and recent operation history;
- edit;
- test API credential;
- test account credential;
- sync models;
- refresh balance or quota;
- check in when supported;
- disable or enable;
- delete.

Deletes should be confirmed and should remove secrets. Disable keeps the account and synced models but excludes it from routing.

### Add/Edit Drawer

Sections:

- basic information: name, platform kind, base URL, priority, note;
- API credential: API key, request header mode, masked display after save;
- account credential: credential kind and secret fields;
- status options: enabled, include in routing, auto sync models, auto refresh quota, auto check-in when possible;
- actions: test API credential, test account credential, save, save and sync models.

The drawer should change helper text and credential fields by platform kind.

## Backend Contract

All endpoints are under authenticated RelayDeck admin routes.

```text
GET    /api/admin/upstreams/accounts
POST   /api/admin/upstreams/accounts
GET    /api/admin/upstreams/accounts/{id}
PUT    /api/admin/upstreams/accounts/{id}
DELETE /api/admin/upstreams/accounts/{id}
POST   /api/admin/upstreams/accounts/{id}/enable
POST   /api/admin/upstreams/accounts/{id}/disable
POST   /api/admin/upstreams/accounts/{id}/test-api
POST   /api/admin/upstreams/accounts/{id}/test-account
POST   /api/admin/upstreams/accounts/{id}/sync-models
GET    /api/admin/upstreams/accounts/{id}/models
POST   /api/admin/upstreams/accounts/{id}/refresh-quota
POST   /api/admin/upstreams/accounts/{id}/checkin
GET    /api/admin/upstreams/accounts/{id}/events
POST   /api/admin/upstreams/accounts/batch/test-api
POST   /api/admin/upstreams/accounts/batch/sync-models
POST   /api/admin/upstreams/accounts/batch/refresh-quota
POST   /api/admin/upstreams/accounts/batch/checkin
```

Responses must never return plaintext secrets. Create and update requests may contain plaintext secrets, but handlers encrypt before persistence.

## Data Model

### upstream_accounts

Durable account metadata:

- `id`
- `name`
- `code`
- `platform_kind`
- `base_url`
- `enabled`
- `include_in_routing`
- `priority`
- `api_key_enc`
- `api_key_prefix`
- `account_credential_kind`
- `account_credential_enc`
- `auto_sync_models`
- `auto_refresh_quota`
- `auto_checkin`
- `note`
- `created_at`
- `updated_at`

### upstream_account_status

Latest status snapshot:

- `upstream_account_id`
- `api_status`
- `account_status`
- `checkin_status`
- `model_count`
- `latency_ms`
- `balance_amount`
- `balance_unit`
- `last_api_checked_at`
- `last_account_checked_at`
- `last_model_synced_at`
- `last_checkin_at`
- `last_error_class`
- `last_error_message`
- `action_required_reason`
- `updated_at`

### upstream_models

Synced model records:

- `id`
- `upstream_account_id`
- `normalized_model_name`
- `upstream_model_name`
- `display_name`
- `native_wire_protocol`
- `supported_wire_protocols`
- `capabilities`
- `status`
- `raw_metadata`
- `last_synced_at`

### upstream_account_events

Operation history:

- `id`
- `upstream_account_id`
- `operation`
- `status`
- `error_class`
- `message`
- `latency_ms`
- `metadata`
- `created_at`

## Adapter Interfaces

The backend should isolate platform differences behind small interfaces.

```go
type PlatformAdapter interface {
    TestAPI(ctx context.Context, account domain.UpstreamAccount) (domain.UpstreamAccountStatus, error)
    TestAccount(ctx context.Context, account domain.UpstreamAccount) (domain.UpstreamAccountStatus, error)
    SyncModels(ctx context.Context, account domain.UpstreamAccount) (domain.ModelSyncResult, error)
    RefreshQuota(ctx context.Context, account domain.UpstreamAccount) (domain.QuotaRefreshResult, error)
    Checkin(ctx context.Context, account domain.UpstreamAccount) (domain.CheckinResult, error)
}
```

Adapters should return normalized statuses and preserve raw upstream payloads in event metadata when useful. Adapter failures should classify errors instead of passing through free-form text only.

## Security

- Encrypt all upstream API keys and account credentials before writing to PostgreSQL.
- Return only masked secret previews to the frontend.
- Redact secrets from logs, events, and failed request payloads.
- Keep account credential configuration optional.
- Do not store 2FA or TOTP secrets.
- Do not attempt to bypass human verification challenges.
- Mark unsupported or missing-credential operations explicitly instead of retrying them blindly.

## Error Handling

Use normalized error classes:

- `auth_error`
- `credential_missing`
- `credential_expired`
- `quota_error`
- `rate_limit`
- `timeout`
- `transport_error`
- `protocol_mismatch`
- `invalid_response`
- `upstream_5xx`
- `unsupported`
- `action_required`
- `unknown_error`

The frontend should show the current normalized status in the table and keep detailed messages in the row detail/history view.

## Testing

Backend tests:

- domain tests for account, credential, status, model, and event types;
- handler tests for list/create/update/delete, secret redaction, status actions, batch actions, and session protection;
- adapter tests with `httptest` for new-api and sub2api model list, usage, account credential, unsupported check-in, and action-required responses;
- store tests for CRUD, status upsert, model sync upsert, and event insertion.

Frontend verification:

- TypeScript build must pass;
- `SitesPage` should render loading, empty, error, and populated states;
- form validation should prevent saving without platform, name, base URL, and API key;
- table filters and batch selections should work against backend data;
- action results should update row status without full-page reload.

End-to-end behavior:

- adding a new-api account with only API key shows API healthy and account credentials not configured;
- adding a new-api account with account credential can refresh account state and attempt check-in;
- adding a sub2api account with only API key can list models and usage;
- unsupported check-in displays unsupported, not failed;
- human verification displays action required, not automatic failure.

## Open Design Notes

- The first implementation can use a simple encryption key from environment configuration. A later milestone can move to KMS or external secret storage.
- Model protocol normalization should align with the platform aggregation design, but this page owns only account and model visibility, not final routing policy.
- Automatic schedules for sync/quota/check-in can be persisted now, but background scheduling may be implemented after the manual page actions are stable.
