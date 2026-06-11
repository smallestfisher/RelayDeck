# Site Management Upstream Accounts Implementation Plan

> **Execution mode:** build by page goal, not by minimum slice. Complete the backend contract, frontend interaction, tests, local commits, and final verification before any remote push.

**Goal:** Implement the `Design_image/5.PNG` site management page as a real upstream ordinary-user-account management workflow for `new-api` and `sub2api`.

**Architecture:** Backend owns upstream account records, encrypted credentials, platform adapters, status snapshots, model sync, and admin APIs. Frontend keeps the "站点管理" product language while using backend `upstream account` semantics.

**Out of scope:** upstream admin APIs, direct provider accounts, 2FA/TOTP, CAPTCHA/Turnstile/QR bypass, billing settlement.

---

## Task 1: Domain And Secret Foundation

**Files**

- `backend/internal/domain/domain.go`
- `backend/internal/domain/domain_test.go`
- `backend/internal/secretbox/secretbox.go`
- `backend/internal/secretbox/secretbox_test.go`
- `backend/internal/config/config.go`

**Work**

- Add domain types for upstream accounts, platform kind, credential kind, API status, account credential status, check-in status, synced models, quota refresh results, check-in results, and event records.
- Add helper semantics for route eligibility and manual-action status.
- Add AES-GCM secret encryption for upstream API keys and account credentials.
- Add `APP_UPSTREAM_SECRET_KEY` configuration.

**Tests**

- Domain tests prove API credential and account credential are independent.
- Domain tests prove missing optional account credential is not a failure.
- Secretbox tests cover round trip, empty secret, invalid key, and wrong-key failure.

**Verify**

- `cd backend && GOCACHE=/tmp/go-build go test ./internal/domain ./internal/secretbox ./internal/config -v`

**Commit**

- `git commit -m "Add upstream account domain and secret handling"`

**Status**

- Completed and committed: `be79eaf Add upstream account domain and secret handling`
- Verified with: `cd backend && GOCACHE=/tmp/go-build go test ./internal/domain ./internal/secretbox ./internal/config -v`

---

## Task 2: Storage And Schema

**Files**

- `backend/migrations/0001_initial.sql`
- `backend/internal/store/upstream.go`
- `backend/internal/store/admin.go`
- `backend/internal/store/postgres/upstreams.go`
- `backend/internal/store/postgres/upstreams_test.go`

**Work**

- Add upstream account store interface.
- Do not extend legacy `MemoryStore`; remove the site-management dependency on prototype memory data.
- Add PostgreSQL tables for upstream accounts, latest status snapshots, synced models, and operation events.
- Implement PostgreSQL-backed upstream account store.
- Extend `AdminStore` so admin handlers can access upstream account storage.
- Keep Redis as the runtime/session/cache direction; do not add mock memory-backed upstream account behavior for the page.

**Tests**

- Store contract tests cover create, update, list, delete, status upsert, model replacement, and event history.
- PostgreSQL tests cover persistence when `DATABASE_URL` points to a test database, otherwise skip.

**Verify**

- `cd backend && GOCACHE=/tmp/go-build go test ./internal/store -v`
- `cd backend && GOCACHE=/tmp/go-build go test ./internal/store/postgres -run TestUpstream -v`

**Commit**

- `git commit -m "Persist upstream account records"`

**Status**

- Completed and committed: `cc7086d Persist upstream account records`
- Legacy `MemoryStore` was removed instead of extended.
- Verified with:
  - `cd backend && GOCACHE=/tmp/go-build go test ./internal/store -v`
  - `cd backend && DATABASE_URL=postgres://postgres:postgres@localhost:5432/relaydeck_test?sslmode=disable GOCACHE=/tmp/go-build go test ./internal/store/postgres -run TestUpstream -v`
  - `cd backend && GOCACHE=/tmp/go-build go test ./internal/http/admin ./internal/http/gateway ./internal/app -v`

