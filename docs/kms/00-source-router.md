# BC Ai Account — Source Router

Start every task at `AGENTS.md`. `docs/kms/` describes what the code does, not business rules; Jead's own specifications live in `mydocs/` (read-only for AI — e.g. `mydocs/specs/gl/spec.md`, SQL in `mydocs/datamodels/gl/`). Use this file only to locate implementation evidence; it defines no system behavior. Ask Jead when a required business decision has no confirmed requirement.

The system runs on PostgreSQL only (+ MinIO/S3 for files). MongoDB, Kafka, Redis, ClickHouse, OpenSearch and the old k8s/`cmd/*` microservices were removed on 2026-09-23 ([ADR](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md)) — never route work back to them.

## Routes

- Organization and tenant isolation: `backend/internal/organization/**`, `backend/internal/shop/**`, central DB schema `backend/internal/centraldb/**`, and the exact workspace/holding/company Frontend path (`frontend/src/app/workspace/**`, `frontend/src/app/holding/**`).
- Login, session, Membership, and permissions: `backend/internal/authentication/**`, `backend/internal/shop/**`, session store `backend/pkg/microservice/cacher.go` (table `cache_entries`), live permission check `backend/pkg/microservice/live_authorization.go`, and the exact login/system-settings Frontend path.
- General Ledger (active module): `backend/internal/generalledger/**` (HTTP `httpapi/`), `frontend/src/app/gl/**`; spec `mydocs/specs/gl/spec.md`.
- Tax reports, RD forms, 50 ทวิ, RD media file: `backend/internal/goapi/handlers/tax_*.go`, `backend/internal/goapi/handlers/wht_certificate.go`, `backend/internal/rdform/**`, `backend/internal/whtcert/**`, `backend/internal/rdfile/**`, `backend/internal/taxaddress/**`, `frontend/src/app/tax/**`; official references in `21-thai-tax-form-references.md`.
- Fixed assets and depreciation: `backend/internal/fixedasset/**`, `frontend/src/app/asset/fixed-assets-screen.tsx`.
- API/MCP tokens and the GL MCP server: `backend/internal/mcptoken/**`, `backend/internal/mcpgateway/**`, `backend/internal/generalledger/httpapi/mcp.go`, `backend/internal/fixedasset/mcp/**`, `frontend/src/app/mcp-tokens/**`.
- Infrastructure, databases, and object storage: `backend/README.md`, `deploy/account/**`, `tools/fast-deploy.py`, file persister `backend/pkg/microservice/persister_file_r2.go`, and the exact configuration path.
- Screens whose backend was removed (product, sales/purchase, stock, AR/AP masters, bank …): `frontend/src/lib/menu-screen-status.ts` (`isMenuBackendRetired`), `frontend/src/lib/erp-transaction.ts`; leftover goapi stock/cost code in `backend/internal/goapi/process/stockengine/**`, `backend/internal/goapi/process/process-stock/**`, `backend/internal/goapi/inventory/**`. Rebuilding one means new PostgreSQL schema + API — ask Jead first.
- Menu and screen text: `frontend/src/lib/menu-data.ts`, `backend/assets/language/languages.tsv`.
- API, retry, and idempotency: the exact caller, handler, request-id/lock path (e.g. `backend/internal/generalledger/models.go`, `postgres_source.go`), and contract tests.
- UAT or release verification: `tools/verify.sh` (the local runner that replaced GitHub CI on 2026-09-09), `backend/.ci/test-quarantine.txt`, PostgreSQL checks via `tests/support/pg.ts` / `psql` → the narrowest relevant browser, API, and database evidence.

Compile-only or quarantined tests do not establish business correctness. Inspect runtime configuration without exposing secrets.
