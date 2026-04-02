# BC AI Backend — Project Rules

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
- Frontend project (bcaiaccount) ตั้งกฏว่า AI ห้ามอ่าน backend code — ต้องผ่าน MCP เท่านั้น
- ถ้า frontend ต้องการ API ใหม่ จะส่ง API Specification Prompt มาในไฟล์ `prompts/api_requests/{feature}.md`
- เมื่อได้รับ prompt จาก frontend → สร้าง API ตาม spec + เพิ่ม MCP tool ถ้าจำเป็น

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

## Jead Skill Library (Global Rules)
**Read `D:\bcdev\clone-skills\` every session:**
- `identity.md` — Jead's identity, style, lessons learned
- `rules/*.md` — All working rules (coding, workflow, security, MCP)
- `skills/*/SKILL.md` — Available slash commands
- `references/` — Domain knowledge (core, flutter, go)

**Auto-update:** When learning new patterns/rules from Jead → update clone-skills immediately
