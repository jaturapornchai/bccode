# โครงสร้างพื้นฐานและการ deploy (Infra & Deploy)
> ตรวจล่าสุด: 2026-09-07 (commit d93a210d) — ผู้เขียน: AI reader; ทุกข้อเท็จจริงอ้าง path:line · ตรวจซ้ำโดย fact-checker

## 1. ภาพรวม
ระบบมี binary หลักตัวเดียว (`backend/main.go`) ที่สลับบทบาทด้วย env `DEV_API_MODE` และมี 3 ชุด compose ที่ใช้งานจริง: local dev (`backend/docker-compose.yml` + overlay `docker-compose.local.yml`), CI integration (`backend/.ci/projection.compose.yml`) และ production (`deploy/account/compose.yml` + override `compose.8gb.yml`). Frontend เป็น Next.js image แยก (`frontend/Dockerfile`) ที่ proxy `/backend/*` ไป mainapi ผ่าน `BCAI_LOCAL_BACKEND_URL` (frontend/next.config.ts:19-21, 74).

### โหมดของ binary (`DEV_API_MODE`)
| ค่า | หน้าที่ | หลักฐาน | ใช้โดย |
|---|---|---|---|
| `""` | รันทั้ง API และ consumers ใน process เดียว | backend/main.go:256, backend/main.go:654 | ไม่มี compose ไหนใช้ค่าว่าง |
| `"2"` | HTTP API (mainapi modules) + goapi embed ที่ `/goapi/*` + `/api/language/:lang` public | backend/main.go:256, backend/main.go:568-580 | local `mainapi` (backend/docker-compose.yml:80), prod `mainapi` (provision-server.sh:103), Dockerfile default (backend/Dockerfile:41) |
| `"1"` | legacy Kafka consumers (gorm → PostgreSQL), `CONSUMER_GROUP_NAME` default `"03"` | backend/main.go:654-662 | prod `worker` (deploy/account/compose.yml:275-276) |
| `"3"` | migrations อย่างเดียวแล้ว `return` | backend/main.go:585-652 | prod `migrate` (deploy/account/compose.yml:212) |

ทุกโหมดจบด้วย `ms.Start()` (backend/main.go:728) ซึ่งเปิด HTTP เมื่อมี route ≥ 1 (backend/pkg/microservice/microservice.go:255-260); โหมด 1 ลงทะเบียน `/healthz` (backend/main.go:658) จึงตอบ healthcheck ของ `worker` ได้ (deploy/account/compose.yml:301). goapi เริ่ม Kafka consumers ของตัวเองเมื่อ `ENABLE_KAFKA=true` (backend/internal/goapi/bootstrap.go:157-161) → prod จึงมี consumer 2 ชุดเขียน PG เดียวกัน (mainapi mode 2 + worker mode 1) ตามที่ docs/handoff/HANDOFF-2026-09-06.md:58 ระบุเป็น P3.

## 2. Local stack (Docker Desktop)
คำสั่งเริ่ม: `cd backend && docker compose -f docker-compose.yml -f docker-compose.local.yml up -d` (docs/handoff/HANDOFF-2026-09-06.md:19-22)

| service | image | port (host) | volume | อ้างอิง |
|---|---|---|---|---|
| redis | redis:7-alpine | 127.0.0.1:6379 | redis-data | backend/docker-compose.yml:29-38 |
| kafka | confluentinc/confluent-local (KRaft, auto-create topics) | 127.0.0.1:9092 (ภายใน `kafka:29092`) | kafka-data | backend/docker-compose.yml:40-57 |
| mainapi | build `Dockerfile.local` (overlay ทับ `Dockerfile`) | 0.0.0.0:8888 | mount bootstrap/custom_config/tdict/swagger | backend/docker-compose.yml:59-108, backend/docker-compose.local.yml:103-124 |
| mongodb | mongo:7 `--replSet rs0` | 127.0.0.1:27017 | mongo-data | backend/docker-compose.local.yml:11-21 |
| postgres | postgres:18-alpine (db `appdb`) | 127.0.0.1:5432 | postgres-data (mount `/var/lib/postgresql`) | backend/docker-compose.local.yml:23-37 |
| minio | minio/minio (pinned digest), console :9001 | 127.0.0.1:9100→9000, 9001 | minio-data, mem 1g, cap_drop ALL | backend/docker-compose.local.yml:39-66 |
| minio-init | minio/mc one-shot: mb + versioning + policy `bcai-account-app` + app user | — | mount `deploy/account/minio-app-policy.json` | backend/docker-compose.local.yml:68-101 |
| clickhouse | **ถอดออกจาก overlay แล้ว 2026-09-06** | — | volume `backend_clickhouse-data` ยังค้าง (runtime `docker volume ls`) | backend/docker-compose.local.yml:6, docs/handoff/HANDOFF-RISKS-2026-09-05.md:22 |

