# RelayDeck LinuxDO OAuth Account Aggregation Design

Date: 2026-06-11

## Context

Most upstream sites RelayDeck aggregates are deployed with `new-api`, with a
smaller number on `sub2api`. On these deployments, end users do not register
with a local username and password. They sign in through **LinuxDO OAuth**
(`https://connect.linux.do`). This raises a specific aggregation problem:

> There is no static account password RelayDeck can store and replay. The only
> thing the OAuth flow produces is a short-lived browser session. RelayDeck
> needs a durable credential it can hold and reuse for account-level operations
> (balance, quota, profile, model sync, and check-in) without driving the OAuth
> flow on every call.

This document explains how the two upstream platforms actually authenticate
after LinuxDO login, and defines a first-version aggregation approach that
fits the existing site management credential model in
[`2026-06-11-site-management-upstream-accounts-design.md`](2026-06-11-site-management-upstream-accounts-design.md).

This milestone is scoped to **manual, on-demand check-in** (single or
one-click batch). Scheduled automatic check-in is explicitly out of scope.

## Key Finding: LinuxDO OAuth Is Only the Entry Point

LinuxDO OAuth is used for **registration and binding**, not for ongoing API
authentication. After the OAuth callback completes, each platform issues its
own native long-lived credential. That native credential, not the LinuxDO
session, is what RelayDeck aggregates.

### new-api authentication model

Verified against `Reference/new-api`:

- The LinuxDO callback (`controller/linuxdo.go` → `setupLogin` in
  `controller/user.go`) writes a **server-side session cookie** holding
  `id`, `username`, `role`, `status`, `group`. This cookie is short-lived.
- Separately, `GET /api/user/token` (`controller.GenerateAccessToken`) issues a
  **29–32 character long-lived access token**, stored on the user row
  (`model/user.go`: `AccessToken *string gorm:"type:char(32);uniqueIndex"`).
  It does not expire unless the user regenerates it.
- Decisive detail: the check-in route is registered under `UserAuth()`:

  ```text
  selfRoute.GET("/checkin", controller.GetCheckinStatus)
  selfRoute.POST("/checkin", middleware.TurnstileCheck(), controller.DoCheckin)
  ```

  And `UserAuth` (`middleware/auth.go`, `authHelper`) accepts the access token
  when no session is present: if the session has no `username`, it reads the
  `Authorization` header and calls `model.ValidateAccessToken`. The same
  handler also requires a `New-Api-User` header that must equal the
  authenticated user id.

This means RelayDeck can perform account-level operations on new-api with a
**stateless header pair**, no cookie and no repeated OAuth:

```text
POST /api/user/checkin
Authorization: <access_token>
New-Api-User: <user_id>
```

The same auth works for `GET /api/user/self`, `GET /api/user/checkin`
(status), and the user-side endpoints that run under `UserAuth`. `GET
/api/user/self` returns the user id, `quota`, and `used_quota`, so RelayDeck can
use it for both credential validation and account-balance refresh.

> Caveat: `POST /api/user/checkin` is wrapped by `middleware.TurnstileCheck()`.
> On deployments with Turnstile enabled for check-in, the request will be
> rejected pending human verification. RelayDeck must classify this as
> `action_required`, not `failed`. See Human Verification below.

### sub2api authentication model

Verified against `Reference/sub2api/backend/internal/service/auth_service.go`:

- sub2api issues **JWT access tokens** plus **refresh tokens** (prefix `rt_`).
  JWT claims carry a `token_version` that is bumped on password change to
  invalidate old tokens.
- Access tokens expire; refresh tokens can mint a new access token via the
  refresh endpoint until they themselves expire or are revoked/reused.
- LinuxDO login on sub2api goes through `LoginOrRegisterOAuth`, after which the
  normal JWT pair is issued — again, the OAuth step is only the entry point.

For sub2api, the durable credential RelayDeck stores is the **refresh token**.
The adapter exchanges it for a short-lived access token when needed, and uses
that access token as a bearer for account-level endpoints.

