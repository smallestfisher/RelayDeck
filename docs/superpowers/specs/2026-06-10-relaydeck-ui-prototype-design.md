# RelayDeck UI Prototype Design

Date: 2026-06-10

## Context

RelayDeck is a unified control platform for managing many upstream large-model sites based on systems such as new-api and sub2api. The first milestone is a high-fidelity frontend UI prototype based on the design images in `RelayDeck/Design_image`.

Reference projects:

- `Reference/new-api`: reference for domain concepts and a React-based modern frontend direction.
- `Reference/sub2api`: reference for site/subscription management concepts and a Vue/Vite frontend.

The actual new work lives under `RelayDeck`.

## Goals

- Build a high-fidelity frontend-only prototype.
- Use local mock data only; do not connect backend APIs in this milestone.
- Implement the pages shown by the design images.
- Support both dark and light themes across the whole app.
- Keep the frontend structure easy to replace with real API data later.

## Non-Goals

- No backend implementation.
- No real authentication, registration, OAuth, SSO, or permission system.
- No persistent database-backed changes.
- No real upstream site testing, routing, quota refresh, or check-in behavior.
- No production-grade charting or table virtualization in the first milestone unless required by layout quality.

## Technology

- React
- Vite
- TypeScript
- Tailwind CSS
- lucide-react for interface icons

The prototype should be scaffolded as an independent frontend app inside `RelayDeck`.

## Architecture

The app will use local state for navigation, theme, and lightweight interactions. It does not need route URLs for the first milestone, so page switching can be driven by an `activePage` state. This keeps the prototype simple while preserving clear page boundaries. `react-router-dom` can be added later when backend integration or shareable URLs become necessary.

Suggested structure:

```text
RelayDeck/
  package.json
  index.html
  vite.config.ts
  tsconfig.json
  tailwind.config.js
  postcss.config.js
  src/
    App.tsx
    main.tsx
    index.css
    components/
      layout/
      ui/
      charts/
      data/
    data/
      mock.ts
    lib/
      format.ts
      theme.ts
    pages/
      LoginPage.tsx
      OverviewPage.tsx
      SitesPage.tsx
      ModelsPage.tsx
      RoutingPage.tsx
      CheckinQuotaPage.tsx
      TestPage.tsx
      PlaceholderPage.tsx
```

## Theme Design

The whole platform supports dark and light themes. Pages are not fixed to one theme. The design images provide examples for both visual modes:

- Dark theme reference: overview, model management, smart routing, invocation testing.
- Light theme reference: site management, check-in/quota management.

Implementation requirements:

- Theme state is global.
- Theme choice is stored in `localStorage`.
- Login page and authenticated shell both respond to the same theme setting.
- Styling uses shared theme tokens through Tailwind classes and CSS variables.
- Core surfaces, tables, drawers, inputs, charts, and status badges must remain readable in both themes.

## Pages

### Login

The login page contains:

- RelayDeck logo and subtitle.
- Product value headline.
- Visual feature orbit or capability cluster.
- Login/register tabs.
- Email and password fields.
- Remember-me checkbox and forgot-password link.
- Google, GitHub, and enterprise SSO buttons.
- Theme switch.

The login button transitions into the management console. Registration is visual only.

### Overview

The overview page contains:

- Key metric cards: site count, model count, daily calls, daily cost, pending check-ins.
- Site status distribution ring chart.
- Seven-day call trend chart.
- Exception reminder list.
- Site status table.
- Model availability cards.
- Refresh control with a mock updated timestamp.

### Site Management

The site management page contains:

- Summary metric cards.
- Search input and filters for status, type, and latency.
- Batch action buttons.
- Site table with status, type, available model count, average latency, balance, and check-in state.
- Add-site drawer with fields for name, URL, type, API key, optional cookie, and notes.
- Test connection button with loading/success/failure visual states.

Table actions are visual in this milestone and do not need persistent data mutation.

### Model Management

The model management page contains:

- Summary metric cards.
- Model matrix table with model name, available site count, recommended site, minimum latency, seven-day success rate, quota condition, routing mode, and status.
- Selected model detail panel with recommended sites, capabilities, success trend, and metadata.

Clicking a model row updates the detail panel.

### Smart Routing

The smart routing page contains:

- Tabs for global routing rules, model routing rules, routing logs, and routing analysis.
- Routing setting cards for routing mode, health score threshold, circuit-break threshold, cooldown time, and minimum candidate count.
- Model selector.
- Candidate site table with manual weight sliders, health score, success rate, latency, load, circuit state, and actions.
- Right-side explanation panel showing why a site was selected, weighted score breakdown, and recent routing history.

Sliders are interactive, but the score does not need to be recalculated in the first milestone.

### Check-In And Quota

The check-in/quota page contains:

- Summary metric cards for checked-in sites, unchecked sites, remaining quota, and today's added quota.
- Check-in progress ring chart.
- Seven-day quota trend chart.
- Task log.
- Check-in record table.
- Site quota distribution chart.
- Refresh, filter abnormal, and one-click check-in buttons.

One-click check-in uses a simulated loading and completion state.

### Invocation Test

The invocation testing page contains:

- Test configuration form for model, site scope, test type, and prompt.
- Advanced options toggles for advanced parameters, streaming, and logs.
- Start test, concurrent test, and save template buttons.
- Result metric cards.
- Result details table.
- Recent test templates.
- Error summary chart.
- Capability coverage panel.

Start test uses a simulated loading state before showing mock results.

### Placeholder Pages

Task logs and system settings remain in the sidebar but render simple placeholder pages in the first milestone.

## Components

Reusable components should be built around the management-console surface:

- `Sidebar`
- `Topbar`
- `PageHeader`
- `MetricCard`
- `StatusBadge`
- `DataTable`
- `SearchInput`
- `SelectControl`
- `ToggleSwitch`
- `RangeSlider`
- `ActionButton`
- `Drawer`
- `RingChart`
- `LineChart`
- `MiniTrend`
- `AlertList`
- `EmptyState`

Charts can be implemented with lightweight SVG/CSS for the prototype. A full charting library is not required unless the SVG approach becomes too limiting during implementation.

## Mock Data

Mock data should be organized by business entity:

- Sites: name, URL, type, region, status, model count, latency, balance, check-in status, notes.
- Models: ID, display name, provider, status, availability, recommended site, success rate, quota condition, capabilities.
- Routing: rules, candidate scores, route history, score breakdown.
- Alerts: severity, title, description, timestamp.
- Quotas: check-in records, remaining quota, added quota, distribution.
- Tests: templates, result metrics, per-site result rows, error categories, capability coverage.

The shape should be close enough to future API data that the mock layer can later be replaced by fetch/query functions without rewriting pages.

## Interaction And State

The prototype includes these interactions:

- Login enters the console.
- Sidebar changes active page.
- Theme toggle applies globally and persists.
- Search and filters update visible table rows where practical.
- Add-site drawer opens and closes.
- Test connection shows temporary loading/result state.
- Model row selection updates model details.
- Routing sliders move visually.
- Refresh buttons update timestamps or show a brief loading state.
- One-click check-in simulates progress and completion.
- Invocation test simulates a run and displays mock results.

## Validation

Verification should include:

- `npm run build`
- Manual inspection in the browser for dark and light themes.
- Responsive checks for desktop and narrower widths.
- If practical, Playwright screenshots for key pages after implementation.

The first milestone is complete when the UI prototype can be run locally and the implemented screens visually match the design direction with coherent behavior in both themes.