ข้อสังเกตจากไฟล์:
- Header ของ base compose บอกว่า Mongo/PG เป็น NATIVE systemd และ container เข้าผ่าน `host.docker.internal` (backend/docker-compose.yml:14-21, 72-73) แต่ overlay local เพิ่ม mongodb/postgres เป็น container และชี้ `bootstrap.local.json` แทน (backend/docker-compose.local.yml:120-124) — comment ใน base ล้าสมัยสำหรับเครื่อง dev
- `mainapi` local รัน `user: root` (backend/docker-compose.yml:76) ทั้งที่ image ตั้ง `USER appuser` (backend/Dockerfile:45); overlay `depends_on` minio-init `service_completed_successfully` (backend/docker-compose.local.yml:118-119)
- `Dockerfile.local`, `bootstrap*.json`, `custom_config*.json`, `*.env` ถูก gitignore (backend/.gitignore:62-66, .gitignore:11-13) — เครื่องใหม่ต้องสร้างเองก่อน `up`
- local ไม่มี job `rs.initiate` (มีเฉพาะ deploy/account/mongo-init.js:7-10 และ backend/.ci/projection.compose.yml:19) — บนเครื่อง dev replica set ถูก init ด้วยมือมาก่อน (runtime: `rs.status()` = `rs0 members=1`)
- local compose ยังส่ง env `R2_*` ให้ mainapi (backend/docker-compose.yml:95-100) ควบคู่กับ `storage.local.env` ที่มี key `S3_*`/`STORAGE_ALLOW_PRESIGNED_URL` (backend/docker-compose.yml:64-66, backend/docker-compose.local.yml:107-108)
- `docker-compose.dev.yml` เป็น overlay เล็กสำหรับ hot-reload ไฟล์ภาษา/ที่อยู่ (backend/docker-compose.dev.yml:5-12)

หลักฐาน runtime (อ่านอย่างเดียว 2026-09-07): container `mainapi` healthy, `postgres/redis/kafka/mongodb/minio` Up; Kafka มี 99 topics; PG มี db `appdb bc001 bctest01 demo postgres qa23995213 test uat260810a`; `docker ps` แสดง `mongodb` ไม่มี host port mapping (แค่ `27017/tcp`) ต่างจาก compose ที่ประกาศ `127.0.0.1:27017:27017` → container ถูกสร้างก่อนแก้ compose (ต้อง recreate ถ้าต้องต่อจาก host); log mainapi (start ล่าสุด 2026-09-05 23:27, container Up ~25 ชม.): loader ของ mainapi โหลด 24 env vars, loader ของ goapi โหลด 41 env vars, worker pool auto = 24 (12 CPU × 2, `docker info` NCPU=12) ตามสูตร backend/internal/goapi/workers/doc_processor.go:48-54, 64-68

## 3. Production compose (`deploy/account/compose.yml`, project `bcai-account`)
| service | image | mem_limit (base → 8gb) | healthcheck | env_file / mode | depends_on | อ้างอิง |
|---|---|---|---|---|---|---|
| mongo | mongo:7.0 (digest) `--replSet rs0` | 1g → 768m (+wiredTiger 0.25 GB) | mongosh ping | — | — | compose.yml:10-24, compose.8gb.yml:7-9 |
| mongo-init | mongo:7.0 one-shot รัน `mongo-init.js` แล้วรอ PRIMARY ≤60s | — | — | — | mongo healthy | compose.yml:26-45 |
| postgres | postgres:18-alpine | 512m → 384m | pg_isready | `/etc/bcai-account/postgres.env` | — | compose.yml:47-61 |
| clickhouse | clickhouse-server:25.8-alpine | 1g → 768m | wget /ping | `clickhouse.env` | — | compose.yml:63-77 (**โค้ด stub แล้ว แต่ยังเป็น dependency**) |
| redis | redis:7-alpine AOF, maxmemory 192mb | 256m → 128m | redis-cli ping | — | — | compose.yml:79-92 |
| minio | minio/minio (digest), cpus 1.0 | 1g → 512m | /minio/health/ready | `minio.env` | — | compose.yml:94-118 |
| minio-init | minio/mc one-shot (เหมือน local) | 256m | — | `minio.env` | minio healthy | compose.yml:120-152 |
| kafka-init | alpine chown 1000:1000 volume | 32m | — | — | — | compose.yml:154-168 |
| kafka | apache/kafka:4.3.1 KRaft single node, heap 768m → 512m | 1280m → 1g | kafka-topics --list | `kafka.env` (CLUSTER_ID) | kafka-init completed | compose.yml:170-206, compose.8gb.yml:18-21 |
| migrate | `${MAINAPI_IMAGE}` `DEV_API_MODE=3` | 1536m → 768m | — | `backend.env` | mongo-init, postgres, clickhouse, redis, kafka | compose.yml:208-232 |
| mainapi | `${MAINAPI_IMAGE}` (mode 2 จาก backend.env) | 1536m → 1g | wget /healthz (start 90s) | `backend.env` | + minio-init, migrate completed | compose.yml:234-269 |
| worker | `${MAINAPI_IMAGE}` `DEV_API_MODE=1`, group `bcai-account-01` | 1g → 512m | wget /healthz | `backend.env` | เหมือน mainapi | compose.yml:271-307 |
| frontend | `${FRONTEND_IMAGE}` | 768m → 512m | wget :3000/ | `frontend.env` | mainapi healthy | compose.yml:309-330 |

