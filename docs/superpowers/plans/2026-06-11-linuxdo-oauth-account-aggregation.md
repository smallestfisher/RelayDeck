# LinuxDO OAuth Account Aggregation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Support LinuxDO-OAuth upstream account operations by storing new-api access-token/user-id credentials and sub2api refresh tokens, then using them for account test, quota refresh, and check-in.

**Architecture:** Keep durable credential storage in the existing `account_credential_kind` + encrypted `account_credential_enc` fields. Adapters parse structured plaintext credentials and return status snapshots plus optional credential rotations; the admin handler remains responsible for encryption, persistence, status/event storage, and API responses. Frontend forms serialize structured credential inputs into the same `account_credential` request field and patch rows from action response snapshots.

**Tech Stack:** Go 1.22 backend with `net/http` adapter tests, PostgreSQL-backed store through existing interfaces, React + TypeScript frontend with existing UI components, Vite build verification.

---

## Files

- Modify: `backend/internal/domain/domain.go`
  - Add `new_api_access_token` and `sub2api_refresh_token` credential kinds.
  - Add `UpstreamCredentialUpdate`.
  - Add `AccountCredentialTestResult`.
  - Extend `QuotaRefreshResult` and `CheckinResult` with optional credential update and account-error fields.
- Modify: `backend/internal/upstream/account_adapter.go`
  - Update `AccountAdapter.TestAccountCredential` signature.
  - Add JSON credential parsing helpers, new-api success/error helpers, and refresh-token response helpers.
  - Add request helper for new-api `New-Api-User` header.
- Modify: `backend/internal/upstream/newapi_account_adapter.go`
  - Use `GET /api/user/self` with structured new-api credentials for account test and quota refresh.
  - Inspect HTTP 200 `success:false`.
  - Use structured headers for check-in and classify Turnstile/auth failures.
- Modify: `backend/internal/upstream/sub2api_account_adapter.go`
  - Exchange stored refresh token through `POST /api/v1/auth/refresh`.
  - Use returned JWT bearer for profile/quota calls.
  - Return rotated refresh-token updates to the handler.
- Modify: `backend/internal/upstream/account_adapter_test.go`
  - Replace old cookie/JWT credential assumptions with structured new-api/sub2api credential tests.
- Modify: `backend/internal/http/admin/upstream_handler.go`
  - Persist adapter-returned credential rotations by encrypting plaintext and updating the account.
  - Include `account_status` snapshots in action results.
  - Preserve stored credentials when editing and plaintext credential is blank.
- Modify: `backend/internal/http/admin/upstream_handler_test.go`
  - Update fake adapter interface.
  - Add handler tests for action snapshot responses and rotated credential persistence.
- Modify: `src/types.ts`
  - Add new credential kind literals.
  - Add `accountStatus?: UpstreamAccountStatusSnapshot` to `UpstreamActionResult`.
- Modify: `src/lib/adminApi.ts`
  - Map raw `account_status` in single and batch action responses.
- Modify: `src/pages/sites/siteOptions.ts`
  - Add `New API Access Token` and `Sub2API Refresh Token` credential options.
- Modify: `src/pages/sites/SiteDrawer.tsx`
  - Split structured credential entry into token/user-id fields for new-api and refresh-token field for sub2api.
  - Serialize structured fields to JSON before save.
  - Keep edit mode blank-field behavior so existing encrypted credentials are preserved.
- Modify: `src/pages/SitesPage.tsx`
  - Patch row status snapshots from single and batch action responses.
  - Show concise batch summaries, especially for check-in.
  - Validate structured credential fields on create and when editing with newly entered values.

## Task 1: Backend Domain And Adapter Contract

**Files:**
- Modify: `backend/internal/domain/domain.go`
- Modify: `backend/internal/upstream/account_adapter.go`
- Test: `backend/internal/upstream/account_adapter_test.go`
- Test: `backend/internal/http/admin/upstream_handler_test.go`

