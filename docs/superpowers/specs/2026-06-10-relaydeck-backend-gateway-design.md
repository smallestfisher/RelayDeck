# RelayDeck Backend Gateway Design

Date: 2026-06-10

## Context

RelayDeck already has a frontend UI prototype for managing upstream AI relay sites, models, routing, API keys, users, logs, and system settings. The backend milestone should now implement the core value of the product: aggregating many scattered AI relay sites based on `new-api`, `sub2api`, or other OpenAI-compatible services into one unified gateway.

The first backend milestone prioritizes the gateway core over a complete management backend. Management APIs should exist where needed to configure and operate the gateway, but the primary success criterion is that clients can call RelayDeck as a unified OpenAI-compatible endpoint.

## Goals

- Provide a unified OpenAI-compatible gateway for downstream clients and SDKs.
- Support API Key authentication issued by RelayDeck, not direct upstream credentials.
- Support routing across multiple upstream relay sites by model, capability, health, latency, and policy.
- Support both `Responses API` and legacy `Chat Completions` flows.
- Support non-streaming and streaming proxy behavior.
- Record request logs, upstream attempts, health status, and basic usage statistics.
- Provide enough admin APIs for the existing frontend to manage sites, models, API keys, users, routing policies, logs, and statistics.

## Non-Goals For MVP

- No microservice split.
- No complex billing or exact USD settlement.
- No full enterprise RBAC beyond basic owner/admin/member roles.
- No deep protocol conversion for every possible OpenAI endpoint.
- No transparent failover after a streaming response has already started writing to the client.
- No direct exposure of upstream credentials to clients or frontend responses.

## Technology Direction

Use a Go backend single service with PostgreSQL as the primary database.

Recommended stack:

- Language: Go
- HTTP routing: `chi` preferred for a lightweight gateway service; `gin` remains acceptable if implementation ergonomics require it.
- Database: PostgreSQL
- Cache/coordination: Redis optional for multi-instance rate limiting and shared circuit-break state.
- Migrations: SQL migrations checked into `backend/migrations`.
- Deployment: one Go binary plus PostgreSQL for MVP; Docker Compose for local development.

Rationale:

- Go is well-suited for high-concurrency HTTP proxying, streaming, timeout control, and low resource usage.
- A single service keeps operational complexity low while still allowing clean internal package boundaries.
- PostgreSQL is sufficient for gateway configuration, logs, health snapshots, and usage aggregation.
- Redis can be introduced only when multi-instance deployment or distributed rate limiting is required.

## High-Level Architecture

```text
Client / OpenAI SDK
  -> RelayDeck Gateway (/v1/*)
    -> API Key authentication and rate limiting
    -> request normalization into GatewayRequest
    -> model and capability routing
    -> upstream adapter
    -> new-api / sub2api / OpenAI-compatible upstreams
```

Internal modules:

- `gateway`: OpenAI-compatible public API under `/v1/*`.
- `admin_api`: RelayDeck management API under `/api/admin/*`.
- `auth`: admin login/session handling and gateway API key verification.
- `router`: model, capability, health, and policy based upstream selection.
- `upstream`: upstream HTTP client, request adaptation, SSE streaming, timeout handling, and upstream error normalization.
- `health`: periodic upstream checks, latency tracking, capability checks, and circuit-break state.
- `quota`: API key usage accounting, rate limits, and soft quota enforcement.
- `logs`: request logs and per-upstream attempt logs.
- `stats`: daily aggregation for frontend dashboards.
- `store`: PostgreSQL repositories.

## Gateway API Scope

The public gateway uses OpenAI-compatible paths.

MVP priority:

- `POST /v1/responses`
- `POST /v1/chat/completions`
- `GET /v1/models`

Prepared extensions:

- `GET /v1/responses/{response_id}`
- `POST /v1/embeddings`
- `POST /v1/images/generations`
- `POST /v1/audio/*`

The design is `Responses API first, Chat Completions compatible`.

Request normalization:

```text
/v1/responses or /v1/chat/completions
  -> parse into internal GatewayRequest
  -> authenticate and authorize
  -> resolve model and capabilities
  -> route to upstream
  -> adapt request to upstream endpoint
  -> return response in the original API shape
```

If an upstream supports `/v1/responses`, RelayDeck can proxy the Responses request directly. If an upstream only supports `/v1/chat/completions`, MVP should only convert simple text-only Responses requests where conversion is safe. Complex Responses requests involving tools, multimodal input, or unsupported input structures should fail with a clear compatibility error rather than silently producing incorrect behavior.

## Gateway Request Flow

For `POST /v1/responses` or `POST /v1/chat/completions`:

1. Parse `Authorization: Bearer rd_*`.
2. Verify API key hash, status, expiration, owner status, scope, IP whitelist, model whitelist, rate limit, and quota.
3. Parse requested model and endpoint type.
4. Derive required capabilities such as `responses`, `chat`, `streaming`, `tools`, `vision`, or `embedding`.
5. Build candidate upstream list from model mapping and policy.
6. Filter disabled, unhealthy, incompatible, circuit-open, or unauthorized candidates.
7. Score candidates and choose an upstream.
8. Forward request using the upstream adapter with mapped model name and upstream credential.
9. For non-streaming requests, read the full upstream response, record logs, update health/usage, and return the response.
10. For streaming requests, stream SSE chunks from upstream to client and record final status when the stream ends.
11. On eligible pre-write failures, retry the next candidate according to policy.

