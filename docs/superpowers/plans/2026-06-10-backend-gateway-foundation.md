# Backend Gateway Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first runnable Go backend slice for RelayDeck: service bootstrap, in-memory gateway configuration, API key auth, model listing, candidate routing, and non-streaming OpenAI-compatible proxying through a fake-upstream-tested path.

**Architecture:** Keep the first slice dependency-light and use Go `net/http` with the Go 1.24 method-aware `ServeMux`; the package boundaries still match the backend spec so `chi`, PostgreSQL, Redis, and persistent repositories can be added without rewriting gateway logic. Use in-memory stores for bootstrap data and fake upstream tests, while preserving domain interfaces that later map to PostgreSQL.

**Tech Stack:** Go 1.24, standard-library HTTP server, standard-library tests, React frontend unchanged.

---

## File Structure

- Create `backend/go.mod`: Go module declaration.
- Create `backend/cmd/relaydeck/main.go`: process entrypoint.
- Create `backend/internal/config/config.go`: environment config loader.
- Create `backend/internal/domain/domain.go`: shared domain types for API keys, models, upstreams, routing, and gateway requests.
- Create `backend/internal/auth/apikey.go`: API key hashing and verification.
- Create `backend/internal/auth/apikey_test.go`: API key verification tests.
- Create `backend/internal/router/router.go`: candidate filtering and smart scoring.
- Create `backend/internal/router/router_test.go`: routing tests.
- Create `backend/internal/upstream/client.go`: upstream proxy client for JSON requests.
- Create `backend/internal/upstream/client_test.go`: fake upstream proxy tests.
- Create `backend/internal/store/memory.go`: in-memory bootstrap store.
- Create `backend/internal/http/gateway/handler.go`: `/v1/models`, `/v1/chat/completions`, and `/v1/responses` handlers.
- Create `backend/internal/http/gateway/handler_test.go`: gateway HTTP tests.
- Create `backend/internal/http/admin/handler.go`: minimal admin health/config summary endpoints for UI readiness.
- Create `backend/internal/app/app.go`: HTTP router wiring.
- Create `backend/README.md`: backend run/test notes and first-slice limitations.
- Create `backend/migrations/0001_initial.sql`: schema draft matching the approved spec, not executed by first slice.

## Task 1: Bootstrap Go Backend

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/relaydeck/main.go`
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/app/app.go`
- Create: `backend/README.md`

- [ ] **Step 1: Write bootstrap files**

Create a Go module under `backend` and a `main.go` that loads config, builds the app handler, and starts `http.Server`.

- [ ] **Step 2: Add config loader**

Implement `config.Load()` reading `HTTP_ADDR`, `APP_SECRET`, `GATEWAY_REQUEST_TIMEOUT`, and defaulting to `:8080`, `dev-secret`, and `30s`.

- [ ] **Step 3: Add app router skeleton**

Implement `app.New()` returning an `http.Handler` with `GET /healthz` and placeholder routing groups.

- [ ] **Step 4: Verify bootstrap**

Run: `cd backend && go test ./...`

Expected: exit code 0.

## Task 2: Domain Types and API Key Auth

**Files:**
- Create: `backend/internal/domain/domain.go`
- Create: `backend/internal/auth/apikey.go`
- Create: `backend/internal/auth/apikey_test.go`

- [ ] **Step 1: Write failing API key tests**

Tests must cover valid key verification, invalid secret rejection, inactive key rejection, expired key rejection, model whitelist rejection, and missing scope rejection.

- [ ] **Step 2: Run auth tests red**

Run: `cd backend && go test ./internal/auth -run TestVerifyGatewayKey -v`

Expected: FAIL before implementation compiles or before verifier exists.

- [ ] **Step 3: Implement domain and auth**

Define `APIKey`, `Scope`, `Model`, `UpstreamSite`, `SiteModel`, `GatewayRequest`, and `GatewayPrincipal`. Implement SHA-256 key hashing with prefix extraction and `VerifyGatewayKey()`.

- [ ] **Step 4: Run auth tests green**

Run: `cd backend && go test ./internal/auth -run TestVerifyGatewayKey -v`

Expected: PASS.

## Task 3: Routing Candidate Selection

**Files:**
- Create: `backend/internal/router/router.go`
- Create: `backend/internal/router/router_test.go`

- [ ] **Step 1: Write routing tests**

Tests must verify candidates are filtered by enabled state, model mapping, endpoint capability, required capability, circuit state, and model whitelist. A second test must verify smart scoring chooses the higher health/success/lower-latency candidate.

- [ ] **Step 2: Run routing tests red**

Run: `cd backend && go test ./internal/router -v`

Expected: FAIL before router implementation exists.