- Network: `edge` (bridge) และ `data` (`internal: true`); เฉพาะ mainapi อยู่ทั้งสอง, frontend อยู่ `edge` (compose.yml:258, 319, 332-337). Port ออก host มีแค่ `127.0.0.1:8888` (mainapi) และ `127.0.0.1:3200→3000` (frontend) (compose.yml:254-255, 317-318)
- `cap_drop: [ALL]` + `no-new-privileges` มีเฉพาะ minio/minio-init/kafka-init/migrate/mainapi/worker/frontend (compose.yml:115-117, 148-150, 163-165, 228-230, 259-261, 297-299, 320-322) — mongo/postgres/clickhouse/redis/kafka ไม่มี; `init: true` มีเฉพาะ mainapi/worker/frontend (compose.yml:253, 293, 316); log json-file 10m×3 ผ่าน anchor ทุก service (compose.yml:3-7); config mount `/var/lib/bcai-account/config → /app/bootstrap` (compose.yml:226, 257, 295)
- ผลรวม mem_limit base ≈ 9 GB จึงมี `compose.8gb.yml` สำหรับ droplet 4 vCPU/8 GB (compose.8gb.yml:1-5)
- Caddy: site `account.bcaicloud.com` → `reverse_proxy 127.0.0.1:3200`, ตอบ 404 ให้ `/backend/goapi/api/health/{kafka,background,queue*,database,system}` (deploy/account/Caddyfile.account:1-13) ซึ่งเป็น route ที่ goapi ลงทะเบียนบน group `g` ตรง ๆ ไม่ใช่ `authGroup` จึงไม่ผ่าน bearer auth (backend/internal/goapi/bootstrap.go:353, 377-382; mainapi ก็ยกเว้น `/goapi/*` จาก auth ที่ backend/main.go:296); frontend เองก็ block `/backend/goapi/get|exec|getdoc`, `/backend/reportm/*`, `/backend/goapi/api/setup|mcp/*`, `/backend/reload-config` และ login routes ที่ระดับ rewrite (frontend/next.config.ts:29-45, 58-72)

