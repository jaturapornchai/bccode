---
kind: external_dependency
name: ClickHouse BI Analytics Store
slug: clickhouse
category: external_dependency
category_hints:
    - vendor_identity
scope:
    - '**'
---

### ClickHouse
- **Role**: BI/analytics/reporting store fed from processed facts/projections for analytical queries and reporting.
- **Integration**: ch-go client library for high-performance columnar storage operations.
- **Usage Pattern**: Consumes processed facts/projections from the data pipeline; not used for operational CRUD operations.
- **Constraints**: Rebuild/sync target derived from MongoDB/system events; conflicts with MongoDB are resolved in MongoDB's favor.
- **Architecture**: Final stage in MongoDB -> Kafka -> PostgreSQL -> ClickHouse pipeline.