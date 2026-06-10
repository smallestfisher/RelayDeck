# Admin Auth MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add real RelayDeck admin login/session handling, protect `/api/admin/*`, and connect the root login page to the backend while keeping registration invitation-only.

**Architecture:** Use a server-side opaque session stored behind an HttpOnly cookie for admin auth. Keep the backend single-instance and memory-backed for this slice, but structure the auth and session code so it can later move behind PostgreSQL or Redis without changing the HTTP contract. The frontend should restore the session on load, submit email/password on login, and treat the register tab as non-functional guidance rather than a sign-up flow.

**Tech Stack:** Go `net/http`, in-memory session/auth helpers, React 18, TypeScript, Vite, Tailwind CSS, existing RelayDeck UI primitives.

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
    Create(session Session)
    Get(token string) (Session, bool)
    Delete(token string)
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

## Self-Review

- Spec coverage: login, logout, session lookup, protected admin routes, bootstrap admin seeding, login page wiring, and invitation-only registration are each covered by a task.
- Placeholder scan: no TBD/TODO/fill-later language is present in the implementation steps.
- Type consistency: `relaydeck_session` is used consistently as the admin cookie name, and the backend auth flow remains separate from `/v1/*` API key auth.

