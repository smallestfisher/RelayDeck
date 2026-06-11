# RelayDeck Agent Notes

## Working Root

All project code, plans, specs, and implementation status live under:

```text
/home/y/manager/RelayDeck
```

Do not treat `/home/y/manager/Reference` as part of the RelayDeck worktree. It is reference material only and should be read only when a task explicitly needs comparison with upstream projects or external client behavior.

Relevant reference directories:

- `/home/y/manager/Reference/new-api`: upstream relay platform source used to inspect API endpoints, account/channel behavior, model discovery, protocol support, and admin integration details that RelayDeck may need to connect to.
- `/home/y/manager/Reference/sub2api`: upstream relay platform source used to inspect API endpoints, account/channel management, model mapping, monitor/sync behavior, and platform-specific integration details.
- `/home/y/manager/Reference/codex`: client source/reference used to understand Codex-style request behavior so RelayDeck can simulate client traffic toward upstreams when needed.
- `/home/y/manager/Reference/claude-code`: client source/reference used to understand Claude Code-style request behavior, headers, auth conventions, streaming expectations, and request shape.

The product intent is for RelayDeck to aggregate and route across upstream platforms while making upstream services see traffic shaped like the original client where appropriate. This helps hide the aggregation relay layer from upstream platforms and keeps protocol/client compatibility explicit.

## Directory Map

```text
RelayDeck/
├── backend/                 # Go backend service
│   ├── cmd/relaydeck/       # backend entrypoint
│   ├── migrations/          # PostgreSQL schema migrations
│   └── internal/
│       ├── app/             # application wiring: stores, sessions, routers
│       ├── auth/            # password hashing, API key auth, session stores
│       ├── config/          # environment and runtime config
│       ├── domain/          # shared domain types
│       ├── http/
│       │   ├── admin/       # admin API handlers
│       │   ├── gateway/     # OpenAI-compatible gateway handlers
│       │   └── middleware/  # HTTP middleware
│       ├── router/          # upstream route selection logic
│       ├── store/           # memory/admin store abstractions
│       │   └── postgres/    # PostgreSQL-backed stores
│       └── upstream/        # upstream HTTP client and sync/integration code
├── src/                     # React/Vite admin console
│   ├── components/
│   │   ├── charts/          # chart widgets
│   │   ├── layout/          # sidebar, topbar, app shell
│   │   └── ui/              # reusable UI controls
│   ├── data/                # temporary mock data for prototype screens
│   ├── lib/                 # frontend utility functions
│   ├── pages/               # page-level admin screens
│   ├── App.tsx              # frontend routing/state shell
│   ├── main.tsx             # React entrypoint
│   ├── types.ts             # frontend shared types
│   └── index.css            # Tailwind/global styles
├── docs/superpowers/
│   ├── specs/               # approved design specs
│   └── plans/               # implementation plans and task status
├── Design_image/            # UI prototype screenshots; 5.PNG is station/site management
├── package.json             # frontend scripts and dependencies
├── vite.config.ts           # Vite config
├── tailwind.config.js       # Tailwind config
├── docker-compose.yml       # local service dependencies
└── README.md                # project overview
```

## Generated Or External Directories

These directories exist locally but are not project source for feature planning:

```text
RelayDeck/.git/          # git metadata
RelayDeck/node_modules/  # installed frontend dependencies
RelayDeck/dist/          # frontend build output
RelayDeck/.claude/       # local agent/tooling metadata
```

## Current Planning Context

- Existing specs are in `docs/superpowers/specs/`.
- Existing implementation plans and checklist state are in `docs/superpowers/plans/`.
- The current platform direction is an aggregation and distribution management platform for upstream accounts, models, routing, health, quota, and admin operations.
- The site management page corresponds to `Design_image/5.PNG`.
- Existing `src/pages/SitesPage.tsx` is primarily a mock-data UI prototype and should be connected to backend admin APIs as the backend contract is implemented.
