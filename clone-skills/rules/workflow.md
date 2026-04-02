# Workflow — Development Patterns

## Development Pipeline (all steps required)
```
Plan → Implement → Build → Deploy → Test → Push
```

1. **Plan**: Plan first for complex tasks — use plan mode
2. **Implement**: Write code immediately (minimize questions)
3. **Build**: `go build ./cmd/goapi/` or flutter build — must pass
4. **Deploy**: `docker compose build mainapi && docker compose up -d`
5. **Test**: Test immediately with curl or browser
6. **Push**: `git push` only when Jead commands (never push autonomously)

## Deploy Flow

### Architecture Overview
```
┌─────────────────────────────────────────────────────────┐
│ Frontend (Flutter Web)                                  │
│ ┌─────────────┐ ┌─────────────┐ ┌────────────────────┐ │
│ │ Firebase DEV │ │ Firebase UAT│ │ Firebase PROD      │ │
│ │ (dev-2.web)  │ │ (uat.web)   │ │ (bcai-cloud.web)   │ │
│ └─────────────┘ └─────────────┘ └────────────────────┘ │
│ ┌────────────────────────────────────────────┐          │
│ │ VPS Hetzner (dev.bcaicloud.com)            │          │
│ │ nginx → static files (SCP upload)          │          │
│ └────────────────────────────────────────────┘          │
├─────────────────────────────────────────────────────────┤
│ Backend (Go + Docker)                                   │
│ ┌──────────────────┐ ┌─────────────────────────┐        │
│ │ Docker Desktop   │ │ VPS (self-deploy API)    │        │
│ │ localhost:8888   │ │ dev.bcaicloud.com:8888   │        │
│ └──────────────────┘ └─────────────────────────┘        │
├─────────────────────────────────────────────────────────┤
│ Database (NATIVE — not in Docker)                       │
│ MongoDB:27017 │ PostgreSQL:5432                          │
├─────────────────────────────────────────────────────────┤
│ Infrastructure (Docker)                                  │
│ Redis:6379 │ ClickHouse:18123 │ Kafka:9092              │
│ SeaweedFS Master:9333 │ Filer:18888 │ S3:18333          │
└─────────────────────────────────────────────────────────┘
```

### 0. Data Sync Rule — MongoDB → Kafka → PostgreSQL + ClickHouse (Critical)

**All master + transaction data must flow through every layer:**
```
MongoDB (save) → Kafka (publish) → PostgreSQL (consumer upsert) + ClickHouse (consumer upsert)
```

**Rules:**
- Every entity saved to MongoDB **must** publish a Kafka message (created/updated/deleted/bulk-*)
- Kafka consumer must upsert to both PostgreSQL **and** ClickHouse
- If PG table missing → create model + migration immediately
- If ClickHouse table missing → create CREATE TABLE statement immediately
- Ensure consistency: MongoDB ↔ PG ↔ CH fields must match

**Reference patterns:**
- Master data (org/product): `internal/product/unit/` pattern (config + MQ repo + publish in service)
- Transaction data (PO/PR): `internal/transaction/transactionconsumer/purchaseorder/` pattern
- Consumer registration: `cmd/transaction_consumer/main.go`

**Deploy Rule for languages.tsv:**
After editing `assets/language/languages.tsv` → deploy Docker Desktop immediately (language data is in Docker image).

### 1. Deploy Backend — Docker Desktop (routine)
**Auto Deploy Rule (mandatory):**
After every backend code change → deploy Docker Desktop immediately without waiting for Jead's command.
Steps: build → docker build → up → verify — complete all in one sequence.

```bash
# 1. Build check (must pass first)
cd D:/bcdev/backend && go build ./cmd/goapi/

# 2. Docker build + deploy
cd D:/bcdev/backend && docker compose build mainapi && docker compose up -d

# 3. Verify
curl http://localhost:8888/goapi/api/health
```

