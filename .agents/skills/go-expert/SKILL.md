---
name: go-expert
description: Auto-activate when writing, reviewing, or modifying Go code.
---

## 1. Mandatory Steps
- **Inspect**: Skip full references check for local/internal components; perform definition lookup only for unfamiliar files.
- **Local Docker Desktop Deploy (auto after backend edits)**: After backend Go-code edits are complete, deploy the local Docker Desktop container with `cd D:\bccode\backend; .\scripts\deploy-mainapi-fast.ps1` and verify `/healthz` — no need to wait for a `rebuild` command. Use full `docker-compose up -d --no-deps --build mainapi` when Dockerfile, dependencies, runtime assets, compose, config, or image contents changed. DEV server deploy still requires an explicit `deploy dev`. See core-rules "Dev Workflow Mode".

## 2. Coding Patterns
- **Context & Errors**: Always pass `ctx`. Return error types, do not `panic` in production.
- **SQL Param**: Parameterize SQL queries (no string concat).
- **Default Data Source**: Unless Jead explicitly asks for a projection, rebuild, sync, relational calculation, migration, or BI/reporting flow, operational APIs must read/write MongoDB first. Do not use PostgreSQL or ClickHouse as the user-facing source for CRUD/list/detail data.
- **CRUD Event Pipeline**: For operational CRUD writes, implement the path `MongoDB -> Kafka -> PostgreSQL -> ClickHouse`. Handler code writes MongoDB first, then publishes Kafka or records a durable MongoDB outbox event for retry. Projection consumers write PostgreSQL, and BI consumers feed ClickHouse. Do not call `DBClient()` from user-facing CRUD handlers except explicit rebuild/sync/projection code.
- **Projection Writes/Rebuilds**: Only read/write PostgreSQL or ClickHouse when working on explicit sync, rebuild, consumer, projection, relational calculation, or BI/reporting code. MongoDB remains the operational source of truth. Treat PostgreSQL and ClickHouse rows as rebuildable derived data; if they conflict with MongoDB, fix sync/rebuild instead of patching projections as truth. Legacy projection tables may use columns without underscores (`unitname`, `groupcode`, `groupnames`, `itemtype`); verify the table before writing.
- **Accounting Decimal Types**: Do not use Go `float32`/`float64` for money, price, cost, VAT/tax, debit, credit, balance, totals, stock value, decimal quantity, unit price, average cost, exchange rate, or rounding. Preserve API decimal strings, convert only to native database decimal/smallest-unit types, and verify migrations with before/after totals.
- **Mongo reorder**: MongoDB `$set` updates must target specific reordered fields (`xsorts`, parent/order keys) instead of saving the full document (full document write causes duplicate `_id` errors).
- **Checks**: Run `go vet ./...` and `go build ./...` before committing.
