# RelayDeck UI Prototype Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first RelayDeck high-fidelity frontend UI prototype under `RelayDeck` with local mock data and global dark/light theme support.

**Architecture:** Create an independent React/Vite/TypeScript/Tailwind app inside `RelayDeck`. Keep business-shaped mock data in `src/data`, reusable console UI in `src/components`, global shell in `src/components/layout`, and each design screen in `src/pages`. Use local React state for authentication, page switching, theme, filters, selected rows, drawers, and simulated loading states.

**Tech Stack:** React, Vite, TypeScript, Tailwind CSS, lucide-react, lightweight SVG/CSS charts.

---

## Current Workspace Notes

`/home/y/manager` is not a valid git repository. The root `.git` directory is empty/read-only, so implementation checkpoints use build verification instead of commits.

## File Structure

Create or modify these files:

- Create `RelayDeck/package.json`: frontend scripts and dependencies.
- Create `RelayDeck/index.html`: Vite HTML entry.
- Create `RelayDeck/vite.config.ts`: React plugin configuration.
- Create `RelayDeck/tsconfig.json`: TypeScript compiler settings.
- Create `RelayDeck/tsconfig.node.json`: Vite config TypeScript settings.
- Create `RelayDeck/tailwind.config.js`: Tailwind content paths and theme extensions.
- Create `RelayDeck/postcss.config.js`: Tailwind and autoprefixer plugins.
- Create `RelayDeck/src/main.tsx`: React mount.
- Create `RelayDeck/src/App.tsx`: top-level state orchestration.
- Create `RelayDeck/src/index.css`: Tailwind layers, CSS variables, global theme classes.
- Create `RelayDeck/src/types.ts`: page, site, model, routing, quota, test result types.
- Create `RelayDeck/src/data/mock.ts`: all mock data.
- Create `RelayDeck/src/lib/format.ts`: currency, percent, latency, class helpers.
- Create `RelayDeck/src/components/layout/Sidebar.tsx`: left navigation and status footer.
- Create `RelayDeck/src/components/layout/Topbar.tsx`: search, refresh cadence, status pill, notifications, theme switch, user menu.
- Create `RelayDeck/src/components/layout/AppLayout.tsx`: authenticated shell.
- Create `RelayDeck/src/components/ui/Logo.tsx`: RelayDeck logo mark and wordmark.
- Create `RelayDeck/src/components/ui/Button.tsx`: shared button variants.
- Create `RelayDeck/src/components/ui/Card.tsx`: shared card container.
- Create `RelayDeck/src/components/ui/MetricCard.tsx`: dashboard metric card.
- Create `RelayDeck/src/components/ui/StatusBadge.tsx`: status and severity labels.
- Create `RelayDeck/src/components/ui/Controls.tsx`: search input, select, toggle, range slider.
- Create `RelayDeck/src/components/ui/Drawer.tsx`: right-side drawer.
- Create `RelayDeck/src/components/ui/DataTable.tsx`: reusable table wrapper.
- Create `RelayDeck/src/components/charts/RingChart.tsx`: SVG donut chart.
- Create `RelayDeck/src/components/charts/LineChart.tsx`: SVG line/area chart.
- Create `RelayDeck/src/components/charts/MiniTrend.tsx`: small sparkline.
- Create `RelayDeck/src/pages/LoginPage.tsx`: login/register screen.
- Create `RelayDeck/src/pages/OverviewPage.tsx`: overview dashboard.
- Create `RelayDeck/src/pages/SitesPage.tsx`: site management with drawer.
- Create `RelayDeck/src/pages/ModelsPage.tsx`: model matrix and detail panel.
- Create `RelayDeck/src/pages/RoutingPage.tsx`: smart routing page.
- Create `RelayDeck/src/pages/CheckinQuotaPage.tsx`: check-in/quota page.
- Create `RelayDeck/src/pages/TestPage.tsx`: invocation testing page.
- Create `RelayDeck/src/pages/EmptyPage.tsx`: simple task logs and settings screens.

## Task 1: Scaffold React/Vite/Tailwind App

