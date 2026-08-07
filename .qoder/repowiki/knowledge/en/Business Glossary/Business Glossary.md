---
kind: business_term
name: Business Glossary
category: business_term
scope:
    - '**'
---

### BC Ai Account
- Definition：The canonical product brand name for this Thai SME accounting and business management ERP system. Used consistently across all user-facing text and documentation. Technical identifiers like bc-account, bc_account, and bcaccount remain unchanged unless explicitly migrated.
- Aliases：bc-account、bc_account、bcaccount

### Holding
- Definition：A group of companies that share master data and administrative settings in the BC Ai Account system. Holdings own shared records like products, customers, debtors, and creditors so companies don't duplicate setup and enable Holding-wide analysis. Access scopes and permissions are configured at the Holding level.
- Aliases：holdingcode

### guidfixed
- Definition：An immutable technical identifier generated exactly once when a MongoDB record is first created and never changed afterward. Used for CRUD operations, audit trails, and event correlation, but must never be used as a business relation key or cross-store join key. Business relations use normalized codes instead.
- Aliases：guid

### businesscode
- Definition：A user-facing business reference code that serves as the primary identifier for business relationships between records. Must be uppercase, space-free, and unique within its scope. Used for all cross-record relations instead of internal GUIDs.
- Aliases：businesscodes

### CRUD workbench
- Definition：A shared UI pattern for data management screens featuring a left-side list/search/filter panel and right-side detail/edit form. Provides consistent behavior for create, read, update, delete operations with common save/delete/confirm/dirty-state handling across all ERP screens.
- Aliases：workbench

### disposable database
- Definition：A pre-launch policy allowing complete database reset and rebuild without migration concerns. During pre-launch phase, all data and structures in MongoDB, PostgreSQL, ClickHouse, and Kafka are considered disposable - agents can drop/recreate collections, tables, and topics directly without backward compatibility scaffolding.
- Aliases：no-migration rule

### operational CRUD pipeline
- Definition：The mandatory data flow pattern: MongoDB -> Kafka -> PostgreSQL -> ClickHouse. Every operational write must follow this sequence, with MongoDB as the authoritative source and downstream stores rebuilt from events rather than direct writes.
- Aliases：data pipeline

### product category groups
- Definition：Usage/device/channel grouping system where groupnumber 1-20 represents different usage contexts (group 1 = ordering/tablet, group 2 = cashier/POS, group 3 = kitchen/KDS, group 4 = delivery). Each group contains a complete category tree with parent-child relationships and sellable leaf categories.
- Aliases：category groups
