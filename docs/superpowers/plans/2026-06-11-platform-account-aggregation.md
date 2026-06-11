# Platform Account Aggregation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add RelayDeck upstream account aggregation, model synchronization, native-protocol-first routing, and rolling health tracking for `new-api`, `sub2api`, and direct provider accounts.

**Architecture:** Store durable upstream account, model, binding, and health history data in PostgreSQL; keep rolling health and circuit state in Redis; expose admin APIs for account/model sync and health inspection; and update the gateway router so protocol choice is derived from model capability metadata rather than from relay-platform convenience conversion.

**Tech Stack:** Go `net/http`, PostgreSQL, Redis, SQL migrations, existing RelayDeck domain/store/router packages, mock HTTP tests with `httptest`, Vite/React admin console only after backend APIs stabilize.

---

### Task 1: Extend the Domain and Migration for Upstream Accounts

**Files:**
- Modify: `backend/internal/domain/domain.go`
- Modify: `backend/migrations/0001_initial.sql`
- Create: `backend/internal/domain/domain_test.go`

- [ ] **Step 1: Write the failing domain tests**

```go
func TestUpstreamAccountCapturesProtocolPreferences(t *testing.T)
func TestRouteCandidateCanRepresentNativeProtocolChoice(t *testing.T)
func TestHealthObservationCarriesNormalizedErrorClass(t *testing.T)
```

- [ ] **Step 2: Run the domain tests and confirm they fail**

Run: `cd backend && go test ./internal/domain -v`

Expected: FAIL because the new upstream account and health types do not exist yet.

- [ ] **Step 3: Add upstream account, model binding, and health types**

```go
type Protocol string

const (
    ProtocolOpenAIResponses   Protocol = "openai_responses"
    ProtocolOpenAIChat        Protocol = "openai_chat_completions"
    ProtocolAnthropicMessages Protocol = "anthropic_messages"
    ProtocolGeminiGenerate    Protocol = "gemini_generate_content"
)

type UpstreamPlatformKind string

type AuthKind string

type UpstreamAccount struct {
    ID                   string
    Name                 string
    PlatformKind         UpstreamPlatformKind
    BaseURL              string
    Enabled              bool
    Priority             int
    AuthKind             AuthKind
    AuthSecretEncrypted  string
    DefaultWireProtocol  Protocol
    SupportedWireProtocols []Protocol
    Metadata             map[string]any
}

type UpstreamModel struct {
    ID                   string
    UpstreamAccountID    string
    NormalizedModelName   string
    UpstreamModelName     string
    DisplayName           string
    NativeWireProtocol    Protocol
    SupportedWireProtocols []Protocol
    Capabilities          []Capability
    Status                string
    SyncSource            string
    RawMetadata           map[string]any
}

type ModelBinding struct {
    ID                  string
    NormalizedModelName string
    UpstreamAccountID   string
    UpstreamModelName   string
    PreferredWireProtocol Protocol
    CapabilityOverride  []Capability
    Weight              int
    Enabled             bool
}

type HealthObservation struct {
    ID                   string
    UpstreamAccountID    string
    NormalizedModelName  string
    UpstreamModelName    string
    WireProtocol         Protocol
    RequestKind          string
    Success              bool
    StatusCode           int
    ErrorClass           string
    LatencyMS            int
    RequestBytes         int
    ResponseBytes        int
    RequestID            string
}
```

- [ ] **Step 4: Add the schema tables and indexes**

Add tables and constraints to `backend/migrations/0001_initial.sql` for:

```sql
upstream_accounts
upstream_models
model_bindings
health_observations
```

Keep the migration aligned with the durable data model from the design. Do not add unrelated billing or user tables here.

- [ ] **Step 5: Run the domain package tests again**

Run: `cd backend && go test ./internal/domain -v`

Expected: PASS after the new types and helpers compile.

- [ ] **Step 6: Commit the domain and migration changes**

```bash
git add backend/internal/domain/domain.go backend/internal/domain/domain_test.go backend/migrations/0001_initial.sql
git commit -m "Add upstream account domain model"
```

### Task 2: Add Repositories for Accounts, Models, Bindings, and Health

