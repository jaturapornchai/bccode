---
name: db-expert
description: Use when writing SQL, schema design, migration, or query optimization for PostgreSQL multi-tenant
---
When working with the Database:

## Multi-tenant Rules (BC Account)
- `tenant_id` is the canonical tenant boundary for new work.
- Existing core `shopid` values are the tenant id for old data. Some GoAPI/MCP modules expose `shop_id` at the API/DTO layer. Do not create a second unrelated tenant identifier.
- Do not rename/add `tenant_id` columns in existing tables only for naming consistency; keep compatibility mapping from logical `tenant_id` to the physical key used by that module, usually `shopid`.
- Every query must always include a logical `tenant_id` filter. For legacy tables/collections, map that value to `shopid` or module-specific `shop_id` in the same query.
- Every new standalone table should have a `tenant_id` column + useful composite index. Tables that extend legacy modules may keep the existing physical key for consistency.
- Never perform cross-tenant queries accidentally

## PostgreSQL Best Practices
- Index: every WHERE clause and JOIN column must be indexed
- EXPLAIN ANALYZE: run before deploying any new query
- Transaction: wrap operations that must be atomic
- Prepared: always use $1,$2 parameterized queries

## Migration Rules
- Every migration must have both up and down scripts
- Never DROP COLUMN/TABLE without a backup plan
- Test migration on staging before production

## ClickHouse (Analytics)
- Use for reporting and analytics only
- INSERT in batch, not row-by-row
- Use MergeTree engine as primary
- Shared tables must include `company_group_id`, `tenant_id`, and `branch_id`; do not create one ClickHouse table/database per tenant.
- Official BI/report facts should come from PostgreSQL posted fact events, not raw MongoDB changes.

## Anti-patterns
- SELECT * in production
- N+1 query problem
- Missing WHERE clause in UPDATE/DELETE
