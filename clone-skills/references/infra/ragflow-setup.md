---
name: RAGFlow + Ollama + Typhoon Setup Guide
description: Complete install guide for the BC Account self-hosted RAG stack — Docker, persistent storage, multi-tenant strategy, backup/restore. Use when setting up a new server or troubleshooting an existing RAGFlow deployment.
---

# RAGFlow + Ollama + Typhoon — BC Account RAG Stack

> **Status**: Integrated into goapi as of 2026-04-08.
> Backend code lives in `backend/internal/goapi/ragflow/` and `backend/internal/goapi/handlers/knowledgebase/`.
> Tool name in "Nong Kung" agent: `query_knowledge_base`.

## Why this stack

- **Self-hosted on Docker** — data never leaves the box, 100% private
- **Thai-first** — Typhoon LLM + BGE-M3 multilingual embedding
- **Free** — no API cost, no per-token fees
- **REST API** — clean integration with Go/Echo + Flutter
- **Multi-tenant** — one dataset per shop, scales to 2,500+

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Docker (host: dual-Xeon, 100GB RAM)                     │
│                                                          │
│  ┌───────────────────┐     ┌──────────────────────┐      │
│  │ RAGFlow            │     │ Ollama                │      │
│  │ - Web UI (:80)     │────▶│ - Typhoon 3.1 (LLM)  │      │
│  │ - REST API (:9380) │     │ - BGE-M3 (Embedding)  │      │
│  │ - Elasticsearch    │     │ - Port: 11434          │      │
│  │ - MySQL            │     └──────────────────────┘      │
│  │ - MinIO            │                                    │
│  └───────┬───────────┘                                    │
│          │ REST API                                        │
└──────────┼────────────────────────────────────────────────┘
           │
    ┌──────┴──────┐
    │ goapi (BC)  │ ─── Flutter app (KB screen + Nong Kung chat)
    │ MongoDB     │     Auto-creates dataset per shop
    │ PostgreSQL  │
    └─────────────┘
```

## Persistent storage layout

**ALL data goes under `/data/ragflow/`** — backup this folder = backup everything.

```
/data/ragflow/
├── elasticsearch/   # vector + full-text index
├── mysql/           # RAGFlow metadata, users, configs
├── minio/           # uploaded source documents (PDF/DOCX/MD/TXT)
└── ollama/          # LLM + embedding models (~6GB)
```

Mapping `shop_id → dataset_id` is **NOT** stored in PostgreSQL of BC Account.
Instead we use **deterministic naming** (`bcacct_shop_<shopID>`) plus an in-memory cache + lookup, so RAGFlow is the source of truth for dataset existence.

Per-document BC metadata (branch, status, schedule) lives in MongoDB collection `kbDocumentMetadata` (see `backend/internal/goapi/handlers/knowledgebase/kb_meta.go`).

## Install steps

### 1. Docker + system tuning

**Linux (production server):**
```bash
sudo apt update && sudo apt install -y curl git git-lfs
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Required for Elasticsearch
sudo sysctl -w vm.max_map_count=262144
echo "vm.max_map_count=262144" | sudo tee -a /etc/sysctl.conf

# Persistent dirs
sudo mkdir -p /data/ragflow/{elasticsearch,mysql,minio,ollama}
sudo chmod -R 777 /data/ragflow
```

**Windows (dev):** install Docker Desktop with WSL2 backend, RAM ≥ 8GB.

### 2. Ollama + Thai models

```bash
docker run -d \
  --name ollama \
  -p 11434:11434 \
  -v /data/ragflow/ollama:/root/.ollama \
  --restart unless-stopped \
  ollama/ollama

# (use --gpus all if NVIDIA GPU available)

docker exec ollama ollama pull typhoon3.1:8b   # ~5GB — Thai LLM
docker exec ollama ollama pull bge-m3          # ~1.2GB — multilingual embedding

curl http://localhost:11434/   # → "Ollama is running"
```

### 3. RAGFlow with persistent storage

```bash
git clone https://github.com/infiniflow/ragflow.git
cd ragflow/docker
```

Create `docker-compose.override.yml` next to `docker-compose.yml`:

```yaml
services:
  es01:
    volumes:
      - /data/ragflow/elasticsearch:/usr/share/elasticsearch/data
  mysql:
    volumes:
      - /data/ragflow/mysql:/var/lib/mysql
  minio:
    volumes:
      - /data/ragflow/minio:/data
```

> Service names may differ between RAGFlow versions — check `docker-compose.yml` of the version you cloned and adjust accordingly.

```bash
docker compose -f docker-compose.yml up -d
docker logs -f ragflow-server   # wait until "server ready"
```

### 4. RAGFlow Web UI configuration

1. Open `http://localhost`
2. Sign up (first user becomes admin)
3. **Profile → Model providers → Ollama**
   - Linux: `http://<host-ip>:11434`
   - Docker Desktop (Win/Mac): `http://host.docker.internal:11434`