- [ ] **Step 1: Write compile-facing test updates**

  Update fake adapter methods in `backend/internal/http/admin/upstream_handler_test.go` to the new contract:

  ```go
  func (fakeAccountAdapter) TestAccountCredential(context.Context, domain.UpstreamAccount, string) (domain.AccountCredentialTestResult, error) {
  	return domain.AccountCredentialTestResult{
  		Status: domain.UpstreamAccountStatus{AccountStatus: domain.AccountCredentialStatusValid},
  	}, nil
  }
  ```

  This is expected to fail compilation until the domain type and interface are added.

- [ ] **Step 2: Add domain types**

  In `backend/internal/domain/domain.go`, extend credential kinds:

  ```go
  const (
  	CredentialKindNone                 UpstreamCredentialKind = "none"
  	CredentialKindCookie               UpstreamCredentialKind = "cookie"
  	CredentialKindAccessToken          UpstreamCredentialKind = "access_token"
  	CredentialKindRefreshToken         UpstreamCredentialKind = "refresh_token"
  	CredentialKindJSON                 UpstreamCredentialKind = "json"
  	CredentialKindNewAPIAccessToken    UpstreamCredentialKind = "new_api_access_token"
  	CredentialKindSub2APIRefreshToken  UpstreamCredentialKind = "sub2api_refresh_token"
  )
  ```

  Add result/update structs near the existing action result structs:

  ```go
  type UpstreamCredentialUpdate struct {
  	Kind      UpstreamCredentialKind
  	Plaintext string
  }

  type AccountCredentialTestResult struct {
  	Status           UpstreamAccountStatus
  	CredentialUpdate *UpstreamCredentialUpdate
  }
  ```

  Extend quota/check-in results:

  ```go
  type QuotaRefreshResult struct {
  	Status           UpstreamAccountStatus
  	BalanceAmount    float64
  	BalanceUnit      string
  	CredentialUpdate *UpstreamCredentialUpdate
  }

  type CheckinResult struct {
  	Status               UpstreamCheckinStatus
  	Message              string
  	AccountStatus        AccountCredentialStatus
  	LastErrorClass       UpstreamErrorClass
  	LastErrorMessage     string
  	ActionRequiredReason string
  	CredentialUpdate     *UpstreamCredentialUpdate
  }
  ```

- [ ] **Step 3: Update adapter interface and helpers**

  Change `AccountAdapter.TestAccountCredential` in `backend/internal/upstream/account_adapter.go` to:

  ```go
  TestAccountCredential(ctx context.Context, account domain.UpstreamAccount, credential string) (domain.AccountCredentialTestResult, error)
  ```

  Add helper structs and functions:

  ```go
  type newAPIAccessCredential struct {
  	AccessToken string `json:"access_token"`
  	UserID      string `json:"user_id"`
  }

  type sub2APIRefreshCredential struct {
  	RefreshToken string `json:"refresh_token"`
  }

  func parseNewAPIAccessCredential(credential string) (newAPIAccessCredential, error)
  func parseSub2APIRefreshCredential(credential string) (sub2APIRefreshCredential, error)
  func encodeSub2APIRefreshCredential(refreshToken string) string
  func responseSucceeded(body []byte) bool
  func responseExplicitlyFailed(body []byte) bool
  func newAPISelfQuotaAmount(body []byte) float64
  func authMessageIndicatesCredentialProblem(message string) bool
  ```

  Parsing rules:

  - `new_api_access_token` JSON requires non-empty `access_token` and `user_id`.
  - `user_id` accepts JSON string or number by unmarshalling into `map[string]any` and normalizing with `fmt.Sprint` for numbers without decimals.
  - `sub2api_refresh_token` JSON requires `refresh_token` beginning with `rt_`; for backwards compatibility, a raw `rt_...` plaintext credential is also accepted.

## Task 2: new-api Structured Credential Behavior

**Files:**
- Modify: `backend/internal/upstream/account_adapter.go`
- Modify: `backend/internal/upstream/newapi_account_adapter.go`
- Test: `backend/internal/upstream/account_adapter_test.go`