Streaming failover rule:

- Failover is allowed only before any bytes are written to the client.
- After SSE output starts, RelayDeck must not transparently switch upstreams. It should terminate the stream, record the failure, and expose an OpenAI-compatible error when possible.

## Routing Design

Routing input includes more than a model name:

```text
model + endpoint_type + required_capabilities + stream + user/key policy
```

Candidate filtering order:

1. Upstream site is enabled and has a usable credential.
2. Site has a model mapping for the requested RelayDeck model.
3. Site supports the requested endpoint type, such as `responses` or `chat_completions`.
4. Site supports required capabilities, such as `streaming`, `tools`, or `vision`.
5. Site is not circuit-open and health score meets the policy threshold.
6. User and API key policies allow this model/site.

Initial routing modes:

- `priority`: fixed priority/weight order for stable primary-line routing.
- `weighted`: weighted random distribution for traffic spreading.
- `smart`: score-based selection using health, recent success rate, latency, manual weight, and quota condition.

Default smart score:

```text
score =
  health_score * 0.40
  + success_rate_5m * 0.25
  + latency_score * 0.20
  + weight * 0.10
  + quota_score * 0.05
```

## Health Checks And Circuit Breaking

Health checks run periodically per upstream site.

Signals:

- `/v1/models` availability.
- Optional lightweight model request for configured test model.
- Response latency.
- HTTP status and normalized error category.
- Capability support for `responses`, `chat`, `streaming`, and other declared capabilities.

Circuit states:

- `closed`: normal traffic allowed.
- `open`: traffic blocked because recent failures exceeded threshold.
- `half_open`: limited probe traffic allowed after cooldown.

Behavior:

- Consecutive failures open the circuit.
- Cooldown moves `open` to `half_open`.
- Successful probe closes the circuit.
- Failed probe returns to `open`.

## Authentication, Authorization, Rate Limits, And Quota

Gateway clients use RelayDeck-issued API keys:

```text
Authorization: Bearer rd_live_xxx
```

Clients never receive or use upstream credentials directly.

API key validation order:

1. Key exists and hash matches.
2. Key status is active.
3. Key is not expired.
4. Owning user is active.
5. Key scope allows the endpoint type, such as `responses`, `chat`, or `embeddings`.
6. Requested model is in the key whitelist.
7. IP/source restrictions pass.
8. Rate limit passes.
9. Quota is not exceeded.

Key storage:

- Store only `key_hash` and `key_prefix` for RelayDeck keys.
- Show the plaintext key only once during creation.
- Store upstream credentials separately in `upstream_credentials`.
- Encrypt upstream credentials with `UPSTREAM_CREDENTIAL_ENCRYPTION_KEY` before storing them.

Rate limiting:

- MVP single-instance mode can use in-memory token buckets.
- Multi-instance mode should use Redis.
- Dimensions include `api_key_id`, `user_id`, and `model`.
- Support RPM and TPM limits.
- Exact TPM accounting can be refined after upstream `usage` data is available; streaming requests may use estimated input tokens and post-stream output token reconciliation.

Quota:

- MVP uses soft token quota or soft USD quota.
- Token quota is more reliable initially because USD requires pricing configuration.
- Requests over quota should be rejected or flagged depending on key policy.
- Precise USD accounting can be added later with `model_pricing`.

Security rules:

- `/api/admin/*` uses admin login/session authentication.
- `/v1/*` uses RelayDeck API key authentication.
- Logs must not store full API keys, full Authorization headers, or upstream credentials.
- Request body logging is disabled by default. Debug logging may be enabled temporarily for a specific user/key.

## PostgreSQL Data Model

Core tables:

- `users`: admin/backend users with email, password hash, role, status, and timestamps.
- `api_keys`: RelayDeck-issued keys with prefix, hash, owner, scopes, model whitelist, limits, status, expiration, and timestamps.
- `upstream_sites`: upstream relay site configuration including name, base URL, type, region, status, weight, timeout, and notes.
- `upstream_credentials`: upstream API key/token records, encrypted at rest and linked to upstream sites.
- `models`: RelayDeck canonical model names and capabilities.
- `site_models`: mapping from RelayDeck model names to upstream model names, with per-site capability flags.
- `routing_policies`: global or per-model routing rules, thresholds, retry count, cooldown, and mode.
- `request_logs`: final log record for each gateway request.
- `request_attempts`: per-upstream attempt records for each gateway request.
- `upstream_health_checks`: health check samples and errors.
- `usage_daily`: daily aggregates by user, API key, model, and upstream.

Recommended later tables:

- `model_pricing`: token pricing and cost rules for USD accounting.
- `audit_events`: admin operations and security-sensitive changes.
- `notifications`: alert and notification delivery state.

## Admin API Scope

Admin APIs use RelayDeck-native JSON shapes and are separate from OpenAI-compatible `/v1/*` responses.