**Files:**
- Create: `RelayDeck/package.json`
- Create: `RelayDeck/index.html`
- Create: `RelayDeck/vite.config.ts`
- Create: `RelayDeck/tsconfig.json`
- Create: `RelayDeck/tsconfig.node.json`
- Create: `RelayDeck/tailwind.config.js`
- Create: `RelayDeck/postcss.config.js`
- Create: `RelayDeck/src/main.tsx`
- Create: `RelayDeck/src/App.tsx`
- Create: `RelayDeck/src/index.css`

- [ ] **Step 1: Add project package metadata**

Create `RelayDeck/package.json` with scripts for local development, build, lint-free type checking through Vite build, and preview.

```json
{
  "name": "relaydeck",
  "private": true,
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite --host 0.0.0.0",
    "build": "tsc -b && vite build",
    "preview": "vite preview --host 0.0.0.0"
  },
  "dependencies": {
    "@vitejs/plugin-react": "^4.3.4",
    "lucide-react": "^0.468.0",
    "vite": "^5.4.11",
    "react": "^18.3.1",
    "react-dom": "^18.3.1"
  },
  "devDependencies": {
    "@types/react": "^18.3.12",
    "@types/react-dom": "^18.3.1",
    "autoprefixer": "^10.4.20",
    "postcss": "^8.4.49",
    "tailwindcss": "^3.4.15",
    "typescript": "^5.6.3"
  }
}
```

- [ ] **Step 2: Add Vite and TypeScript config**

Create `index.html`, `vite.config.ts`, `tsconfig.json`, and `tsconfig.node.json` so Vite can serve and build the React app from `src/main.tsx`.

- [ ] **Step 3: Add Tailwind config**

Create `tailwind.config.js` with `darkMode: 'class'`, scan `index.html` and `src/**/*.{ts,tsx}`, and extend colors with CSS variables for app background, surface, border, text, muted text, primary, success, warning, danger, and info.

- [ ] **Step 4: Add temporary app entry**

Create `src/main.tsx`, `src/App.tsx`, and `src/index.css` with a temporary RelayDeck heading and Tailwind imports.

- [ ] **Step 5: Install dependencies**

Run: `npm install`

Expected: dependencies install and `RelayDeck/package-lock.json` is created.

- [ ] **Step 6: Verify scaffold build**

Run: `npm run build`

Expected: TypeScript completes and Vite creates `dist/`.

## Task 2: Add Theme Tokens, Types, Formatting, And Mock Data

**Files:**
- Modify: `RelayDeck/src/index.css`
- Create: `RelayDeck/src/types.ts`
- Create: `RelayDeck/src/lib/format.ts`
- Create: `RelayDeck/src/data/mock.ts`

- [ ] **Step 1: Define global theme CSS variables**

Add dark and light theme variables in `src/index.css`. Use `.dark` on the root element for dark mode. Define base body colors, font smoothing, table scrollbar styling, and utility classes for glass panels, elevated cards, and focus rings.

- [ ] **Step 2: Define TypeScript domain types**

Create `src/types.ts` with explicit union types for page IDs, statuses, site types, capabilities, chart points, and table entities. Include interfaces for `Site`, `ModelInfo`, `AlertItem`, `RoutingCandidate`, `RouteHistory`, `QuotaRecord`, `TestTemplate`, and `TestResultRow`.

- [ ] **Step 3: Add formatting helpers**

Create `src/lib/format.ts` with:

```ts
export function cn(...classes: Array<string | false | null | undefined>): string {
  return classes.filter(Boolean).join(' ');
}

export function formatCurrency(value: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat('en-US').format(value);
}

export function formatPercent(value: number, digits = 1): string {
  return `${value.toFixed(digits)}%`;
}

export function formatLatency(value?: number): string {
  return typeof value === 'number' ? `${value} ms` : '-';
}
```

- [ ] **Step 4: Add mock data**

Create `src/data/mock.ts` with realistic mock arrays for seven sites, six models, overview metrics, alerts, route candidates, route history, quota records, quota distribution, test templates, test result rows, error summary, and capability coverage. Use Chinese labels matching the design images.

- [ ] **Step 5: Verify type build**

Run: `npm run build`

Expected: no TypeScript errors.

## Task 3: Build Shared UI Components