- [ ] **Step 1: Add failing adapter tests**

  Add tests that demonstrate the required behavior:

  ```go
  func TestNewAPIAdapterAccountCredentialUsesAccessTokenUserHeaders(t *testing.T)
  func TestNewAPIAdapterAccountCredentialSuccessFalseExpiresCredential(t *testing.T)
  func TestNewAPIAdapterRefreshQuotaUsesUserSelfQuotaMinusUsedQuota(t *testing.T)
  func TestNewAPIAdapterCheckinClassifiesTurnstileAndAuthBodies(t *testing.T)
  ```

  Core assertions:

  - `GET /api/user/self` receives `Authorization: token-123` and `New-Api-User: 42`.
  - HTTP 200 with `{"success":false,"message":"access token invalid"}` returns `AccountCredentialStatusExpired`.
  - Quota refresh maps `{"success":true,"data":{"quota":100,"used_quota":25}}` to `BalanceAmount == 75` and `BalanceUnit == "quota"`.
  - Check-in returns `CheckinStatusActionRequired` for Turnstile bodies and sets `AccountStatusExpired` for auth bodies.

- [ ] **Step 2: Implement new-api request helper**

  Add a helper that builds new-api account requests:

  ```go
  func (a accountHTTPAdapter) doNewAPIUser(ctx context.Context, method string, account domain.UpstreamAccount, path string, credential newAPIAccessCredential, body io.Reader) (adapterResponse, error)
  ```

  It sets:

  ```text
  Accept: application/json
  Authorization: <access_token>
  New-Api-User: <user_id>
  Content-Type: application/json when body is present
  ```

- [ ] **Step 3: Implement `TestAccountCredential`**

  In `NewAPIAccountAdapter.TestAccountCredential`:

  - If credential is blank, return `AccountCredentialTestResult{Status: missingCredentialStatus(account.ID)}`.
  - If `account.AccountCredentialKind == CredentialKindNewAPIAccessToken`, parse JSON and call `/api/user/self` via `doNewAPIUser`.
  - Keep old `doCookie` behavior for existing `cookie`, `access_token`, `refresh_token`, and `json` credentials.
  - For HTTP 401/403 or 200 `success:false` auth messages, set `AccountStatusExpired` and `LastErrorClass=credential_expired`.
  - For Turnstile/captcha messages, set `AccountStatusActionRequired` and `LastErrorClass=action_required`.
  - For success, set `AccountStatusValid`, `CheckinStatusUnchecked`, `LastAccountCheckedAt`, and latency.

- [ ] **Step 4: Implement quota refresh**

  In `NewAPIAccountAdapter.RefreshQuota`:

  - If structured new-api credential is present, call `GET /api/user/self`.
  - Map `quota - used_quota` from either root payload or `data`.
  - If only `quota` exists, use it as remaining balance.
  - Preserve old API-key `/api/usage/token` fallback when no structured account credential exists.

- [ ] **Step 5: Implement check-in**

  In `NewAPIAccountAdapter.Checkin`:

  - For structured credentials, call `POST /api/user/checkin` via `doNewAPIUser`.
  - Keep old `doCookie` behavior for generic credentials.
  - HTTP 200 `success:true` returns `CheckinStatusChecked`.
  - HTTP 200 `success:false` with Turnstile/captcha returns `CheckinStatusActionRequired`.
  - HTTP 200 `success:false` auth body or HTTP 401/403 returns `CheckinStatusActionRequired`, `AccountStatusExpired`, `LastErrorClass=credential_expired`.

## Task 3: sub2api Refresh Token Flow

**Files:**
- Modify: `backend/internal/upstream/sub2api_account_adapter.go`
- Modify: `backend/internal/upstream/account_adapter.go`
- Test: `backend/internal/upstream/account_adapter_test.go`

- [ ] **Step 1: Add failing adapter tests**

  Add tests:

  ```go
  func TestSub2APIAdapterRefreshesTokenForProfileAndReturnsRotation(t *testing.T)
  func TestSub2APIAdapterRefreshQuotaUsesPlatformQuotasWithBearer(t *testing.T)
  func TestSub2APIAdapterInvalidRefreshTokenExpiresCredential(t *testing.T)
  ```

  Core assertions:

  - Adapter posts `{"refresh_token":"rt_old"}` to `/api/v1/auth/refresh`.
  - Profile/quota calls use `Authorization: Bearer jwt-new`.
  - A returned `refresh_token:"rt_new"` becomes `CredentialUpdate{Kind: CredentialKindSub2APIRefreshToken, Plaintext: "{\"refresh_token\":\"rt_new\"}"}`.
  - 401/403 refresh failure returns expired account status.