`auth`:

- `POST /api/admin/auth/login`
- `POST /api/admin/auth/logout`
- `GET /api/admin/auth/me`

`sites`:

- `GET /api/admin/sites`
- `POST /api/admin/sites`
- `PATCH /api/admin/sites/{id}`
- `POST /api/admin/sites/{id}/test`
- `POST /api/admin/sites/{id}/enable`
- `POST /api/admin/sites/{id}/disable`

`models`:

- `GET /api/admin/models`
- `POST /api/admin/models`
- `PATCH /api/admin/models/{id}`
- `GET /api/admin/models/{id}/sites`
- `PUT /api/admin/models/{id}/sites`

`routing`:

- `GET /api/admin/routing/policies`
- `PUT /api/admin/routing/policies/{id}`
- `GET /api/admin/routing/candidates?model=...`
- `GET /api/admin/routing/history`

`api-keys`:

- `GET /api/admin/api-keys`
- `POST /api/admin/api-keys`
- `PATCH /api/admin/api-keys/{id}`
- `POST /api/admin/api-keys/{id}/revoke`
- `POST /api/admin/api-keys/{id}/rotate`

`users`:

- `GET /api/admin/users`
- `POST /api/admin/users`
- `PATCH /api/admin/users/{id}`
- `POST /api/admin/users/{id}/disable`

`logs`:

- `GET /api/admin/logs/requests`
- `GET /api/admin/logs/requests/{id}`
- `GET /api/admin/logs/attempts?request_id=...`

`stats`:

- `GET /api/admin/stats/overview`
- `GET /api/admin/stats/usage`
- `GET /api/admin/stats/upstreams`
- `GET /api/admin/stats/models`

MVP admin APIs should support pagination, basic filtering, and masked credential/key responses. Sorting can be added where needed by the frontend.

## Backend Project Structure

Create backend code under `backend/`:

```text
backend/
  cmd/relaydeck/
    main.go
  internal/
    app/
    config/
    http/
      gateway/
      admin/
      middleware/
    auth/
    router/
    upstream/
    health/
    quota/
    logs/
    stats/
    store/
      postgres/
    domain/
  migrations/
  configs/
  tests/
```

Responsibilities:

- `cmd/relaydeck`: process entrypoint and dependency initialization.
- `internal/app`: application wiring and lifecycle.
- `internal/config`: environment and config loading.
- `internal/http/gateway`: `/v1/*` handlers.
- `internal/http/admin`: `/api/admin/*` handlers.
- `internal/http/middleware`: request IDs, auth, logging, recovery, CORS, and rate-limit middleware.
- `internal/auth`: password/session and API key verification.
- `internal/router`: route policy and candidate selection.
- `internal/upstream`: upstream clients, adapters, streaming, and normalized errors.
- `internal/health`: health checks and circuit-break state.
- `internal/quota`: usage and rate-limit enforcement.
- `internal/logs`: request and attempt logging.
- `internal/stats`: aggregation queries.
- `internal/store/postgres`: repository implementations.
- `internal/domain`: core entities and interfaces.

## Configuration

Environment variables:

- `DATABASE_URL`
- `REDIS_URL` optional
- `APP_SECRET`
- `UPSTREAM_CREDENTIAL_ENCRYPTION_KEY`
- `HTTP_ADDR`
- `LOG_LEVEL`
- `GATEWAY_REQUEST_TIMEOUT`
- `HEALTH_CHECK_INTERVAL`

Local development should provide a Docker Compose setup for backend and PostgreSQL. Redis should be optional until distributed rate limiting is implemented.

## Implementation Slices

Recommended implementation order:

1. Initialize Go service, config, health endpoint, PostgreSQL connection, and migrations.
2. Implement minimal admin APIs for upstream sites, models, and API keys.
3. Implement `GET /v1/models` from canonical model and site mapping data.
4. Implement `POST /v1/chat/completions` non-streaming proxy.
5. Implement `POST /v1/chat/completions` streaming proxy.
6. Implement `POST /v1/responses` direct upstream passthrough and safe simple conversion fallback.
7. Add routing policies, health checks, circuit breaking, pre-write failover, request logs, attempts, and basic in-memory rate limiting.
8. Connect frontend pages to real admin APIs.

## Testing And Verification

Backend verification should include:

- Unit tests for API key verification.
- Unit tests for route candidate filtering and scoring.
- Unit tests for circuit-break state transitions.
- Integration tests for PostgreSQL migrations and repositories.
- HTTP tests for `/v1/models`, non-streaming chat completions, streaming chat completions, and responses passthrough using a fake upstream server.
- Admin API tests for site, model, API key, logs, and stats endpoints.

Minimum acceptance criteria for MVP:

- A downstream OpenAI-compatible client can call RelayDeck with a RelayDeck API key.
- RelayDeck can route to at least one configured upstream and return a valid response.
- Streaming responses are proxied without buffering the full response.
- Failed pre-write upstream attempts can fail over to another candidate.
- Request logs and attempt logs are persisted.
- `/api/admin/*` can manage enough configuration for the gateway to operate.
