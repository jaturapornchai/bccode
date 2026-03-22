# Workflow — Jead's Preferred Work Patterns

## Development Pipeline (ต้องทำครบทุกขั้น)
```
Plan → Implement → Build → Deploy → Test → Push
```

1. **Plan**: วางแผนก่อน (ถ้างานซับซ้อน) — ใช้ plan mode
2. **Implement**: ลงมือเขียน code ทันที (ไม่ถามมากเกินไป)
3. **Build**: `go build ./cmd/goapi/` หรือ flutter build — ต้องผ่าน
4. **Deploy**: `docker compose build mainapi && docker compose up -d`
5. **Test**: ทดสอบด้วย curl หรือ browser ทันที
6. **Push**: `git push` เมื่อ Jead สั่ง (ห้าม push เอง)

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
│ Database (NATIVE — ไม่อยู่ใน Docker)                     │
│ MongoDB:27017 │ PostgreSQL:5432                          │
├─────────────────────────────────────────────────────────┤
│ Infrastructure (Docker)                                  │
│ Redis:6379 │ ClickHouse:18123 │ Kafka:9092              │
│ SeaweedFS Master:9333 │ Filer:18888 │ S3:18333          │
└─────────────────────────────────────────────────────────┘
```

### 0. Data Sync Rule — MongoDB → Kafka → PostgreSQL + ClickHouse (กฏสำคัญมาก)

**ทุก master data + transaction data ต้องไหลครบทุกชั้น:**
```
MongoDB (save) → Kafka (publish) → PostgreSQL (consumer upsert) + ClickHouse (consumer upsert)
```

**กฏ:**
- ทุก entity ที่ save ลง MongoDB **ต้อง** publish Kafka message (created/updated/deleted/bulk-*)
- Kafka consumer ต้อง upsert ทั้ง PostgreSQL **และ** ClickHouse
- ถ้า PG table ยังไม่มี → AI ต้องสร้าง model + migration ทันที
- ถ้า ClickHouse table ยังไม่มี → AI ต้องสร้าง CREATE TABLE statement ทันที
- ตรวจสอบ consistency: MongoDB ↔ PG ↔ CH ต้องสัมพันธ์กันเสมอ (fields ตรงกัน)

**Pattern ต้นแบบ:**
- Master data (org/product): `internal/product/unit/` pattern (config + MQ repo + publish in service)
- Transaction data (PO/PR): `internal/transaction/transactionconsumer/purchaseorder/` pattern
- Consumer registration: `cmd/transaction_consumer/main.go`

**Deploy Rule สำหรับ languages.tsv:**
ทุกครั้งที่แก้ไข `assets/language/languages.tsv` → **AI ต้อง deploy Docker Desktop เองทันที** (ไม่ต้องรอ Jead สั่ง) เพราะ language data อยู่ใน Docker image

### 1. Deploy Backend — Docker Desktop (ทำประจำ)
**Auto Deploy Rule (สำคัญมาก — ห้ามลืมเด็ดขาด)**:
ทุกครั้งที่แก้ไข backend code เสร็จแล้ว → **AI ต้อง deploy Docker Desktop เองทันที โดยไม่ต้องรอ Jead สั่ง**
ขั้นตอน: build → docker build → up → verify — ทำต่อเนื่องจบครบในทีเดียว

```bash
# 1. Build check (ต้องผ่านก่อน)
cd D:/bcdev/backend && go build ./cmd/goapi/

# 2. Docker build + deploy (ทำต่อทันที)
cd D:/bcdev/backend && docker compose build mainapi && docker compose up -d

