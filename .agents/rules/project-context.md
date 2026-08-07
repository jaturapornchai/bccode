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
- Primary: `#812920` (terracotta primary; `primary-container` `#a04035`) — see formdesign skill palette
- Accent: HSL tailored colors (sky blue, emerald green, amber yellow)
- Font heading: Outfit
- Font body: Inter
- Radius: token scale only — `--radius-xs..2xl` = 2/3/5/7/10/14px (never hardcode px; see formdesign skill)
- Tone: premium, modern, clean dark/light mode

## Do NOT touch
- `backend/vendor/`
- `frontend/node_modules/`
- `package-lock.json`
- `*.gen.*`, `*.lock`