## Goals

- Define how RelayDeck obtains and stores a durable credential for LinuxDO-OAuth
  upstream accounts on new-api and sub2api.
- Make account-level operations (profile, balance/quota, model sync, check-in)
  work using that durable credential, without re-running OAuth per call.
- Support manual single check-in and manual one-click batch check-in across
  selected accounts, with per-account success/failure results.
- Fit entirely within the existing two-layer credential model and the existing
  `/api/admin/upstreams/...` contract. No new top-level concepts.
- Make failure modes explicit and actionable (expired credential, Turnstile
  required) so an administrator can fix the affected accounts by hand.
- Detect an expired/invalid credential during normal operations and only then
  prompt the operator with a guided re-acquisition flow, instead of polling or
  nagging on a schedule.

## Non-Goals

- No scheduled/automatic check-in in this milestone. Check-in is triggered
  manually (single or batch) only.
- No headless-browser automation of the LinuxDO OAuth flow.
- No bypass of CAPTCHA, Turnstile, QR, or email-code challenges.
- No attempt to retrieve or store the user's LinuxDO password or LinuxDO session.
- No parsing of the LinuxDO OAuth callback URL to recover a platform credential.
  The callback `code` is issued for the upstream site's `client_id`, redirects
  to the site's own `redirect_uri` (`/api/oauth/linuxdo`), and can only be
  exchanged with that site's `client_secret` — none of which RelayDeck holds.
  Even a captured `code` would yield a LinuxDO identity token, not the site's
  own session/access token. Credentials are obtained from the logged-in browser
  instead (see Credential Acquisition).
- No change to the runtime `/v1/*` gateway forwarding path; this milestone is
  about account-level management operations only.

## Credential Acquisition

This milestone uses **manual credential entry** (operator pastes a durable
credential once per account). Automated OAuth-proxy acquisition is recorded as
a future option but is not built here.

### new-api: paste access token + user id

The operator, who controls the upstream user account, generates an access token
once on their new-api site and records it in RelayDeck:

1. On the new-api site, the user calls `GET /api/user/token` (the site UI
   exposes this as "generate access token") to obtain the 29–32 char token.
2. The user notes their numeric new-api user id (visible in their profile /
   returned by `GET /api/user/self`).
3. In RelayDeck's Add/Edit Site drawer, the operator selects account credential
   kind **`new_api_access_token`** and provides both values.

Because new-api access tokens do not expire on their own, this is a one-time
manual step per account.

### sub2api: paste refresh token

1. After signing in to the sub2api site (via LinuxDO), the operator retrieves
   the refresh token issued to the browser session (cookie/local storage,
   prefix `rt_`).
2. In the drawer, the operator selects account credential kind
   **`sub2api_refresh_token`** and provides it.

The adapter will use the refresh token to obtain access tokens on demand. If
the platform returns a rotated refresh token, the adapter reports it to the
admin handler so the handler can re-encrypt and persist the replacement.

### Guided re-acquisition on credential expiry

RelayDeck should not nag the operator on a schedule. Instead it detects an
expired/invalid credential during normal operations and, only then, surfaces a
guided re-acquisition flow for the affected account.

Detection (no extra polling — reuse signals the adapters already produce):

- An account-level call (`test-account`, `refresh-quota`, `checkin`, or a
  background account read) returns `401/403`, or a sub2api refresh fails with an
  expired/reused/revoked refresh token.
- The adapter sets `account_status = expired` and records the reason in
  `last_error_class` / `last_error_message`.

Guidance (frontend reaction to the `expired` state):

- The row's account-credential badge shows `expired` and exposes a
  **"重新获取凭据"** affordance (e.g. on the row action or in the inspect drawer).
- Activating it opens a guided panel pre-filled for the account's platform kind.
  The panel restates the steps and the exact command to run, then provides the
  same credential field(s) used at creation so the operator can paste the fresh
  value in place. Saving re-encrypts `account_credential_enc` and clears the
  `expired` status on the next successful account call.

