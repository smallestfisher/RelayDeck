# Admin Auth MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add real RelayDeck admin login/session handling, protect `/api/admin/*`, connect the root login page to the backend, persist admin users in PostgreSQL, and store production sessions in Redis while keeping registration invitation-only.

**Architecture:** Use a server-side opaque session stored behind an HttpOnly cookie for admin auth. Admin users are loaded from PostgreSQL when `DATABASE_URL` is configured; Redis stores sessions when `REDIS_URL` is configured. In-memory stores remain only as local development fallback paths. The frontend restores the session on load, submits email/password on login, and treats the register tab as invitation-only guidance.

**Tech Stack:** Go `net/http`, PostgreSQL via pgx, Redis via go-redis, React 18, TypeScript, Vite, Tailwind CSS, existing RelayDeck UI primitives.

**Status:** Implemented and verified on 2026-06-11. Follow-up work should move gateway configuration, API keys, model mappings, logs, and usage aggregates from memory seed data into PostgreSQL repositories.

---

### Task 1: Define Backend Auth Primitives

**Files:**
- Create: `backend/internal/auth/password.go`
- Create: `backend/internal/auth/password_test.go`
- Create: `backend/internal/auth/session.go`
- Create: `backend/internal/auth/session_test.go`

- [ ] **Step 1: Write password and session tests**

```go
func TestHashPasswordAndVerify(t *testing.T)
func TestSessionLifecycle(t *testing.T)
```

- [ ] **Step 2: Run the new auth tests and confirm they fail**

Run: `cd backend && go test ./internal/auth -run 'TestHashPasswordAndVerify|TestSessionLifecycle' -v`

Expected: FAIL because the new files do not exist yet.

- [ ] **Step 3: Implement password hashing and opaque sessions**

```go
package auth

func HashPassword(password string) (string, error)
func VerifyPassword(hash string, password string) bool

type Session struct {
    Token     string
    UserID    string
    Email     string
    Role      string
    IssuedAt  time.Time
    ExpiresAt time.Time
    LastSeenAt time.Time
}

type SessionStore interface {
    Create(session Session) error
    Get(token string) (Session, bool)
    Delete(token string) error
}
```

- [ ] **Step 4: Run the auth tests and confirm they pass**

Run: `cd backend && go test ./internal/auth -run 'TestHashPasswordAndVerify|TestSessionLifecycle' -v`

Expected: PASS.

- [ ] **Step 5: Commit the auth primitives**

```bash
git add backend/internal/auth/password.go backend/internal/auth/password_test.go backend/internal/auth/session.go backend/internal/auth/session_test.go
git commit -m "Add admin auth primitives"
```

### Task 2: Add Admin Login Handler and Middleware

**Files:**
- Create: `backend/internal/http/admin/auth_handler.go`
- Create: `backend/internal/http/middleware/admin_auth.go`
- Modify: `backend/internal/http/admin/handler.go`
- Modify: `backend/internal/app/app.go`

- [ ] **Step 1: Write handler tests first**

```go
func TestAdminLoginSetsSessionCookie(t *testing.T)
func TestAdminMeRequiresSession(t *testing.T)
func TestAdminLogoutClearsSession(t *testing.T)
func TestProtectedAdminRouteRejectsNoSession(t *testing.T)
```

- [ ] **Step 2: Run the admin HTTP tests and confirm they fail**

Run: `cd backend && go test ./internal/http/admin ./internal/http/middleware -v`

Expected: FAIL because the auth handler and middleware do not exist yet.

- [ ] **Step 3: Implement login/logout/me and session protection**

```go
POST /api/admin/auth/login
POST /api/admin/auth/logout
GET  /api/admin/auth/me
```

The login handler should:

- accept JSON `{ "email": "...", "password": "..." }`
- verify against stored users
- create a session token
- set `relaydeck_session` as an HttpOnly cookie
- return a minimal user payload

The middleware should:

- allow `/api/admin/auth/login`, `/api/admin/auth/logout`, and `/api/admin/auth/me`
- require a valid session for all other `/api/admin/*` routes
- attach the current admin user to request context for handlers that need it

- [ ] **Step 4: Wire the middleware into the app**

Update `backend/internal/app/app.go` so `/api/admin/*` runs through the auth gate, while `/v1/*` remains on API key auth only.

- [ ] **Step 5: Run the admin tests and confirm they pass**

Run: `cd backend && go test ./internal/http/admin ./internal/http/middleware -v`

Expected: PASS.

- [ ] **Step 6: Commit the admin HTTP changes**

```bash
git add backend/internal/http/admin/auth_handler.go backend/internal/http/middleware/admin_auth.go backend/internal/http/admin/handler.go backend/internal/app/app.go
git commit -m "Add admin login session flow"
```

### Task 3: Bootstrap the First Admin User

