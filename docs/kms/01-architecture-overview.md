# ภาพรวมสถาปัตยกรรมระบบ BC Ai Account
> ตรวจล่าสุด: 2026-09-07 (commit d93a210d) — ผู้เขียน: AI reader; ทุกข้อเท็จจริงอ้าง path:line · ตรวจซ้ำโดย fact-checker

## 1. สรุปสั้น
- ระบบเป็น monorepo 2 ส่วนหลัก: Go backend (module `smlcloudplatform`, binary เดียวจาก `backend/main.go`) และ Next.js frontend (`frontend/`); ทั้งคู่ deploy เป็น Docker container ตาม `deploy/account/compose.yml`
- binary backend ตัวเดียวทำได้ 3 บทบาทตาม env `DEV_API_MODE` (อ่านที่ `backend/main.go:238`): HTTP API + goapi (`""`/`"2"`, `backend/main.go:256`), Kafka consumer รุ่นเก่า (`""`/`"1"`, `backend/main.go:654`), migration (`"3"`, `backend/main.go:585`) — **ค่าว่าง `""` รัน HTTP และ consumer พร้อมกันใน process เดียว** เพราะเป็น `if` แยกกัน 2 ก้อน ไม่ใช่ `else`
- goapi ไม่ใช่ service แยกใน deploy ปัจจุบัน แต่เป็น sub-server ที่ mount ใต้ `/goapi` ของ mainapi (`backend/main.go:568-576`); `cmd/goapi` + `Dockerfile.goapi` เป็น entry สำรองแบบ standalone ที่ไม่มี compose ไหนอ้างถึง (ดู §4)
- browser ไม่คุย mainapi ตรง: ทุก request ผ่าน Next.js (route handler ใต้ `frontend/src/app/api/**` หรือ rewrite `/backend/*`) แล้ว Next.js ค่อย fetch ไป `BCAI_LOCAL_BACKEND_URL` ฝั่ง server (`frontend/next.config.ts:19-21,74`, `frontend/src/lib/backend-url.ts:117-127`)

