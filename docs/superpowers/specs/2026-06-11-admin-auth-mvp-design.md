# RelayDeck Admin Auth MVP Design

Date: 2026-06-11

## Context

RelayDeck now has a Go backend slice for the OpenAI-compatible gateway and a root `src/` React prototype for the management console. The current gateway already uses RelayDeck-issued API keys for `/v1/*`, but the admin surface still has no real login/session boundary. The login page in `src/pages/LoginPage.tsx` is still a local prototype, and the old Ant Design `frontend/` app has been removed.

This milestone is about making the management console real enough to operate: an authenticated admin session, protected `/api/admin/*` endpoints, and a login page that talks to the backend. Registration stays invitation-only and is not opened to the public in this slice.

## Goals

- Provide admin login, logout, and session lookup APIs.
- Protect `/api/admin/*` with server-side session authentication.
- Allow the root login page to restore a session on refresh and submit credentials to the backend.
- Keep the admin user model aligned with the existing `users` table in `backend/migrations/0001_initial.sql`.
- Preserve the current product direction: no public signup flow in this slice.

## Non-Goals

- No public registration endpoint.
- No OAuth, SSO, MFA, or password reset flow.
- No role management UI beyond the existing mock console pages.
- No full PostgreSQL-backed auth repository yet.
- No distributed session store or Redis-backed session sharing in this slice.

## Authentication Model

Admin authentication uses an opaque session token stored in an HttpOnly cookie.

Flow:

1. The user submits email and password to `POST /api/admin/auth/login`.
2. The backend verifies the password against the stored hash for the matching admin user.
3. The backend creates a random session token and stores the session server-side.
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

- If the user table is empty, the service seeds one active owner/admin user at startup.
- If the user table already has data, bootstrap seeding is skipped.
- The bootstrap password is hashed before storage and never echoed back in API responses or logs.

## Session Scope

Sessions are only for admin login and only authorize `/api/admin/*` routes.

- `/api/admin/auth/login`, `/api/admin/auth/logout`, and `/api/admin/auth/me` are public entry points.
- All other `/api/admin/*` routes require a valid session.
- `/v1/*` continues to use RelayDeck API keys and does not accept admin sessions.

## Data Model

The existing `users` table remains the source of truth for admin identity.

For the MVP, session state can live in memory with the following shape:

- session token
- user id
- email
- role
- issued at
- expires at
- last seen at

This is enough for a single-instance backend and keeps the implementation simple while preserving a path to a durable repository later.

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
- Session cookies are HttpOnly and same-site.
- Login, logout, and session lookup never expose the password hash or session token.
- Admin-only routes must reject missing or expired sessions with 401.
- Login failures should not reveal whether the email or password was wrong.

## Testing And Verification

The MVP should be verified with backend unit tests and a frontend build.

Backend tests should cover:

- password hashing and verification
- session creation, lookup, expiry, and logout
- login success and login failure
- protected admin route rejection without a session
- `/api/admin/auth/me` returning the current admin user

Frontend verification should cover:

- login page submits credentials to the backend
- session restoration on refresh works
- logout returns the app to the login screen
- register tab does not create a real account