- [ ] **Step 3: Implement router**

Implement `SelectCandidate(req, principal, sites, mappings, policy)` returning the best candidate and a rejection reason when none are available.

- [ ] **Step 4: Run routing tests green**

Run: `cd backend && go test ./internal/router -v`

Expected: PASS.

## Task 4: Upstream Client

**Files:**
- Create: `backend/internal/upstream/client.go`
- Create: `backend/internal/upstream/client_test.go`

- [ ] **Step 1: Write fake upstream tests**

Use `httptest.Server` to verify the proxy sends the mapped upstream model, uses the upstream bearer token, preserves JSON response status/body, and normalizes upstream 500 errors.

- [ ] **Step 2: Run upstream tests red**

Run: `cd backend && go test ./internal/upstream -v`

Expected: FAIL before client implementation exists.

- [ ] **Step 3: Implement upstream client**

Implement `Client.DoJSON(ctx, upstream, path, body)` using a configured `http.Client`, upstream `BaseURL`, `Credential`, and request timeout.

- [ ] **Step 4: Run upstream tests green**

Run: `cd backend && go test ./internal/upstream -v`

Expected: PASS.

## Task 5: Gateway Handlers with In-Memory Store

**Files:**
- Create: `backend/internal/store/memory.go`
- Create: `backend/internal/http/gateway/handler.go`
- Create: `backend/internal/http/gateway/handler_test.go`
- Modify: `backend/internal/app/app.go`

- [ ] **Step 1: Write gateway HTTP tests**

Tests must verify `GET /v1/models` returns canonical models, unauthorized requests receive 401, authorized `POST /v1/chat/completions` proxies to fake upstream, and `POST /v1/responses` returns a compatibility error when no upstream supports responses.

- [ ] **Step 2: Run gateway tests red**

Run: `cd backend && go test ./internal/http/gateway -v`

Expected: FAIL before handler implementation exists.

- [ ] **Step 3: Implement memory store**

Seed one active RelayDeck API key, two canonical models, one chat-capable upstream, and one site-model mapping. The seed key should be documented as `rd_live_dev_test_secret` for local testing only.

- [ ] **Step 4: Implement gateway handlers**

Implement OpenAI-shaped `GET /v1/models`, `POST /v1/chat/completions`, and `POST /v1/responses` using auth, routing, and upstream client packages.

- [ ] **Step 5: Run gateway tests green**

Run: `cd backend && go test ./internal/http/gateway -v`

Expected: PASS.

## Task 6: Minimal Admin and Schema Draft

**Files:**
- Create: `backend/internal/http/admin/handler.go`
- Create: `backend/migrations/0001_initial.sql`
- Modify: `backend/internal/app/app.go`

- [ ] **Step 1: Add admin summary endpoint**

Implement `GET /api/admin/summary` returning counts for sites, models, and API keys from the memory store. This gives the frontend a realistic API shape without pretending CRUD is complete.

- [ ] **Step 2: Add migration draft**

Create SQL tables for `users`, `api_keys`, `upstream_sites`, `upstream_credentials`, `models`, `site_models`, `routing_policies`, `request_logs`, `request_attempts`, `upstream_health_checks`, and `usage_daily`.

- [ ] **Step 3: Verify all backend tests**

Run: `cd backend && go test ./...`

Expected: PASS.

## Task 7: Final Verification and Commit

**Files:**
- All backend files
- Plan document

- [ ] **Step 1: Run backend tests**

Run: `cd backend && go test ./...`

Expected: PASS.

- [ ] **Step 2: Run frontend build**

Run: `npm run build`

Expected: PASS.

- [ ] **Step 3: Review git state**

Run: `git status --short`

Expected: only intended backend files and this plan are changed.

- [ ] **Step 4: Commit and push**

Run: `git add . && git commit -m "Add backend gateway foundation" && git push`

Expected: commit and push succeed.

## UI Contract Note

The existing frontend is a high-fidelity prototype. Some visible operations are only styled and do not represent valid backend flows yet. This backend slice should not blindly mirror those invalid interactions. Instead, it establishes realistic API operations: gateway API key auth, upstream model routing, model listing, and admin summary. Later frontend work should replace prototype-only flows with calls to these backend contracts.

## Self-Review

- Spec coverage: this plan covers the first implementation slice for gateway bootstrap, API key auth, model listing, routing, upstream proxying, and first admin contract. Streaming, PostgreSQL execution, Redis, health checks, and full CRUD are intentionally deferred to later slices.
- Placeholder scan: no unresolved placeholder markers are present.
- Type consistency: package names are `domain`, `auth`, `router`, `upstream`, `store`, `gateway`, `admin`, and `app`; route paths match the approved backend spec.
