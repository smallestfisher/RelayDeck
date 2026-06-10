# RelayDeck Backend

First backend slice for the RelayDeck gateway. This service currently uses in-memory bootstrap data so the gateway path can be tested before PostgreSQL repositories are wired.

## Run

```bash
GOCACHE=/tmp/go-build go run ./cmd/relaydeck
```

Defaults:

- `HTTP_ADDR=:8080`
- `APP_SECRET=dev-secret`
- `GATEWAY_REQUEST_TIMEOUT=30s`

## Test

```bash
GOCACHE=/tmp/go-build go test ./...
```

`GOCACHE` is set under `/tmp` because the Codex sandbox may make the default home cache read-only.

## Local Seed Key

The in-memory store uses this development-only RelayDeck API key:

```text
rd_live_dev_test_secret
```

This key is for local tests only and must not be used as a production secret.
