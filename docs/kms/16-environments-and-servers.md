# สภาพแวดล้อมและเซิร์ฟเวอร์ (dev / on-prem / production)
> ตรวจล่าสุด: 2026-09-07 (commit d93a210d) — ที่มา: ความรู้ปฏิบัติงานที่เคยอยู่แค่ใน memory ของ Claude (2026-06 → 2026-09) ย้ายมารวมที่นี่ตามคำสั่งลุงจืด; ค่าที่เป็น secret ไม่บันทึกที่นี่ (อยู่ใน env/bootstrap บนเครื่องนั้น ๆ)

## ภาพรวม: 3 ระบบแยกกัน ไม่แชร์ข้อมูล

| ระบบ | ที่อยู่ | ใช้ทำอะไร | สถานะ 2026-09-07 |
|---|---|---|---|
| **DEV** | เครื่องลุงจืด (Windows 11, Docker Desktop) | พัฒนาเร็ว ข้อมูลทิ้งได้ | ใช้งานอยู่ — stack ตาม `backend/docker-compose.yml` + `docker-compose.local.yml` |
| **on-prem "deploy dev"** | `192.168.2.202` (LAN, ssh user `smlsoft`) | ให้ทีมเล่น (shared) | **ssh timeout 2026-09-06 — ไม่ทราบสถานะ** |
| **PROD ใหม่** | DigitalOcean SGP1 `159.223.43.229` (root, key-only) | production `account.bcaicloud.com` | provision + deploy ครั้งแรก 2026-09-02/03 (images r20260902-11 / r20260903-1); DNS ชี้แล้วหรือยัง = ยังไม่ตรวจ |
| PROD เก่า | `188.212.158.39` (host `bcsoft`) | production เดิม | **เข้าไม่ได้ตั้งแต่ 2026-09-02** (ping/ssh/https ตาย) |

กฎ: DB/Kafka data **disposable ทุก env จนกว่าจะ go-live** (ดู `18-decisions-and-agreements.md`)

## DEV (เครื่องลุงจืด)

- backend stack: `cd backend && docker compose -f docker-compose.yml -f docker-compose.local.yml up -d` → `mongodb` (mongo:7, replica set `rs0` — จำเป็นสำหรับ transaction/outbox), `postgres` (18-alpine, 127.0.0.1:5432, DB ต่อ holding), `kafka` (confluent-local, 127.0.0.1:9092), `redis`, `minio` (S3 127.0.0.1:9100, console :9001, bucket จาก `minio.local.env`), `mainapi` (:8888, DEV_API_MODE=2, mount `bootstrap.local.json` + `custom_config.local.json`)
- **ClickHouse ถูกพักตั้งแต่ 2026-09-06** (ลบออกจาก compose local; volume `backend_clickhouse-data` ยังอยู่) — เหตุผลใน `15-known-issues.md` / `docs/handoff/HANDOFF-2026-09-06.md`
- frontend: `cd frontend && npm run dev` (Next 16, Turbopack, :3000) — ถ้าเห็น UI เก่าให้เช็ค `netstat -ano | grep :3000` ว่ามี `next start` ค้างไหม
- env ที่ frontend ต้องมี (ชื่อเท่านั้น): `BCAI_LOCAL_BACKEND_URL=http://localhost:8888`, `NEXT_PUBLIC_GOOGLE_CLIENT_ID`; Dev Login: `BC_ENV=dev`, `BCAI_DEV_LOGIN_ENABLED`, `BCAI_DEV_LOGIN_SECRET` (≥32 ตัวอักษร) ต้องตรงกันระหว่าง Windows User env (ที่ `next dev` สืบทอด) กับ container `mainapi` — ไม่ตรง = 401 ใน ~20µs; 404 = backend ไม่ได้ register `/dev-login`; 503 = ฝั่ง frontend ไม่มี env
- Demo Login (ไม่ต้องมี secret): ปุ่ม "ทดลองใช้ระบบ (Demo)" → `POST /api/auth/demo-login` → holding `demo` (บริษัท C01–C03) — ใช้ UAT
- เครื่องมือเสริมบนเครื่องนี้: `D:\mongomodel` (MongoModel MCP, container port 3100 — ถ้า port ค้าง TIME_WAIT ให้ `docker-compose down` รอ `netstat` ว่างก่อน `up`), `codebase-memory-mcp` (knowledge graph ของ repo, re-index ด้วย `codebase-memory-mcp.exe cli index_repository '{"repo_path":"D:/bccode"}'`), Stagehand UAT harness ที่ `scratch/stagehand-uat/`
- **Go build/test บน Windows ตรง ๆ ไม่ได้** (CGO/librdkafka) → ใช้ container `golang:1.26` + `librdkafka-dev` (คำสั่งเต็มใน `11-testing-quality.md` / `docs/handoff/HANDOFF-2026-09-06.md`)

