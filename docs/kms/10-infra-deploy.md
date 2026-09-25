# โครงสร้างพื้นฐานและการ deploy (Infra & Deploy)
> ตรวจล่าสุด: 2026-09-25 — ผู้เขียน: AI reader; ทุกข้อเท็จจริงอ้าง path:line ตรวจกับซอร์สจริงในรอบนี้
> **สถาปัตยกรรมเปลี่ยนใหญ่ 2026-09-23**: ถอด MongoDB, Kafka, Redis และ ClickHouse ออกจากระบบทั้งหมด — เหลือ **PostgreSQL ตัวเดียว** (+ MinIO สำหรับไฟล์/รูปภาพ) ดูเหตุผลและรายการที่ถอดที่ ADR [`decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md`](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md) — **ห้ามเพิ่มสี่ตัวนี้กลับเข้าระบบ**

## 1. ภาพรวม
ระบบมี binary หลักตัวเดียว (`backend/main.go`, 192 บรรทัด) รันเป็น process เดียว ไม่มีการแยกโหมดทำงานแล้ว — env `DEV_API_MODE` ยังถูกตั้งใน Dockerfile/compose/`provision-server.sh` และมี mapping ใน loader ทั้งสองตัว (§6) แต่ไม่มีโค้ดใดอ่านค่านี้ (grep `DEV_API_MODE` ใน `backend/**/*.go` เจอแค่ mapping) มี 2 ชุด compose ที่ใช้งานจริง: local dev (`backend/docker-compose.yml` + overlay `backend/docker-compose.local.yml`) และ production (`deploy/account/compose.yml` + override `deploy/account/compose.8gb.yml`). Frontend เป็น Next.js image แยก (`frontend/Dockerfile`) ที่ proxy `/backend/*` ไป mainapi ผ่าน `BCAI_LOCAL_BACKEND_URL` (ค่าตั้งต้นเมื่อไม่ส่ง: production = `http://mainapi:8888`, dev = `http://localhost:8888` — frontend/next.config.ts:19, rewrite ที่บรรทัด 71).

### สิ่งที่ `main.go` ทำตอน start (backend/main.go:76-181)
- `setupconfig.LoadBootstrapConfig()` โหลด `bootstrap.json`/`custom_config.json` ทับ env (บรรทัด 77; ดู §6)
- เปิดฐานควบคุมกลางด้วย `mypg.PgSqlFastConnect(mcptoken.ControlDatabase)` (main.go:90) แล้วสร้าง `microservice.NewCacher(controlDB)` เก็บ session/token ในตาราง PostgreSQL `cache_entries` (main.go:94-98; ตารางสร้างเองด้วย `CREATE TABLE IF NOT EXISTS` — backend/pkg/microservice/cacher.go:45-53, `NewCacher` บรรทัด 62) — งานพื้นหลังล้าง session หมดอายุทุก 10 นาที (main.go:99-112)
- ตรวจสิทธิ์สดทุกคำขอผ่าน `microservice.NewAuthService(...)` ที่ผูกกับ `controlDB` เดียวกัน (main.go:118; ตรรกะจริงอยู่ที่ `backend/pkg/microservice/live_authorization.go:37-150`) — ผู้ใช้ถูกปิด/ถูกถอดจากกลุ่มกิจการ/บริษัท-สาขาถูกปิด → session เดิมใช้ไม่ได้ทันที
- ลงทะเบียน HTTP module ของ mainapi (authentication, shop + shop member, employee, fixedasset, generalledger, mcptoken, company, branch, businesstype, rolepermission — main.go:153-165 และ media upload ที่บรรทัด 166) แล้ว mount goapi ทั้งชุดใต้ `/goapi/*` (main.go:168-178, log ยืนยัน "GoAPI routes registered under /goapi/\*" ที่บรรทัด 176)
- จบด้วย `ms.Start()` (main.go:181)

## 2. Local stack (Docker Desktop)
คำสั่งเริ่ม: `cd backend && docker compose -f docker-compose.yml -f docker-compose.local.yml up -d --build`