## 2. แผนภาพ (ASCII)
```
 Browser (Thai 40+ users)
   |  HTTPS account.bcaicloud.com  (prod: Caddy -> 127.0.0.1:3200, deploy/account/Caddyfile.account:1-13)
   v
 Next.js frontend :3000 (prod publish 127.0.0.1:3200, deploy/account/compose.yml:317-318)
   |-- route handlers  frontend/src/app/api/**  (36 route.ts)  -- forward Authorization: Bearer
   `-- rewrites        /backend/:path*  ->  ${BCAI_LOCAL_BACKEND_URL}/:path*   (next.config.ts:74)
   v
 mainapi :8888  (backend/main.go, DEV_API_MODE=2)  -- Echo via pkg/microservice
   |-- /login /holding /shop /... (132 module HTTP, main.go:357-548, auth middleware main.go:343)
   |-- /goapi/*  <- GoAPIServer sub-server (internal/goapi/bootstrap.go:350) own auth (bootstrap.go:208)
   |-- /api/language/:lang (public alias, main.go:580)   /healthz  /metrics  /swagger/*
   `-- background: outbox worker (product_http.go:54), coupon cleanup 5 min (main.go:565),
                   goapi WorkerManager poll queue (doc_processor.go:239/252/272), goapi Kafka consumers (kafka.go:26)
   |
   |--> MongoDB :27017   (source of truth; MongoPersister main.go:263, goapi InitMongoAtlas bootstrap.go:68)
   |--> Redis :6379      (auth token cache, main.go:262-264)
   |--> Kafka :9092      (topics when-*-created/updated/deleted; producers in modules, consumers in goapi + worker)
   |--> PostgreSQL :5432 (per-holding projection + queues table; goapi InitManagerPool bootstrap.go:99)
   |--> MinIO/S3 :9100   (images/files; InitR2Client bootstrap.go:82, proxy /goapi/s3/file/* bootstrap.go:509)
   `--X ClickHouse       (STUBBED: myclickhouse/utils.go:234,239 returns "clickhouse is disabled")

 worker  (prod only, DEV_API_MODE=1, deploy/account/compose.yml:271-276)  -- legacy Kafka consumers main.go:654-725
 migrate (prod one-shot, DEV_API_MODE=3, deploy/account/compose.yml:208-212) -- main.go:585-652
```

## 3. Process entrypoints ของ backend
| โหมด | เงื่อนไข | ทำอะไร | อ้างอิง |
|---|---|---|---|
| ทุกโหมด | ก่อน `NewConfig` | `setupconfig.LoadBootstrapConfig()` โหลด bootstrap.json → env, แล้ว `config.NewConfig()` + `microservice.NewMicroservice(cfg)`, เปิด Prometheus | `backend/main.go:236,248-249,254` และ `backend/pkg/microservice/microservice.go:80` |
| HTTP API | `DEV_API_MODE == "" หรือ "2"` | swagger, Redis cacher + MongoPersister + AuthService, publicPath/exceptShopPath, auth middleware `MWFuncWithRedisMixShop`, `/healthz`, CORS, `migration.StartMigrateModel` (`:353`), ลงทะเบียน 132 module `*.New*Http(ms,cfg)` (grep ได้ 133 บรรทัดใน `:357-547`, 1 บรรทัดเป็น comment) ผ่าน `serviceStartHttp` (`:548`), migrationAPI + media upload, coupon cleanup scheduler, แล้ว mount goapi | `backend/main.go:256-583` (goapi block `:568-581`) |
| Migration | `DEV_API_MODE == "3"` | เรียก `MigrationDatabase(ms,cfg)` ของ journal/chartofaccount/productbarcode/payment และ consumer ทุกตัว (purchase/sales/stock/ap/ar/warehouse/stockbalance) | `backend/main.go:585-652` |
| Consumers (legacy) | `DEV_API_MODE == "" หรือ "1"` | `/healthz`, `CONSUMER_GROUP_NAME` (default `"03"`), `ms.RegisterConsumer(...)` 27 ตัว + chartofaccount 4 starter (Created/Updated/Deleted/BlukCreated, `:667-670`) + `serviceStartConsumer(task, productbarcode)` (`:725`) | `backend/main.go:654-727` |
| Start | ทุกโหมด | `ms.Start()` — spawn backgroundWorkers, ถ้ามี route ก็ `startHTTP`, รอ SIGTERM/SIGINT | `backend/main.go:728`, `backend/pkg/microservice/microservice.go:233-262` |

หมายเหตุ mode: ใน local `DEV_API_MODE: "2"` (`backend/docker-compose.yml:80`) จึง **ไม่มี consumer รุ่นเก่ารันบนเครื่อง dev** — มีแต่ consumer ของ goapi ที่เปิดด้วย `ENABLE_KAFKA` (`backend/internal/goapi/bootstrap.go:157-160`)

## 4. binary / Dockerfile / compose ที่ใช้จริง
| artifact | build จาก | mode | ถูกอ้างโดย | สถานะ |
|---|---|---|---|---|
| `backend/Dockerfile` | `go build … main.go` (`:21`), `ENV DEV_API_MODE=2` (`:41`), `EXPOSE 8888` (`:47`), healthcheck `/healthz` (`:50`) | HTTP+goapi | `backend/docker-compose.yml:59-62`; prod ใช้ image `${MAINAPI_IMAGE}` ตัวเดียวสำหรับ mainapi/worker/migrate (`deploy/account/compose.yml:209,235,272`) | LIVE |
| `backend/Dockerfile.local` | `main.go` (`:18`), `DEV_API_MODE=2` (`:34`) | HTTP+goapi | `backend/docker-compose.local.yml:103-106` (mount `bootstrap.local.json` → `/app/bootstrap.json` `:122`) | LIVE (dev) |
| `backend/Dockerfile.goapi` | `go build ./cmd/goapi/` (`:15`), healthcheck `/version` (`:42`) | goapi standalone `:8888` (`backend/cmd/goapi/main.go:21-26,69-71`) | ไม่พบ compose/workflow ใดอ้าง (grep ทั้ง repo) | DEAD (สำรอง) |
| `backend/Dockerfile-consumer` / `-migration` | `main.go` + `DEV_API_MODE=1` (`Dockerfile-consumer:29`) / `=3` (`Dockerfile-migration:28`), base `golang:1.21-alpine3.17` | consumer / migrate | เฉพาะ workflow เก่าใต้ `backend/.github/workflows/build_consumer.yaml:35`, `build_deploy_consumer_dev.yaml:35`, `build_migration.yaml:34` | DEAD (prod ใช้ image เดียว + override env แทน) |
| `backend/Dockerfile-member` | `cmd/member/main.go` (`:20`), `golang:1.20.2` | member service | `backend/.github/workflows/build_api_member.yaml:35`, `build_deploy_api_member_dev.yaml:35` | DEAD |
| `backend/DockerfileM1` | `main.go`, `golang:1.18` | — | `backend/Makefile:101` | DEAD |
| `backend/cmd/app/main.go` | entry คู่ขนานที่สร้าง `NewMicroservice` เอง (`:120-123`) + เขียน `routes.json` (`:389`) | — | `backend/dev.sh:3` (`cd cmd/app/`) | ยังไม่ตรวจว่ายัง build ผ่าน |
| `backend/cmd/*` อื่น (30 โฟลเดอร์, 52 ไฟล์) | — | — | `.github/workflows/ci.yml:36,39` compile ด้วย `go test -run "^$" ./cmd/...` เท่านั้น; `backend/Makefile` อ้าง `cmd/inventoryservice`, `cmd/uploadmediaservice` ฯลฯ ที่ **ไม่มีอยู่ใน tree** | ส่วนใหญ่ DEAD; `cmd/migrationapi/api` ถูก import โดย `backend/main.go:10,550` (LIVE) |

Stack prod (`deploy/account/compose.yml`): mongo 7.0 (`:10`), postgres 18 (`:47`), clickhouse 25.8 (`:63`, ยังประกาศอยู่แม้โค้ด stub), redis 7 (`:79`), minio (`:94`), kafka `apache/kafka:4.3.1` (`:170`), migrate (`:208`), mainapi publish `127.0.0.1:8888` (`:254-255`), worker (`:271`), frontend publish `127.0.0.1:3200` (`:317-318`); edge = Caddy บน host proxy ไป `:3200` และตอบ 404 ให้ `/backend/goapi/api/health/*` (`deploy/account/Caddyfile.account:4-12`). service ส่วนใหญ่อ่าน `env_file: /etc/bcai-account/*.env` ที่ไม่อยู่ใน repo (postgres `:49`, clickhouse `:65`, minio `:97,122`, kafka `:173`, migrate/mainapi/worker ใช้ `backend.env` `:210,236,273`, frontend `:311`); mongo กับ redis ไม่มี `env_file`

## 5. goapi sub-server (`backend/internal/goapi/bootstrap.go`)
ลำดับใน `Init()` (`:55`): `setupconfig.LoadBootstrapConfig()` ของ goapi เอง (`:59`) → `handlers.InitMongoAtlas()` (`:68`) → `handlers.InitR2Client()` S3 (`:82`) → `mydb.InitManagerPool` PostgreSQL+ClickHouse config (`:99`) → `mypostgres.InitQueueSchema` (`:128`) → `workers.NewWorkerManager(0)` + `Start()` (`:136-138`) → goroutine health checker / background task / ticker 30 นาที (`:146-149`) → ถ้า `ENABLE_KAFKA == "true"` เรียก `handlers.StartConsumers()` (`:157-160`)

`RegisterRoutes(g, "/goapi", pst)` (`:350`): route สาธารณะ `/goapi/`, `/goapi/version`, `/goapi/api/health*` (`:356-382`), `/goapi/api/approval/*` (`:465-467`), `/goapi/api/language/:lang`, `/goapi/api/address/thailand` (`:520-521`); ที่เหลืออยู่ใต้ `authGroup` ซึ่งตรวจ Bearer token กับ Redis ผ่าน `microservice.AuthService` (`:208-262`) เช่น `/goapi/api/*` inventory (`:385`), `/goapi/s3/file/*` (`:509`), chatbot/kb/aichat/ai-provider/unified (`:542-567`); rate limiter แบบ tier (`:595`). mainapi ปล่อยให้ `/goapi/*` ผ่าน middleware ของตัวเองเพราะอยู่ใน `publicPath` (`backend/main.go:296`). `Shutdown()` ปิด worker, Mongo, DB manager, ClickHouse, PG pools (`:577-589`)

## 6. pkg/microservice (framework กลาง, 44 ไฟล์)
- `NewMicroservice(cfg)` สร้าง Echo + logger (`backend/pkg/microservice/microservice.go:80`); `Echo()` เปิดให้ main.go เพิ่ม route เอง (`:643`)
- `RegisterHttp(h)` แค่เรียก `h.RegisterHttp()` (`backend/pkg/microservice/microservice_http.go:168-170`); `startHTTP` ฟัง `0.0.0.0:<port>` (`:141,154`) โดย port มาจาก `SERVICE_PORT` ก่อน แล้วค่อย `HTTP_PORT` default `8080` (`backend/internal/config/config_http.go:20-27`) — compose ตั้ง `SERVICE_PORT: "8888"` ให้ mainapi (`backend/docker-compose.yml:81`) และ log จริงยืนยัน `Listening: 8888` (docker logs mainapi, บรรทัด log เวลา 2026-09-05T23:27:40Z อ่านซ้ำ 2026-09-07)
- `RegisterConsumer(c)` → `c.RegisterConsumer(ms)` (`backend/pkg/microservice/microservice_consumer.go:106-108`); `Consume(servers, topic, groupID, …)` (`:54`) — ถ้าเป็น barcode topic (`projection.IsBarcodeTopic`) จะลงทะเบียนเป็น background worker (`:55-56`), topic อื่นแค่ `go ms.consumeSingle(...)` เป็น goroutine ธรรมดา (`:59`)
- `RegisterBackgroundWorker(fn)` เก็บลง `ms.backgroundWorkers` (`microservice.go:228-229`) แล้ว `Start()` spawn goroutine ให้ทุกตัวพร้อม context ยกเลิกตอนปิด (`microservice.go:236-246`); ผู้ใช้ปัจจุบัน: product outbox (`backend/internal/product/product/product_http.go:54-57`), productbarcode (`backend/internal/product/productbarcode/productbarcode_http.go:81`), productimport (`backend/internal/productimport/productimport_http.go:111`), barcode projection consumer (`microservice_consumer.go:56`)
- persister หลายชนิดอยู่ในแพ็กเกจเดียว: mongo, clickhouse, elk, opensearch, file/R2, image (`backend/pkg/microservice/persister_*.go`) — ตัวไหนยังถูกเรียกจริงนอกจาก mongo/file: ยังไม่ตรวจ

## 7. เส้นทาง request: browser → Next.js → mainapi
1. browser เรียก `/api/<domain>/...` ของ Next.js เอง; route handler เช่น `frontend/src/app/api/product/[[...productPath]]/route.ts:15-104` (GET/POST/PUT/DELETE) ดึง `Authorization` header ด้วย `requireBearerToken` (`frontend/src/lib/workspace-api.ts:8-13`) แล้ว `fetch(baseUrl + path)` พร้อม header เดิม (`route.ts:159-164`)
2. `baseUrl` มาจาก `getMainApiUrl(getBackendUrlFromRequest(...))` — ค่าที่ client ส่งมา (body / header `x-bc-backend-url` / query) ถูก validate แล้ว **ทิ้ง**; ฝั่ง server ใช้ `serverMainApiBase()` = env `BCAI_LOCAL_BACKEND_URL` เสมอ (`frontend/src/lib/workspace-api.ts:16-30`, `frontend/src/lib/backend-url.ts:117-123`; goapi base = `${BCAI_LOCAL_BACKEND_URL}/goapi` `:125-127`)
3. เส้นทางที่สอง: rewrite `/backend/:path*` → `${BCAI_LOCAL_BACKEND_URL}/:path*` (`frontend/next.config.ts:74`) โดยบล็อกเส้นทาง auth และเส้นทางอันตราย (`/backend/login*`, `/backend/goapi/exec`, `/backend/goapi/api/setup/*`, `/backend/reload-config` ฯลฯ) ไปที่ `/_blocked-auth-route` ก่อน (`frontend/next.config.ts:30-72`)
4. ที่ mainapi: middleware `MWFuncWithRedisMixShop(cacher, exceptShopPath, publicPath...)` (`backend/main.go:343`) — publicPath ไม่ต้อง token (`:265-297`), exceptShopPath ต้อง token แต่ไม่ต้องเลือก shop (`:300-320`), ที่เหลือต้องมี shop context
5. `BCAI_LOCAL_BACKEND_URL` เป็น build ARG/ENV ของ image frontend (`frontend/Dockerfile:26-29`), image ฟัง `:3000` ด้วย `npm run start` (`frontend/Dockerfile:52,57`; scripts `frontend/package.json:10-16`)

## 8. Background workers / consumers
| ตัวประมวลผล | รันที่ไหน | กลไก | สถานะ | อ้างอิง |
|---|---|---|---|---|
| Legacy Kafka consumers (journal, chartofaccount, purchase*/saleinvoice*/rfq/stock*/stockbalance/stockprocess, appurchasereceive, warehouse, creditor/debtor + creditorpayment/debtorpayment, shift, pay/paid, BOM, saleinvoicebomprice, task, productbarcode — 27 `RegisterConsumer` + 4 starter + 2 `serviceStartConsumer`) | prod `worker` (mode 1) เท่านั้น; local ไม่รัน | `ms.RegisterConsumer` + consumer group จาก `CONSUMER_GROUP_NAME` (prod ตั้ง key นี้ที่ `deploy/account/compose.yml:276`) | LIVE ใน prod; debtor/creditor/erp_user projection พังจาก schema drift ตาม lead (ยังไม่ตรวจซ้ำในบทความนี้) | `backend/main.go:654-725` |
| goapi Kafka consumers | ใน mainapi เมื่อ `ENABLE_KAFKA=true` | `StartConsumers()` spawn goroutine ต่อ topic, 79 จุด `ConsumeMessage(...)` ครอบคลุม `when-<entity>-{created,updated,deleted,bulk-*}` (saleinvoice, product, product-barcode, warehouse, saleorder, purchase*, rfq, stock*, creditor, customer, employee, debtor); group id = `<base>-<KAFKA_CONSUMER_GROUP_VERSION>` default `v1` | LIVE (log 2026-09-05 "Kafka consumers initialized") | `backend/internal/goapi/handlers/kafka.go:19-26`, `backend/internal/goapi/config/config.go:119` |
| goapi WorkerManager | ใน mainapi | `NewWorkerManager(0)` → `OptimalWorkerCount(4,32,2.0)` (log จริง: 24 workers / 12 cores); `monitorActiveShops` ticker 5 วินาที; loop worker `time.Sleep(100ms)` ระหว่างรอบ poll ตาราง queue | LIVE แต่ busy-poll เมื่อ queue ว่าง | `backend/internal/goapi/workers/doc_processor.go:64-68,160-163,239,252,272` |
| Product outbox worker | ใน mainapi | เก็บ `outboxevents` ใน Mongo transaction เดียวกับ product/barcode แล้ว worker ส่ง Kafka + retry/backoff | LIVE (module ลงทะเบียนที่ `backend/main.go:371` ผ่าน alias `products`; บรรทัด `:374` ที่ comment ไว้เป็นชื่อซ้ำ) | `backend/internal/product/product/product_http.go:38,54-57`, `backend/internal/product/product/outbox/README.md:1-12` |
| Coupon reservation cleanup | ใน mainapi | goroutine ทุก 5 นาที holdingCode ว่าง (= ทุกร้าน) | LIVE | `backend/main.go:553-565` |
| ClickHouse | — | `ClickHouseFastConnect` / `CreateClickHouseConnection` คืน error "clickhouse is disabled"; 15 ไฟล์ใน goapi ยัง import `myclickhouse` | STUBBED | `backend/internal/goapi/myclickhouse/utils.go:234-240` |