**Files:**
- Modify: `backend/internal/store/memory.go`
- Modify: `backend/internal/domain/domain.go`
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/config/config.go`

- [ ] **Step 1: Write bootstrap tests**

```go
func TestBootstrapOwnerSeededWhenStoreIsEmpty(t *testing.T)
func TestBootstrapSkippedWhenStoreAlreadyHasUsers(t *testing.T)
```

- [ ] **Step 2: Run the bootstrap tests and confirm they fail**

Run: `cd backend && go test ./internal/app ./internal/store -v`

Expected: FAIL until bootstrap seeding is implemented.

- [ ] **Step 3: Add bootstrap owner seeding**

Seed a single active owner/admin from config when the store is empty. The seeded user should use the same `users` table shape already drafted in `backend/migrations/0001_initial.sql`.

- [ ] **Step 4: Run the bootstrap tests and confirm they pass**

Run: `cd backend && go test ./internal/app ./internal/store -v`

Expected: PASS.

- [ ] **Step 5: Commit the bootstrap change**

```bash
git add backend/internal/store/memory.go backend/internal/domain/domain.go backend/internal/app/app.go backend/internal/config/config.go
git commit -m "Seed bootstrap admin user"
```

### Task 4: Connect the Root Login Page

**Files:**
- Modify: `src/pages/LoginPage.tsx`
- Modify: `src/App.tsx`
- Modify: `src/types.ts`
- Modify: `src/lib/format.ts` if needed for auth messages

- [ ] **Step 1: Add login page tests or build checks**

Add the smallest local state and HTTP call coverage needed to prove the page can submit login and restore the session.

- [ ] **Step 2: Run the frontend build and confirm the current app still compiles**

Run: `npm run build`

Expected: PASS before wiring the new HTTP calls.

- [ ] **Step 3: Implement session restore and login submit**

Make the login page call `GET /api/admin/auth/me` on mount, `POST /api/admin/auth/login` on submit, and `POST /api/admin/auth/logout` from the authenticated shell when logout is added.

- [ ] **Step 4: Keep registration invitation-only**

Leave the register tab non-functional and show a clear invitation-only message instead of creating accounts locally.

- [ ] **Step 5: Run the frontend build again**

Run: `npm run build`

Expected: PASS.

- [ ] **Step 6: Commit the login page wiring**

```bash
git add src/pages/LoginPage.tsx src/App.tsx src/types.ts src/lib/format.ts
git commit -m "Wire login page to admin auth"
```

### Task 5: Final Verification

**Files:**
- All files changed in this plan

- [ ] **Step 1: Run the backend test suite**

Run: `cd backend && go test ./...`

Expected: PASS.

- [ ] **Step 2: Run the frontend build**

Run: `npm run build`

Expected: PASS.

- [ ] **Step 3: Review git state**

Run: `git status --short`

Expected: only the intended admin auth files, frontend wiring files, and this plan/spec set are changed.

- [ ] **Step 4: Commit the completed slice**

```bash
git add .
git commit -m "Add admin auth MVP"
```

### Task 6: PostgreSQL Admin Users, Redis Sessions, and Environment Management

**Files:**
- Create: `.env.example`
- Create: `docker-compose.yml`
- Create: `backend/internal/config/env.go`
- Create: `backend/internal/store/admin.go`
- Create: `backend/internal/store/postgres/users.go`
- Create: `backend/internal/store/postgres/users_test.go`
- Create: `backend/internal/auth/redis_session.go`
- Create: `backend/internal/auth/redis_session_test.go`
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/auth/session.go`
- Modify: `backend/internal/http/admin/auth_handler.go`
- Modify: `backend/internal/http/middleware/admin_auth.go`
- Modify: `README.md`
- Modify: `backend/README.md`
- Modify: `backend/go.mod`
- Create: `backend/go.sum`

- [x] **Step 1: Document environment variables**

Create `.env.example` as the source of truth for `DATABASE_URL`, `REDIS_URL`, `APP_BOOTSTRAP_OWNER_EMAIL`, `APP_BOOTSTRAP_OWNER_PASSWORD`, and service runtime settings. Local `.env` is ignored by Git and loaded automatically by backend startup and integration tests.

- [x] **Step 2: Add Docker local infrastructure**

Create `docker-compose.yml` with PostgreSQL 16 and Redis 7 services, health checks, and persistent local volumes.

- [x] **Step 3: Persist admin users in PostgreSQL**

Implement `backend/internal/store/postgres.UserStore` with `EnsureSchema`, `BootstrapOwner`, `Users`, `UserByEmail`, and `UserByID`. The bootstrap owner is inserted only when the `users` table is empty.

- [x] **Step 4: Store sessions in Redis**

Implement `auth.RedisSessionStore` behind the existing `auth.SessionStore` interface. `SessionStore.Create` and `Delete` now return errors so login can fail closed if Redis writes fail.

- [x] **Step 5: Wire infrastructure in `app.New`**

Use PostgreSQL for admin users when `DATABASE_URL` is configured. Use Redis for admin sessions when `REDIS_URL` is configured. Keep memory stores as local development fallback only.

- [x] **Step 6: Verify with Docker-backed integration tests**

Run PostgreSQL and Redis containers, then run:

```bash
cd backend && GOTOOLCHAIN=local GOCACHE=/tmp/go-build go test ./internal/store/postgres ./internal/auth -v
```

Expected: PostgreSQL owner bootstrap and Redis session lifecycle tests pass using `.env` values.

- [x] **Step 7: Run full verification**

```bash
cd backend && GOTOOLCHAIN=local GOCACHE=/tmp/go-build go test ./...
npm run build
```

Expected: both commands pass.

## Self-Review

- Spec coverage: login, logout, session lookup, protected admin routes, PostgreSQL bootstrap owner, Redis session storage, environment management, login page wiring, and invitation-only registration are each covered by a task.
- Placeholder scan: no TBD/TODO/fill-later language is present in the implementation steps.
- Type consistency: `relaydeck_session` is used consistently as the admin cookie name, and the backend auth flow remains separate from `/v1/*` API key auth.
