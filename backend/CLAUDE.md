# BC Ai Account Backend — Project Rules

## Overview
Unified Go backend: **mainapi** (cloud platform) + **goapi** (BI/analytics API) in one module.

Read `D:\bccode\.agents\rules\bc-account-core-rules.md` and `D:\bccode\.agents\wiki\llm-index.md` before changing backend runtime, database, storage, secret, DEV deployment behavior, or reusable backend LLM knowledge.

- **Module:** `smlcloudplatform`
- **Go version:** 1.26
- **Framework:** Echo v4
- **Databases:** MongoDB, PostgreSQL, ClickHouse, Kafka, Redis

## Project Structure

```
├── cmd/
│   ├── app/              # mainapi entry point (port 8080)
│   ├── goapi/            # goapi entry point (port 8888)
│   └── .../              # 23+ other microservices
├── internal/
│   ├── goapi/            # goapi packages (BI/analytics)
│   │   ├── handlers/         # HTTP handlers (50+ files)
│   │   │   ├── aichat/       # AI chat handlers
│   │   │   ├── approval/     # Approval workflow
│   │   │   ├── datahistory/  # Data history
│   │   │   ├── dataimport/   # Data import
│   │   │   ├── gen-trans-pdf/# PDF generation (transactions)
│   │   │   ├── kafka/        # Kafka bridge handlers
│   │   │   ├── lineoa/       # LINE OA integration
│   │   │   └── unified/      # Unified handlers
│   │   ├── models/           # Data models
│   │   ├── mydb/             # Database manager (PG + CH + Circuit Breaker)
│   │   ├── mypg/             # PostgreSQL utilities
│   │   ├── mypostgres/       # PostgreSQL (extended utilities)
│   │   ├── myclickhouse/     # ClickHouse utilities
│   │   ├── mykafkaconsumer/  # Kafka consumer framework
│   │   ├── mydlq/            # Dead Letter Queue
│   │   ├── myretry/          # Retry utilities
│   │   ├── myglobal/         # Global state + MongoDB connection
│   │   ├── myollama/         # Ollama local AI integration
│   │   ├── mythaitokenizer/  # Thai NLP tokenizer
│   │   ├── gemini/           # Google Gemini AI integration
│   │   ├── config/           # goapi config
│   │   ├── setupconfig/      # bootstrap.json loader
│   │   ├── logger/           # goapi logger
│   │   ├── dataimport/       # Data import pipeline
│   │   ├── datainfo/         # Data info
│   │   ├── process/          # Stock/document processing
│   │   ├── workers/          # Background workers
│   │   │   ├── db_log_cleaner.go
│   │   │   └── doc_processor.go
│   │   └── cache/            # Unified cache manager
│   ├── config/               # mainapi config
│   ├── microservice/         # mainapi microservice framework
│   ├── smlaiproduct/         # AI product features
│   ├── stockbalanceimport/   # Stock balance import
│   └── .../                  # 60+ mainapi internal packages
├── pkg/                  # Shared utilities (mainapi)
├── assets/
│   ├── fonts/            # PDF fonts
│   └── language/         # i18n
├── migrations/           # SQL migrations
├── prompts/              # AI API request specs (frontend → backend)
│   └── api_requests/     # Feature API specs from frontend
├── Dockerfile            # mainapi Docker (CGO + librdkafka)
├── Dockerfile.goapi      # goapi Docker (pure Go, no CGO)
└── docker-compose.yml    # Full stack (infra + apps)
```

## Critical Rules

### Config: bootstrap.json ONLY — NO .env files
- ALL config reads from `bootstrap.json` via `internal/goapi/setupconfig/loader.go`
- bootstrap.json location (priority):
  1. `/app/bootstrap/bootstrap.json`
  2. `/app/bootstrap.json` (Docker mount)
  3. `bootstrap.json`
  4. `config/bootstrap.json`
- Docker mount: `./bootstrap.json:/app/bootstrap.json:rw` (relative path — ใช้ได้ทุก OS)
- `bootstrap.json` อยู่ใน `.gitignore` (ห้าม commit เพราะมี credentials)
- **NEVER use** `env_file`, `.env`, `.env.*` in docker-compose or code
- **NEVER use** `godotenv.Load()` or read from `.env` files

### Docker Stack — Port มาตรฐาน (ใช้ทุกที่เหมือนกัน)
ใช้ `docker-compose.yml` ไฟล์เดียว ทุก environment (Docker Desktop, VPS, Production)

Local backend development runs on Docker Desktop. Do not require host Go/CGO/librdkafka setup for normal local backend runs; host Go is for focused compiler/toolchain work only.
Kafka and Redis are mandatory runtime services for MainAPI. Local Docker Desktop runs must include Kafka and Redis, with MainAPI using Docker-network addresses such as `KAFKA_SERVER_URL=kafka:29092` and `REDIS_CACHE_URI=redis:6379`.

