---
name: go-expert
description: Auto-activate when writing, reviewing, or modifying Go code.
---

## 1. Mandatory Steps
- **Inspect**: Skip full references check for local/internal components; perform definition lookup only for unfamiliar files.
- **Local Rebuild**: Rebuild backend container (`cd D:\bccode\backend; docker-compose up -d --no-deps --build mainapi`) only when functional changes are completed or explicitly requested.

## 2. Coding Patterns
- **Context & Errors**: Always pass `ctx`. Return error types, do not `panic` in production.
- **SQL Param**: Parameterize SQL queries (no string concat).
- **PostgreSQL Projection Writes**: Only write PostgreSQL tables when working on explicit sync/consumer/projection code. MongoDB remains the operational source of truth. Legacy projection tables may use columns without underscores (`unitname`, `groupcode`, `groupnames`, `itemtype`); verify the table before writing.
- **Mongo reorder**: MongoDB `$set` updates must target specific reordered fields (`xsorts`, parent/order keys) instead of saving the full document (full document write causes duplicate `_id` errors).
- **Checks**: Run `go vet ./...` and `go build ./...` before committing.
