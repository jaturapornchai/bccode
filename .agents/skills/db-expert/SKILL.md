---
name: db-expert
description: Use when writing SQL, schema design, migration, or query optimization for PostgreSQL.
---

## 1. Multi-Tenant Isolation
- **Boundary**: Filter every query by `tenant_id` (physically mapped to `shopid` or `shop_id`).
- **No Leaks**: Never perform cross-tenant queries unless explicitly auditing.
- **Indexes**: Ensure every `WHERE` and `JOIN` column is indexed.

## 2. Postgres & ClickHouse Rules
- **Postgres Role**: Use PostgreSQL for rebuildable relational processing/projections only: postings, balances, stock costing, VAT/tax, AR/AP, GL, strict relational calculations, and integration-ready relational outputs. Do not design PostgreSQL as the direct CRUD source; operational data starts and is verified in MongoDB.
- **Postgres Querying**: Parameterize queries (`$1`, `$2`), wrap in transactions, run `EXPLAIN ANALYZE` for new queries.
- **ClickHouse Role**: Use ClickHouse for rebuildable BI/analytics/reporting facts fed from processed projections. It is not transactional source of truth.
- **ClickHouse Implementation**: Use MergeTree engine. Insert in batch. Track `company_group_id`, `tenant_id`, and `branch_id`.
- **Pipeline Contract**: Derived stores are fed by `MongoDB -> Kafka -> PostgreSQL -> ClickHouse`. PostgreSQL schema/migrations support projection consumers only; ClickHouse schema supports BI/reporting consumers only. Never design a CRUD write path that starts in PostgreSQL or ClickHouse.
- **Projection Conflict Rule**: If PostgreSQL or ClickHouse data differs from MongoDB, MongoDB wins. Repair the sync/rebuild path; do not manually patch derived stores as operational truth.
- **Migrations**: Always provide both UP and DOWN SQL files. Transaction-wrapped.
