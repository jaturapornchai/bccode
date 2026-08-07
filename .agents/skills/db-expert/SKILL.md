---
name: db-expert
description: Use when writing SQL, schema design, migration, or query optimization for PostgreSQL.
---

## 1. Multi-Tenant Isolation
- **Boundary**: Filter every company-owned operational query by validated `holdingcode + businesscode`; branch-owned data also includes `branchcode`.
- **No Leaks**: Cross-company queries are allowed only through authorized read-only Holding aggregation and must retain/group by `businesscode`. Never merge equal business keys from different companies.
- **Indexes**: Ensure every `WHERE` and `JOIN` column is indexed.

## 2. Postgres & ClickHouse Rules
- **Postgres Role**: Use PostgreSQL for rebuildable relational processing/projections only: postings, balances, stock costing, VAT/tax, AR/AP, GL, strict relational calculations, and integration-ready relational outputs. Do not design PostgreSQL as the direct CRUD source; operational data starts and is verified in MongoDB.
- **Accounting Decimal Rule**: For money, price, cost, discount, VAT/tax, debit, credit, balance, totals, stock value, decimal quantity, unit price, average cost, exchange rate, and rounding, never use PostgreSQL `real`/`double precision`/`float`/`money` or ClickHouse `Float32`/`Float64`. Use PostgreSQL `numeric(P,S)` or `bigint` smallest units, and ClickHouse `Decimal(P,S)` or `Int64` smallest units. Default precision is money `18,2`, unit price `20,4`, quantity `18,4`, exchange rate `28,8`; project field names remain lowercase (underscore allowed) such as `netamount`, `unitprice`, `exchangerate`, and `amountsatang`.
- **Postgres Querying**: Parameterize queries (`$1`, `$2`), wrap in transactions, run `EXPLAIN ANALYZE` for new queries.
- **ClickHouse Role**: Use ClickHouse for rebuildable BI/analytics/reporting facts fed from processed projections. It is not transactional source of truth.
- **ClickHouse Implementation**: Use MergeTree engine. Insert in batch. Track `companygroupid`, `tenantid`, and `branchid`.
- **Pipeline Contract**: Derived stores are fed by `MongoDB -> Kafka -> PostgreSQL -> ClickHouse`. PostgreSQL schema/migrations support projection consumers only; ClickHouse schema supports BI/reporting consumers only. Never design a CRUD write path that starts in PostgreSQL or ClickHouse.
- **Projection Conflict Rule**: If PostgreSQL or ClickHouse data differs from MongoDB, MongoDB wins. Repair the sync/rebuild path; do not manually patch derived stores as operational truth.
- **Portable Relationship Keys**: PostgreSQL/ClickHouse may carry Mongo `guidfixed` for traceability, but company-owned foreign keys, joins, and relationships include `businesscode` with normalized keys such as `code`, `itemcode`, `unitcode`, or document number. Never use GUID or a business code alone across companies.
- **Rebuild Driving Set**: Rebuild projections from active same-company MongoDB rows, retain `businesscode`, and remove projection-only rows. Product projection is driven by MongoDB `products`; `productbarcodes` never creates Product masters because every Barcode is linked at creation.
- **Timezone = UTC+0 (set 2026-06-22)**: Store every timestamp/datetime column in MongoDB, PostgreSQL, and ClickHouse in **UTC+0** (PostgreSQL `timestamptz` / store UTC, ClickHouse `DateTime`/`DateTime64` in UTC). Never store branch/local time in the DB; the frontend converts UTC → the active branch `timezone` for display. See core-rules "Timezone Iron Rule".
- **Migrations**: Always provide both UP and DOWN SQL files. Transaction-wrapped.
