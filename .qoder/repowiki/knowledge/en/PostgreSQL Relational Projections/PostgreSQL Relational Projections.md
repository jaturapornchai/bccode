---
kind: external_dependency
name: PostgreSQL Relational Projections
slug: postgresql
category: external_dependency
category_hints:
    - vendor_identity
    - client_constraint
scope:
    - '**'
---

### PostgreSQL
- **Role**: Relational processing and projection engine for postings, balances, tax/VAT, AR/AP, GL, strict relational calculations, and integration-ready relational outputs.
- **Integration**: GORM ORM with lib/pq driver; rebuildable projections derived from MongoDB/system events.
- **Usage Pattern**: Not operational truth - PostgreSQL is a processing/projection target derived from MongoDB events via Kafka consumers.
- **Constraints**: Must be native installation (systemd service) outside Docker; not used for operational CRUD writes.
- **Architecture**: Part of MongoDB -> Kafka -> PostgreSQL -> ClickHouse pipeline for data flow.