**Files:**
- Create: `RelayDeck/src/components/ui/Logo.tsx`
- Create: `RelayDeck/src/components/ui/Button.tsx`
- Create: `RelayDeck/src/components/ui/Card.tsx`
- Create: `RelayDeck/src/components/ui/MetricCard.tsx`
- Create: `RelayDeck/src/components/ui/StatusBadge.tsx`
- Create: `RelayDeck/src/components/ui/Controls.tsx`
- Create: `RelayDeck/src/components/ui/Drawer.tsx`
- Create: `RelayDeck/src/components/ui/DataTable.tsx`
- Create: `RelayDeck/src/components/charts/RingChart.tsx`
- Create: `RelayDeck/src/components/charts/LineChart.tsx`
- Create: `RelayDeck/src/components/charts/MiniTrend.tsx`

- [ ] **Step 1: Build logo and base buttons**

Implement `Logo` with a CSS gradient lightning mark. Implement `Button` variants for primary, secondary, ghost, danger, and icon.

- [ ] **Step 2: Build cards, metric cards, and badges**

Implement `Card`, `MetricCard`, and `StatusBadge`. Badges must map site/model/test statuses to consistent colors in both themes.

- [ ] **Step 3: Build controls**

Implement `SearchInput`, `SelectControl`, `ToggleSwitch`, and `RangeSlider` in `Controls.tsx`. Use lucide icons for search and chevrons.

- [ ] **Step 4: Build drawer and table wrapper**

Implement `Drawer` with right-side fixed positioning and overlay. Implement `DataTable` as a styled wrapper around normal table markup with overflow handling.

- [ ] **Step 5: Build lightweight charts**

Implement `RingChart`, `LineChart`, and `MiniTrend` using SVG. They must accept data props and use theme-readable colors.

- [ ] **Step 6: Verify component type build**

Run: `npm run build`

Expected: no TypeScript errors.

## Task 4: Build Authenticated Layout

**Files:**
- Create: `RelayDeck/src/components/layout/Sidebar.tsx`
- Create: `RelayDeck/src/components/layout/Topbar.tsx`
- Create: `RelayDeck/src/components/layout/AppLayout.tsx`
- Modify: `RelayDeck/src/App.tsx`

- [ ] **Step 1: Build sidebar**

Add the RelayDeck brand, navigation items for overview, sites, models, routing, check-in, quota, tests, logs, and settings. Include a system status footer and collapse icon.

- [ ] **Step 2: Build topbar**

Add global search, refresh cadence, green running-status pill, notification icon with badge, theme toggle, and user menu.

- [ ] **Step 3: Build app layout shell**

Compose sidebar and topbar in `AppLayout`, with a fixed sidebar, sticky topbar, and scrollable content area.

- [ ] **Step 4: Wire temporary page switching**

Update `App.tsx` with `activePage`, `theme`, and `isAuthenticated` state. Apply `dark` class to the root container based on theme. Render a simple page title inside `AppLayout`.

- [ ] **Step 5: Verify layout build**

Run: `npm run build`

Expected: no TypeScript errors.

## Task 5: Build Login, Overview, Sites, And Models Pages

**Files:**
- Create: `RelayDeck/src/pages/LoginPage.tsx`
- Create: `RelayDeck/src/pages/OverviewPage.tsx`
- Create: `RelayDeck/src/pages/SitesPage.tsx`
- Create: `RelayDeck/src/pages/ModelsPage.tsx`
- Modify: `RelayDeck/src/App.tsx`

- [ ] **Step 1: Build login page**

Implement the design-image login layout with brand, headline, feature orbit, login/register tabs, email/password inputs, remember checkbox, third-party login buttons, and theme toggle. The login button calls `onLogin`.

- [ ] **Step 2: Build overview page**

Implement metric cards, status distribution ring, call trend line chart, alert list, site status table, and model availability cards using mock data.

- [ ] **Step 3: Build sites page**

Implement summary cards, search and filters, site table, and add-site drawer. Search and filters should update visible rows. Test connection should show a temporary loading/result label.

- [ ] **Step 4: Build models page**

Implement model metrics, model table, selected row styling, right-side model detail panel, capabilities, recommended sites, and success trend.

