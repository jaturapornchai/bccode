# BC Ai Account Backend — Project Rules

## Overview
Unified Go backend: **mainapi** (cloud platform) + **goapi** (BI/analytics API) in one module.

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
│   │   ├── mcp/              # MCP server (AI tools bridge)
│   │   │   ├── server.go     # SSE server
│   │   │   ├── sse_handler.go
│   │   │   ├── auth/         # API key auth
│   │   │   ├── mongodb/      # MongoDB tools
│   │   │   ├── redis/        # Redis tools
│   │   │   └── tools/        # Tool implementations (15 files)
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

**MainAPI — จุดเข้าเดียว (port 8888):**
- MainAPI มี GoAPI embedded อยู่ข้างใน (ไม่มี reverse proxy)
- Local: `http://localhost:8888`
- `/goapi/*` → GoAPI (embedded)
- `/*` → MainAPI routes

| Service | Host Port | Bind | Note |
|---------|-----------|------|------|
| **MainAPI** (gateway) | **8888** | `0.0.0.0` | จุดเข้าเดียว (GoAPI embedded) |
| MongoDB | 27017 | — | **NATIVE** (ไม่อยู่ Docker) |
| PostgreSQL | 5432 | — | **NATIVE** (ไม่อยู่ Docker) |
| Redis | 6379 | `127.0.0.1` | redis:6379 |
| ClickHouse HTTP | 8123 | `127.0.0.1` | clickhouse:8123 |
| ClickHouse Native | 9000 | `127.0.0.1` | clickhouse:9000 |
| Kafka | 9092 | `127.0.0.1` | kafka:29092 |
| Zookeeper | — | docker only | zookeeper:2181 |
| SeaweedFS Master | 9333 | `127.0.0.1` | seaweedfs-master:9333 |
| SeaweedFS Filer | 18888 | `127.0.0.1` | seaweedfs-filer:8888 |
| SeaweedFS S3 | 18333 | `127.0.0.1` | seaweedfs-filer:8333 |

**MongoDB + PostgreSQL = NATIVE:**
- ลง native บน host (ไม่อยู่ใน Docker — ข้อมูลปลอดภัย)
- Docker containers เข้าถึงผ่าน `host.docker.internal`
- MongoDB ต้องมี replica set `rs0` (goapi ใช้ change streams)
- Docker Compose ใช้ `extra_hosts: ["host.docker.internal:host-gateway"]`

**กฏ Port:**
- **MainAPI (port 8888) เท่านั้น** ที่เปิดออกนอก
- Infrastructure ทุกตัว bind `127.0.0.1` (internal only — ไม่เปิดออกนอก)
- **Port เดียวกันทุก environment** ไม่ว่า Docker Desktop, VPS, หรือ server ตัวไหน
- ห้ามเปลี่ยน port mapping — ถ้าต้องการเปลี่ยนต้องแก้ทั้ง compose + CLAUDE.md + Memory

**Domain (กฏ):**
- **VPS dev:** `api.bcaicloud.com` → `5.223.69.66` (Hetzner)
- ห้ามใช้ IP ตรง ให้ใช้ domain เสมอ (ยกเว้น SSH)

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

# Test — VPS
curl http://api.bcaicloud.com:8888/goapi/version
curl http://api.bcaicloud.com:8888/goapi/api/health
```

### Import Paths
goapi packages use: `smlcloudplatform/internal/goapi/...`
```go
import "smlcloudplatform/internal/goapi/handlers"
import "smlcloudplatform/internal/goapi/mydb"
import "smlcloudplatform/internal/goapi/myglobal"
```

### MCP Server — ตัวเชื่อมข้าม Project
MCP เป็น bridge ให้ AI tools ฝั่ง frontend เข้าถึง backend ได้โดยไม่ต้องอ่าน code โดยตรง

- SSE endpoint: `GET /goapi/mcp/sse` (for Claude Desktop / Claude Code)
- Health: `GET /goapi/mcp/health`
- Tools: `GET /goapi/mcp/tools`
- API keys: `POST/GET/PUT/DELETE /api/mcp/keys`
- **Tools (15 files):** sales, dashboard, financial, inventory, customers, products, comparison, database, model_schema, enum_catalog, clickhouse_query, mongodb_query, api_catalog, api_spec, token_export

**กฏสำคัญ:**
- Frontend target ใหม่คือ Next.js ที่ `D:\bccode\frontend` จาก `https://github.com/jaturapornchai/bccode`
- Flutter ที่ `D:\bcdev\frontend` จาก `https://github.com/jaturapornchai/bcdev` โดยเฉพาะ `bcaiaccount` เป็น reference/template สำหรับ migration เท่านั้น
- AI ฝั่ง frontend ห้ามเดา backend behavior/schema — ต้องใช้ MCP, API docs, หรือ API Specification Prompt
- ถ้า frontend ต้องการ API ใหม่ ให้ส่ง API Specification Prompt มาในไฟล์ `prompts/api_requests/{feature}.md`
- เมื่อได้รับ prompt จาก frontend → สร้าง API ตาม spec + เพิ่ม MCP tool ถ้าจำเป็น

