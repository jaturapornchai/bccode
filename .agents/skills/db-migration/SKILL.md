---
name: db-migration
description: Create database migrations for PostgreSQL multi-tenant.
---

- **Format**: `YYYYMMDDHHMMSS_description.sql` with UP and DOWN sections.
- **Tenant Scope**: Include `tenant_id` (mapped physically to `shopid` or `shop_id`).
- **Safety**: No DROP columns/tables without backups. Index query/join columns. Use transaction wrapper.
- **Testing**: Test against DEV/copy data before applying migrations.