- [ ] **Step 2: Implement refresh exchange**

  In `Sub2APIAccountAdapter`, add:

  ```go
  type sub2APIRefreshResult struct {
  	AccessToken      string
  	RefreshToken     string
  	CredentialUpdate *domain.UpstreamCredentialUpdate
  }

  func (a *Sub2APIAccountAdapter) refreshAccessToken(ctx context.Context, account domain.UpstreamAccount, credential string) (sub2APIRefreshResult, domain.UpstreamAccountStatus, error)
  ```

  The function calls `POST /api/v1/auth/refresh`, supports response shapes with token fields at root or under `data`, and classifies invalid refresh tokens as `AccountCredentialStatusExpired`.

- [ ] **Step 3: Use refresh result for profile/quota**

  Update:

  - `TestAccountCredential` to refresh first, then call `GET /api/v1/user/profile` with bearer JWT.
  - `RefreshQuota` to refresh first, then call `GET /api/v1/user/platform-quotas`; parse remaining quota with existing `quotaAmount`.
  - `Checkin` remains unsupported but can return a credential-expired result if the credential is missing.

## Task 4: Handler Persistence And Action Snapshots

**Files:**
- Modify: `backend/internal/http/admin/upstream_handler.go`
- Modify: `backend/internal/http/admin/upstream_handler_test.go`

- [ ] **Step 1: Add failing handler tests**

  Add:

  ```go
  func TestUpstreamActionResponseIncludesAccountStatusSnapshot(t *testing.T)
  func TestUpstreamActionPersistsRotatedAccountCredential(t *testing.T)
  ```

  Assertions:

  - Single action response includes `account_status.account_status`.
  - Batch response includes per-result `account_status`.
  - When adapter returns `CredentialUpdate`, handler encrypts the new plaintext and `UpdateUpstreamAccount` stores it.
  - Response and events do not contain the plaintext rotated token.

- [ ] **Step 2: Extend `actionResult`**

  In `upstream_handler.go`:

  ```go
  type actionResult struct {
  	ID            string                        `json:"id"`
  	Status        string                        `json:"status"`
  	Message       string                        `json:"message,omitempty"`
  	AccountStatus *domain.UpstreamAccountStatus `json:"account_status,omitempty"`
  }
  ```

- [ ] **Step 3: Persist credential update in handler**

  Add:

  ```go
  func (h *Handler) applyCredentialUpdate(upstreams store.UpstreamAccountStore, account domain.UpstreamAccount, update *domain.UpstreamCredentialUpdate) (domain.UpstreamAccount, error)
  ```

  Behavior:

  - If update is nil, return original account.
  - Validate `Kind != ""` and plaintext non-empty.
  - Encrypt plaintext through `h.encryptSecret`.
  - Set `AccountCredentialKind` and `AccountCredentialEncrypted`.
  - Call `upstreams.UpdateUpstreamAccount`.

- [ ] **Step 4: Wire result handling**

  In `runUpstreamAction`:

  - For `test-account`, `refresh-quota`, and `checkin`, call `applyCredentialUpdate` when results include updates.
  - For `checkin`, build a status snapshot with `CheckinStatus`, optional `AccountStatus`, error class/message, action-required reason, and `LastCheckinAt` on checked results.
  - Return snapshots from `storeActionStatus`.

## Task 5: Frontend Structured Credentials And Row Patching

**Files:**
- Modify: `src/types.ts`
- Modify: `src/lib/adminApi.ts`
- Modify: `src/pages/sites/siteOptions.ts`
- Modify: `src/pages/sites/SiteDrawer.tsx`
- Modify: `src/pages/SitesPage.tsx`

- [ ] **Step 1: Update TypeScript types**

  In `src/types.ts`:

  ```ts
  export type UpstreamCredentialKind =
    | 'none'
    | 'cookie'
    | 'access_token'
    | 'refresh_token'
    | 'json'
    | 'new_api_access_token'
    | 'sub2api_refresh_token';

  export interface UpstreamActionResult {
    id: string;
    status: 'success' | 'failed' | 'not_found';
    message?: string;
    accountStatus?: UpstreamAccountStatusSnapshot;
  }
  ```