The panel can pre-build a convenience link to the upstream site's login page
(`<base_url>` is already stored on the account), so the operator lands on the
right site quickly. RelayDeck does not automate the login itself; it only
shortens the manual path.

#### new-api re-acquisition guidance

1. Open `<base_url>` and sign in (LinuxDO OAuth) if the session has lapsed.
2. Generate a fresh access token (site UI "generate access token", i.e.
   `GET /api/user/token`).
3. Paste the token and the numeric user id back into the panel.

Because new-api access tokens do not self-expire, an `expired` result here
usually means the token was regenerated/reset upstream, so re-pasting resolves
it.

#### sub2api re-acquisition guidance

1. Open `<base_url>` and sign in (LinuxDO OAuth).
2. In the browser console on that site, copy the current refresh token:

   ```js
   localStorage.getItem('refresh_token')
   ```

   (Verified storage location: `sub2api/frontend/src/api/auth.ts` stores
   `refresh_token` in `localStorage`.)
3. Paste the `rt_...` value back into the panel.

### Future option (not in this milestone): OAuth-proxy acquisition

RelayDeck could host its own callback and drive the LinuxDO authorization flow
to capture the platform credential automatically, eliminating the manual paste.
This requires handling LinuxDO `state`/`redirect_uri` validation and, for
new-api, calling `GET /api/user/token` after obtaining a session. It is
deferred; the manual path above is the supported first version.

> Why parsing the OAuth callback URL is not a viable shortcut: the authorize URL
> new-api builds (`web/.../helpers/api.js`) uses the **site's own** `client_id`
> and a `redirect_uri` pointing back to `<site>/api/oauth/linuxdo`. The callback
> `code` is delivered to the upstream site, is exchanged with the site's
> `client_secret` (which RelayDeck does not hold), and even then yields only a
> LinuxDO identity — not the site's own login credential. The durable
> credential is produced once, inside the site, during that callback. So the
> supported path is to read the credential the site already issued to the
> browser (access token / refresh token), not to intercept the OAuth flow.

## Credential Storage Model

The existing `account_credential_kind` + `account_credential_enc` columns are
reused. The change is that some kinds carry **structured** secrets, stored as
encrypted JSON in `account_credential_enc`.

New/confirmed account credential kinds:

- `new_api_access_token` — structured: `{ "access_token": "...", "user_id": 123 }`
- `sub2api_refresh_token` — structured: `{ "refresh_token": "rt_..." }`
- existing generic kinds (`cookie`, `access_token`, `refresh_token`, `json`)
  remain valid for non-LinuxDO or manual cases.

Rules:

- The structured JSON is encrypted as a whole via the existing secretbox before
  persistence. No plaintext component is ever returned to the frontend.
- Responses expose only capability/state booleans (e.g. `has_account_credential`,
  `account_credential_kind`) and a masked hint, never the token or user id.
- For sub2api, when a refresh produces a rotated refresh token, the adapter
  returns a credential update to the admin handler. The handler, not the
  transport adapter, re-encrypts and updates `account_credential_enc` in place
  and records an event. This keeps encryption and persistence in the HTTP/store
  layer that already owns account writes.

## Adapter Behavior

Adapters live in `backend/internal/upstream`. This milestone modifies the
existing `applyAccountCredential` path and the new-api/sub2api account adapters.

### new-api adapter

For account-level calls, build the stateless header pair from the decrypted
structured credential:

```text
Authorization: <access_token>
New-Api-User: <user_id>
Accept: application/json
```

Endpoint mapping:

- `TestAccountCredential` → `GET /api/user/self`
  - HTTP 2xx with `success=true` or a plain user object ⇒
    `account_status = valid`
  - HTTP 2xx with `success=false` and an auth/credential message ⇒
    `account_status = expired`
  - HTTP 401/403 ⇒ `account_status = expired`
  - challenge-indicating body ⇒ `action_required`