**Files:**
- Create: `backend/internal/store/postgres/upstream_accounts.go`
- Create: `backend/internal/store/postgres/upstream_models.go`
- Create: `backend/internal/store/postgres/model_bindings.go`
- Create: `backend/internal/store/postgres/health_observations.go`
- Create: `backend/internal/store/postgres/upstream_accounts_test.go`
- Create: `backend/internal/store/postgres/upstream_models_test.go`

- [ ] **Step 1: Write repository tests against Docker-backed PostgreSQL**

```go
func TestUpstreamAccountCRUD(t *testing.T)
func TestModelSyncUpsertsAndPreservesManualBindings(t *testing.T)
func TestHealthObservationInsertAndList(t *testing.T)
```

- [ ] **Step 2: Run the new repository tests and confirm they fail**

Run: `cd backend && GOTOOLCHAIN=local GOCACHE=/tmp/go-build go test -count=1 ./internal/store/postgres -v`

Expected: FAIL because the new repository files do not exist yet.

- [ ] **Step 3: Implement the PostgreSQL repositories**

```go
type UpstreamAccountStore interface {
    List(ctx context.Context) ([]domain.UpstreamAccount, error)
    Get(ctx context.Context, id string) (domain.UpstreamAccount, bool, error)
    Upsert(ctx context.Context, account domain.UpstreamAccount) error
    Delete(ctx context.Context, id string) error
}
```

Add equivalent repository interfaces for upstream models, model bindings, and health observations. Keep the repository API small and batch-friendly so sync can upsert many rows without making a round trip per model.

- [ ] **Step 4: Run the repository tests and confirm they pass**

Run: `cd backend && GOTOOLCHAIN=local GOCACHE=/tmp/go-build go test -count=1 ./internal/store/postgres -v`

Expected: PASS.

- [ ] **Step 5: Commit the repository layer**

```bash
git add backend/internal/store/postgres/upstream_accounts.go backend/internal/store/postgres/upstream_models.go backend/internal/store/postgres/model_bindings.go backend/internal/store/postgres/health_observations.go backend/internal/store/postgres/upstream_accounts_test.go backend/internal/store/postgres/upstream_models_test.go
git commit -m "Persist upstream account metadata"
```

### Task 3: Build Upstream Sync Adapters

**Files:**
- Create: `backend/internal/upstream/newapi_adapter.go`
- Create: `backend/internal/upstream/sub2api_adapter.go`
- Create: `backend/internal/upstream/sync.go`
- Create: `backend/internal/upstream/sync_test.go`
- Modify: `backend/internal/upstream/client.go`

- [ ] **Step 1: Write sync adapter tests with mocked HTTP servers**

```go
func TestSyncNewAPIModelsMapsNativeProtocols(t *testing.T)
func TestSyncSub2APIAccountsMapsChannelModes(t *testing.T)
func TestSyncPreservesRawMetadataAndManualBindings(t *testing.T)
```

- [ ] **Step 2: Run the sync tests and confirm they fail**

Run: `cd backend && go test ./internal/upstream -v`

Expected: FAIL until the adapters and sync orchestration exist.

- [ ] **Step 3: Implement discovery and normalization**

```go
type SyncResult struct {
    AccountID     string
    CreatedModels int
    UpdatedModels int
    SkippedModels []string
}

type AccountSyncAdapter interface {
    DiscoverModels(ctx context.Context, account domain.UpstreamAccount) ([]domain.UpstreamModel, error)
}
```

`new-api` should discover model metadata and protocol compatibility from its model/admin endpoints. `sub2api` should discover channel/account data, model mappings, and the equivalent `api_mode` or protocol choice fields where available.

- [ ] **Step 4: Run the sync tests and confirm they pass**

Run: `cd backend && go test ./internal/upstream -v`

Expected: PASS.

- [ ] **Step 5: Commit the sync layer**

```bash
git add backend/internal/upstream/newapi_adapter.go backend/internal/upstream/sub2api_adapter.go backend/internal/upstream/sync.go backend/internal/upstream/sync_test.go backend/internal/upstream/client.go
git commit -m "Add upstream sync adapters"
```

### Task 4: Make Routing Prefer the Native Wire Protocol

**Files:**
- Modify: `backend/internal/domain/domain.go`
- Modify: `backend/internal/router/router.go`
- Modify: `backend/internal/router/router_test.go`
- Modify: `backend/internal/http/gateway/handler.go`
- Modify: `backend/internal/store/memory.go`

- [ ] **Step 1: Write routing tests that prove protocol preference**