## 4. สคริปต์บนเซิร์ฟเวอร์ (`deploy/account/`)
`provision-server.sh` (ต้อง root, ต้องมี env `MAINAPI_IMAGE`/`FRONTEND_IMAGE`) ทำตามลำดับ (deploy/account/provision-server.sh):
1. สร้าง `/etc/bcai-account` 0700, `/opt/bcai-account` 0750, `/var/lib/bcai-account/config` owner 10001 (บรรทัด 22-25)
2. ถ้ายังไม่มี `secrets.env` → gen `POSTGRES_PASSWORD`, `CLICKHOUSE_PASSWORD`, `JWT_SECRET_KEY`, `RELOAD_CONFIG_SECRET`, `KAFKA_CLUSTER_ID` แล้วเติม `MINIO_ROOT_USER/PASSWORD`, `MINIO_APP_ACCESS_KEY/SECRET_KEY` ถ้าขาด (27-62)
3. สร้าง `bootstrap.json` และ `custom_config.json` เป็น `{}` ถ้ายังไม่มี, chown 10001, chmod 0600 (64-71) → บน prod config ทั้งหมดมาจาก env_file ไม่ใช่ bootstrap
4. เขียน `postgres.env` (db `bcai_projection`, user `bcai`), `clickhouse.env` (db `bcai_analytics`), `kafka.env`, `minio.env` (bucket `bcai-account`, `MINIO_BROWSER=off`) (73-97)
5. เขียน `backend.env` — key ที่ตั้ง: `MODE/GO_ENV/BC_ENV=production`, `DEV_API_MODE=2`, `SERVICE_PORT`, `LOG_LEVEL=INFO`, `TZ`, `HOST_API`, `HTTP_CORS`, `MONGODB_PRO_URI/DB`, `POSTGRES_*` (ทั้ง `USER`/`USERNAME`, `DB`/`DATABASE`/`DB_NAME` เพราะ mainapi กับ goapi อ่านคนละชื่อ), `CH_*`/`CLICKHOUSE_*`, `REDIS_CACHE_URI`, `ENABLE_KAFKA=true`, `KAFKA_SERVER_URL`, `JWT_SECRET_KEY`, `RELOAD_CONFIG_SECRET`, `GOOGLE_CLIENT_ID`, `BCAI_DEV_LOGIN_ENABLED=false`, `BCAI_DEMO_LOGIN_ENABLED=true`, `BCAI_DEMO_USERNAME`, `GODEBUG`, `S3_ENDPOINT/REGION/ACCESS_KEY_ID/SECRET_ACCESS_KEY/BUCKET_NAME/FORCE_PATH_STYLE`, `STORAGE_ALLOW_PRESIGNED_URL=false` (99-147)
6. เขียน `frontend.env` (`NODE_ENV`, `BCAI_LOCAL_BACKEND_URL=http://mainapi:8888`, `NEXT_PUBLIC_GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_ID`, `BCAI_DEV_LOGIN_ENABLED=false`) และ `release.env` (`MAINAPI_IMAGE`, `FRONTEND_IMAGE`) แล้ว chmod 0600 ทุก env (149-162)

`rotate-postgres-password.sh`: gen รหัสใหม่ → `ALTER ROLE bcai` ผ่าน `compose exec postgres psql` (23-30) → เขียนทับใน `secrets.env` (37-47) → รัน provision ใหม่ (54) → `rm -sf mainapi worker migrate`, recreate postgres, รัน migrate รอ `exited:0` ≤120s (56-77) → `up -d --no-deps mainapi worker` แล้วรอ healthy ≤180s ต่อ service (79-102)

`sshd-hardening.conf`: ปิด password/kbd-interactive, `PermitRootLogin prohibit-password`, `MaxAuthTries 3`, `LoginGraceTime 30` (deploy/account/sshd-hardening.conf:3-8). `minio-app-policy.json`: สิทธิ์ app user จำกัดที่ bucket `bcai-account` (list + get/put/delete/multipart) (deploy/account/minio-app-policy.json:6-26).

`docs/runbooks/RECOVERY-READINESS.md` (2026-09-05) เป็น runbook เตรียมการเท่านั้น — ยืนยันว่า **ยังไม่มีหลักฐาน backup/restore จริง**, ต้องตกลง RPO/RTO/ปลายทางก่อน, และต้อง restore Mongo+outbox+PG fences+Kafka offsets+MinIO objects (คู่ต้นฉบับ/thumbnail) เป็นชุดเดียวกัน (docs/runbooks/RECOVERY-READINESS.md:3, 7, 17, 24, 26, 29-30, 40)

