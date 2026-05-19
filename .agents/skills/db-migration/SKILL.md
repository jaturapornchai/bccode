---
name: db-migration
description: >
  Create database migrations for PostgreSQL multi-tenant.
  Use when: adding tables, modifying schema, creating migrations, changing database.
  Trigger: "migration", "เพิ่ม column", "สร้างตาราง", "alter table", "add column", "create table"
---

When creating a migration:

1. Name format: YYYYMMDDHHMMSS_description.sql
2. Every migration must have UP and DOWN
3. Account for multi-tenant data: `tenant_id` is canonical at the API/service boundary; existing `shop_id` / `shopid` values are the tenant id for old data
4. Never DROP COLUMN without a backup plan
5. Add indexes for columns used in WHERE clauses
6. Always use a transaction wrapper
7. Test against a copy of production data before running for real
8. Every new standalone table should have a `tenant_id` column + useful composite index. Do not backfill/rename existing `shop_id` / `shopid` columns only for naming consistency.