- `RefreshQuota` → `GET /api/user/self` using the same account credential
  headers, normalized as `balance_amount = quota - used_quota` and
  `balance_unit = quota`. If a deployment returns only `quota`, use that as the
  remaining balance and leave `used_quota` out of the event metadata.
- `Checkin` → `POST /api/user/checkin`
  - HTTP 2xx with `success=true` ⇒ `checkin_status = checked`, record
    `quota_awarded` in event metadata when present
  - HTTP 2xx with `success=false` and an auth/credential message ⇒
    `account_status = expired`, `checkin_status = action_required`
  - Turnstile/verification rejection ⇒ `checkin_status = action_required`
  - 401/403 ⇒ credential expired, surface as `action_required` for manual fix
- `GetCheckinStatus` (optional, for display) → `GET /api/user/checkin`.

Model sync and API health continue to use the **API key** (`Authorization:
Bearer <api_key>` against `GET /v1/models`), unchanged. The access token is for
account-level/session endpoints only.

### sub2api adapter

1. If no cached access token (or it is expired), call `POST
   /api/v1/auth/refresh` with JSON body `{ "refresh_token": "rt_..." }` to
   obtain a new JWT access token.
2. Use `Authorization: Bearer <jwt>` for account-level endpoints
   (`/api/v1/user/profile`, `/api/v1/user/platform-quotas`, usage dashboards).
3. If the refresh response returns a new `refresh_token`, return a credential
   update containing `{ "refresh_token": "new_rt_..." }`; the admin handler
   encrypts and persists it on the upstream account.
4. If refresh fails with an expired/reused/revoked refresh token, set
   `account_status = expired` and surface `action_required` so the operator can
   paste a fresh refresh token.

sub2api check-in is not assumed to exist. If a deployment exposes a user-side
check-in endpoint, the adapter can add it without changing the page contract;
until then sub2api check-in status stays `unsupported`.

### Header shaping note

Per the project's client-compatibility intent, account-level requests should
look like ordinary user-dashboard traffic from the original client where
practical (standard `Accept`, no RelayDeck-specific headers leaking upstream).
This milestone does not require deep client emulation for management calls, but
adapters must not add headers that reveal the aggregation layer unnecessarily.

## Check-in UX

The site management page already exposes per-row and batch actions. This
milestone clarifies their behavior for LinuxDO-OAuth accounts.

- **Single check-in**: existing per-row check-in action calls
  `POST /api/admin/upstreams/accounts/{id}/checkin`. The action response must
  include the updated status snapshot so the frontend can update that row
  without reloading the full list.