---

## Task 3: Platform Adapters

**Files**

- `backend/internal/upstream/account_adapter.go`
- `backend/internal/upstream/newapi_account_adapter.go`
- `backend/internal/upstream/sub2api_account_adapter.go`
- `backend/internal/upstream/account_adapter_test.go`

**Work**

- Add a platform adapter interface for API test, account credential test, model sync, quota refresh, and check-in.
- Implement `new-api` adapter using user/API-key endpoints only.
- Implement `sub2api` adapter using user/API-key endpoints only.
- Normalize unsupported operations, missing credentials, auth errors, rate limits, upstream errors, invalid responses, and human-verification requirements.
- Do not use upstream administrator APIs.
- Do not add 2FA/TOTP support.

**Tests**

- `new-api` adapter tests use `httptest` for model list, token usage, account credential failure, and check-in action-required behavior.
- `sub2api` adapter tests use `httptest` for model list, `/v1/usage`, JWT/cookie profile access, and unsupported check-in.

**Verify**

- `cd backend && GOCACHE=/tmp/go-build go test ./internal/upstream -v`

**Commit**

- `git commit -m "Add upstream account platform adapters"`

**Status**

- Completed and committed: `50fc19b Add upstream account platform adapters`
- Verified with: `cd backend && GOCACHE=/tmp/go-build go test ./internal/upstream -v`

---

## Task 4: Admin API

**Files**

- `backend/internal/http/admin/handler.go`
- `backend/internal/http/admin/upstream_handler.go`
- `backend/internal/http/admin/upstream_handler_test.go`
- `backend/internal/app/app.go`
- `backend/internal/app/app_test.go`
- `.env.example`

**Work**

- Mount authenticated admin routes for upstream accounts.
- Add CRUD endpoints for accounts.
- Add action endpoints for API test, account credential test, model sync, quota refresh, check-in, event history, and batch actions.
- Encrypt incoming secrets before persistence.
- Redact secrets in every response.
- Wire app dependencies: secretbox, upstream store, and platform adapter registry.
- Add `APP_UPSTREAM_SECRET_KEY` to `.env.example`.

**API Surface**

- `GET /api/admin/upstreams/accounts`
- `POST /api/admin/upstreams/accounts`
- `GET /api/admin/upstreams/accounts/{id}`
- `PUT /api/admin/upstreams/accounts/{id}`
- `DELETE /api/admin/upstreams/accounts/{id}`
- `POST /api/admin/upstreams/accounts/{id}/test-api`
- `POST /api/admin/upstreams/accounts/{id}/test-account`
- `POST /api/admin/upstreams/accounts/{id}/sync-models`
- `GET /api/admin/upstreams/accounts/{id}/models`
- `POST /api/admin/upstreams/accounts/{id}/refresh-quota`
- `POST /api/admin/upstreams/accounts/{id}/checkin`
- `GET /api/admin/upstreams/accounts/{id}/events`
- Batch equivalents for `test-api`, `sync-models`, `refresh-quota`, and `checkin`.

**Tests**

- Missing session returns 401.
- Create/list/update responses never leak plaintext API keys or account credentials.
- CRUD persists expected data.
- Action endpoints update status and append events.
- Batch endpoints return per-account results.
- App route test proves upstream routes are mounted after admin login.

**Verify**

- `cd backend && GOCACHE=/tmp/go-build go test ./internal/http/admin ./internal/app -v`

**Commit**

- `git commit -m "Add upstream account admin APIs"`

**Status**

- Completed and committed: `efd43f1 Add upstream account admin APIs`
- Verified with:
  - `cd backend && GOCACHE=/tmp/go-build go test ./internal/http/admin ./internal/app -v`
  - `cd backend && GOCACHE=/tmp/go-build go test ./...`

---

## Task 5: Frontend API Client And Types

**Files**

