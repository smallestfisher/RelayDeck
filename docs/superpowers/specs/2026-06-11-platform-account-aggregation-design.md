# RelayDeck Platform Account Aggregation Design

Date: 2026-06-11

## Context

RelayDeck is moving from admin login/session groundwork into the upstream management layer. The product now needs to aggregate and operate multiple upstream accounts from relay platforms such as `new-api` and `sub2api`, keep their model inventories synchronized, and route downstream traffic through the most appropriate upstream protocol instead of relying on whatever protocol conversion a relay platform happens to offer.

The reference material is split across three sources:

- `Reference/new-api`: upstream relay platform source with explicit support for OpenAI, Anthropic, Gemini, Responses, and other endpoint families.
- `Reference/sub2api`: upstream relay platform source with account/channel management, model mapping, and monitor history.
- `Reference/codex` and `Reference/claude-code`: client-side behavior and configuration references for OpenAI Responses and Anthropic-style requests.

The `claude-code` reference directory currently provides public package/configuration artifacts and behavioral clues rather than a complete private client tree. That is enough to define the compatibility surface, but not enough to pretend we have a full internal client implementation.

## Goals

- Manage upstream accounts from relay platforms and direct providers in one administration model.
- Sync upstream model inventories into RelayDeck with normalized model records and per-model protocol metadata.
- Prefer the model's native wire protocol when RelayDeck talks upstream.
- Support protocol conversion only when it is explicit and lossless enough to preserve request semantics.
- Track upstream health from real call outcomes, not only from scheduled probes.
- Make Codex and Claude Code client profiles first-class compatibility inputs for gateway behavior.
- Keep secrets and transient health state separated, with PostgreSQL for durable metadata and Redis for ephemeral state.

## Non-Goals

- No attempt to fully reimplement new-api or sub2api internally.
- No public account onboarding flow for external users in this milestone.
- No billing engine, quota settlement, or payout logic in this slice.
- No attempt to support every exotic upstream provider on day one.
- No dependence on upstream platform-side protocol conversion as the primary path.
- No speculative support for a client behavior that is not visible in the reference material.

## Core Decision

RelayDeck treats protocol choice as a property of the upstream model and account, not as a global platform setting.

That means:

- the account record knows what the upstream can speak;
- the model record knows what it natively expects;
- the request router chooses the direct protocol first;
- lossy conversion is a fallback only when the chosen account cannot speak the native protocol directly.

This is the key product rule behind the request: even if `new-api` or `sub2api` can translate between protocols, RelayDeck should still prefer to call the model in its native shape whenever possible.

## Architecture

RelayDeck will use three cooperating layers.

### 1. Upstream account registry

This is the administrative source of truth for upstream connections. Each record represents either a relay-platform account or a direct provider account. It stores the base URL, authentication material, platform kind, enabled state, and protocol preferences.

### 2. Model capability catalog

Each synchronized model record stores a normalized RelayDeck model name, the upstream account it came from, the upstream model identifier, the model's native protocol, the set of supported wire protocols, and a capability set such as chat, responses, streaming, tools, vision, embeddings, or realtime where applicable.

### 3. Runtime routing and health control

Every downstream request resolves a model into one or more upstream candidates. Candidate scoring uses model compatibility, protocol match quality, recent health, latency, circuit state, and account policy. Real request outcomes feed the health tracker, which updates the account/model score and circuit state.

## Data Model

The PostgreSQL schema should store durable metadata in normalized tables.

### Upstream accounts

Fields:

- `id`
- `name`
- `platform_kind` - for example `new_api`, `sub2api`, `direct_openai`, `direct_anthropic`
- `base_url`
- `enabled`
- `priority`
- `auth_kind` - for example bearer token, api key header, oauth token, or signed credentials
- `auth_secret_enc`
- `default_wire_protocol`
- `supported_wire_protocols`
- `metadata`
- `created_at`
- `updated_at`

The secret is encrypted before storage. PostgreSQL should never store a plaintext upstream credential.

### Synchronized models

Fields:

