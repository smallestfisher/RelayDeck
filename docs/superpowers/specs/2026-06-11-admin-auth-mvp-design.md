# RelayDeck Admin Auth MVP Design

Date: 2026-06-11

## Context

RelayDeck now has a Go backend slice for the OpenAI-compatible gateway and a root `src/` React management console. The gateway uses RelayDeck-issued API keys for `/v1/*`; the admin surface uses email/password login, HttpOnly cookies, PostgreSQL-backed admin users, and Redis-backed sessions when configured.

This milestone is about making the management console real enough to operate: an authenticated admin session, protected `/api/admin/*` endpoints, and a login page that talks to the backend. Registration stays invitation-only and is not opened to the public in this slice.

## Goals

- Provide admin login, logout, and session lookup APIs.
- Protect `/api/admin/*` with server-side session authentication.
- Persist admin users in PostgreSQL when `DATABASE_URL` is configured.
- Store admin sessions in Redis when `REDIS_URL` is configured.
- Allow the root login page to restore a session on refresh and submit credentials to the backend.
- Keep the admin user model aligned with the existing `users` table in `backend/migrations/0001_initial.sql`.
- Preserve the current product direction: no public signup flow in this slice.

## Non-Goals

- No public registration endpoint.
- No OAuth, SSO, MFA, or password reset flow.
- No role management UI beyond the existing mock console pages.
- No persistence migration for gateway configuration, API keys, model mappings, logs, or usage aggregates in this slice.
- No Redis-backed rate limiting or circuit breaker state in this slice.

## Authentication Model

Admin authentication uses an opaque session token stored in an HttpOnly cookie.

Flow:

1. The user submits email and password to `POST /api/admin/auth/login`.
2. The backend verifies the password against the stored hash for the matching admin user.
3. The backend creates a random session token and stores the session server-side. Redis is used when `REDIS_URL` is configured; memory is only the local fallback.
4. The backend returns a minimal user payload and sets `relaydeck_session` as an HttpOnly cookie.
5. The frontend calls `GET /api/admin/auth/me` on startup to restore the session.
6. `POST /api/admin/auth/logout` invalidates the session and clears the cookie.

The cookie is the only credential the browser stores. The frontend never stores the raw password after submission and never stores the session token in localStorage.

## Bootstrap Admin

The first slice should support a bootstrap owner/admin account so the system can be initialized without a separate provisioning UI.

Recommended bootstrap inputs:

- `APP_BOOTSTRAP_OWNER_EMAIL`
- `APP_BOOTSTRAP_OWNER_PASSWORD`

Behavior:

- If the PostgreSQL `users` table is empty, the service seeds one active owner/admin user at startup.
- If the user table already has data, bootstrap seeding is skipped.
- The bootstrap password is hashed before storage and never echoed back in API responses or logs.

## Session Scope

Sessions are only for admin login and only authorize `/api/admin/*` routes.

- `/api/admin/auth/login`, `/api/admin/auth/logout`, and `/api/admin/auth/me` are public entry points.
- All other `/api/admin/*` routes require a valid session.
- `/v1/*` continues to use RelayDeck API keys and does not accept admin sessions.

## Data Model

The existing `users` table remains the source of truth for admin identity when `DATABASE_URL` is configured.

Admin session state is stored in Redis when `REDIS_URL` is configured. Session records have this shape:

- session token
- user id
- email
- role
- issued at
- expires at
- last seen at

The memory store remains available only as a local development fallback when Redis or PostgreSQL are intentionally not configured.

## Environment Management

Environment variables are documented in `.env.example`. Developers copy it to `.env` for local development and integration tests. Backend startup and Go integration tests load `.env` automatically.

Core variables:

- `DATABASE_URL`
- `REDIS_URL`
- `APP_BOOTSTRAP_OWNER_EMAIL`
- `APP_BOOTSTRAP_OWNER_PASSWORD`
- `GATEWAY_REQUEST_TIMEOUT`

## Login Page Behavior

The root login page becomes a real entry point for the admin console.

- On mount, it checks `GET /api/admin/auth/me`.
- If the session exists, it transitions into the authenticated console.
- If the session is absent or invalid, it stays on the login view.
- The login form posts email/password to the backend and shows errors inline or in a banner area.
- The register tab does not create accounts; it shows an invitation-only message or a disabled state.

## Security Rules

- Passwords are hashed before storage.
- Sessions are opaque and server-side.
- Production session storage should use Redis via `REDIS_URL`.
- Session cookies are HttpOnly and same-site.
- Login, logout, and session lookup never expose the password hash or session token.
- Admin-only routes must reject missing or expired sessions with 401.
- Login failures should not reveal whether the email or password was wrong.

## Testing And Verification

The MVP should be verified with backend unit tests and a frontend build.

Backend tests should cover:

- password hashing and verification
- session creation, lookup, expiry, and logout
- Redis session lifecycle with Docker-backed Redis
- PostgreSQL owner bootstrap with Docker-backed PostgreSQL
- login success and login failure
- protected admin route rejection without a session
- `/api/admin/auth/me` returning the current admin user

Frontend verification should cover:

- login page submits credentials to the backend
- session restoration on refresh works
- logout returns the app to the login screen
- register tab does not create a real account
