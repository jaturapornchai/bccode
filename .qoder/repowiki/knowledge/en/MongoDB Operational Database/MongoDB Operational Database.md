---
kind: external_dependency
name: MongoDB Operational Database
slug: mongodb
category: external_dependency
category_hints:
    - vendor_identity
    - client_constraint
scope:
    - '**'
---

### MongoDB
- **Role**: Primary operational database for all business CRUD operations, master data, transactions, and user-entered data in BC Ai Account ERP system.
- **Integration**: Configured via environment variables (MONGODB_DEV_URI, MONGODB_UAT_URI, MONGODB_PRO_URI) with separate databases per environment (DEV/UAT/PRO).
- **Usage Pattern**: All operational writes go to MongoDB first, then propagate through Kafka to PostgreSQL projections and ClickHouse BI store.
- **Constraints**: Must use separate MongoDB instances/databases per environment; cannot share credentials or data between environments.
- **Deployment**: Native installation (systemd service) outside Docker containers, accessed via host.docker.internal from containers.