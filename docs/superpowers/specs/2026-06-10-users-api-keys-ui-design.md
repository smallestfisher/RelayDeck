# Users and API Keys UI Design

## Context

RelayDeck currently has a primary React/Tailwind prototype under `src/` with routed pages controlled by `PageId`, `App`, and `Sidebar`. Two new design references were added under `Design_image/8.png` and `Design_image/9.png`:

- `8.png` shows the user management experience with metrics, filters, a user table, a right-side detail panel, and an invite-user modal.
- `9.png` shows the API Key management experience with metrics, filters, a key table, a right-side detail panel, and a create-key modal.

The dark and light appearances in the references represent theme variants, not page-specific color assignments. Both new pages must follow the existing global theme switch.

## Scope

Add two first-class UI pages to the main `src/` application:

- `用户管理` (`users`)
- `API Keys` (`apiKeys`)

The existing `frontend/` Ant Design app is not part of this change. It can remain as a separate legacy or alternate prototype.

## User Management Page

The page should mirror the structure of the design reference while fitting the current RelayDeck component language.

Required sections:

- Summary metrics for total users, active users, admin users, and monthly usage.
- Search and filter toolbar for keyword, role, status, and model access.
- User table with avatar, name, email, role, status, allowed models, API key count, monthly quota, used quota, usage percentage, and last login.
- Right-side detail drawer showing selected user profile, permissions, quota, usage, model access, recent activity, and primary actions.
- Invite-user modal opened from the page action button.

Initial data can be static mock data in `src/data/mock.ts`. Interactions can be prototype-level state only: selecting rows, opening/closing drawer, and opening/closing modal.

## API Keys Page

The page should mirror the structure of the design reference while fitting the current RelayDeck component language.

Required sections:

- Summary metrics for total keys, active keys, expiring keys, and blocked keys.
- Search and filter toolbar for keyword, status, permissions, owner, and expiry window.
- Key table with key name, owner, prefix, permission scopes, rate limits, allowed model count, created time, expiry time, recent use, status, and row actions.
- Right-side detail drawer showing selected key metadata, scopes, allowed models, IP whitelist, rate limits, usage, recent calls, and primary actions.
- Create-key modal opened from the page action button.

Initial data can be static mock data in `src/data/mock.ts`. Interactions can be prototype-level state only: selecting rows, opening/closing drawer, and opening/closing modal.

## Navigation and Routing

Extend the main app routing model:

- Add `users` and `apiKeys` to `PageId`.
- Add labels to `pageTitles`.
- Add sidebar entries for `用户管理` and `API Keys`.
- Render `UsersPage` and `ApiKeysPage` from `App`.

The sidebar order should follow the product management flow: dashboard and core routing pages first, then testing, users, API keys, logs, and settings.

## Components and Styling

Use the existing design system under `src/components/ui` where practical:

- Reuse `Card`, `Button`, `MetricCard`, `StatusBadge`, `Drawer`, and chart/table helpers where they fit.
- Keep styles Tailwind-based and theme-token-based so both pages render correctly in dark and light modes.
- Do not introduce Ant Design into `src/`.
- Avoid adding backend calls; this is a static UI prototype change.

If a small local helper is needed for repeated chips, progress bars, or action groups, keep it inside the page file unless it is clearly reusable by other pages.

## Data Model

Add explicit TypeScript interfaces in `src/types.ts` for:

- User management records.
- API key records.
- Recent activity or recent key call records if needed by detail panels.

Add mock arrays in `src/data/mock.ts` using realistic RelayDeck values and existing model/site terminology.

## Testing and Verification

Verification must include:

- `npm run build` from the repository root.
- A visual sanity pass by running the app locally if feasible, checking both dark and light themes for the two new pages.

Success criteria:

- Both pages are reachable from the sidebar.
- Both pages track the global theme.
- Tables, detail drawers, and modals match the layout intent of the new design references.
- Build completes without TypeScript errors.