- `id`
- `upstream_account_id`
- `normalized_model_name`
- `upstream_model_name`
- `display_name`
- `native_wire_protocol`
- `supported_wire_protocols`
- `capabilities`
- `status`
- `sync_source`
- `last_synced_at`
- `raw_metadata`

The normalized model name is what downstream clients use. The upstream model name is what RelayDeck sends to the upstream account after mapping.

### Model bindings

RelayDeck should allow explicit bindings between normalized models and one or more upstream accounts. This is how the system expresses that the same RelayDeck model may be available through multiple upstreams.

Fields:

- `id`
- `normalized_model_name`
- `upstream_account_id`
- `upstream_model_name`
- `preferred_wire_protocol`
- `capability_override`
- `weight`
- `enabled`

### Health observations

Durable history should record one row per meaningful upstream outcome.

Fields:

- `id`
- `upstream_account_id`
- `normalized_model_name`
- `upstream_model_name`
- `wire_protocol`
- `request_kind`
- `success`
- `status_code`
- `error_class`
- `latency_ms`
- `request_bytes`
- `response_bytes`
- `observed_at`
- `request_id`

Redis keeps the current rolling health summary and circuit state. PostgreSQL keeps the event history.

## Protocol Selection

Protocol selection is a request-time decision, not a permanent platform setting.

### Priority order

1. Use the model's native protocol if the selected upstream account supports it.
2. If the native protocol is unavailable on that account, use a lossless compatible protocol that preserves the request shape.
3. If neither is available, reject the request with a clear compatibility error instead of silently converting away semantics.

### Expected protocol families

The initial set should cover:

- `openai_responses`
- `openai_chat_completions`
- `anthropic_messages`
- `gemini_generate_content`

RelayDeck should treat protocol support as explicit metadata per account and per model. The platform must not assume that a relay platform can safely translate every request shape just because it exposes a convenient compatibility endpoint.

### Conversion rule

Conversion is allowed only when all of the following are true:

- the requested endpoint can be represented faithfully in the target wire format;
- the upstream account does not already support the native protocol;
- the target account/model combination is marked as compatible for that conversion path;
- the converted request preserves tool calls, streaming semantics, and message boundaries.

This rule is intentionally strict. If a request cannot be translated without semantic loss, RelayDeck should fail fast.

## Health Model

Health must come from actual traffic, not just scheduled probes.

### Inputs

- successful downstream calls
- failed downstream calls
- latency per call
- error classification
- consecutive failure count
- recovery success streak
- periodic probe result

### Error classes

Use a small normalized set instead of free-form strings:

- `auth_error`
- `quota_error`
- `rate_limit`
- `timeout`
- `transport_error`
- `protocol_mismatch`
- `invalid_response`
- `upstream_5xx`
- `unknown_error`

### Scoring behavior

The current health score should be a rolling value computed from recent observations. One practical model is:

- success rate over a recent window as the base signal;
- latency trend as a secondary signal;
- consecutive failures as a strong penalty;
- auth and protocol mismatches as immediate heavy penalties;
- successful recovery probes as gradual recovery.

Redis should hold the live summary with a short TTL so the router can read it quickly. PostgreSQL should store the observation log so the score can be recomputed if needed.

### Circuit behavior

- a few consecutive hard failures should open the circuit for that account/model combination;
- the half-open state should allow limited probes after cooldown;
- a successful probe should close the circuit;
- protocol mismatch should count as a hard failure and should not be retried blindly on the same path.

## Client Profiles

RelayDeck should model downstream client behavior explicitly because the gateway will eventually need to forward requests that look like Codex or Claude Code.

### Built-in profiles

Initial profiles:

- `codex`
- `claude_code`

Each profile defines:

- request envelope shape;
- default endpoint family;
- header conventions;
- streaming expectations;
- retry behavior;
- model naming and selection hints;
- tool payload conventions where applicable.

### Observed compatibility inputs

From the reference material, the following are relevant:

- Codex leans on OpenAI Responses-style traffic and related model/provider metadata.
- Claude Code leans on Anthropic-style auth and header conventions such as `anthropic-version`, `anthropic-beta`, and bearer or API-key auth variants.

