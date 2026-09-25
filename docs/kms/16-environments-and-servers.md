# สภาพแวดล้อมและเซิร์ฟเวอร์ (dev / production)
> ตรวจล่าสุด: 2026-09-25 (commit c58f0b62) — PostgreSQL ตัวเดียว: MongoDB, Kafka, Redis, ClickHouse ถูกถอดออกจากระบบเมื่อ 2026-09-23 ([ADR](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md)) ห้ามเพิ่มกลับทุก environment; ค่าที่เป็น secret ไม่บันทึกที่นี่ (อยู่ใน env/bootstrap บนเครื่องนั้น ๆ)

## ภาพรวม

| ระบบ | ที่อยู่ | ใช้ทำอะไร |
|---|---|---|
| **DEV** | เครื่องลุงจืด (Windows 11, Docker Desktop) | พัฒนาเร็ว ข้อมูลทิ้งได้ |
| **PROD** | DigitalOcean SGP1 `159.223.43.229` (root, key-only ssh) | production `https://account.bcaicloud.com/` |

กฎ: DB **disposable ทุก env จนกว่าจะ go-live** (ดู `18-decisions-and-agreements.md`)

หมายเหตุ: on-prem `192.168.2.202` ที่เอกสารรุ่นก่อนเคยใช้เป็นสภาพแวดล้อมที่ 3 ไม่มีสคริปต์หรือ compose ใดใน `tools/`/`deploy/` อ้างถึงแล้ว — ในโค้ดเหลือเพียง `frontend/src/lib/backend-url.ts:134-144` (`migrateRuntimeBackendUrl`) ที่ถือว่า URL backend ที่ผู้ใช้เคยบันทึกไว้ซึ่งชี้ `192.168.2.202`/`dev.bcaicloud.com`/`api.bcaicloud.com` เป็นที่อยู่เก่า แล้วย้ายไปใช้ same-origin

## DEV (เครื่องลุงจืด)

- backend stack: `cd backend && docker compose -f docker-compose.yml -f docker-compose.local.yml up -d --build` → `postgres` (postgres:18-alpine, `127.0.0.1:5432`), `minio` + `minio-init` (S3 `127.0.0.1:9100`, console `127.0.0.1:9001`, สร้าง bucket/policy อัตโนมัติ), `mainapi` (build `Dockerfile.local`, `:8888`, mount `bootstrap.local.json` + `custom_config.local.json`, env จาก `storage.local.env`)
- frontend: `cd frontend && npm run dev` (Next 16, `next dev`, `:3000`) — ถ้าเห็น UI เก่าให้เช็คว่ามี `next start` ค้างพอร์ต 3000 อยู่หรือไม่
- env ที่ frontend ต้องมี (ชื่อเท่านั้น): `BCAI_LOCAL_BACKEND_URL=http://localhost:8888`, `NEXT_PUBLIC_GOOGLE_CLIENT_ID`
- Dev Login (`/api/auth/dev-login` → mainapi `/dev-login`): ฝั่ง backend ต้องตั้งครบ 4 ค่า — `BC_ENV=dev`, `BCAI_DEV_LOGIN_ENABLED=true`, `BCAI_DEV_LOGIN_USER_UID`, `BCAI_DEV_LOGIN_SECRET` (≥32 ตัวอักษร) ไม่ครบ = mainapi ไม่ลงทะเบียน `/dev-login` (`backend/internal/authentication/authentication_http.go:102-120,132-134`); ฝั่ง frontend ต้องมี `BCAI_DEV_LOGIN_ENABLED=true` + `BCAI_DEV_LOGIN_SECRET` ค่าเดียวกัน และเรียกจาก localhost เท่านั้น (404 = ปิด, 403 = ไม่ใช่ localhost, 503 = secret สั้น/ไม่มี — `frontend/src/app/api/auth/dev-login/route.ts:9-28`)
- Demo Login (ไม่ต้องมี secret): ปุ่ม "ทดลองใช้ระบบ (Demo)" → `POST /api/auth/demo-login` → เปิดด้วย `BCAI_DEMO_LOGIN_ENABLED=true` (ค่าเริ่มต้นใน `backend/docker-compose.yml:50`) + ชื่อผู้ใช้จาก `BCAI_DEMO_USERNAME` (default `demo`) (`backend/internal/demo/demo.go:12-25`)
- Go build/test บน Windows รันตรงได้ (backend เป็น Pure Go, `CGO_ENABLED=0` ใน `backend/Dockerfile:22`)

## PROD DigitalOcean `159.223.43.229`

