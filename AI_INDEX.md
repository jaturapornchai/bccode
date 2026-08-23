# BC Ai Account — Source Router

Start every task at `docs/README.md`, then read the matching Jead-authored files under `docs/**`. Use this file only to locate implementation evidence; it defines no system behavior.

## Routes

- Organization and tenant isolation: `docs/organization.md`, `docs/login.md` → `backend/internal/organization/**`, `backend/internal/shop/**`, and the exact workspace/company Frontend path.
- Login, User, Membership, and permissions: `docs/login.md`, `docs/organization.md` → `backend/internal/authentication/**`, `backend/internal/shop/**`, and the exact login/user-management Frontend path.
- Infrastructure, databases, Kafka topology, cache, and object storage: `docs/system.md` plus any matching Domain file → the exact configuration, producer/consumer, storage, or integration path.
- MongoDB model and technical-workflow alignment: `docs/README.md` and matching files under `docs/**` → live MongoModel MCP project `BC Ai Account`; use `.agents/skills/audit-mongomodel-sync/SKILL.md` only for the audit procedure, then inspect the exact source/tests/runtime path.
- Product, Barcode, Unit, and media: matching files under `docs/**` → `backend/internal/product/**`, Product/Barcode handlers, and the exact Product/Barcode Frontend path.
- Sales, purchase, and stock documents: matching files under `docs/**` → the matching `backend/internal/transaction/**` parser/service and Kafka handler.
- Inventory costing and stock movement: matching files under `docs/**` → `backend/internal/goapi/process/process-stock/**`, relevant Kafka handlers, stock-process code, migrations, and focused reconciliation tests.
- API, outbox, retry, and idempotency: `docs/system.md` plus any more specific matching file → the exact caller, handler, outbox/consumer/config path, and contract tests.
- Record lifecycle changes: matching files under `docs/**` → exact source/tests/runtime; schema changes: live MongoModel MCP → exact model, repository, and target runtime.
- UAT or release verification: applicable files under `docs/**` and affected technical workflows from live MongoModel MCP → the narrowest relevant browser, API, database, and event evidence.
- MongoModel MCP maintenance: `docs/README.md`, then `D:\mongomodel\AGENTS.md` → exact MongoModel source, tests, build, container, and MCP runtime evidence.