## 9. การโหลด config (bootstrap.json → env)
- มี loader 2 ชุดที่ทำงานเหมือนกันแต่คนละแพ็กเกจ: `backend/internal/setupconfig/loader.go` (mainapi, `:148`) และ `backend/internal/goapi/setupconfig/loader.go` (goapi, `:196`) — goapi เรียกซ้ำอีกครั้งใน `Init()` (`bootstrap.go:59`)
- ลำดับหาไฟล์: `/app/bootstrap/bootstrap.json` → `/app/bootstrap.json` → `bootstrap.json` → `config/bootstrap.json` (`backend/internal/setupconfig/loader.go:99-104`, goapi `:146-151`); ถ้ามี `custom_config.json` ข้าง ๆ จะ merge ทับ (`:106-113,177-187`); log จริงบน dev: "อ่าน bootstrap.json จาก /app/bootstrap.json"
- section ที่ map: mongodb, mongodbdev/uat/pro/production, postgresql, clickhouse, service, integrations, storage, kafka (`backend/internal/setupconfig/loader.go:193-203`); key ใน JSON ถูก map เป็นชื่อ env เช่น `serviceport → SERVICE_PORT` (`:59`), goapi มี `enablekafka → ENABLE_KAFKA`, `kafkaconsumergroupversion → KAFKA_CONSUMER_GROUP_VERSION` (goapi loader `:78-79`) โดยตัด `_` ออกก่อน lookup (goapi `:284`)
- **precedence จริง: ค่าใน bootstrap.json ชนะ env ของ process** เพราะ `applyBootstrapSection` เรียก `os.Setenv` ทุก key ที่ไม่ว่างโดยไม่เช็คว่ามี env อยู่ก่อน (`backend/internal/setupconfig/loader.go:226-248`); env จาก compose จึงมีผลเฉพาะ key ที่ไม่อยู่ใน JSON (เช่น `DEV_API_MODE`, `CONSUMER_GROUP_NAME`) — Redis ไม่มีใน mapping ของ bootstrap; Kafka มี section `kafka.serverurl → KAFKA_SERVER_URL` ทั้ง 2 loader (mainapi `:82-83,203`, goapi `:118,251`) แม้ comment ใน goapi loader `:17` จะบอกว่าไม่มี; ถ้า env ยังว่างค่อยเติม default `redis:6379`, `kafka:9092` (mainapi loader `:267-274`, goapi `:329-340`)
- goapi ยังเขียนกลับไฟล์ได้ (`UpdateBootstrapJSON` `:514`) และ `ReloadAndReconnect` + แจ้ง mainapi ผ่าน `POST <MAINAPI_INTERNAL_URL>/reload-config` default `http://mainapi:8080` (`:370,398-404`) ซึ่ง mainapi รับที่ `backend/main.go:323-340` — default port 8080 ไม่ตรงกับ 8888 ที่ mainapi ฟังจริง ถ้าไม่ตั้ง `MAINAPI_INTERNAL_URL` (ยังไม่ตรวจว่า prod/local ตั้งไว้)
- ไฟล์ `backend/bootstrap.json`, `bootstrap.local.json`, `custom_config.json`, `custom_config.local.json` มีอยู่ใน working tree แต่ **ไม่ถูก track ใน git** (`git ls-files` ว่าง, `git check-ignore -v` ชี้ `backend/.gitignore:62,64`; กฎอยู่ที่ `.gitignore:13`, `backend/.gitignore:62-64`) — บทความนี้ไม่ได้เปิดดูค่าในไฟล์