RelayDeck does not need to perfectly clone every internal client detail on day one. It does need to preserve the request shape and headers that upstream platforms or direct providers use to distinguish client families, especially when those details affect protocol acceptance or model selection.

### Profile extensibility

The profile registry should allow future additions without changing the router contract. The registry can be seeded in code first and later exposed through admin configuration if needed.

## Upstream Discovery And Sync

Discovery is split by platform type.

### new-api

Use the upstream admin/model endpoints to discover:

- supported models;
- model-to-endpoint compatibility;
- upstream capabilities;
- account-level protocol support;
- any explicit provider or vendor metadata the platform exposes.

### sub2api

Use the upstream channel/account and model-mapping endpoints to discover:

- upstream channels/accounts;
- platform membership;
- model mappings;
- monitoring metadata;
- per-channel protocol selection or `api_mode` equivalent fields where exposed.

### Sync policy

- sync should be idempotent;
- sync should not delete a model mapping blindly if it disappears once;
- manual bindings should win over auto-discovered defaults;
- raw upstream metadata should be retained for debugging and later reconciliation.

## Admin APIs

RelayDeck needs operational endpoints for the new account and model layer.

Recommended admin surface:

- `GET /api/admin/upstreams/accounts`
- `POST /api/admin/upstreams/accounts`
- `GET /api/admin/upstreams/accounts/{id}`
- `PUT /api/admin/upstreams/accounts/{id}`
- `DELETE /api/admin/upstreams/accounts/{id}`
- `POST /api/admin/upstreams/accounts/{id}/sync-models`
- `GET /api/admin/upstreams/accounts/{id}/models`
- `POST /api/admin/upstreams/accounts/{id}/probe`
- `GET /api/admin/models`
- `GET /api/admin/models/{name}`
- `PUT /api/admin/models/{name}/bindings`
- `GET /api/admin/health/summary`
- `GET /api/admin/health/events`

These APIs should stay admin-only and should never expose plaintext upstream credentials.

## Error Handling

Upstream failures should be normalized before they reach the router or the admin UI.

Rules:

- auth failures should be classified as credential problems, not generic network failures;
- protocol mismatch should be surfaced as a compatibility problem, not a retryable outage;
- upstream 429s should be distinguished from 5xx and timeouts;
- malformed upstream responses should be recorded as invalid responses and penalized;
- downstream error bodies should be sanitized so they do not leak secrets or internal implementation details.

When a request cannot be routed because no compatible upstream exists, the API should return a deterministic compatibility error rather than a vague 500.

## Testing And Verification

This feature needs backend tests and reference-driven compatibility checks.

### Backend tests

- upstream account CRUD and secret encryption
- model sync from mocked new-api/sub2api responses
- native-protocol selection for a model that supports multiple protocols
- compatibility rejection when a request cannot be translated losslessly
- health score updates for success, timeout, auth failure, protocol mismatch, and recovery
- Redis-backed live health cache
- PostgreSQL-backed health observation history
- admin API authorization on the new endpoints

### Compatibility tests

- Codex-style Responses request shaping
- Claude Code-style Anthropic header shaping
- upstream selection when a model exists on multiple upstream accounts
- routing preference for native protocol over conversion endpoint

### Verification criteria

- the router chooses the native upstream protocol whenever it exists;
- the health score changes after real request outcomes;
- admin APIs can list accounts and model bindings;
- credentials remain encrypted at rest;
- no test relies on temporary `test_*` environment names instead of `.env` plus `.env.example`.

## Rollout Sequence

The implementation should land in this order:

1. schema and repository support for upstream accounts, model bindings, and health observations;
2. account/model sync adapters for `new-api` and `sub2api`;
3. routing updates that prefer native protocol selection;
4. health scoring and Redis cache wiring;
5. Codex and Claude Code profile emulation;
6. admin UI/API wiring for account and model management.

This order keeps the system usable at each step and avoids building protocol conversion logic before the model catalog exists.