## on-prem `192.168.2.202`

- ทุกอย่างเป็น Docker บน network `bc-backend_app-network`: `postgres` (18), `clickhouse` (25.5), `mongodb` (7, replica set `rs0` + keyFile), `minio` (bucket `app-images`), `mainapi` :8888, `redis`, `kafka`; frontend container `bc-frontend` :3000 build จาก `frontend/Dockerfile.onprem` (**ห้าม `output: standalone`**)
- deploy backend = tar source (ไม่รวม `bootstrap.json`/`custom_config.json` ของเซิร์ฟเวอร์) → ssh → `/home/smlsoft/bc-backend` → `docker compose up -d --build --no-deps --force-recreate mainapi` → เช็ค `/healthz` 200 และ md5 ของ `bootstrap.json` ไม่เปลี่ยน
- Cloudflare Tunnel `bcaicloud` (systemd `cloudflared`, token = secret) เคยให้บริการ `app.bcaicloud.com` — **โดเมนนี้เลิกใช้ 2026-06-28**; tunnel ยัง active ใช้ซ้ำกับโดเมนใหม่ได้; โซน `bcaicloud.com` อยู่บน Cloudflare NS แบบ Full
- ไม่มี port-forward (CGNAT) — เข้าจาก LAN/tunnel เท่านั้น

## PROD DigitalOcean `159.223.43.229`

- spec: 4 vCPU / 7.8 GB RAM / 160 GB; Ubuntu 24.04; Docker 29 + compose v5; Caddy 2.11; ufw 22/80/443; fail2ban; swap 4 GB; ssh key-only (`deploy/account/sshd-hardening.conf` → `/etc/ssh/sshd_config.d/00-bcai-hardening.conf`); ผู้ใช้เพิ่ม `smlsupport`, `goh` (sudo + docker)
- แอป: `/opt/bcai-account/deploy` = สำเนา `deploy/account/`; รันด้วย `docker compose -p bcai-account --env-file /etc/bcai-account/release.env -f compose.yml -f compose.8gb.yml up -d` (9 services); config ที่ `/etc/bcai-account/*.env` + `/var/lib/bcai-account/config` (secret ทั้งหมดอยู่ที่นั่น)
- **มี stack อื่นอยู่บนเครื่องเดียวกัน (`bcmk-*`, `bctms-*`) — ห้าม `docker compose down` แบบไม่ระบุ project**
- release = build image local → `docker save | ssh docker load` → แก้ `FRONTEND_IMAGE`/`MAINAPI_IMAGE` ใน `release.env` (เก็บ `.bak-<date>`) → `up -d --no-deps <service>`; rollback = คืน `release.env` แล้ว `up` ซ้ำ (image เก่ายังอยู่บนเครื่อง)
- build args ของ frontend ที่ **ต้องใส่ทั้งคู่**: `NEXT_PUBLIC_GOOGLE_CLIENT_ID`, `BCAI_LOCAL_BACKEND_URL=http://mainapi:8888` (`.env*` ถูก dockerignore)
- Demo login เปิดใน prod (`BCAI_DEMO_LOGIN_ENABLED=true` ใน backend.env); seed holding demo ด้วย `SEED_BASE=<url> node scripts/seed-demo.mjs`
- RAM 8 GB < ผลรวม mem_limit ของ compose (~9 GB) → ใช้ `compose.8gb.yml` ลดเพดาน; ClickHouse ยังจอง 768 MB ทั้งที่ไม่ได้ใช้ (รอตัดสินใจถอด)

## Google Sign-In

Login จริงใช้ Google Identity Services → ID token → `frontend/src/app/api/auth/google/verify` (ตรวจ aud/iss/exp/email_verified กับ tokeninfo) → mainapi `/googlelogin`; `/googlelogin` **ไม่รับแค่ email แล้ว** (security fix 2026-06-21) — ต้องมี credential จริง หรือใช้ demo/dev login แทน

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ

- สถานะปัจจุบันของ .202 ทั้งหมด (ssh timeout)
- DNS `account.bcaicloud.com` ชี้ไป 159.223.43.229 แล้วหรือยัง; ข้อมูลจาก prod เก่าย้ายหรือทิ้ง (disposable)
- Backup/RPO/RTO ของ prod — ยังไม่มีคำตอบจากลุงจืด (`docs/runbooks/RECOVERY-READINESS.md` รอข้อมูลนี้)