## 10. พอร์ตบนเครื่อง dev (compose + docker ps 2026-09-07, read-only)
| service | พอร์ต host | ที่มา | runtime |
|---|---|---|---|
| mainapi | `8888:8888` | `backend/docker-compose.yml:70-71` | up, healthy |
| mongodb | `127.0.0.1:27017` | `backend/docker-compose.local.yml:15-16` | container up 7 วัน แต่ `docker port mongodb` ว่าง; `docker inspect` (read-only) พบ `HostConfig.PortBindings` ตั้ง `127.0.0.1:27017` ไว้ แต่ `NetworkSettings.Ports["27017/tcp"]` ว่าง → binding ไม่ active; container สร้างจาก `docker-compose.yml` + `docker-compose.local.yml` (label compose) — สาเหตุยังไม่ตรวจ (ต้อง recreate ซึ่งเป็น write) |
| postgres | `127.0.0.1:5432` | `backend/docker-compose.local.yml:30-31` | up |
| minio | `127.0.0.1:9100→9000`, `9001` console | `backend/docker-compose.local.yml:45-47` | up |
| redis | `127.0.0.1:6379` | `backend/docker-compose.yml:32-33` | up |
| kafka | `127.0.0.1:9092` (`confluentinc/confluent-local`) | `backend/docker-compose.yml:40-44` | up; `kafka-topics --list` (2026-09-07) พบ topic `when-*` 97 topic (+ `TEST-CONNECT`, `__consumer_offsets`) — ยังไม่ได้เทียบว่าครบตาม producer ทุกตัว |
| frontend | `next dev` :3000 (ไม่อยู่ใน compose local) | `frontend/package.json:10` | `curl localhost:3000/` ตอบ HTTP 200 (2026-09-07) — รันอยู่ แต่ยังไม่ตรวจว่าเป็น `next dev` หรือ `next start` |

