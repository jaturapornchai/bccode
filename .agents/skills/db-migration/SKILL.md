---
name: db-migration
description: Create database migrations for PostgreSQL multi-tenant.
---

- **Format**: `YYYYMMDDHHMMSS_description.sql` with UP and DOWN sections.
- **Tenant Scope**: Company-owned operational/projection tables require `holdingcode + businesscode` in rows, unique keys, foreign keys, joins, and indexes; branch-owned data also requires `branchcode`. Holding aggregate tables/views are read-only and retain `businesscode`. New database identifiers remain lowercase.
- **Source of Truth**: PostgreSQL migrations create or maintain rebuildable projection/relational-processing tables only. Do not introduce PostgreSQL tables as the operational CRUD source unless Jead explicitly approves a projection/rebuild/sync design.
- **Pipeline**: Migration design must support `MongoDB -> Kafka -> PostgreSQL -> ClickHouse`; PostgreSQL tables are populated by sync/projection consumers, not by user-facing CRUD handlers.
- **Safety**: While PRE-LAUNCH the Disposable Database Rule applies — drop/rebuild is allowed without backups; backups become mandatory only once real production data exists at go-live. Index query/join columns. Use transaction wrapper.
- **Testing**: Test against DEV/copy data before applying migrations.
