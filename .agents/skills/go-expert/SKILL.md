---
name: go-expert
description: Auto-activate when writing, reviewing, or modifying Go code.
---

## 1. Mandatory Steps
- **Inspect**: Skip full references check for local/internal components; perform definition lookup only for unfamiliar files.
- **Local Rebuild (on request ONLY)**: Backend is batch mode — do NOT auto-rebuild. Rebuild the container (`cd D:\bccode\backend; docker-compose up -d --no-deps --build mainapi`) and verify `/healthz` ONLY when Jead says `rebuild`/`deploy`. See core-rules "Dev Workflow Mode".

## 2. Coding Patterns
- **Context & Errors**: Always pass `ctx`. Return error types, do not `panic` in production.
- **SQL Param**: Parameterize SQL queries (no string concat).
- **Default Data Source**: Unless Jead explicitly asks for a projection, rebuild, sync, relational calculation, migration, or BI/reporting flow, operational APIs must read/write MongoDB first. Do not use PostgreSQL or ClickHouse as the user-facing source for CRUD/list/detail data.
- **Projection Writes/Rebuilds**: Only read/write PostgreSQL or ClickHouse when working on explicit sync, rebuild, consumer, projection, relational calculation, or BI/reporting code. MongoDB remains the operational source of truth. Treat PostgreSQL and ClickHouse rows as rebuildable derived data; if they conflict with MongoDB, fix sync/rebuild instead of patching projections as truth. Legacy projection tables may use columns without underscores (`unitname`, `groupcode`, `groupnames`, `itemtype`); verify the table before writing.
- **Mongo reorder**: MongoDB `$set` updates must target specific reordered fields (`xsorts`, parent/order keys) instead of saving the full document (full document write causes duplicate `_id` errors).
- **Checks**: Run `go vet ./...` and `go build ./...` before committing.