- โดเมน: `https://account.bcaicloud.com/` (Caddy proxy → `127.0.0.1:3200` ตาม `deploy/account/Caddyfile.account`)
- ติดตั้งครั้งแรกด้วย `deploy/account/provision-server.sh` (สร้าง `/etc/bcai-account/*.env` + secret, `/opt/bcai-account`, `/var/lib/bcai-account/config`); ssh key-only ตาม `deploy/account/sshd-hardening.conf`
- แอป: `/opt/bcai-account/deploy` = สำเนา `deploy/account/`; รันด้วย `docker compose -p bcai-account --env-file /etc/bcai-account/release.env -f compose.yml -f compose.8gb.yml up -d` (`compose.8gb.yml` ลด `mem_limit` ให้พอดีเครื่อง 4 vCPU / 8 GB)
- **service 5 ตัว** (`deploy/account/compose.yml`): `postgres` (18-alpine), `minio`, `minio-init` (สร้าง bucket/policy แล้วจบ), `mainapi` (publish `127.0.0.1:8888`), `frontend` (publish `127.0.0.1:3200`)
- config ที่ `/etc/bcai-account/*.env` (`postgres.env`, `minio.env`, `backend.env`, `frontend.env`, `release.env`) + `/var/lib/bcai-account/config` (mount เป็น `/app/bootstrap` ของ mainapi) — secret ทั้งหมดอยู่ที่นั่น ไม่อยู่ใน repo
- **มี stack อื่นอยู่บนเครื่องเดียวกัน (`bcmk-*`, `bctms-*`) — ห้าม `docker compose down` แบบไม่ระบุ project**
- Deploy: คำสั่งเดียว `py tools/fast-deploy.py --all --tag rYYYYMMDD-N` จากเครื่อง dev (`SERVER_HOST` ที่ `tools/fast-deploy.py:25`): build image local → preflight backup (`pg_dumpall` + tar ของ `/etc/bcai-account`, `/var/lib/bcai-account/config`, `/opt/bcai-account/deploy` และสำเนา `release.env.before` ไว้ที่ `/opt/bcai-account/releases/<tag>/`) → `docker save | ssh -C docker load` → สลับ `MAINAPI_IMAGE`/`FRONTEND_IMAGE` ใน `release.env` แบบ atomic → `up -d --no-deps <service>` → รอ container healthy → เช็ค URL จริง → ลบ image เก่ากว่า 72 ชม. (`tools/fast-deploy.py:113-160,206-297`)
- Rollback (สคริปต์ไม่ทำให้อัตโนมัติ): คืน `/opt/bcai-account/releases/<tag>/release.env.before` เป็น `/etc/bcai-account/release.env` แล้ว `up -d --no-deps` ซ้ำ ภายใน 72 ชม. ที่ image เก่ายังอยู่
- build args ของ frontend ที่ต้องใส่ทั้งคู่: `NEXT_PUBLIC_GOOGLE_CLIENT_ID`, `BCAI_LOCAL_BACKEND_URL=http://mainapi:8888` (`tools/fast-deploy.py:97-98`; `.env*` ถูก `frontend/.dockerignore`)
- Demo login เปิดใน prod เช่นกัน (`BCAI_DEMO_LOGIN_ENABLED=true` ใน `backend.env` ตาม `deploy/account/provision-server.sh`) ให้ AI ทดสอบจอได้เองผ่านปุ่ม Demo (ดูกฎ "ทดสอบหน้าจอด้วยปุ่ม Demo" ใน `AGENTS.md`)
- MCP: `https://account.bcaicloud.com/mcp/gl` (Streamable HTTP, token จากหน้า `/mcp-tokens` ระดับ Holding)

## Google Sign-In

Login จริงใช้ Google Identity Services → ID token → `frontend/src/app/api/auth/google/verify/route.ts` (ตรวจ aud/iss/exp/email_verified กับ Google tokeninfo ก่อนเรียก mainapi) → mainapi `/googlelogin` ตรวจ credential ซ้ำอีกชั้น (`verifyGoogleIDToken`, `backend/internal/authentication/authentication_http.go:426`) และผูกบัญชีด้วย issuer+subject (`LoginWithGoogleIdentity`, `:392`) ไม่ใช่แค่ email ที่เบราว์เซอร์ส่งมา — ใช้ demo/dev login แทนได้ถ้าไม่มี Google credential จริง

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ

- schema PostgreSQL จริงบน prod เทียบกับ `centraldb.go`/`schema.sql`/`subledger.sql`/`budget.sql` เวอร์ชันปัจจุบัน — ยังไม่ได้ ssh ตรวจในรอบนี้ (ดู `02-data-stores.md` ช่องว่างข้อ 2)
- stack `bcmk-*`/`bctms-*` บนเครื่อง prod ยืนยันจากประสบการณ์ปฏิบัติงานเท่านั้น ไม่มีใน repo — ตรวจด้วย `docker ps` ก่อนทำงานที่กระทบทั้งเครื่อง
- Backup/RPO/RTO ของ prod แบบเป็นทางการ — ยังไม่มีคำตอบจากลุงจืด (`docs/runbooks/RECOVERY-READINESS.md` รอข้อมูลนี้)