## 5. Dockerfiles
| ไฟล์ | base / build | entry | สถานะ | อ้างอิง |
|---|---|---|---|---|
| backend/Dockerfile | golang:1.26-alpine + librdkafka, `CGO_ENABLED=1 -tags musl`, BuildKit cache mounts → alpine:3.21 uid/gid 10001, `DEV_API_MODE=2`, `SERVICE_PORT=8888`, HEALTHCHECK `/healthz` start 120s | `/app/go-app` (`main.go`) | LIVE (prod image `MAINAPI_IMAGE`) | backend/Dockerfile:6-21, 26-30, 41-52 |
| backend/Dockerfile.local | เหมือนบนแต่ไม่มี `--mount=type=cache` (เครื่อง dev ไม่มี buildx); uid ไม่ fix | เดียวกัน | LIVE local เท่านั้น, **ไม่ tracked** | backend/Dockerfile.local:1-5, 23; backend/.gitignore:65 |
| backend/Dockerfile.goapi | `CGO_ENABLED=0` build `./cmd/goapi/`, HEALTHCHECK `/version` | `/app/goapi` | ยังไม่ตรวจว่ามีใครใช้ (ไม่มี compose อ้าง; `backend/cmd/goapi/` มีอยู่) | backend/Dockerfile.goapi:15, 41-44 |
| backend/Dockerfile-consumer | golang:1.21-alpine3.17, librdkafka 1.9.2, `DEV_API_MODE=1` | `/root/go-app` | DEAD (go.mod ประกาศ `go 1.26` ที่ backend/go.mod:3; ไม่มี compose อ้าง — อ้างเฉพาะ workflow ค้างเก่า backend/.github/workflows/build_consumer.yaml:35, build_deploy_consumer_dev.yaml:35 ซึ่งอยู่นอก root `.github/` จึงไม่ถูก GitHub รัน) | backend/Dockerfile-consumer:6-9, 29 |
| backend/Dockerfile-member | golang:1.20.2-alpine3.17, build `cmd/member/main.go` | `/root/go-app` | DEAD (ไม่มี compose อ้าง — เหลือแค่ workflow ค้างเก่า backend/.github/workflows/build_api_member.yaml:35, build_deploy_api_member_dev.yaml:35) | backend/Dockerfile-member:6, 20 |
| backend/Dockerfile-migration | golang:1.21-alpine3.17, `DEV_API_MODE=3` | `/root/go-app` | DEAD — prod ใช้ image เดียวกับ mainapi + env แทน (compose.yml:208-212); เหลือแค่ workflow ค้างเก่า backend/.github/workflows/build_migration.yaml:34 | backend/Dockerfile-migration:6, 28 |
| backend/DockerfileM1 | golang:1.18.0-alpine3.15, `SERVERLESS=serverless` | `/root/go-app` | DEAD (อ้างจาก backend/Makefile:101 เท่านั้น — target push `smlsoft/smlcloudplatform:apidev` ยุคเก่า) | backend/DockerfileM1:6, 14 |
| frontend/Dockerfile | node:24.18.0-alpine 4 stage; build args `NEXT_PUBLIC_GOOGLE_CLIENT_ID`, `BCAI_LOCAL_BACKEND_URL`, `NEXT_PUBLIC_DEFAULT_BACKEND_URL`; runner uid 1001 `npm run start` :3000 (ไม่ใช่ standalone) | `npm run start` | LIVE (prod `FRONTEND_IMAGE`) | frontend/Dockerfile:2, 25-32, 41-57 |

ไม่มี `frontend/Dockerfile.onprem` ใน repo (`git ls-files frontend/Dockerfile*` มีแค่ `frontend/Dockerfile`) — memory เก่าที่อ้างชื่อนี้ล้าสมัย. Build frontend ต้องส่ง `BCAI_LOCAL_BACKEND_URL` เสมอ ไม่งั้น `next build` โยน error ที่ rewrites (frontend/next.config.ts:19-21; `start` = `next start` ตาม frontend/package.json:13, ไม่มี `output: standalone` ใน next.config.ts); `NEXT_PUBLIC_DEFAULT_BACKEND_URL` ไม่ตั้งจะ fallback เป็น `http://192.168.2.202:8888/goapi` ในหน้า settings (frontend/src/app/settings/settings-screen.tsx:77) — ค่า default ชี้เครื่อง on-prem เก่า

## 6. การโหลด config (bootstrap.json → env)
มี loader **สองตัว** ที่ทำงานตามลำดับใน process เดียว:
1. mainapi: `setupconfig.LoadBootstrapConfig()` ก่อน `config.NewConfig()` (backend/main.go:236, 248). ค้นหาไฟล์ตามลำดับ `/app/bootstrap/bootstrap.json` → `/app/bootstrap.json` → `bootstrap.json` → `config/bootstrap.json` (backend/internal/setupconfig/loader.go:99-104), merge `custom_config.json` ข้าง ๆ ทับ (loader.go:106-111, 177-186), แล้ว `os.Setenv` **ทับ env เดิมโดยไม่เช็ค** (loader.go:247-248) → ค่าใน bootstrap ชนะ env_file เสมอ; ตั้ง default `REDIS_HOST/PORT`, `REDIS_CACHE_URI=redis:6379`, `KAFKA_SERVER_URL=kafka:9092` ถ้าว่าง (loader.go:266-281); compose `CH_SERVER_ADDRESS=host:port` (loader.go:284-299)
2. goapi: `setupconfig.LoadBootstrapConfig()` ของตัวเองใน `Init()` (backend/internal/goapi/bootstrap.go:59) path list เดียวกัน (backend/internal/goapi/setupconfig/loader.go:146-150) แต่ mapping กว้างกว่า (AI provider keys, `r2*`/`s3*` ใต้ `integrations` บรรทัด 101-106, `KAFKA_CONSUMER_GROUP_VERSION` บรรทัด 79, `CORS_ALLOWED_ORIGINS` บรรทัด 88) (goapi/setupconfig/loader.go:60-129)

