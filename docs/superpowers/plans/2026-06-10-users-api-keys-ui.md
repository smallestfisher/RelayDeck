# Users and API Keys UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add theme-aware 用户管理 and API Keys pages to the main RelayDeck React/Tailwind prototype.

**Architecture:** Extend the existing `src/` app rather than the alternate `frontend/` app. Add typed mock records, two focused page components, and route/sidebar wiring through the existing `PageId` model.

**Tech Stack:** React 18, TypeScript, Vite, Tailwind CSS, lucide-react, existing RelayDeck UI primitives.

---

## File Structure

- Modify: `src/types.ts` to add user/key status unions and mock-data interfaces.
- Modify: `src/data/mock.ts` to add `userRecords`, `apiKeyRecords`, and small recent-activity arrays embedded in records.
- Create: `src/pages/UsersPage.tsx` for metrics, filters, table, detail drawer, and invite modal.
- Create: `src/pages/ApiKeysPage.tsx` for metrics, filters, table, detail drawer, and create modal.
- Modify: `src/App.tsx` to import and render the two pages and title labels.
- Modify: `src/components/layout/Sidebar.tsx` to add the two navigation entries.
- Modify: `src/lib/format.ts` only if status display text needs new labels.

## Task 1: Add Types and Mock Data

**Files:**
- Modify: `src/types.ts`
- Modify: `src/data/mock.ts`

- [ ] **Step 1: Extend types**

Add unions and interfaces for `UserRecord`, `UserActivity`, `ApiKeyRecord`, and `ApiKeyCall`. Keep fields aligned with the design references: role/status/scopes/quota/usage/last activity.

- [ ] **Step 2: Add mock data**

Add realistic RelayDeck mock records using existing models (`GPT-4o`, `Claude-3.5`, `Gemini Pro`, `Embedding-3-large`) and users (`管理员`, `张晓明`, `李婷婷`, etc.). Include enough rows to make filtering and tables meaningful.

- [ ] **Step 3: Run type check through build**

Run: `npm run build`

Expected: either pass, or fail only for not-yet-created page references that later tasks introduce. Do not leave type errors from the new mock data.

## Task 2: Implement Users Page

**Files:**
- Create: `src/pages/UsersPage.tsx`

- [ ] **Step 1: Build page shell**

Create `UsersPage` with title, subtitle, primary invite button, four `MetricCard`s, filter toolbar using `SearchInput` and `SelectControl`, and `DataTable`.

- [ ] **Step 2: Add prototype interactions**

Use local state for query, role, status, model filter, selected user, drawer visibility, and invite modal visibility. Selecting a row opens the drawer.

- [ ] **Step 3: Add detail drawer and invite modal**

Use existing `Drawer` for details. Implement the invite modal as an in-page fixed overlay using existing tokens and `Button`, matching the existing static prototype style.

- [ ] **Step 4: Verify page compiles in isolation**

Run: `npm run build`

Expected: no TypeScript errors from `UsersPage.tsx`.

## Task 3: Implement API Keys Page

**Files:**
- Create: `src/pages/ApiKeysPage.tsx`

- [ ] **Step 1: Build page shell**

Create `ApiKeysPage` with title, subtitle, create-key button, secondary actions, four `MetricCard`s, filter toolbar, and `DataTable`.

- [ ] **Step 2: Add prototype interactions**

Use local state for query, status, scope, owner, expiry filter, selected key, drawer visibility, and create modal visibility. Selecting a row opens the drawer.

- [ ] **Step 3: Add detail drawer and create modal**

Use existing `Drawer` for details. Implement the create modal as an in-page fixed overlay with form-like controls and scope checkboxes.

- [ ] **Step 4: Verify page compiles in isolation**

Run: `npm run build`

Expected: no TypeScript errors from `ApiKeysPage.tsx`.

## Task 4: Wire Navigation and Status Labels

**Files:**
- Modify: `src/types.ts`
- Modify: `src/App.tsx`
- Modify: `src/components/layout/Sidebar.tsx`
- Modify: `src/lib/format.ts`

- [ ] **Step 1: Extend routing type**

Add `users` and `apiKeys` to `PageId`.

- [ ] **Step 2: Render pages**

Import `UsersPage` and `ApiKeysPage`, add labels to `pageTitles`, and return the pages from `renderPage()`.

- [ ] **Step 3: Add sidebar entries**

Add `用户管理` and `API Keys` entries after `调用测试`, before logs/settings.

- [ ] **Step 4: Add readable status text**

Map new statuses such as `active`, `inactive`, `blocked`, `expired`, and `expiring` in `statusText()`.

## Task 5: Verify, Commit, and Push

**Files:**
- All changed implementation files
- `Design_image/8.png`
- `Design_image/9.png`
- `docs/superpowers/plans/2026-06-10-users-api-keys-ui.md`

- [ ] **Step 1: Build root app**

Run: `npm run build`

Expected: exit code 0.

- [ ] **Step 2: Review git status**

Run: `git status --short`

Expected: only intended implementation files, plan, and new design images are changed/untracked.

- [ ] **Step 3: Commit**

Run: `git add . && git commit -m "Add users and API keys UI"`

Expected: commit succeeds.

- [ ] **Step 4: Push**

Run: `git push`

Expected: branch updates `origin/master` successfully.

## Self-Review

- Spec coverage: user page, API key page, theme-aware styling, navigation, typed mock data, and build verification are covered.
- Placeholder scan: no TBD/TODO/fill-later placeholders are present.
- Type consistency: route IDs are `users` and `apiKeys`; page component names are `UsersPage` and `ApiKeysPage`; mock data names are `userRecords` and `apiKeyRecords`.