- [ ] **Step 5: Wire pages into App**

Render the correct page for `overview`, `sites`, and `models`. Keep other nav items mapped to the simple page component until their pages are built.

- [ ] **Step 6: Verify first page batch**

Run: `npm run build`

Expected: no TypeScript errors and Vite production bundle succeeds.

## Task 6: Build Routing, Check-In/Quota, Test, And Simple Pages

**Files:**
- Create: `RelayDeck/src/pages/RoutingPage.tsx`
- Create: `RelayDeck/src/pages/CheckinQuotaPage.tsx`
- Create: `RelayDeck/src/pages/TestPage.tsx`
- Create: `RelayDeck/src/pages/EmptyPage.tsx`
- Modify: `RelayDeck/src/App.tsx`

- [ ] **Step 1: Build smart routing page**

Implement rule tabs, routing setting cards, model selector, candidate site table with sliders, score explanation panel, score breakdown bars, and route history.

- [ ] **Step 2: Build check-in/quota page**

Implement metric cards, check-in progress ring, quota trend chart, task log, check-in record table, quota distribution chart, and one-click check-in simulated loading/completion.

- [ ] **Step 3: Build invocation test page**

Implement model/site/test-type selectors, prompt textarea, advanced option toggles, action buttons, result metrics, result table, recent templates, error summary ring, and capability coverage bars. Start test should simulate loading before showing results.

- [ ] **Step 4: Build simple pages**

Implement `EmptyPage` for task logs and system settings, with concise page copy and a visual empty state that matches the console theme.

- [ ] **Step 5: Wire all pages into App**

Map `routing`, `checkin`, `quota`, `testing`, `logs`, and `settings` to their page components. `checkin` and `quota` can share `CheckinQuotaPage` for the first milestone.

- [ ] **Step 6: Verify second page batch**

Run: `npm run build`

Expected: no TypeScript errors and Vite production bundle succeeds.

## Task 7: Polish Responsive Layout, Theme Consistency, And Interaction States

**Files:**
- Modify: `RelayDeck/src/index.css`
- Modify: `RelayDeck/src/components/layout/Sidebar.tsx`
- Modify: `RelayDeck/src/components/layout/Topbar.tsx`
- Modify: page files under `RelayDeck/src/pages/`
- Modify: UI components under `RelayDeck/src/components/ui/`

- [ ] **Step 1: Check theme contrast**

Review every page in dark and light theme. Adjust CSS variables and component classes so backgrounds, text, badges, borders, inputs, charts, and drawers remain readable.

- [ ] **Step 2: Check responsive behavior**

At desktop width, preserve the design image density. At narrower widths, allow content grids to stack, tables to scroll horizontally, and the sidebar to remain usable.

- [ ] **Step 3: Stabilize fixed-format UI**

Set explicit dimensions or min-widths for icon buttons, status badges, charts, table cells, nav items, and metric cards so dynamic text does not resize controls.

- [ ] **Step 4: Verify polished build**

Run: `npm run build`

Expected: no TypeScript errors and production bundle succeeds.

## Task 8: Run Final Verification And Start Dev Server

**Files:**
- No source changes expected unless verification finds issues.

- [ ] **Step 1: Run production build**

Run: `npm run build`

Expected: TypeScript and Vite build complete successfully.

- [ ] **Step 2: Start local dev server**

Run: `npm run dev`

Expected: Vite prints a local URL. Use a free port if the default is occupied.

- [ ] **Step 3: Manual UI smoke check**

Open the dev URL and check:

- login enters console
- sidebar switches pages
- theme toggle works on login and console pages
- site drawer opens and closes
- model row selection updates details
- routing sliders move
- one-click check-in simulates loading
- invocation test simulates loading and results

- [ ] **Step 4: Report result**

Report changed files, build result, and dev server URL.

## Self-Review

- Spec coverage: all approved pages are represented in Tasks 5 and 6; global theme and mock data are represented in Tasks 2, 4, and 7; validation is represented in Task 8.
- Placeholder scan: this plan avoids open-ended implementation gaps and names exact files and commands.
- Type consistency: page IDs, statuses, and data entities are centralized in `src/types.ts`, then consumed by mock data, layout, and pages.