### ERP Language Source Of Truth
- `assets/language/languages.tsv` is the single source of truth for ERP UI labels, field labels, report names, report headers, report columns, status text, and repeated business terms.
- Backend reports, PDF generation, API metadata, and frontend screens must use the same keys from `languages.tsv`; do not maintain separate report-only JSON dictionaries or frontend-only business-label dictionaries that can drift.
- Language APIs must normalize aliases such as `zh -> cn`, `jp -> ja`, `kr -> ko`, and `tl -> fil`.
- Fallback order is requested language → English → Thai → key.
- Report request payloads must pass `language_code`; generated report titles/headers/columns must match frontend screen labels for the same selected language.
- When adding any visible ERP text, add the key to `assets/language/languages.tsv` and add/update a narrow test for the backend/frontend path using it.

### Multi-Tenant With tenant_id
- `tenant_id` is the canonical tenant boundary for new backend/API/report work.
- `tenant_id` represents one company/business/legal entity/workspace. It must not represent the owner user because one user can own or access many companies.
- For existing production data, `tenant_id` is a logical alias whose value is the existing core `shopid`. Do not create a second tenant id for old records.
- User/company access must be modeled through membership/role data: `user_id` -> many `tenant_id`; each `tenant_id` -> many `branch_id`.
- Owner-level overview across many companies uses `company_group_id` above many `tenant_id` values. Keep authorization tenant-scoped and pass only authorized tenant lists to ClickHouse.
- Existing core storage uses `shopid` as the physical tenant identity in many current tables, collections, ClickHouse rows, and Kafka payloads. Some GoAPI/MCP/AI/approval DTOs use `shop_id`; inspect the module before choosing the physical key. Keep existing field names as-is unless a later migration has a functional reason beyond naming consistency.
- Every handler must derive `tenant_id` from authenticated user/workspace membership, validate access, and pass it through repository/service/report/job layers.
- Every customer-data query must filter by the logical `tenant_id`. In legacy repositories, map that value to physical `shopid` or module-specific `shop_id` and keep the filter in the same query.
- New standalone schema should include `tenant_id` and composite indexes such as `(tenant_id, id)`, `(tenant_id, branch_id, doc_no)`, or the best key for the access pattern. Legacy schema may keep its real physical tenant key, usually `shopid`.
- Do not remove or rename existing `shopid` / `shop_id` fields until all callers, migrations, indexes, tests, reports, Kafka consumers, and object paths are verified and there is a real functional benefit.
- Cross-tenant admin/report operations require explicit admin permission, audit log, and clear code-level naming.
- Use `architecture/high-scale-multitenant-bi.md` as the blueprint for MongoDB -> Kafka -> PostgreSQL -> Kafka -> ClickHouse at high concurrency.
- Use `architecture/admin-access-control.md` for platform admins, group owners, tenant admins, branch grants, first-admin bootstrap, and policy-based access resolution.

### K3s Production Scaling
- Docker Desktop/docker-compose remains local/dev only. Use `cluster/k3s` for production/high-concurrency deployment.
- Do not promise 10,000 concurrent screens until a load test proves the full path: Ingress → mainapi/goapi → PostgreSQL/MongoDB/ClickHouse/Kafka/object storage/report generation.
- K3s production must use HA control-plane design, SSD-backed datastore, secrets encryption at rest, resource requests/limits, HPA, probes, PodDisruptionBudget, and no hardcoded secrets in manifests.
- `mainapi` is stateless enough to scale horizontally only when production config is read from Kubernetes Secret/ConfigMap or a shared config service. Mutable per-pod `bootstrap.json` writes are not safe for multiple replicas.
- Stateful dependencies must be external HA services or dedicated clustered deployments. Do not copy single-node docker-compose infrastructure into production K3s as-is.

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
- Use this `CLAUDE.md`, backend source code, `prompts/`, `assets/language/`, and MCP docs as the source of truth.
- Do not rely on deleted shared skill folders.
- When Jead asks to keep a new backend rule, update the relevant project-local docs or prompts.