- `src/types.ts`
- `src/lib/adminApi.ts`
- `src/lib/format.ts`
- `src/components/ui/StatusBadge.tsx`

**Work**

- Add TypeScript types for upstream account records, status snapshots, platform kinds, credential kinds, and create/update inputs.
- Add admin API client methods for account CRUD and actions.
- Add status labels for `healthy`, `unknown`, `not_configured`, `valid`, `action_required`, and `unsupported`.
- Update status badge tone mapping for the new statuses.

**Tests**

- TypeScript build catches type mismatches.

**Verify**

- `npm run build`

**Commit**

- `git commit -m "Add upstream account frontend API types"`

**Status**

- Completed and committed: `60041e6 Add upstream account frontend API types`
- Verified with: `npm run build`

---

## Task 6: Site Management Page

**Files**

- `src/pages/SitesPage.tsx`
- `src/pages/sites/siteOptions.ts`
- `src/pages/sites/SiteDrawer.tsx`
- `src/pages/sites/SiteTable.tsx`

**Work**

- Replace mock-only site data with backend-loaded upstream accounts.
- Keep the 5.PNG layout structure: header, four metric cards, filters, selectable table, pagination, right-side drawer, row icon actions.
- Add filters for search, platform kind, API status, account credential status, check-in status, and latency band.
- Add add/edit drawer with platform-aware credential fields.
- Require name, code, base URL, and API key for new accounts.
- Support optional account credential kind/value.
- Exclude 2FA/TOTP fields entirely.
- Add row actions for edit, API test, account credential test, model sync, quota refresh, check-in, enable/disable, delete, and event/history view where implemented.
- Add batch actions for selected accounts.
- Show `未配置账号凭据`, `不支持`, and `需人工处理` as distinct states instead of generic failure.

**Tests**

- Build verifies type correctness.
- Manual browser check covers loading, empty, error, populated, add/edit drawer validation, row actions, and batch actions.

**Verify**

- `npm run build`

**Commit**

- `git commit -m "Connect site management page to upstream APIs"`

**Status**

- Completed and committed: `a6388f4 Connect site management page to upstream APIs`
- Verified with:
  - `cd backend && GOCACHE=/tmp/go-build go test ./internal/domain ./internal/http/admin ./internal/app -v`
  - `npm run build`

---

## Task 7: Full Verification

**Work**

- Run backend test suite.
- Run frontend build.
- Start the dev server for final UI inspection.
- Verify login and navigation to `站点管理`.
- Verify no 2FA/TOTP controls are present.
- Verify account-only fields do not show as failure when account credentials are missing.
- Verify unsupported check-in displays as unsupported.
- Verify action-required state displays distinctly.
- Review git status.

**Verify**

- `cd backend && GOCACHE=/tmp/go-build go test ./...`
- `npm run build`
- `npm run dev`
- `git status --short`

**Commit**

- If final verification required fixes, commit them with `git commit -m "Finalize site management upstream accounts"`.

**Status**

- In progress.
- Completed:
  - `cd backend && GOCACHE=/tmp/go-build go test ./...`
  - `npm run build`
  - `git status --short` before this status update was clean.
- Pending:
  - Start backend and frontend dev servers.
  - Browser-check login and navigation to `站点管理`.
  - Browser-check no 2FA/TOTP controls are present.
  - Browser-check missing account credentials show as `未配置`, unsupported check-in shows as `不支持`, and human verification states show as `需人工处理`.
- Remote push remains intentionally deferred until final feature verification and user confirmation.

**Remote Push**

- Do not push during implementation.
- Push only after the feature is complete, tests pass, and the user confirms it is ready to publish.

---

## Implementation Notes

- Work through the tasks in order.
- Use test-first development for backend behavior changes.
- Keep commits local until final approval.
- Keep reference code read-only.
- `Reference/new-api` and `Reference/sub2api` are for ordinary user/API-key endpoint behavior, not administrator integration.