**MainAPI — จุดเข้าเดียว (port 8888):**
- MainAPI มี GoAPI embedded อยู่ข้างใน (ไม่มี reverse proxy)
- Local: `http://localhost:8888`
- `/goapi/*` → GoAPI (embedded)
- `/*` → MainAPI routes

| Service | Host Port | Bind | Note |
|---------|-----------|------|------|
| **MainAPI** (gateway) | **8888** | `0.0.0.0` | จุดเข้าเดียว (GoAPI embedded) |
| MongoDB | 27017 | local docker | DEV ใช้ local container (replica set rs0) — Atlas เลิกใช้แล้ว ตาม Environment Topology |
| PostgreSQL | 5432 | local docker | DEV ใช้ local container |
| Redis | 6379 | `0.0.0.0` | redis:6379 |
| ClickHouse HTTP | 8123 | local docker | DEV ใช้ local container |
| ClickHouse Native | 9000 | local docker | DEV ใช้ local container |
| Kafka | 9092 | `0.0.0.0` | kafka:29092 |
| MinIO | 9000 | local docker | S3-compatible image/file storage |
| Zookeeper | — | docker only | zookeeper:2181 |

**DEV data services:**
- DEV stack ทั้งหมดรัน local (Docker Desktop): MongoDB/PostgreSQL/ClickHouse/Kafka/Redis/MinIO — ตาม Environment Topology ใน core-rules (MongoDB Atlas เลิกใช้แล้ว; stack บน `192.168.2.202` คือ DEPLOY ไม่ใช่ DEV local)
- Supply DB/storage credentials from local env/secret files only.
- System images/files เก็บใน S3-compatible storage (MinIO local/`.202`, Cloudflare R2 เมื่อ config `S3_*`/`R2_*`)
- Do not commit server passwords, MongoDB connection strings with credentials, Cloudflare tokens, or storage access keys.

**กฏ Port:**
- **MainAPI (port 8888) เท่านั้น** ที่เปิดออกนอก
- Infrastructure ทุกตัว bind `127.0.0.1` (internal only — ไม่เปิดออกนอก)
- **Port เดียวกันทุก environment** ไม่ว่า Docker Desktop, VPS, หรือ server ตัวไหน
- ห้ามเปลี่ยน port mapping — ถ้าต้องการเปลี่ยนต้องแก้ทั้ง compose + CLAUDE.md + Memory

**DEV server (on-prem) rule:**
- DEV server: `192.168.2.202`, SSH user `smlsoft` (key-based / passwordless), on-prem Docker stack.
- Backend entrypoint: `http://192.168.2.202:8888` (mainapi, GoAPI embedded).
- Frontend runs on the dev machine pointing at the LOCAL backend `http://localhost:8888` via `BCAI_LOCAL_BACKEND_URL` (fast iteration with `next dev` HMR; `next build && next start` for production-like verification; never `output:"standalone"`). It points at `http://192.168.2.202:8888` only when deliberately testing against the DEPLOY stack.
- A public domain/Caddy is optional and only applies if Jead exposes one; agents must not require it.

### Build Commands
```bash
# GoAPI (no CGO needed)
go build ./cmd/goapi/

# MainAPI (needs CGO + librdkafka)
CGO_ENABLED=1 go build -tags musl main.go

# Docker
docker compose build mainapi
docker compose up -d

# Test — Local (HTTP)
curl http://localhost:8888/goapi/version
curl http://localhost:8888/goapi/api/health

# Test — DEV public route
curl https://dev.bcaicloud.com/backend/goapi/version
curl https://dev.bcaicloud.com/backend/goapi/api/health
```

### Import Paths
goapi packages use: `smlcloudplatform/internal/goapi/...`
```go
import "smlcloudplatform/internal/goapi/handlers"
import "smlcloudplatform/internal/goapi/mydb"
import "smlcloudplatform/internal/goapi/myglobal"
```

### Frontend ↔ Backend Contract

**กฏสำคัญ:**
- Frontend target ใหม่คือ Next.js ที่ `D:\bccode\frontend` จาก `https://github.com/jaturapornchai/bccode`
- Flutter ที่ `D:\bcdev\frontend` จาก `https://github.com/jaturapornchai/bcdev` โดยเฉพาะ `bcaiaccount` เป็น reference/template สำหรับ migration เท่านั้น
- AI ฝั่ง frontend ห้ามเดา backend behavior/schema — ต้องใช้ API docs หรือ API Specification Prompt
- ถ้า frontend ต้องการ API ใหม่ ให้ส่ง API Specification Prompt มาในไฟล์ `prompts/api_requests/{feature}.md`
- เมื่อได้รับ prompt จาก frontend → สร้าง API ตาม spec