## 11. โครงสร้าง directory ระดับบน (git ls-files, commit d93a210d)
| directory | หน้าที่ | ไฟล์ |
|---|---|---|
| `backend/` | Go monolith: `internal/` 67 แพ็กเกจ 1,686 ไฟล์ (ใหญ่สุด `internal/transaction` 543, `internal/goapi` 269, `internal/product` 122), `pkg/` 59, `cmd/` 52, `assets/` 33, `http_test/` 17, `cluster/` 16, `migrations/` 8, Dockerfile ×7, compose ×3, `.github/workflows` เก่า 9 ไฟล์ | 1,944 (1,780 .go) |
| `frontend/` | Next.js: `src/app` 129 (route handlers 36 `route.ts` + test), `src/lib` 58, `src/components` 41, `src/locales` 12 | 311 |
| `scratch/` | งานทดลอง / UAT harness | 18 |
| `tests/` | Playwright spec ระดับ repo (`uat-crud.spec.ts`, employee/login/currency) + `playwright.config.ts` ที่ root | 15 |
| `tools/` | สคริปต์ช่วยงาน | 13 |
| `deploy/` | `deploy/account/*`: compose prod, Caddyfile, provision/rotate scripts, mongo-init, docs/runbooks/RECOVERY-READINESS.md | 9 |
| `docs/` | `docs/kms/` (ชุดนี้) + `docs/skills/{ui-scale-polish,audit-mongomodel-sync}` | 6 |
| `scripts/` | สคริปต์ระดับ repo | 5 |
| `.github/` | `workflows/ci.yml` jobs: backend-test, frontend-test, backend-outbox-integration, backend-projection-kafka-integration (`ci.yml:10,42,82,138`) — ล็อกเพราะ billing ตั้งแต่ 2026-09-03 (`docs/handoff/HANDOFF-2026-09-06.md:13`) จึงยังไม่มีหลักฐาน run จริง | 1 |
| root | `AGENTS.md`, `CLAUDE.md`, `docs/kms/00-source-router.md`, `docs/reference/CODE-MAP.md`, `docs/handoff/HANDOFF-*.md` ×2, `package.json`, `package-lock.json`, `playwright.config.ts`, `.mcp.json`, `.gitignore`, `.ignore` (exclude `**/build/`), `.mongo_all_cols.txt`, `.mongo_audit_out.txt` | 14 |

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ
- `/etc/bcai-account/backend.env` และ `frontend.env` ของ prod ไม่อยู่ใน repo → ไม่รู้ว่า mainapi prod ใช้ `DEV_API_MODE` จาก Dockerfile (=2) หรือถูก override, และ `MAINAPI_INTERNAL_URL` / `ENABLE_KAFKA` ตั้งค่าอย่างไร — ถามลุงจืด
- ยังไม่ได้นับว่า route ใน goapi กี่เส้นยังเรียก `myclickhouse` จริง (พบ 15 ไฟล์ import); lead ระบุ 3 legacy route — ต้อง trace ต่อในบทความฐานข้อมูล/analytics
- `cmd/app/main.go` (ใช้โดย `backend/dev.sh`) และ `cmd/*` อีก ~28 ตัวยัง build ผ่านหรือไม่ ตรวจได้เฉพาะใน container `golang:1.26` (ยังไม่รัน); `backend/Makefile` อ้าง `cmd/` ที่หายไปแล้วหลายตัว — ควรลบหรือไม่ ให้ลุงจืดตัดสิน
- `pkg/microservice/persister_{clickhouse,elk,opensearch}.go` มีผู้เรียกจริงหรือไม่ ยังไม่ตรวจ
- mongodb container บนเครื่อง dev: `docker inspect` ยืนยันว่า PortBindings ตั้ง `127.0.0.1:27017` แต่ port ไม่ active (ดู §10) — สาเหตุ (เช่น สร้าง container ตอน port ถูกยึด) ยังไม่ตรวจ เพราะต้อง recreate container
- `backend/bootstrap*.json` + `custom_config*.json` ไม่ได้ถูก track (ตรวจแล้ว §9) — ค่าในไฟล์ยังไม่ได้เปิดดู จึงยังไม่รู้ว่า local/prod ตั้ง `enablekafka`, `kafkaconsumergroupversion`, `MAINAPI_INTERNAL_URL` อย่างไร
- จำนวน module HTTP นับได้ 132 (grep `New*Http(ms, cfg)` ใน `backend/main.go:357-547` = 133 บรรทัด, 1 บรรทัดเป็น comment) — ยังไม่ได้ไล่ว่าแต่ละ module ลงทะเบียน route กี่เส้น
- ยังไม่ตรวจว่า `/reload-config` ของ mainapi (`backend/main.go:323`) ถูก frontend/goapi เรียกจริงในการทำงานปกติหรือไม่ (frontend บล็อก `/backend/reload-config` ไว้ที่ `frontend/next.config.ts:66`)