| service | image/build | port (host) | volume | อ้างอิง |
|---|---|---|---|---|
| mainapi | build `Dockerfile.local` (overlay ทับ `Dockerfile`) | 0.0.0.0:8888 | mount `bootstrap.local.json`, `custom_config.local.json` | backend/docker-compose.yml:21-65, backend/docker-compose.local.yml:90-106 |
| postgres | postgres:18-alpine (db `appdb`) | 127.0.0.1:5432 | postgres-data (mount `/var/lib/postgresql`) | backend/docker-compose.local.yml:10-25 |
| minio | minio/minio (pinned digest), console :9001 | 127.0.0.1:9100→9000, 9001 | minio-data, mem 1g, cap_drop ALL | backend/docker-compose.local.yml:26-54 |
| minio-init | minio/mc one-shot: mb + versioning + policy `bcai-account-app` + app user | — | mount `deploy/account/minio-app-policy.json` | backend/docker-compose.local.yml:55-89 |

ข้อสังเกตจากไฟล์ปัจจุบัน:
- Base compose (`backend/docker-compose.yml`) มี service เดียวคือ `mainapi` — คอมเมนต์หัวไฟล์บอกว่า PostgreSQL เป็น "NATIVE (systemd) — ไม่อยู่ Docker" (บรรทัด 14) แต่ overlay local เพิ่ม `postgres` เป็น container จริงเสมอ (docker-compose.local.yml:10-25) → คอมเมนต์นี้ล้าสมัยสำหรับ workflow dev ปัจจุบัน
- `mainapi` local รัน `user: root` (backend/docker-compose.yml:35) ทั้งที่ image ตั้ง `USER appuser` (backend/Dockerfile:46, uid/gid 10001 สร้างที่บรรทัด 30)
- `Dockerfile.local`, `bootstrap*.json`, `custom_config*.json`, `*.env` ถูก gitignore (backend/.gitignore:62-66, root `.gitignore:11-13) — เครื่องใหม่ต้องสร้างเองก่อน `up`
- base compose ยังส่ง env `R2_*` ให้ mainapi (backend/docker-compose.yml:52-57) ควบคู่กับ `env_file: storage.local.env` (บรรทัด 26-28) — persister อ่าน `S3_*` ก่อนแล้ว fallback `R2_*` (§7)
- `docker-compose.dev.yml` เป็น overlay เล็กสำหรับ hot-reload ไฟล์ภาษา/ที่อยู่ (mount `assets/language`, `assets/address` เป็น read-only — backend/docker-compose.dev.yml:1-12)

## 3. Production compose (`deploy/account/compose.yml`, project `bcai-account`)
| service | image | mem_limit | healthcheck | env_file | depends_on | อ้างอิง |
|---|---|---|---|---|---|---|
| postgres | postgres:18-alpine (digest) | 1024m | pg_isready | `/etc/bcai-account/postgres.env` | — | compose.yml:10-24, compose.8gb.yml:4-5 |
| minio | minio/minio (digest), cpus 1.0 | 512m | /minio/health/ready | `/etc/bcai-account/minio.env` | — | compose.yml:26-50, compose.8gb.yml:6-7 |
| minio-init | minio/mc one-shot (เหมือน local) | 256m | — | `/etc/bcai-account/minio.env` | minio healthy | compose.yml:52-83 |
| mainapi | `${MAINAPI_IMAGE}` | 1536m | wget /healthz (start 60s) | `/etc/bcai-account/backend.env` | postgres healthy, minio-init completed | compose.yml:86-110, compose.8gb.yml:8-9 |
| frontend | `${FRONTEND_IMAGE}` | 768m | wget :3000/ | `/etc/bcai-account/frontend.env` | mainapi healthy | compose.yml:113-133, compose.8gb.yml:10-11 |

ชื่อ project `bcai-account` มาจาก `name:` ที่ compose.yml:1. ตัวเลข `mem_limit` ใน `compose.8gb.yml` (postgres 1024m, minio 512m, mainapi 1536m, frontend 768m — compose.8gb.yml:4-11) **ตรงกับค่าใน `compose.yml` เองทุกตัว** (compose.yml:23, 40, 110, 133) — override นี้จึงไม่ได้เปลี่ยนค่าใดจากฐานแล้ว (vestigial) แต่ `tools/fast-deploy.py` ยังส่ง `-f compose.8gb.yml` เสมอ (fast-deploy.py:216). backend มี service เดียวคือ `mainapi` — ไม่มี service migration/worker แยกแล้วตั้งแต่ 2026-09-23

- Network: `edge` (bridge) และ `data` (`internal: true`) (compose.yml:136-141); เฉพาะ mainapi อยู่ทั้งสอง (compose.yml:100), frontend อยู่ `edge` เท่านั้น (compose.yml:123). Port ออก host มีแค่ `127.0.0.1:8888` (mainapi, compose.yml:96-97) และ `127.0.0.1:3200→3000` (frontend, compose.yml:121-122)
- `cap_drop: [ALL]` + `no-new-privileges` มีทุก service ยกเว้น postgres (compose.yml:47-49, 81-82, 102-103, 125-126); `init: true` มีเฉพาะ mainapi/frontend (compose.yml:95, 120); log json-file 10m×3 ผ่าน anchor ทุก service (compose.yml:3-7)
- Caddy: site `account.bcaicloud.com` → `reverse_proxy 127.0.0.1:3200`, ตอบ 404 ให้ path `/backend/goapi/api/health/{background,queue*,database,system}` (deploy/account/Caddyfile.account:1-12) — ในจำนวนนี้ `/api/health/background` **ไม่มี route จริงใน goapi ปัจจุบัน** (goapi ลงทะเบียนแค่ `/api/health`, `/api/health/queue[/:holdingcode]`, `/api/health/database`, `/api/health/system` — backend/internal/goapi/bootstrap.go:323-335) จึง 404 อยู่แล้วโดยธรรมชาติ ไม่ต่างจากที่ Caddy บล็อกไว้; route เหล่านี้ลงทะเบียนบน group `g` ตรง ๆ ไม่ผ่าน `authGroup` (bootstrap.go:308 เทียบ 323-335) และ mainapi ก็ยกเว้น `/goapi/*` จากการตรวจ auth ทั้งกลุ่ม (main.go:133)

## 4. สคริปต์บนเซิร์ฟเวอร์ (`deploy/account/`)
`provision-server.sh` (ต้อง root, ต้องมี env `MAINAPI_IMAGE`/`FRONTEND_IMAGE`) ทำตามลำดับ (deploy/account/provision-server.sh):
1. สร้าง `/etc/bcai-account` 0700, `/opt/bcai-account` 0750, `/var/lib/bcai-account/config` owner 10001 (บรรทัด 22-24)
2. ถ้ายังไม่มี `secrets.env` → gen `POSTGRES_PASSWORD`, `JWT_SECRET_KEY`, `RELOAD_CONFIG_SECRET` แล้วเติม `MINIO_ROOT_USER/PASSWORD`, `MINIO_APP_ACCESS_KEY/SECRET_KEY` ถ้าขาด (27-55)
3. สร้าง `bootstrap.json` และ `custom_config.json` เป็น `{}` ถ้ายังไม่มี, chown 10001, chmod 0600 (57-64)
4. เขียน env ของ dependency 2 ไฟล์: `postgres.env` (db `bcai_projection`, user `bcai`) และ `minio.env` (bucket `bcai-account`, `MINIO_BROWSER=off`) (66-79)
5. เขียน `backend.env` — key ที่ตั้ง: `MODE/GO_ENV/BC_ENV=production`, `DEV_API_MODE=2` (ไม่มีโค้ดอ่าน — ดู §1), `SERVICE_PORT`, `LOG_LEVEL=INFO`, `TZ`, `HOST_API`, `HTTP_CORS`, `POSTGRES_*` (ทั้ง `USER`/`USERNAME`, `DB`/`DATABASE`/`DB_NAME` เพราะ mainapi กับ goapi อ่านคนละชื่อ — internal/config/config_postgresql.go:32,40 เทียบ internal/goapi/config/config.go:40-53), `JWT_SECRET_KEY`, `RELOAD_CONFIG_SECRET`, `GOOGLE_CLIENT_ID`, `BCAI_DEV_LOGIN_ENABLED=false`, `BCAI_DEMO_LOGIN_ENABLED=true`, `BCAI_DEMO_USERNAME`, `GODEBUG`, `S3_ENDPOINT/REGION/ACCESS_KEY_ID/SECRET_ACCESS_KEY/BUCKET_NAME/FORCE_PATH_STYLE`, `STORAGE_ALLOW_PRESIGNED_URL=false` (81-115)
6. เขียน `frontend.env` (`NODE_ENV`, `BCAI_LOCAL_BACKEND_URL=http://mainapi:8888`, `NEXT_PUBLIC_GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_ID`, `BCAI_DEV_LOGIN_ENABLED=false`) และ `release.env` (`MAINAPI_IMAGE`, `FRONTEND_IMAGE`) แล้ว chmod 0600 ทุก env (117-130)

`rotate-postgres-password.sh`: gen รหัสใหม่ → `ALTER ROLE bcai` ผ่าน `compose exec postgres psql` (บรรทัด 26-34) → เขียนทับใน `secrets.env` (37-47) → รัน provision ใหม่ ซึ่งเขียน `postgres.env`/`backend.env` ด้วยรหัสใหม่ (54) → `docker compose rm -sf mainapi worker migrate` (56) → recreate postgres (57) → `up -d migrate` แล้วรอ `exited:0` (58-77) → `up -d --no-deps mainapi worker` แล้วรอ healthy (79-102). **สคริปต์นี้ใช้ไม่ได้กับ compose ปัจจุบัน:** `compose.yml` ไม่มี service `worker`/`migrate` แล้ว — ทดสอบกับ Docker Compose v5.3.1 บนโปรเจกต์ทดลอง: `docker compose rm -sf mainapi worker migrate` ตอบ `no such service: worker` exit 1 และ **ไม่แตะ `mainapi` เลย** (container ยัง running) → `set -euo pipefail` หยุดสคริปต์ที่บรรทัด 56 หลังจากรหัสใน PostgreSQL, `secrets.env` และ env files ถูกเปลี่ยนไปแล้ว แต่ `mainapi` ยังรันด้วย env รหัสเก่า (อนุมานจากโค้ด ยังไม่ได้ทดสอบบน prod: connection ใหม่ของ `mainapi` จะ login PostgreSQL ไม่ผ่านจนกว่าจะ recreate `mainapi`) — ห้ามรันจนกว่าจะแก้สคริปต์ให้เหลือแค่ `mainapi` (R1)

`sshd-hardening.conf`: ปิด password/kbd-interactive, `PermitRootLogin prohibit-password`, `MaxAuthTries 3`, `LoginGraceTime 30` (deploy/account/sshd-hardening.conf:3-8). `minio-app-policy.json`: สิทธิ์ app user จำกัดที่ bucket `bcai-account` (list + get/put/delete/multipart) (deploy/account/minio-app-policy.json:6-26).

## 5. Dockerfiles
| ไฟล์ | base / build | entry | สถานะ | อ้างอิง |
|---|---|---|---|---|
| backend/Dockerfile | golang:1.26-alpine, `CGO_ENABLED=0` (Pure Go), BuildKit cache mounts → alpine:3.21 uid/gid 10001, `DEV_API_MODE=2` (ไม่มีโค้ดอ่าน), `SERVICE_PORT=8888`, HEALTHCHECK `/healthz` start 120s | `/app/go-app` (`main.go`) | LIVE (prod image `MAINAPI_IMAGE`) | backend/Dockerfile:7-53 |
| backend/Dockerfile.local | เหมือนบนแต่ไม่มี `--mount=type=cache` (เครื่อง dev ไม่มี buildx); uid ไม่ fix | เดียวกัน | LIVE local เท่านั้น, **ไม่ tracked** (gitignore) | backend/Dockerfile.local:1-5, 18; backend/.gitignore:65 |
| frontend/Dockerfile | node:24.18.0-alpine 5 stage (base/deps/prod-deps/builder/runner); build args `NEXT_PUBLIC_GOOGLE_CLIENT_ID`, `BCAI_LOCAL_BACKEND_URL`; runner uid 1001 `npm run start` :3000 (ไม่ใช่ standalone) | `npm run start` | LIVE (prod `FRONTEND_IMAGE`) | frontend/Dockerfile:2, 25-30, 39-55 |

Dockerfile อื่นใน `backend/` และ `backend/Makefile` ถูกลบพร้อม `backend/cmd/*` microservice เก่าเมื่อ 2026-09-23 (`ls backend/Dockerfile*` เหลือ 2 ไฟล์ข้างต้น; ดู `14-cmd-tools-scripts.md` §1–§2)

## 6. การโหลด config (bootstrap.json → env)
มี loader **สองตัว** ที่ทำงานตามลำดับใน process เดียว (ไม่มี section ของระบบที่ถอดไปแล้ว 2026-09-23 เหลือใน `configMapping` ของทั้งสองไฟล์):
1. mainapi: `setupconfig.LoadBootstrapConfig()` ก่อน `config.NewConfig()` (backend/main.go:77). ค้นหาไฟล์ตามลำดับ `/app/bootstrap/bootstrap.json` → `/app/bootstrap.json` → `bootstrap.json` → `config/bootstrap.json` (backend/internal/setupconfig/loader.go:71-76), merge `custom_config.json` ข้าง ๆ ทับ (loader.go:78-83, 85-109), แล้ว `os.Setenv` ทับ env เดิมโดยไม่เช็คใน `applyBootstrapSection` (loader.go:178-213, `os.Setenv` ที่บรรทัด 200) → ค่าใน bootstrap ชนะ env_file เสมอ. `configMapping` ปัจจุบันมีแค่ 4 section: `postgresql`, `service`, `integrations` (Gemini API key/model เท่านั้น), `storage` (STORAGE_DATA_PATH/URI, Azure, S3_*) (loader.go:20-56)
2. goapi: `setupconfig.LoadBootstrapConfig()` ของตัวเองเรียกใน `Init()` (backend/internal/goapi/bootstrap.go:45, 49) path list เดียวกัน (backend/internal/goapi/setupconfig/loader.go:116-119) แต่ mapping กว้างกว่า (AI provider keys, `r2*`/`s3*` ใต้ `integrations` บรรทัด 73-83, `CORS_ALLOWED_ORIGINS` บรรทัด 61) (goapi/setupconfig/loader.go:42-83)

**กับดักที่ยังอยู่:** loader ของ goapi ตัด `_` ออกจาก key ก่อน lookup (`strings.ReplaceAll(rawKey, "_", "")` — goapi/setupconfig/loader.go:233) แต่ loader ของ mainapi ใช้ key ดิบไม่มีตัวคั่น เช่น `dbname`, `serviceport`, `devapimode` (backend/internal/setupconfig/loader.go:21-36) — ถ้า `bootstrap.json` ใช้ snake_case (`db_name`) mainapi loader จะข้ามเงียบ ๆ ต่างจาก goapi ที่ normalize ให้ก่อน — จุดนี้ยังไม่ถูกแก้ ยังเป็นความเสี่ยงเดิม

Reload runtime: mainapi เปิด `POST /reload-config` (public path, main.go:131, handler main.go:138-146) ตรวจ `RELOAD_CONFIG_SECRET` แล้วเรียก `setupconfig.ReloadConfig()`; goapi ฝั่งเขียน `UpdateBootstrapJSON` แล้วเรียก mainapi ให้ reload — บน prod ไฟล์เป็น `{}` และ owner 10001 จึงเขียนได้จาก container

## 7. Object storage (รูปภาพ/ไฟล์)
- mainapi: `NewFilePersister()` คืน persister S3/MinIO ตัวเดียว (backend/pkg/microservice/persister_factory.go:4) อ่าน `S3_ACCOUNT_ID|R2_ACCOUNT_ID`, `S3_ENDPOINT|R2_ENDPOINT`, `S3_ACCESS_KEY_ID|R2_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY|R2_SECRET_ACCESS_KEY`, `S3_BUCKET_NAME|R2_BUCKET_NAME`, region default `us-east-1` (persister_file_r2.go:51-68) — ถ้ามี endpoint จะใช้ BaseEndpoint (MinIO path-style)
- goapi: `handlers.InitR2Client()` (backend/internal/goapi/bootstrap.go:52) ใช้ชุด env เดียวกัน + `S3_FORCE_PATH_STYLE|R2_FORCE_PATH_STYLE` (backend/internal/goapi/handlers/image_r2.go:75, `InitR2Client` บรรทัด 176); presigned URL ตรงต้องเปิด `STORAGE_ALLOW_PRESIGNED_URL=true` (image_r2.go:148, handlers/storage_private.go:14) — prod ตั้ง `false`
- bucket: local จาก `minio.local.env` (`MINIO_BUCKET_NAME`) และ prod = `bcai-account` เปิด versioning ผ่าน minio-init ใช้ app user สิทธิ์จำกัด (minio-app-policy.json) — PostgreSQL เก็บแค่ URI ของไฟล์ตามกฎใน AGENTS.md ไม่เก็บ binary ในฐานข้อมูล

## 8. ชุดตรวจ (local) — ไม่มี CI ฝั่ง GitHub แล้ว
- `tools/verify.sh` (แทน `.github/workflows/ci.yml` ที่ถูกลบ 2026-09-09) มี target เดี่ยว 5 ตัว: `codemap`, `frontend` (lint/typecheck/vitest), `frontend-build`, `backend` (compile + unit test ยกเว้น `backend/.ci/test-quarantine.txt` ใน container `golang:1.26`) และ `postgres` (integration tests กับ PostgreSQL 18 ชั่วคราว) + ชุดรวม 2 ตัว: `fast` (ค่าเริ่มต้น) = codemap + frontend, `all` = ทั้ง 5 ตัว (tools/verify.sh:189-203) — `backend/.ci/` เหลือแค่ `test-quarantine.txt`
- `t_postgres` สร้าง container `postgres:18-alpine` แยก (`POSTGRES_HOST_AUTH_METHOD=trust`) แล้วรัน `go test -tags=integration ./pkg/... ./internal/...` โดยตั้ง `BC_GL_TEST_POSTGRES_DSN`/`GL_AUTH_TEST_DSN`/`BC_TAXFORM_TEST_POSTGRES_DSN` ชี้ container นั้น แล้วลบทิ้งเสมอ (tools/verify.sh:154-175)
- **ไม่มี CI/CD ฝั่ง GitHub เลย** ตั้งแต่ 2026-09-09 (มติลุงจืด: GitHub = ที่เก็บโค้ดอย่างเดียว) → หลักฐาน build/test ทั้งหมดเป็น local เท่านั้น ต้องรัน `npm run verify`/`npm run verify:all` เอง

## 9. เซิร์ฟเวอร์ที่รู้จัก (ไม่มี secret)
| เครื่อง | บทบาท | อ้างอิง |
|---|---|---|
| dev Windows (Docker Desktop) | local stack §2 + `next dev` :3000 | backend/docker-compose*.yml |
| DigitalOcean SGP1 159.223.43.229 (4 vCPU/8 GB) | prod สำหรับ `account.bcaicloud.com` ด้วย compose.yml + compose.8gb.yml, deploy ผ่าน `tools/fast-deploy.py` | tools/fast-deploy.py:25, 216; deploy/account/compose.yml |

เครื่องอื่นที่เอกสารรุ่นก่อนเคยกล่าวถึง (192.168.2.202, 188.212.158.39) ไม่มีหลักฐานสถานะปัจจุบันใน repo — ไม่ยืนยัน

## 10. โครง runbook deploy (สกัดจากไฟล์ — ทุกขั้นที่แตะ prod เป็น R0 ต้องถามลุงจืดก่อน)
1. **Build image บนเครื่อง dev** (ต้องมี BuildKit สำหรับ `backend/Dockerfile`): `docker build -f backend/Dockerfile -t bcai-account-mainapi:rYYYYMMDD-N backend` และ `docker build --build-arg NEXT_PUBLIC_GOOGLE_CLIENT_ID=… --build-arg BCAI_LOCAL_BACKEND_URL=http://mainapi:8888 -t bcai-account-frontend:rYYYYMMDD-N frontend`
2. **ครั้งแรกบนเซิร์ฟเวอร์**: copy `deploy/account/*` ไป `/opt/bcai-account/deploy`, วาง `sshd-hardening.conf` ใน `/etc/ssh/sshd_config.d/` ให้เรียงก่อน cloud-init, รัน `MAINAPI_IMAGE=… FRONTEND_IMAGE=… ./provision-server.sh` (สร้าง secrets/env/bootstrap `{}`), ติดตั้ง `Caddyfile.account` ใน Caddy
3. **ส่ง image**: ไม่มี registry ส่วนกลาง → สตรีมผ่าน `docker save <image> | ssh -C root@<host> docker load` แล้วแก้ `MAINAPI_IMAGE`/`FRONTEND_IMAGE` ใน `/etc/bcai-account/release.env` (สำรองไฟล์เดิมไว้ก่อน) — งานประจำใช้ `tools/fast-deploy.py` แทนการทำมือทุกขั้น
4. **up** (ใน `/opt/bcai-account/deploy`): `docker compose --env-file /etc/bcai-account/release.env -f compose.yml -f compose.8gb.yml up -d` — ลำดับตาม `depends_on` §3: postgres → minio → minio-init → mainapi (รอ postgres healthy + minio-init completed) → frontend (รอ mainapi healthy); ไม่มีขั้น migration แยก — backend สร้างโครงสร้างเองด้วย `CREATE ... IF NOT EXISTS` (เช่น `cache_entries` ตอน start — cacher.go:45-53). งานประจำ `fast-deploy.py` สั่ง `up -d --no-deps` เฉพาะ `frontend` หรือ `mainapi frontend` แล้วรอ healthy (fast-deploy.py:230-256)
5. **ตรวจรับ**: รอ healthcheck `healthy` ทุกตัว (`docker compose ps`), `curl 127.0.0.1:8888/healthz` และ `127.0.0.1:3200/`, ดู `docker logs` ของ mainapi ว่ามีบรรทัด `GoAPI routes registered under /goapi/*` (backend/main.go:176) และ `[Bootstrap]` ไม่ฟ้อง
6. **Rollback**: คืน `release.env` เดิม (`fast-deploy.py` สำรองไว้ที่ `/opt/bcai-account/releases/<tag>/release.env.before` พร้อม `pg_dumpall` ใน `backup/` — fast-deploy.py:119-151) แล้ว `up -d --no-deps mainapi frontend` (image เก่ายังอยู่บนเครื่อง); DB data ถือว่า disposable ก่อน go-live
7. **หมุนรหัส PG**: `rotate-postgres-password.sh` — **ห้ามใช้จนกว่าจะแก้** (ทดสอบแล้วว่าหยุดกลางทางที่บรรทัด 56 — ดู §4)

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ
- สถานะจริงของ 159.223.43.229 ปัจจุบัน (image tag ล่าสุด, Caddyfile/DNS ตรงกับไฟล์ใน repo หรือไม่) — งานนี้ไม่ได้ ssh เข้าเครื่องใด อ่านจาก repo อย่างเดียว
- ไม่ได้รัน `docker build` ทั้ง backend/frontend ในรอบนี้ — ยืนยันแค่คำสั่ง build ของ `backend/Dockerfile:22` ด้วย Go บนเครื่อง dev (2026-09-25: `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" main.go` ผ่าน ได้ binary ~55 MB)
- `rotate-postgres-password.sh` ต้องแก้ให้เหลือแค่ `mainapi` (ตัดขั้น `migrate`/`worker`) ก่อนใช้งานครั้งต่อไป (§4 — ยืนยันแล้วว่าหยุดกลางทาง; R1 เพราะเป็นสคริปต์บนเครื่อง prod)
- `docs/runbooks/RECOVERY-READINESS.md` ยังอธิบาย backup/restore ของระบบที่ถอดไปแล้ว — ต้องเขียนใหม่ให้ตรงกับ PostgreSQL + MinIO
- Backup/restore: มีแค่ preflight ใน `fast-deploy.py` (`pg_dumpall` + tar config — fast-deploy.py:147-150) ไม่มี backup ของ MinIO และยังไม่มีหลักฐานว่าเคยทดสอบ restore จริง