- [ ] **Step 2: Map action status snapshots**

  In `src/lib/adminApi.ts`, add `RawUpstreamActionResult`, map `account_status` with `mapUpstreamStatus`, and use it in `runUpstreamAction` and `runBatchUpstreamAction`.

- [ ] **Step 3: Add credential options**

  In `src/pages/sites/siteOptions.ts`, add:

  ```ts
  { label: 'New API Access Token', value: 'new_api_access_token' },
  { label: 'Sub2API Refresh Token', value: 'sub2api_refresh_token' },
  ```

- [ ] **Step 4: Split structured fields in drawer**

  In `SiteDrawer.tsx`:

  - Add local state for `newAPIAccessToken`, `newAPIUserID`, and `sub2APIRefreshToken`.
  - Reset those fields to blank when opening an existing account so secrets are not displayed.
  - When saving/testing, call `buildCredentialPayload(form)`:

    ```ts
    function buildCredentialPayload(input: UpstreamAccountInput): UpstreamAccountInput {
      if (input.accountCredentialKind === 'new_api_access_token') {
        return {
          ...input,
          accountCredential:
            input.accountCredential?.trim() ||
            JSON.stringify({ access_token: newAPIAccessToken.trim(), user_id: newAPIUserID.trim() }),
        };
      }
      if (input.accountCredentialKind === 'sub2api_refresh_token') {
        return {
          ...input,
          accountCredential:
            input.accountCredential?.trim() ||
            JSON.stringify({ refresh_token: sub2APIRefreshToken.trim() }),
        };
      }
      return input;
    }
    ```

  - Render two inputs for `new_api_access_token`, one input for `sub2api_refresh_token`, and the existing textarea for generic kinds.

- [ ] **Step 5: Patch rows from action responses**

  In `SitesPage.tsx`, add:

  ```ts
  function applyActionResults(results: UpstreamActionResult[]) {
    setAccounts((current) =>
      current.map((account) => {
        const result = results.find((item) => item.id === account.id);
        return result?.accountStatus ? { ...account, status: result.accountStatus } : account;
      })
    );
  }
  ```

  Use it in `runAction`, `runBatch`, and edit-drawer API test before deciding whether a full reload is required. For create/update/delete keep full reload.

- [ ] **Step 6: Add batch summary**

  Add `batchNotice` state and render it near the existing error banner. After batch actions, compute counts from returned results:

  ```ts
  const success = results.filter((item) => item.status === 'success').length;
  const failed = results.filter((item) => item.status !== 'success').length;
  setBatchNotice(`批量操作完成：成功 ${success}，失败 ${failed}`);
  ```

## Task 6: Final Verification And Commit

**Files:**
- All modified files.

- [ ] **Step 1: Format**

  Run:

  ```bash
  gofmt -w backend/internal/domain/domain.go backend/internal/upstream/account_adapter.go backend/internal/upstream/newapi_account_adapter.go backend/internal/upstream/sub2api_account_adapter.go backend/internal/upstream/account_adapter_test.go backend/internal/http/admin/upstream_handler.go backend/internal/http/admin/upstream_handler_test.go
  ```

- [ ] **Step 2: Run backend tests**

  Run once after implementation:

  ```bash
  cd backend && GOCACHE=/tmp/go-build go test ./...
  ```

  Expected: all packages pass.

- [ ] **Step 3: Run frontend build**

  Run:

  ```bash
  npm run build
  ```

  Expected: TypeScript and Vite build pass.

- [ ] **Step 4: Check whitespace**

  Run:

  ```bash
  git diff --check
  ```

  Expected: no whitespace errors.

- [ ] **Step 5: Review diff**

  Run:

  ```bash
  git status --short
  git diff --stat
  ```

  Confirm the diff includes the previous Claude issue fixes plus this LinuxDO implementation, with no unrelated destructive changes.

- [ ] **Step 6: Commit**

  Run:

  ```bash
  git add .
  git commit -m "feat: support linuxdo upstream account credentials"
  ```

  Report the commit hash and verification commands in the final response.