### 2. Deploy Frontend — Firebase (DEV / UAT / PROD)
**Run from:** `D:\bcdev\frontend\bcaiaccount\`

| Environment | Script | Entry Point | Firebase Target | URL |
|-------------|--------|-------------|-----------------|-----|
| DEV | `scripts\deploy_dev.bat` | `main_bcaidev.dart` | `bcai-dev` | smlai-cloud-dev-2.web.app |
| UAT | `scripts\deploy_uat.bat` | `main_bcaiuat.dart` | `bcai-uat` | smlai-cloud-uat.web.app |
| PROD | `scripts\deploy_prod.bat` | `main_bcaiprod.dart` | `bcaicloud` | bcai-cloud.web.app |

**Steps (same for all environments):**
```
1. generate_build_info.ps1 → creates web/build-info.json (version, git hash, auto-increment build number)
2. flutter build web -t lib/main_bcai{env}.dart --release --no-tree-shake-icons --pwa-strategy none
3. firebase deploy --only hosting:{target}
```

**PROD requires typing YES to confirm before deploy.**

### 3. Deploy Frontend — VPS Hetzner (dev.bcaicloud.com)
**Run from:** `D:\bcdev\frontend\bcaiaccount\`
**Script:** `scripts\deploy_dev_vps.bat`

```
VPS: 5.223.69.66 (Hetzner)
SSH Key: %USERPROFILE%\.ssh\hetzner_deploy
Web Dir: /opt/bcaicloud/frontend/web
URL: https://dev.bcaicloud.com/
```

**Steps:**
```
1. generate_build_info.ps1 → build-info.json
2. flutter build web -t lib/main_bcaidev.dart --release
3. Create config.json → {"goapi_url": "https://dev.bcaicloud.com/goapi"}
4. scp -i ~/.ssh/hetzner_deploy -r build/web/* root@5.223.69.66:/opt/bcaicloud/frontend/web/
5. SSH: fix font filenames (space → %20) + nginx -s reload
```

**Note:** Frontend (nginx on host) and Backend (Docker) run independently.

### 4. Deploy Backend — VPS (Self-Deploy API)
Backend on VPS has a built-in deploy API (no SSH needed):

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/deploy/backend?token=xxx` | POST | git pull → docker compose build → up -d |
| `/api/deploy/frontend?token=xxx` | POST | git pull → copy web files |
| `/api/deploy/status?token=xxx` | GET | Check latest deploy status |

**Set DEPLOY_TOKEN env var** — if unset, deploy feature is disabled.

### Common Deploy Commands
```bash
# === Backend (Local) ===
cd D:\bcdev\backend
docker compose build mainapi && docker compose up -d

# === Frontend (Firebase DEV) ===
cd D:\bcdev\frontend\bcaiaccount
scripts\deploy_dev.bat

# === Frontend (VPS dev.bcaicloud.com) ===
cd D:\bcdev\frontend\bcaiaccount
scripts\deploy_dev_vps.bat

# === Frontend (Firebase PROD) ===
cd D:\bcdev\frontend\bcaiaccount
scripts\deploy_prod.bat
```

## Cleanup Rule
- **Never create temp scripts**: No .py / .sh / temporary scripts (e.g., fix_xxx.py) — use Claude Code tools (Edit, Grep, Bash sed) directly
- For batch replace → use Edit tool with replace_all or Bash sed/awk
- Verify no temp scripts remain before commit

## Git Workflow
- **Branch**: Use `dev` branch as primary
- **Auto Push Rule**: Commit + push after every code change to prevent code loss and enable rollback
- **Commit messages**: English, concise, to the point
- **Git config**: Never modify git config (user.email, user.name)

## Frontend-Backend Communication (Critical)

**Two projects work in coordination — watch data structures carefully.**

| Project | Path | AI Access |
|---------|------|-----------|
| Frontend | `D:\bcdev\bcaiaccount` | Read + Write |
| Backend | `D:\bcdev\backend` | Read + Write |

### Coordination Rules:
1. **Always verify data structure first** — use MCP `get_model_schema`, `get_database_schema` for actual fields/types
2. **Never assume field names** — verify via MCP or source code
3. **JSON field names must match** — check if backend uses snake_case (`trans_flag`) or camelCase (`transFlag`)
4. **Null safety** — backend may return null → frontend must handle with `?.` / `?? default`
5. **Type matching** — backend int64 → Dart int, float64 → Dart double, time.Time → Dart DateTime

### After completing backend work that frontend needs:
1. Create prompt spec in `prompts/api_requests/{feature}.md`
2. Tell Jead "need prompt for frontend changes"
3. Prompt must include: endpoint, request, response, fields, error handling

### When frontend needs a new API:
1. Search via MCP first for existing APIs
2. If none → write API Specification Prompt
3. Build frontend code with placeholder/empty state
4. Notify Jead that backend is pending

### Data structure warnings:
- **MongoDB → PostgreSQL sync** — field names may differ (mongo: `guidfixed`, pg: `id`)
- **ClickHouse** — analytics only, may have data delay
- **Enum values** — must match backend (use `/enum-list` to check)
- **Date format** — backend sends ISO 8601 (`2026-03-03T10:30:00Z`) → Dart `DateTime.parse()`

## Skill Auto-Update Rule (Mandatory)
After modifying code related to any skill → update that skill immediately.

Examples:
- Change theme/colors → update `/theming` skill
- Change data list pattern → update `/data-list` skill
- Change API pattern → update references
- Discover new pitfall → add to relevant skill

**Reason:** Without updates, the next AI session repeats the same mistakes endlessly.

## Task Completion Checklist
After every task:
- [ ] Code compiles/builds successfully
- [ ] Deploy succeeded (if required)
- [ ] Summarize what was done
- [ ] Note any required frontend changes
- [ ] **Update related skills** (if any) — never skip

## Forbidden Actions
- Never force push / reset hard
- Never deploy without building first
- Never delete files without asking (unless Jead commands)
- Never change docker-compose.yml ports without notifying
- **Never deploy to VPS** — unless Jead gives explicit special command (Docker Desktop only)

## Troubleshooting
1. **Rate limit**: Use fallback provider (switch immediately, don't wait)
2. **Build error**: Fix root cause (never skip lint/vet)
3. **Deploy fail**: Check `docker compose logs mainapi`
4. **API error**: Test with curl + check logs