**กับดักที่ตรวจพบ:** loader ของ goapi ตัด `_` ออกจาก key ก่อน lookup (goapi/setupconfig/loader.go:283-285) แต่ loader ของ mainapi ใช้ key ดิบ (backend/internal/setupconfig/loader.go:242-244) และ mapping เป็นแบบไม่มีตัวคั่น เช่น `dbname`, `databasename`, `serverurl`, `enablekafka`, `loglevel` (loader.go:37-84). ไฟล์ `backend/bootstrap.json` บนเครื่อง dev ใช้ key snake_case (`db_name`, `database_name`, `server_url`, `enable_kafka`, `log_level`, `ssl_mode`) จึงถูก mainapi loader ข้ามเงียบ ๆ; `bootstrap.local.json` แก้ปัญหาด้วยการใส่ทั้งสองแบบซ้ำกัน (สังเกตจากชื่อ key ในไฟล์, ไม่ใช่ค่า) — ควรแก้ที่ loader ให้ normalize เหมือน goapi. อีกจุด: `custom_config.json` วาง `r2_*` ใต้ `integrations` ซึ่ง mainapi loader ไม่มี mapping (loader.go:65-68) แต่ goapi loader มี (goapi loader:101-104).

ตาราง section → env (mainapi loader, backend/internal/setupconfig/loader.go:28-84): `mongodb.{uri,database,host,port,username,password}` → `MONGODB_*` ทุก suffix (DEV/UAT/PRO/PRODUCTION พร้อมกัน); `postgresql.{host,port,user,password,sslmode,dbname,timezone,loggerlevel}` → `POSTGRES_HOST/PORT/USERNAME/PASSWORD/SSL_MODE/DB_NAME/TIMEZONE/LOGGER_LEVEL`; `clickhouse.*` → `CH_SERVER_ADDRESS/CLICKHOUSE_PORT/CH_USERNAME/CH_PASSWORD/CH_DATABASE_NAME`; `service.*` → `ENABLE_KAFKA, LOG_LEVEL, JWT_SECRET_KEY, DEV_API_MODE, SERVICE_PORT, HOST_API, MODE, HTTP_CORS, FIREBASE_PROJECT_ID`; `integrations.*` → `GEMINI_API_KEY/MODEL`; `storage.*` → `STORAGE_DATA_PATH/URI, AZURE_*, S3_ENDPOINT/PUBLIC_ENDPOINT/ACCESS_KEY_ID/SECRET_ACCESS_KEY/BUCKET_NAME`; `kafka.serverurl` → `KAFKA_SERVER_URL`.

Reload runtime: mainapi เปิด `POST /reload-config` (อยู่ใน public path บรรทัด 294, handler backend/main.go:323) ตรวจ `RELOAD_CONFIG_SECRET` (backend/main.go:326) แล้วเรียก `setupconfig.ReloadConfig()` (backend/main.go:333-335); goapi ฝั่งเขียน `UpdateBootstrapJSON` + `notifyMainAPIReload` (goapi/setupconfig/loader.go:398, 415, 514) — บน prod ไฟล์เป็น `{}` และ owner 10001 (provision-server.sh:64-71) จึงเขียนได้จาก container.

## 7. Object storage (รูปภาพ/ไฟล์)
- mainapi: `NewFilePersister()` คืน persister S3/R2 ตัวเดียว (backend/pkg/microservice/persister_factory.go:4-6) อ่าน `S3_ACCOUNT_ID|R2_ACCOUNT_ID`, `S3_ENDPOINT|R2_ENDPOINT`, `S3_ACCESS_KEY_ID|R2_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY|R2_SECRET_ACCESS_KEY`, `S3_BUCKET_NAME|R2_BUCKET_NAME`, region default `us-east-1` (persister_file_r2.go:51-58, 68) — ถ้ามี endpoint จะใช้ BaseEndpoint (MinIO path-style)
- goapi: `handlers.InitR2Client()` (backend/internal/goapi/bootstrap.go:82) ใช้ชุด env เดียวกัน + `S3_FORCE_PATH_STYLE|R2_FORCE_PATH_STYLE` (backend/internal/goapi/handlers/image_r2.go:75-86, 111); presigned URL ตรงต้องเปิด `STORAGE_ALLOW_PRESIGNED_URL=true` (image_r2.go:154, backend/internal/goapi/handlers/storage_private.go:14) — prod ตั้ง `false` (provision-server.sh:146)
- bucket: local จาก `minio.local.env` (`MINIO_BUCKET_NAME`) และ prod = `bcai-account` เปิด versioning ผ่าน minio-init (compose.yml:131-132) ใช้ app user สิทธิ์จำกัด (minio-app-policy.json) — Mongo เก็บแค่ URI ตามกฎใน AGENTS.md

