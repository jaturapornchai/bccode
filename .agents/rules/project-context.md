# Project Context (CACHED — stable content)

## Domain
BC Ai Account is a modern multilingual multi-tenant ERP and accounting platform built for Thai SMEs. It handles operations, procurement, stock, sales, table zone layouts, kitchen queues, and permissions.

## Folder Map
- `backend/cmd/` — Go entry points
- `backend/internal/` — private Go backend (shop, authentication, product, warehouse, zone, table, kitchen)
- `backend/pkg/` — public API
- `backend/migrations/` — SQL migrations
- `frontend/src/` — Next.js 14 frontend application
- `frontend/src/app/` — app router screens and API routes
- `backend/assets/` — address database, languages TSV

## Conventions
- Error: `fmt.Errorf("context: %w", err)`
- Logging: slog structured logging
- Data roles: MongoDB operational source of truth; R2/S3 images/files; PostgreSQL relational processing/projections; ClickHouse BI/analytics.
- Test: table-driven, 1 happy + 2 error
- Naming: snake_case for files, PascalCase for exports

## Brand (for UI work)
- Primary: `#a04035` (terracotta/brick red)
- Accent: HSL tailored colors (sky blue, emerald green, amber yellow)
- Font heading: Outfit
- Font body: Inter
- Radius: 12px (rounded-xl) and 16px (rounded-2xl)
- Tone: premium, modern, clean dark/light mode

## MCP Tools Available (any agent may use these)
- `<db-mcp>` — query MongoDB, PostgreSQL processing projections, and ClickHouse BI stores
- `<github-mcp>` — manage PRs and issues

## Do NOT touch
- `backend/vendor/`
- `frontend/node_modules/`
- `package-lock.json`
- `*.gen.*`, `*.lock`
