---
name: db-migration
description: Create database migrations for PostgreSQL multi-tenant.
---

- **Format**: `YYYYMMDDHHMMSS_description.sql` with UP and DOWN sections.
- **Tenant Scope**: Include `tenant_id` (mapped physically to `holding_code` or `holding_code`).
- **Source of Truth**: PostgreSQL migrations create or maintain rebuildable projection/relational-processing tables only. Do not introduce PostgreSQL tables as the operational CRUD source unless Jead explicitly approves a projection/rebuild/sync design.
- **Pipeline**: Migration design must support `MongoDB -> Kafka -> PostgreSQL -> ClickHouse`; PostgreSQL tables are populated by sync/projection consumers, not by user-facing CRUD handlers.
- **Safety**: No DROP columns/tables without backups. Index query/join columns. Use transaction wrapper.
- **Testing**: Test against DEV/copy data before applying migrations.