# 3. Verify (ทำต่อทันที)
curl http://localhost:8888/goapi/api/health
```

**เหตุผล:** Jead ไม่อยากต้องสั่ง deploy เอง — พอ backend code เสร็จ AI ต้อง deploy ให้จบ เพื่อจะได้ทดสอบได้เลย

### 2. Deploy Frontend — Firebase (DEV / UAT / PROD)
**ทำจาก:** `D:\bcdev\frontend\bcaiaccount\`

| Environment | Script | Entry Point | Firebase Target | URL |
|-------------|--------|-------------|-----------------|-----|
| DEV | `scripts\deploy_dev.bat` | `main_bcaidev.dart` | `bcai-dev` | smlai-cloud-dev-2.web.app |
| UAT | `scripts\deploy_uat.bat` | `main_bcaiuat.dart` | `bcai-uat` | smlai-cloud-uat.web.app |
| PROD | `scripts\deploy_prod.bat` | `main_bcaiprod.dart` | `bcaicloud` | bcai-cloud.web.app |

**ขั้นตอน (ทุก environment เหมือนกัน):**
```
1. generate_build_info.ps1 → สร้าง web/build-info.json (version, git hash, auto-increment build number)
2. flutter build web -t lib/main_bcai{env}.dart --release --no-tree-shake-icons --pwa-strategy none
3. firebase deploy --only hosting:{target}
```

**PROD ต้องพิมพ์ YES ยืนยันก่อน deploy**

### 3. Deploy Frontend — VPS Hetzner (dev.bcaicloud.com)
**ทำจาก:** `D:\bcdev\frontend\bcaiaccount\`
**Script:** `scripts\deploy_dev_vps.bat`

```
VPS: 5.223.69.66 (Hetzner)
SSH Key: %USERPROFILE%\.ssh\hetzner_deploy
Web Dir: /opt/bcaicloud/frontend/web
URL: https://dev.bcaicloud.com/
```

**ขั้นตอน:**
```
1. generate_build_info.ps1 → build-info.json
2. flutter build web -t lib/main_bcaidev.dart --release
3. สร้าง config.json → {"goapi_url": "https://dev.bcaicloud.com/goapi"}
4. scp -i ~/.ssh/hetzner_deploy -r build/web/* root@5.223.69.66:/opt/bcaicloud/frontend/web/
5. SSH: fix font filenames (space → %20) + nginx -s reload
```

**สำคัญ:** Frontend (nginx on host) กับ Backend (Docker) ทำงานแยกกัน

### 4. Deploy Backend — VPS (Self-Deploy API)
Backend บน VPS มี deploy API ในตัว (ไม่ต้อง SSH เข้าไปเอง):

| Endpoint | Method | หน้าที่ |
|----------|--------|---------|
| `/api/deploy/backend?token=xxx` | POST | git pull → docker compose build → up -d |
| `/api/deploy/frontend?token=xxx` | POST | git pull → copy web files |
| `/api/deploy/status?token=xxx` | GET | ดูสถานะ deploy ล่าสุด |

**ต้องตั้ง DEPLOY_TOKEN env var** — ถ้าไม่ตั้ง = ปิดฟีเจอร์ deploy
**Backend deploy ทำ:** git pull origin develop → docker compose build goapi mainapi → up -d

### สรุปคำสั่ง Deploy ที่ใช้บ่อย
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

## Cleanup Rule (สำคัญ)
- **ห้ามสร้าง temp scripts เด็ดขาด**: AI ห้ามสร้างไฟล์ .py / .sh / script ชั่วคราว (เช่น fix_xxx.py) — ให้ใช้ Claude Code tools (Edit, Grep, Bash sed) จัดการโดยตรงแทน
- ถ้าต้อง batch replace → ใช้ Edit tool กับ replace_all หรือ Bash sed/awk
- ตรวจสอบก่อน commit ว่าไม่มี temp script หลงเหลือ

## Git Workflow
- **Branch**: ใช้ `dev` branch เป็นหลัก
- **Auto Push Rule (สำคัญมาก)**: ทุกครั้งที่มีการแก้ไข code → ต้อง commit + push to GitHub เสมอ เพื่อป้องกัน code หาย และสามารถเรียกกลับมาได้กรณี AI เข้าใจผิด หรือ user เข้าใจผิดทำให้ code พัง
- **Commit message**: ภาษาอังกฤษ กระชับ ตรงประเด็น
- **Git config**: ห้ามแก้ git config (user.email, user.name)

## Frontend-Backend Communication (สำคัญมาก)

**2 project ทำงานประสานกัน — ต้องระวังเรื่องโครงสร้างข้อมูล**

| Project | Path | AI Access |
|---------|------|-----------|
| Frontend | `D:\bcdev\bcaiaccount` | อ่าน+แก้ได้ |
| Backend | `D:\bcdev\backend` | อ่าน+แก้ได้ |

### กฏการประสานงาน:
1. **ตรวจสอบ data structure ก่อนเสมอ** — ใช้ MCP `get_model_schema`, `get_database_schema` เพื่อดู fields/types จริง
2. **ห้าม assume field names** — ต้อง verify กับ backend ผ่าน MCP หรืออ่าน source code
3. **JSON field names ต้องตรงกัน** — backend ใช้ snake_case (`trans_flag`) หรือ camelCase (`transFlag`) ต้องเช็คให้ตรง
4. **Null safety** — backend อาจ return null → frontend ต้อง handle ด้วย `?.` / `?? default`
5. **Type matching** — backend int64 → Dart int, backend float64 → Dart double, backend time.Time → Dart DateTime

### เมื่อทำ backend เสร็จ + frontend ต้องใช้:
1. สร้าง prompt spec ใน `prompts/api_requests/{feature}.md`
2. บอก Jead ว่า "ขอ prompt ไปแก้ frontend"
3. Prompt ต้องมี: endpoint, request, response, fields, error handling

### เมื่อทำ frontend + ต้องการ API ใหม่:
1. ใช้ MCP ค้นหาก่อนว่ามี API อยู่แล้วหรือไม่
2. ถ้าไม่มี → เขียน API Specification Prompt
3. สร้าง frontend code ไว้ก่อน (with placeholder/empty state)
4. แจ้ง Jead ว่ายังรอ backend

### ข้อควรระวังเรื่องโครงสร้างข้อมูล:
- **MongoDB → PostgreSQL sync** — data อาจมีชื่อ field ต่างกัน (mongo: `guidfixed`, pg: `id`)
- **ClickHouse** — ใช้สำหรับ analytics เท่านั้น อาจมี data delay
- **Enum values** — ต้องใช้ค่าเดียวกับ backend (ใช้ `/enum-list` ดูค่าจริง)
- **Date format** — backend ส่ง ISO 8601 (`2026-03-03T10:30:00Z`) → Dart parse ด้วย `DateTime.parse()`

## Skill Auto-Update Rule (สำคัญมาก — ห้ามลืม)
**ทุกครั้งที่แก้ไข code ที่เกี่ยวข้องกับ skill → ต้อง update skill ทันที**

เช่น:
- แก้ theme/สี → update `/theming` skill
- แก้ data list pattern → update `/data-list` skill
- แก้ API pattern → update references
- พบ pitfall ใหม่ → เพิ่มใน skill ที่เกี่ยวข้อง

**เหตุผล:** ถ้าไม่ update skill → AI ตัวต่อไปจะแก้ซ้ำแบบเดิม → วนแก้ไม่จบ

**วิธี:** หลังแก้ code เสร็จ → ตรวจว่า skill ไหนเกี่ยวข้อง → update pattern/pitfall/checklist ใน skill นั้น

## Task Completion Checklist
ทุกครั้งที่ทำงานเสร็จ ต้อง:
- [ ] Code compile/build ผ่าน
- [ ] Deploy สำเร็จ (ถ้า Jead ต้องการ)
- [ ] สรุปสิ่งที่ทำ
- [ ] บอกว่า frontend ต้องแก้อะไร (ถ้ามี)
- [ ] **update skill ที่เกี่ยวข้อง** (ถ้ามี) — ห้ามข้ามเด็ดขาด

## สิ่งที่ห้ามทำ (เด็ดขาด)
- ห้าม force push / reset hard
- ห้าม deploy โดยไม่ build ก่อน
- ห้ามลบไฟล์โดยไม่ถาม (ยกเว้น Jead สั่ง)
- ห้ามแก้ docker-compose.yml ports โดยไม่บอก
- **ห้าม deploy ขึ้น VPS เด็ดขาด** — ยกเว้น Jead สั่งเป็นคำสั่งพิเศษ (deploy Docker Desktop เท่านั้น)

## Troubleshooting Patterns
เมื่อเจอปัญหา:
1. **Rate limit**: ใช้ fallback provider (อย่ารอ — switch เลย)
2. **Build error**: แก้ที่ต้นเหตุ (ห้ามข้าม lint/vet)
3. **Deploy fail**: ดู `docker compose logs mainapi`
4. **API error**: ทดสอบด้วย curl + ดู log