4. **Set default models**:
   - Chat model: `typhoon3.1:8b`
   - Embedding model: `bge-m3`
5. **Profile → API → Create API Key** → copy this string

### 5. Wire goapi to RAGFlow

Set in `backend/docker-compose.yml` (or `.env`):

```yaml
RAGFLOW_BASE_URL: "http://host.docker.internal:9380"   # or production host
RAGFLOW_API_KEY: "ragflow-XXXXXXXXXXXXXXXXXXXXXX"
```

Restart mainapi:
```bash
cd backend && docker compose up -d mainapi
```

Verify:
```bash
curl http://localhost:8888/goapi/api/v1/kb/health
# expect: {"status":"ok","configured":true,...}
```

## REST endpoints (goapi → RAGFlow)

| Method | Endpoint | Purpose |
|---|---|---|
| GET | `/goapi/api/v1/kb/health` | Status check (used by Flutter banner) |
| POST | `/goapi/api/v1/kb/list` | List shop's documents |
| POST | `/goapi/api/v1/kb/upload` | Base64 upload + auto-parse |
| POST | `/goapi/api/v1/kb/delete` | Delete by filename |
| POST | `/goapi/api/v1/kb/update-status` | Toggle on/off |
| POST | `/goapi/api/v1/kb/update-allday` | Toggle 24/7 vs scheduled |
| POST | `/goapi/api/v1/kb/update-schedule` | Set start/end datetime |

The AI agent (Nong Kung) uses tool `query_knowledge_base` which calls RAGFlow `/api/v1/retrieval` directly — returns raw chunks for the agent's LLM to synthesize.

## Backup & restore

**Daily automated backup** — `/opt/scripts/ragflow-backup.sh`:

```bash
#!/bin/bash
set -e
BACKUP_DIR="/backups/ragflow"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p "$BACKUP_DIR"

# Hot backup of data dir
tar -czf "$BACKUP_DIR/ragflow-backup-${DATE}.tar.gz" -C / data/ragflow/

# Extra MySQL dump (safer for relational data)
docker exec ragflow-mysql mysqldump -u root -pinfini_rag --all-databases \
  > "$BACKUP_DIR/ragflow-mysql-${DATE}.sql" 2>/dev/null

# Retention: 30 days
find "$BACKUP_DIR" -name "ragflow-backup-*.tar.gz" -mtime +30 -delete
find "$BACKUP_DIR" -name "ragflow-mysql-*.sql" -mtime +30 -delete

echo "Backup complete: $(du -sh $BACKUP_DIR/ragflow-backup-${DATE}.tar.gz)"
```

```bash
sudo crontab -e
# Daily 02:00
0 2 * * * /opt/scripts/ragflow-backup.sh >> /var/log/ragflow-backup.log 2>&1
```

**Restore** (after reinstalling Docker + Ollama + RAGFlow on a new box):
```bash
sudo tar -xzf /backups/ragflow/ragflow-backup-YYYYMMDD_HHMMSS.tar.gz -C /
cd ragflow/docker && docker compose up -d
# Optional: docker exec -i ragflow-mysql mysql -u root -pinfini_rag < backup.sql
```

## Production tips (2,500 shops)

1. **Disk planning** — 2,500 shops × 100 docs × 1MB = ~250GB. Allocate 500GB+ with monitoring.
2. **Disk alert** — page someone when free space < 20%.
3. **Elasticsearch snapshot repo** — set up native ES snapshot/restore in addition to tar backup.
4. **MinIO → S3** — for true HA, swap MinIO for AWS S3 / MinIO cluster mode with replication.
5. **Move MySQL/ES out of Docker** eventually — decouple from container lifecycle.
6. **API key rotation** — rotate `RAGFLOW_API_KEY` quarterly; goapi reads it from env at startup so just `docker compose restart mainapi`.

## Useful URLs

| Service | URL |
|---|---|
| RAGFlow Web UI | `http://localhost:80` |
| RAGFlow REST API | `http://localhost:9380` |
| Ollama API | `http://localhost:11434` |
| RAGFlow docs | https://ragflow.io/docs |
| RAGFlow source | https://github.com/infiniflow/ragflow |

## Files in this repo touching RAGFlow

- `backend/internal/goapi/ragflow/` — Go client (client/dataset/document/retrieval)
- `backend/internal/goapi/handlers/knowledgebase/` — REST handlers + Mongo metadata
- `backend/internal/goapi/handlers/aichat/kb_tool.go` — Nong Kung agent tool
- `backend/internal/goapi/bootstrap.go` — route registration
- `backend/docker-compose.yml` — RAGFLOW_BASE_URL / RAGFLOW_API_KEY env vars
- `frontend/bcaiaccount/lib/bloc/knowledge_base/knowledge_base_cubit.dart` — calls goapi
- `frontend/bcaiaccount/lib/screens/knowledge_base/knowledge_base_screen.dart` — UI + health banner
