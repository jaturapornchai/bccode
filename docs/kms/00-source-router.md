# BC Ai Account — Source Router

Start every task at `AGENTS.md`. Business-rule documentation in `docs/` was removed on 2026-09-03 for redesign; its replacement is pending. Use this file only to locate implementation evidence; it defines no system behavior. Ask Jead when a required business decision has no confirmed requirement. Historical code and MongoModel are implementation/design evidence, not approval of missing business rules.

## Routes

- Organization and tenant isolation: `backend/internal/organization/**`, `backend/internal/shop/**`, and the exact workspace/company Frontend path.
- Login, User, Membership, and permissions: `backend/internal/authentication/**`, `backend/internal/shop/**`, and the exact login/user-management Frontend path.
- Infrastructure, databases, Kafka topology, cache, and object storage: `backend/README.md`, `deploy/account/**`, and the exact configuration, producer/consumer, storage, or integration path.
- MongoDB model and technical-workflow alignment: live MongoModel MCP project `BC Ai Account`; use `docs/skills/audit-mongomodel-sync/SKILL.md` only for the audit procedure, then inspect the exact source/tests/runtime path.
- Product, Barcode, Unit, and media: `backend/internal/product/**`, Product/Barcode handlers, and the exact Product/Barcode Frontend path.
- Sales, purchase, and stock documents: the matching `backend/internal/transaction/**` parser/service and Kafka handler.
- Inventory costing and stock movement: `backend/internal/goapi/process/process-stock/**`, relevant Kafka handlers, stock-process code, migrations, and focused reconciliation tests.
- API, outbox, retry, and idempotency: the exact caller, handler, outbox/consumer/config path, and contract tests.
- Record lifecycle changes: exact source/tests/runtime; schema changes: live MongoModel MCP → exact model, repository, and target runtime.
- UAT or release verification: `tools/verify.sh` (the local runner that replaced GitHub CI on 2026-09-09), `backend/.ci/test-quarantine.txt`, and affected technical workflows from live MongoModel MCP → the narrowest relevant browser, API, database, and event evidence.
- MongoModel MCP maintenance: `D:\mongomodel\AGENTS.md` → exact MongoModel source, tests, build, container, and MCP runtime evidence.

The missing-docs rule in `AGENTS.md` takes precedence over stale docs links in supporting instructions. Compile-only or quarantined tests do not establish business correctness. Inspect runtime configuration without exposing secrets.