```go
func TestSelectCandidatePrefersNativeProtocolOverConversion(t *testing.T)
func TestSelectCandidateRejectsUnsupportedProtocolConversion(t *testing.T)
func TestSelectCandidateKeepsExistingHealthFilters(t *testing.T)
```

- [ ] **Step 2: Run router tests and confirm they fail**

Run: `cd backend && go test ./internal/router -v`

Expected: FAIL until the candidate model and protocol selection logic is updated.

- [ ] **Step 3: Extend the router to score protocol compatibility**

```go
type RouteCandidate struct {
    Site              UpstreamSite
    Mapping           SiteModel
    PreferredProtocol Protocol
    ConversionNeeded  bool
    Score             float64
}
```

Route selection should:

- prefer a candidate whose `PreferredProtocol` matches the model's native protocol;
- allow a fallback conversion only if the conversion path is explicitly marked compatible;
- reject the request if every remaining candidate requires a lossy conversion.

- [ ] **Step 4: Update the gateway handler to pass protocol requirements into routing**

The gateway should derive required protocol and capability metadata from the request shape before it chooses a candidate, rather than selecting a site first and figuring out the protocol afterward.

- [ ] **Step 5: Run router tests and confirm they pass**

Run: `cd backend && go test ./internal/router -v`

Expected: PASS.

- [ ] **Step 6: Commit the routing changes**

```bash
git add backend/internal/domain/domain.go backend/internal/router/router.go backend/internal/router/router_test.go backend/internal/http/gateway/handler.go backend/internal/store/memory.go
git commit -m "Prefer native upstream protocols"
```

### Task 5: Add Rolling Health Scoring and Redis Cache State

**Files:**
- Create: `backend/internal/health/score.go`
- Create: `backend/internal/health/score_test.go`
- Create: `backend/internal/health/cache.go`
- Create: `backend/internal/health/cache_test.go`
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/upstream/client.go`

- [ ] **Step 1: Write health scoring tests**

```go
func TestHealthScorePenalizesAuthAndProtocolFailures(t *testing.T)
func TestHealthScoreRecoversAfterSuccessfulObservations(t *testing.T)
func TestHealthScoreTracksLatencyAndFailureStreak(t *testing.T)
```

- [ ] **Step 2: Run the health tests and confirm they fail**

Run: `cd backend && go test ./internal/health -v`

Expected: FAIL because the scoring and cache packages do not exist yet.

- [ ] **Step 3: Implement the rolling health model**

```go
type Observation struct {
    Success      bool
    ErrorClass   string
    StatusCode   int
    LatencyMS    int
    ObservedAt   time.Time
}