- **One-click batch check-in**: existing
  `POST /api/admin/upstreams/accounts/batch/checkin` runs check-in for all
  selected accounts and returns a per-account result array. Each successful or
  handled result includes the updated status snapshot for that account. The
  frontend should surface this as a summary (e.g. "12 checked, 3 action
  required, 1 failed") and patch visible rows from the returned statuses. The
  batch button should be present in the toolbar alongside the existing batch
  actions.
- **Failures are left to the operator**: accounts returning `action_required`
  (Turnstile) or `expired` (bad credential) are listed so the operator can
  handle them manually. RelayDeck does not retry or auto-resolve them.

Check-in status values reuse the existing set: `unsupported`,
`not_configured`, `checked`, `unchecked`, `failed`, `action_required`.

## Backend Contract

No new endpoints are required. The relevant existing routes are reused:

```text
POST /api/admin/upstreams/accounts/{id}/test-account
POST /api/admin/upstreams/accounts/{id}/refresh-quota
POST /api/admin/upstreams/accounts/{id}/checkin
POST /api/admin/upstreams/accounts/batch/checkin
```

Create/update requests may include the structured account credential in
plaintext; handlers encrypt before persistence and never echo it back.

Action responses extend the existing shape:

```json
{
  "id": "account-id",
  "status": "success",
  "message": "",
  "account_status": {
    "api_status": "healthy",
    "account_status": "valid",
    "checkin_status": "checked",
    "balance_amount": 123,
    "balance_unit": "quota"
  }
}
```

Batch responses return the same per-account result objects in `results`.

## Security

- Encrypt the full structured credential JSON (access token + user id, or
  refresh token) before writing to PostgreSQL, via the existing secretbox.
- Never return token values or user ids in API responses, logs, or events.
- Redact credential values from any captured upstream response bodies stored in
  event metadata.
- On sub2api refresh-token rotation, ensure the old value is overwritten, not
  appended, so a stale token cannot be reused.
- Treat `401/403` on account endpoints as a credential problem to be fixed by an
  operator, not as a transient error to retry blindly.

## Error Handling

Reuse the normalized error classes from the site management design. Mapping
specific to this milestone:

- new-api access token rejected (`401/403`) ⇒ `credential_expired` ⇒
  `account_status = expired`, action surfaced as `action_required`.
- new-api access token rejected by a HTTP 200 response with `success=false` and
  an auth/credential message ⇒ same as credential expired. new-api commonly
  uses this shape for invalid access tokens, so adapters must inspect the body
  instead of relying only on status code.
- new-api check-in blocked by Turnstile/verification ⇒ `action_required` ⇒
  `checkin_status = action_required` (never `failed`).
- sub2api refresh token expired/reused/revoked ⇒ `credential_expired`.
- missing `New-Api-User` or id mismatch (misconfigured credential) ⇒
  `invalid_response` with a clear operator-facing message, treated as a
  credential configuration error.

## Testing

Backend (`httptest`-based adapter tests, no live upstream):

- new-api `TestAccountCredential`: 2xx user object or `success=true` self ⇒
  valid; HTTP 200 `success=false` auth body ⇒ expired; 401 ⇒ expired.
- new-api `RefreshQuota`: `GET /api/user/self` uses account credential headers
  and maps `quota - used_quota` into `balance_amount`.
- new-api `Checkin`: HTTP 200 `success=true` ⇒ checked with awarded-quota
  metadata; HTTP 200 `success=false` auth body ⇒ action_required plus expired
  account credential; Turnstile rejection body ⇒ action_required; 401 ⇒
  action_required.
- new-api header assertion: requests carry both `Authorization: <access_token>`
  and `New-Api-User: <user_id>` derived from the structured credential.
- sub2api refresh flow: `POST /api/v1/auth/refresh` with stored refresh token
  yields a usable bearer; rotated refresh token is returned as a credential
  update and persisted by the admin handler; invalid refresh token ⇒ expired.
- credential encryption round-trip: structured JSON encrypts/decrypts and is
  never present in serialized responses.

Frontend:

- drawer offers `new_api_access_token` (two fields: token + user id) and
  `sub2api_refresh_token` (one field) credential kinds, with platform-specific
  helper text and a hint on where to generate the new-api token;
- batch check-in button present in the toolbar; batch result summary rendered;
  per-row status patches from action responses without full reload;
- TypeScript build passes.

End-to-end behavior:

- a new-api account with API key + `new_api_access_token` can test account,
  refresh quota, and check in;
- check-in on a Turnstile-protected new-api deployment shows `action_required`,
  not `failed`;
- a sub2api account with `sub2api_refresh_token` can read profile/quota and, on
  refresh-token expiry, shows `expired` for manual re-entry.

## Open Design Notes

- Scheduled automatic check-in is intentionally deferred. The structured
  credential model defined here is sufficient for a future scheduler to reuse
  without schema changes.
- OAuth-proxy credential acquisition (auto-capture instead of manual paste) is
  the natural next step if manual entry proves too tedious at scale.
- If a future new-api version changes the access-token header contract
  (`New-Api-User`, `Auth-Version`), the new-api adapter is the only place that
  needs updating.
