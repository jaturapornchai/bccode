# Database Naming Reset Audit

Date: 2026-06-06

## Rule

> **UPDATED 2026-06-21 by Jead:** the rule now allows underscore. New or changed MongoDB, PostgreSQL, and ClickHouse contracts must use **lowercase** names — digits and underscore `_` are allowed (snake_case OK); only uppercase/camelCase/PascalCase is a violation. The historical entries below describe the earlier "no underscore" reset and are kept for record. See `.agents/rules/bc-account-core-rules.md` "Lowercase Database Naming Iron Rule (underscore allowed)".

New or changed MongoDB, PostgreSQL, and ClickHouse contracts must use lowercase names (underscore allowed).

This includes database-facing function names, MCP tool/action names, API route/query/body keys tied to database fields, variables/constants that represent database identifiers, table names, collection names, column names, field names, index names, SQL aliases, Kafka event keys, and persisted/API contract strings.

Examples: `holdingcode`, `guidfixed`, `businesscode`, `branchcode`, `businesscodes`, `unitcode`, `createdat`, `updatedat`.

## Source Audit

Read-only scan over `backend`, `frontend`, and `.agents` found about 3,313 underscore-style database/API contract hits across 278 files.

Additional read-only pattern audit after the rule was expanded to function/MCP/API/variable names:

| Hits | Pattern |
|---:|---|
| 23 | Go function names containing `_` |
| 4 | Go variable names containing `_` |
| 93 | Go constant names containing `_` |
| 16 | API route strings containing `_` |
| 57 | MCP/tool related lines containing underscore-style names |
| 3,172 | JSON tags with underscore-style names |
| 1,075 | BSON tags with underscore-style names |
| 239 | GORM column tags with underscore-style names |

Most affected areas:

| Count | Area |
|---:|---|
| 36 | `backend\internal\transaction\models` |
| 30 | `backend\internal\goapi\mcp\tools` |
| 23 | `backend\internal\goapi\handlers` |
| 19 | `backend\internal\goapi\models` |
| 15 | `backend\internal\goapi\handlers\aichat` |
| 7 | `backend\internal\models` |
| 5 | `backend\internal\product\productbarcode\models` |
| 4 | `backend\internal\goapi\aiprovider` |
| 3 | `backend\internal\organization\branch\models` |
| 3 | `backend\internal\goapi\ragflow` |

## Reset Guidance

- Do not perform broad blind replacement. Some underscore JSON keys belong to external APIs such as LINE/RAGFlow and are not database contracts.
- For project database contracts, update database-facing functions, MCP tool/action names, API route/query/body keys, variables/constants, MongoDB `bson` tags, PostgreSQL/ClickHouse table and column names, indexes, Kafka event keys, frontend API payload mapping, and query/filter keys together.
- If Jead approves destructive reset, old data can be dropped and rebuilt instead of dual-read migration, but the exact environment and stores must be explicit before running any destructive command.
- If data must be preserved, use a staged rename with backfill, projection rebuild, index rebuild, dual-read/dual-write during migration, and rollback notes.
- **Update 2026-06-28:** per the *No-Migration / Disposable Database Rule* (`bc-account-core-rules.md`), while PRE-LAUNCH every store (Mongo/PG/ClickHouse/Kafka) in ALL environments incl. production is disposable — drop+rebuild from code, no preserve/backfill/dual-read. The "preserve data → staged migration" path above applies only once real production data exists at go-live.

## DEV Reset Executed

Date: 2026-06-06

Jead explicitly confirmed: `ยืนยันลบ DEV ทั้งหมด: MongoDB, PostgreSQL, ClickHouse`.

Reset result:

| Store | DEV reset action | Post-reset verification |
|---|---|---|
| MongoDB | Dropped configured DEV operational database | 0 collections |
| PostgreSQL | Reset non-system schemas in 28 application databases | 0 tables |
| ClickHouse | Dropped all objects in configured DEV database, including dictionary dependencies | 0 tables |

Next rebuild work must create only lowercase database contracts (underscore allowed, per the 2026-06-21 rule update above).

## Code Migration Executed

Date: 2026-06-06

Applied lowercase no-underscore migration across project-controlled database/API contracts:

| Area | Result |
|---|---|
| Core tenant/identity keys | Renamed `holdingcode`, `guidfixed`, `businesscode`, `businesscodes`, `branchcode`, `createdat`, `updatedat`, `deletedat`, and `deletedby` to no-underscore forms such as `holdingcode`, `guidfixed`, `businesscode`, `businesscodes`, `branchcode`, `createdat`, `updatedat`, `deletedat`, and `deletedby`. |
| MCP contracts | Renamed MCP tool names, sandbox helper names, parameters, JSON/BSON tags, and examples to no-underscore names such as `executepython`, `querymongo`, `querypg`, `querych`, `resultvalue`, `getlowstockalerts`, and `querypostgresql`. |
| AI-chat internal contracts | Renamed internal request/response keys and SQL aliases, while keeping OpenAI-compatible fields inside OpenAI adapter files unchanged. |
| Knowledge Base metadata | Renamed MongoDB/API metadata keys to `holdingcode`, `ragflowdocid`, `docid`, `docname`, `datasetid`, `topk`, `chunkcount`, and related no-underscore names. RAGFlow SDK payload names stay isolated in the adapter. |
| Struct tags | Converted project-controlled `json`, `bson`, `gorm column`, and `ch` tags to lowercase no-underscore names. |

Post-migration scan:

| Scan | Result |
|---|---|
| Core legacy names (`holdingcode`, `guidfixed`, old MCP/query helper names) | 0 hits |
| Project-controlled struct tags, excluding external adapter boundaries | 0 hits |
| Remaining underscore struct tags in whole repo | 53 hits, all in external adapter boundary files for OpenAI-compatible, RAGFlow, Gemini, or BCProxy payloads |

Verification run:

| Command | Result |
|---|---|
| `go test ./internal/goapi/mcp/... ./internal/goapi/handlers/aichat ./internal/goapi/handlers/knowledgebase ./internal/goapi/ragflow` | Passed |
| `go test ./internal/goapi/aiprovider ./cmd/branchcode_audit ./cmd/dbtest` | Passed |
| `npm run typecheck` in `frontend` | Passed |
| `git diff --check` | Passed with line-ending warnings only |

Known verification limits:

- `go test ./internal/...` is still not a clean verifier for this repo because it pulls existing environment/dependency failures: Kafka symbols in `pkg/microservice`, Firebase/LINE network/config tests, and date-sensitive coupon tests.
- Product route/screen paths that were touched in this change now use lowercase no-underscore names such as `/productbarcode`, `/productserialregistry`, and `/channelprice`. Other legacy UI route compatibility contracts need their own route migration plan before renaming.

## R0 Blocker

Deleting old data is destructive. Before running any drop/delete/truncate command, Jead must explicitly name:

1. Environment: local, DEV, UAT, production, or all.
2. Stores: MongoDB, PostgreSQL, ClickHouse, Kafka topics, Redis/cache, or a subset.
3. Scope: all tenant data or specific modules such as product/master data only.