## 8. CI / test infra
- `.github/workflows/ci.yml`: job `backend-test` (compile ทุก package + unit test ยกเว้น `backend/.ci/test-quarantine.txt` + compile integration tags ใน `golang:1.26`) (ci.yml:10-41), `frontend-test` (node 24.18.0 lint/typecheck/vitest/build ด้วย `BCAI_LOCAL_BACKEND_URL=http://localhost:8888`) (ci.yml:42-81, 45, 53), `backend-outbox-integration` (mongo rs0 + postgres:17-alpine ชั่วคราว) (ci.yml:82-137, 112), `backend-projection-kafka-integration` ใช้ `backend/.ci/projection.compose.yml` (ci.yml:138-165)
- `projection.compose.yml`: mongo rs0 + mongo-init idempotent, postgres:18-alpine trust, apache/kafka:4.3.1 auto-create **ปิด**, service `tests` profile `test` รัน integration tests ชุด outbox/projection/barcode (backend/.ci/projection.compose.yml:3-75)
- GitHub Actions **ล็อกเพราะ billing ตั้งแต่ 2026-09-03** → หลักฐาน build/test ทั้งหมดเป็น local (docs/handoff/HANDOFF-2026-09-06.md:13); ไม่มี workflow deploy/registry ใน root `.github/` (`git ls-files .github` มีแค่ `.github/workflows/ci.yml`); ส่วน `backend/.github/workflows/*.yaml` (9 ไฟล์ tracked: build_api_image, build_api_member, build_consumer, build_deploy_*_dev, build_migration, ci.yml) เป็นของค้างยุค repo แยก อยู่นอก root `.github/` จึงไม่ถูก GitHub Actions รัน — ยังไม่ตรวจเนื้อหาว่า push ไป registry ไหน

## 9. เซิร์ฟเวอร์ที่รู้จัก (ไม่มี secret)
| เครื่อง | บทบาท | สถานะล่าสุดที่มีหลักฐาน | อ้างอิง |
|---|---|---|---|
| dev Windows (Docker Desktop) | local stack §2 + `next dev` :3000 | ใช้งานอยู่ (runtime 2026-09-07) | docs/handoff/HANDOFF-2026-09-06.md:17-23 |
| on-prem 192.168.2.202 | Docker stack เดิม (คนละ compose กับ local) | ssh timeout 2026-09-06 → ไม่ทราบสถานะ | docs/handoff/HANDOFF-2026-09-06.md:76 |
| DigitalOcean SGP1 159.223.43.229 (4 vCPU/8 GB) | prod ใหม่สำหรับ `account.bcaicloud.com` ด้วย compose.yml + compose.8gb.yml | **ขัดกัน**: docs/handoff/HANDOFF-2026-09-06.md:77 บอก "provision แล้ว ยังไม่ deploy app" แต่ memory `do-sgp1-new-prod-server` บันทึกว่า deploy ครบ 9 service เมื่อ 2026-09-02/03 → ยังไม่ตรวจ (ไม่ได้ ssh ในงานนี้) | docs/handoff/HANDOFF-2026-09-06.md:77 |
| 188.212.158.39 (prod เก่า) | host เดิมของ `account.bcaicloud.com` | เข้าไม่ได้แล้ว | docs/handoff/HANDOFF-2026-09-06.md:77 |

