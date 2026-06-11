# RelayDeck Backend

First backend slice for the RelayDeck gateway. Gateway configuration still uses in-memory bootstrap data, while admin users can be persisted to PostgreSQL when `DATABASE_URL` is configured and admin sessions can be shared through Redis when `REDIS_URL` is configured.

## Run

Copy the environment template first:

```bash
cp ../.env.example ../.env
```

Start local PostgreSQL and Redis:

```bash
docker compose up -d postgres redis
```

The compose file uses a mirror registry for local development because direct Docker Hub pulls can be unreliable in this environment.

```bash
GOCACHE=/tmp/go-build go run ./cmd/relaydeck
```

Defaults:

- `HTTP_ADDR=:8080`
- `APP_SECRET=dev-secret`
- `DATABASE_URL=` keeps admin users in memory
- `REDIS_URL=` keeps admin sessions in memory
- `GATEWAY_REQUEST_TIMEOUT=30s`
- `APP_BOOTSTRAP_OWNER_EMAIL=owner@example.com`
- `APP_BOOTSTRAP_OWNER_PASSWORD=change-me`

When `DATABASE_URL` is set, the backend ensures the `users` table exists and bootstraps the owner account only when the table is empty. When `REDIS_URL` is set, admin sessions are stored under `relaydeck:session:*` keys in Redis. Empty values keep local development on in-memory fallback stores.

## Test

```bash
GOCACHE=/tmp/go-build go test ./...
```

`GOCACHE` is set under `/tmp` because the Codex sandbox may make the default home cache read-only.

PostgreSQL and Redis integration tests read `DATABASE_URL` and `REDIS_URL` from `.env`. The PostgreSQL integration test requires `DATABASE_URL` to point to a database name containing `test` before it resets the `users` table.

## Local Admin And Gateway Credentials

Default local admin login:

```text
owner@example.com / change-me
```

Override it with `APP_BOOTSTRAP_OWNER_EMAIL` and `APP_BOOTSTRAP_OWNER_PASSWORD` before first PostgreSQL bootstrap.

Local gateway API key:

The in-memory store uses this development-only RelayDeck API key:

```text
rd_live_dev_test_secret
```

This key is for local tests only and must not be used as a production secret.