type Summary struct {
    HealthScore      float64
    SuccessRate      float64
    LatencyMS        int
    CircuitState     domain.CircuitState
    FailureStreak    int
    RecoveryStreak   int
}
```

Use recent success/failure, latency, and error class to update the score. Hard auth and protocol failures should drop the score quickly; successful probes should recover it gradually.

- [ ] **Step 4: Add Redis-backed cache state**

Store the live health summary and circuit state in Redis so routing can read it without hitting PostgreSQL on every request.

- [ ] **Step 5: Run the health tests and confirm they pass**

Run: `cd backend && go test ./internal/health -v`

Expected: PASS.

- [ ] **Step 6: Commit health scoring and cache state**

```bash
git add backend/internal/health/score.go backend/internal/health/score_test.go backend/internal/health/cache.go backend/internal/health/cache_test.go backend/internal/app/app.go backend/internal/upstream/client.go
git commit -m "Add upstream health scoring"
```

### Task 6: Expose Admin APIs for Accounts, Sync, and Health

**Files:**
- Create: `backend/internal/http/admin/upstream_handler.go`
- Create: `backend/internal/http/admin/upstream_handler_test.go`
- Modify: `backend/internal/http/admin/handler.go`
- Modify: `backend/internal/app/app.go`

- [ ] **Step 1: Write admin API tests**

```go
func TestAdminListUpstreamAccounts(t *testing.T)
func TestAdminSyncModelsReturnsSummary(t *testing.T)
func TestAdminHealthSummaryReturnsCachedState(t *testing.T)
```

- [ ] **Step 2: Run the admin tests and confirm they fail**

Run: `cd backend && go test ./internal/http/admin -v`

Expected: FAIL before the new admin upstream handler exists.

- [ ] **Step 3: Implement the upstream admin endpoints**

```text
GET  /api/admin/upstreams/accounts
POST /api/admin/upstreams/accounts
PUT  /api/admin/upstreams/accounts/{id}
DELETE /api/admin/upstreams/accounts/{id}
POST /api/admin/upstreams/accounts/{id}/sync-models
GET  /api/admin/upstreams/accounts/{id}/models
POST /api/admin/upstreams/accounts/{id}/probe
GET  /api/admin/health/summary
GET  /api/admin/health/events
```

These handlers should use the new repositories and return sanitized metadata only.

- [ ] **Step 4: Wire the endpoints into the admin router**

Keep the existing session middleware in front of all `/api/admin/*` routes.

- [ ] **Step 5: Run the admin tests and confirm they pass**

Run: `cd backend && go test ./internal/http/admin -v`

Expected: PASS.

- [ ] **Step 6: Commit the admin API surface**

```bash
git add backend/internal/http/admin/upstream_handler.go backend/internal/http/admin/upstream_handler_test.go backend/internal/http/admin/handler.go backend/internal/app/app.go
git commit -m "Add upstream admin APIs"
```

### Task 7: Add Client Profile Metadata For Codex and Claude Code

**Files:**
- Create: `backend/internal/clientprofile/profile.go`
- Create: `backend/internal/clientprofile/profile_test.go`
- Modify: `backend/internal/http/gateway/handler.go`
- Modify: `backend/internal/upstream/client.go`

- [ ] **Step 1: Write profile tests from reference behavior**

```go
func TestCodexProfileDefaultsToResponsesStyleRequests(t *testing.T)
func TestClaudeCodeProfileAppliesAnthropicHeaders(t *testing.T)
func TestProfileRegistryCanResolveByClientName(t *testing.T)
```

- [ ] **Step 2: Run the profile tests and confirm they fail**

Run: `cd backend && go test ./internal/clientprofile -v`

Expected: FAIL because the profile package does not exist yet.

- [ ] **Step 3: Implement the profile registry**

```go
type Profile struct {
    Name              string
    DefaultEndpoint   string
    DefaultProtocol   domain.Protocol
    HeaderTemplate    map[string]string
    StreamingStyle    string
    RetryPolicy       string
}
```

Seed `codex` and `claude_code` profiles. Keep them as metadata objects that inform request shaping and header selection; do not bake them into routing logic directly.

- [ ] **Step 4: Run the profile tests and confirm they pass**

Run: `cd backend && go test ./internal/clientprofile -v`

Expected: PASS.

- [ ] **Step 5: Commit the client profile layer**

```bash
git add backend/internal/clientprofile/profile.go backend/internal/clientprofile/profile_test.go backend/internal/http/gateway/handler.go backend/internal/upstream/client.go
git commit -m "Add client profile metadata"
```

### Task 8: Full Verification and Documentation Sync

**Files:**
- All files changed in this plan
- Modify: `README.md` if the feature status needs to be updated
- Modify: `backend/README.md` if run instructions change

- [ ] **Step 1: Run the backend test suite**

Run: `cd backend && GOTOOLCHAIN=local GOCACHE=/tmp/go-build go test ./...`

Expected: PASS.

- [ ] **Step 2: Run the frontend build**

Run: `npm run build`

Expected: PASS.

- [ ] **Step 3: Verify Docker-backed PostgreSQL and Redis behavior**

Run: `cd backend && GOTOOLCHAIN=local GOCACHE=/tmp/go-build go test -count=1 ./internal/store/postgres ./internal/auth -v`

Expected: PASS using the Docker containers already wired through `.env` and `docker-compose.yml`.

- [ ] **Step 4: Review git state**

Run: `git status --short`

Expected: only intended backend files, docs, and this plan are changed.

- [ ] **Step 5: Commit and push**

```bash
git add .
git commit -m "Add upstream account aggregation foundation"
git push
```

## Self-Review

- Spec coverage: this plan covers durable upstream account storage, model sync, native-protocol-first routing, rolling health, Redis cache state, admin APIs, and client profile metadata. UI wiring is intentionally deferred until the backend contract is stable.
- Placeholder scan: no `TBD`, `TODO`, or vague pseudo-steps remain.
- Type consistency: the new domain types, repository names, and route paths are scoped to the feature and do not conflict with the existing admin auth or gateway foundation code.