## 10. โครง runbook deploy (สกัดจากไฟล์ — ทุกขั้นที่แตะ prod เป็น R0 ต้องถามลุงจืดก่อน)
1. **Build image บนเครื่อง dev** (ต้องมี BuildKit สำหรับ `backend/Dockerfile`): `docker build -f backend/Dockerfile -t bcai-account-mainapi:rYYYYMMDD-N backend` และ `docker build --build-arg NEXT_PUBLIC_GOOGLE_CLIENT_ID=… --build-arg BCAI_LOCAL_BACKEND_URL=http://mainapi:8888 -t bcai-account-frontend:rYYYYMMDD-N frontend` (frontend/Dockerfile:25-32; provision-server.sh:151 กำหนด URL ภายในเป็น `http://mainapi:8888`)
2. **ครั้งแรกบนเซิร์ฟเวอร์**: copy `deploy/account/*` ไป `/opt/bcai-account/deploy`, วาง `sshd-hardening.conf` ใน `/etc/ssh/sshd_config.d/` ให้เรียงก่อน cloud-init, รัน `MAINAPI_IMAGE=… FRONTEND_IMAGE=… ./provision-server.sh` (สร้าง secrets/env/bootstrap `{}`), ติดตั้ง `Caddyfile.account` ใน Caddy (rotate-postgres-password.sh:5-8 ยืนยัน path `/opt/bcai-account/deploy/compose.yml`)
3. **ส่ง image**: ไม่มี registry → `docker save <image> | ssh root@<host> docker load` แล้วแก้ `MAINAPI_IMAGE`/`FRONTEND_IMAGE` ใน `/etc/bcai-account/release.env` (สำรองไฟล์เดิมไว้ก่อน)
4. **up**: `docker compose --env-file /etc/bcai-account/release.env -f compose.yml -f compose.8gb.yml up -d` (คำสั่งตาม compose.8gb.yml:4-5; ชื่อ project `bcai-account` มาจาก `name:` ที่ compose.yml:1 ไม่ต้องใส่ `-p`) — ลำดับจะเป็น mongo→mongo-init, kafka-init→kafka, minio→minio-init, migrate (mode 3) → mainapi/worker → frontend ตาม `depends_on` §3; release เฉพาะ service ใช้ `up -d --no-deps <svc>` เหมือนใน rotate-postgres-password.sh:79
5. **ตรวจรับ**: รอ healthcheck `healthy` ทุกตัว (`docker compose ps`), `curl 127.0.0.1:8888/healthz` และ `127.0.0.1:3200/`, ดู `docker logs` ของ mainapi ว่ามีบรรทัด `GoAPI routes registered under /goapi/*` (backend/main.go:576) และ `[Bootstrap]` ไม่ฟ้อง; เปิดโดเมนผ่าน Caddy
6. **Rollback**: คืน `release.env` เดิมแล้ว `up -d --no-deps mainapi worker frontend` (image เก่ายังอยู่บนเครื่อง); DB data ถือว่า disposable ก่อน go-live (docs/handoff/HANDOFF-2026-09-06.md:14)
7. **หมุนรหัส PG**: `rotate-postgres-password.sh` (§4) — ทำให้ mainapi/worker/migrate ถูกลบและสร้างใหม่ มี downtime สั้น

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ
- สถานะจริงของ 159.223.43.229 (deploy แล้วหรือยัง, image tag ปัจจุบัน, Caddyfile/DNS) และของ 192.168.2.202 — ต้อง ssh ตรวจเอง; งานนี้ไม่ได้เชื่อมต่อเครื่องใดนอกจาก local docker
- ไม่ได้รัน `docker build` ทั้ง backend/frontend ในงานนี้ → ยังไม่ยืนยันว่า `backend/Dockerfile` (BuildKit) build ผ่านบน go.mod ปัจจุบัน; Dockerfile-consumer/-member/-migration/M1 ใช้ Go 1.18–1.21 ซึ่งน่าจะ build ไม่ผ่านแล้ว แต่ยังไม่ตรวจ
- `backend/Dockerfile.goapi` + `backend/cmd/goapi/` มีใครใช้ deploy แยกหรือไม่ — ไม่มี compose/CI อ้าง; ถ้าไม่ใช้ควรลบ (คำถามลุงจืด)
- ผลกระทบจริงของ key snake_case ใน `bootstrap.json` ต่อ mainapi loader (§6) — ยืนยันจากโค้ดและชื่อ key เท่านั้น ยังไม่ได้เขียน test/พิสูจน์ด้วย log เฉพาะ key; ควรตัดสินว่าจะ normalize ใน `internal/setupconfig/loader.go` หรือแก้ไฟล์
- `worker` (mode 1) บน prod ยังจำเป็นไหม เมื่อ goapi consumers ใน mainapi ทำงานอยู่แล้ว (HANDOFF P3) — เป็น R1 รอลุงจืด
- ClickHouse ยังเป็น `depends_on: service_healthy` ของ migrate/mainapi/worker บน prod (compose.yml:216-219, 240-243, 280-283) ทั้งที่โค้ด stub → ถ้าถอด container บน prod โดยไม่แก้ compose stack จะไม่ขึ้น; การถอดถาวรเป็น R1 (docs/handoff/HANDOFF-2026-09-06.md:57)
- Backup/restore: ไม่มี job/หลักฐานใด ๆ ใน repo (docs/runbooks/RECOVERY-READINESS.md:40); ต้องกำหนด RPO/RTO/ปลายทาง
- `mongodb` local ไม่มี host port mapping ที่ runtime แม้ compose ประกาศไว้ — ยังไม่ recreate (งานนี้อ่านอย่างเดียว)
- ไม่ได้ตรวจ `deploy/account/*` เทียบกับสำเนาบนเซิร์ฟเวอร์ (drift), และไม่ได้ตรวจ Caddy config จริงบนเครื่อง