### Language Source Of Truth
- `assets/language/languages.tsv` is the single source of truth for UI labels, field labels, report names, report headers, report columns, status text, and repeated business terms.
- Backend reports, PDF generation, API metadata, and frontend screens must use the same keys from `languages.tsv`; do not maintain separate report-only JSON dictionaries or frontend-only business-label dictionaries that can drift.
- Language APIs must normalize aliases such as `zh -> cn`, `jp -> ja`, `kr -> ko`, and `tl -> fil`.
- Fallback order is requested language → English → Thai → key.
- Report request payloads must pass `language_code`; generated report titles/headers/columns must match frontend screen labels for the same selected language.
- When adding any visible business text, add the key to `assets/language/languages.tsv` and add/update a narrow test for the backend/frontend path using it.

### Multi-Tenant With tenant_id
- `tenant_id` is the canonical tenant boundary for new backend/API/report work.
- `tenant_id` represents one company/business/legal entity/workspace. It must not represent the owner user because one user can own or access many companies.
- For existing production data, `tenant_id` is a logical alias whose value is the existing core `holdingcode`. Do not create a second tenant id for old records.
- User/company access must be modeled through membership/role data: `user_id` -> many `tenant_id`; each `tenant_id` -> many `branch_id`.
- Owner-level overview across many companies uses `company_group_id` above many `tenant_id` values. Keep authorization tenant-scoped and pass only authorized tenant lists to ClickHouse.
- Existing core storage uses `holdingcode` as the physical tenant identity in many current tables, collections, ClickHouse rows, and Kafka payloads. Some GoAPI/AI/approval DTOs use `holdingcode`; inspect the module before choosing the physical key. Keep existing field names as-is unless a later migration has a functional reason beyond naming consistency.
- Every handler must derive `tenant_id` from authenticated user/workspace membership, validate access, and pass it through repository/service/report/job layers.
- Every customer-data query must filter by the logical `tenant_id`. In legacy repositories, map that value to physical `holdingcode` or module-specific `holdingcode` and keep the filter in the same query.
- New standalone schema should include `tenant_id` and composite indexes such as `(tenant_id, id)`, `(tenant_id, branch_id, doc_no)`, or the best key for the access pattern. Legacy schema may keep its real physical tenant key, usually `holdingcode`.
- Do not remove or rename existing `holdingcode` / `holdingcode` fields until all callers, migrations, indexes, tests, reports, Kafka consumers, and object paths are verified and there is a real functional benefit. **Pre-launch exception:** under the *No-Migration / Disposable Database Rule* (`.agents/rules/bc-account-core-rules.md`), while the system is PRE-LAUNCH all database + Kafka data is disposable in ALL environments (incl. production, which has no real customer data yet) — you may change/remove/rename fields, collections, and topics freely and let everything be recreated from code. This verification ritual re-applies to production data only at go-live (real customer/prod data exists); if unsure whether real prod data exists, stop and ask Jead.
- Cross-tenant admin/report operations require explicit admin permission, audit log, and clear code-level naming.
- Use `architecture/high-scale-multitenant-bi.md` as the blueprint for MongoDB -> Kafka -> PostgreSQL -> Kafka -> ClickHouse at high concurrency.
- Use `architecture/admin-access-control.md` for platform admins, group owners, tenant admins, branch grants, first-admin bootstrap, and policy-based access resolution.

### API Version Compatibility
- Backend API contracts start at `v1`; future breaking changes must add `v2`, `v3`, and so on.
- Supported API versions must run side-by-side in the same deployed backend. When `v2` is added, keep `v1` routes/handlers/adapters active at the same time.
- Route or negotiate the API version per request. Do not rely on a global backend version switch that would break old clients.
- Keep `v1` backward-compatible for web, iOS, and Android clients that have not updated yet.
- Version-aware clients should send `X-BC-Required-Backend-Version`, `X-BC-Client-Platform`, and `X-BC-Client-Version`.
- Backend handlers that participate in versioned contracts should expose active/supported backend versions when practical and reject clearly incompatible clients with a safe, user-readable error.
- Do not remove or change `v1` behavior until every dependent client version is migrated or a compatibility adapter exists.

### Kafka
- `ENABLE_KAFKA=true` in bootstrap.json `service.enable_kafka`
- Broker: `kafka:29092` (internal Docker), `localhost:9092` (external)
- Auto-create topics enabled
- Consumer groups: sale-invoice, sale-return, purchase, inventory, warehouse, etc.

## Code Conventions
- ใช้ภาษาไทยใน log messages และ comments ได้
- Logger: `logger.Info()`, `logger.Error()`, `logger.Success()`, `logger.Warn()`
- Error handling: return error, don't panic
- Database: use `mydb.DatabaseManager` with circuit breaker + retry
- HTTP responses: use Echo `c.JSON()` with proper status codes

## Known Issues
- `internal/goapi/handlers/genpdf_handler.go:439` — `fmt.Sprintf("%,.2f", amount)` uses invalid Go format verb `%,`
- mainapi Dockerfile requires CGO + librdkafka (confluent-kafka-go)

## Project Rules Source
- Use this `CLAUDE.md`, backend source code, `prompts/`, and `assets/language/` as the source of truth.
- Do not rely on deleted shared skill folders.
- When Jead asks to keep a new backend rule, update the relevant project-local docs or prompts.
