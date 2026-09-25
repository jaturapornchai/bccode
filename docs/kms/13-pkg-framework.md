# แพ็กเกจ framework กลางของ backend: pkg/microservice, pkg/*, internal/config, utils, models
> ตรวจล่าสุด: 2026-09-25 (commit c58f0b62) — ผู้เขียน: AI reader (กลุ่มงาน G2-auth-framework + รอบ review); ทุกข้อเท็จจริงอ้าง path:line ยืนยันด้วย `ls`/`grep -n` จริงในรอบนี้
> persister/driver ของ MongoDB, ClickHouse, OpenSearch/Elasticsearch, Kafka และ Redis รวมถึงแพ็กเกจ/ไฟล์ที่พึ่งพาทั้งหมด ถูกลบออกเมื่อ 2026-09-23 ([decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md)) — framework ปัจจุบันมี **PostgreSQL ตัวเดียว**; **ห้ามเพิ่มระบบเหล่านี้กลับเข้ามา**

## 1. ภาพรวม
- module ชื่อ `smlcloudplatform` บน Go 1.26 (`backend/go.mod:1-3`) — ทุก path ในบทความนี้ relative จาก repo root
- `main.go` โหลด bootstrap.json → env ก่อน แล้วสร้าง config + Microservice (`backend/main.go:77,83-84`); `NewMicroservice` มีผู้เรียก**จุดเดียว**คือ `backend/main.go:84` (`backend/cmd/` เหลือแค่ `cmd/glseed/` ซึ่งเป็นสคริปต์ seed ไม่ใช่ service)
- Microservice ถือ echo, persister (PostgreSQL/gorm), cacher (`cache_entries`), websocket pool (ไม่มีผู้เรียก), logger, middleware manager (`backend/pkg/microservice/microservice.go:48-62,93-113`) และมี `Echo()` ให้ goapi mount route ตรง ๆ — `ms.Echo()` ถูกเรียก 5 จุด (ไม่นับ test); goapi import แพ็กเกจนี้ไฟล์เดียว (`backend/internal/goapi/bootstrap.go`)
- โมดูลธุรกิจ (auth, shop, organization, GL) **ไม่ใช้** `Persister` ของแพ็กเกจนี้ — เปิด `*sql.DB` เองผ่าน `centraldb.Open()` (`backend/internal/centraldb/centraldb.go:26`) หรือ `internal/goapi/mypg` แล้วเขียน SQL ตรง

## 2. pkg/microservice — Inventory
| module | หน้าที่ | entry | สถานะ | อ้างอิง |
|---|---|---|---|---|
| Microservice core | สร้าง echo + validator + CORS `*`, อ่าน `MODE`, boot แล้วทดสอบต่อ PostgreSQL | `NewMicroservice`, `CheckReadyToStart`, `RegisterBackgroundWorker`, `Start`, `Stop`, `Cleanup` | LIVE | `backend/pkg/microservice/microservice.go:66-113,115-131,160,165-247` |
| HTTP wrapper | GET/POST/PUT/PATCH/DELETE ห่อ echo + alias `/v1/...` อัตโนมัติ, listen `0.0.0.0:<port>` | `ms.GET(...)` ฯลฯ, `RegisterHttp` | LIVE — ทุก route ถูก register 2 ครั้ง (legacy + `/v1`) | `backend/pkg/microservice/microservice_http.go:16-140,154,168` |
| Persister (PostgreSQL/gorm) | CRUD/Where*/Raw/AutoMigrate/Transaction | `ms.Persister`, `NewPersister`, `NewPersisterWithDB` | ใช้แค่ `CheckReadyToStart` และ liveness (`TestConnect`) — ไม่มีผู้เรียกนอกแพ็กเกจ | `backend/pkg/microservice/persister.go:22-68,338-356`; `backend/pkg/microservice/microservice.go:310` |
| `PersisterTenant` | แคช persister ต่อ `host/dbName` | `PersisterTenant`, `RegisterTenantModel` | DEAD (ไม่มีผู้เรียก) | `backend/pkg/microservice/microservice.go:275-308` |
| Cacher | KV/Hash + `ConsumeHash` (อ่าน+บันทึก marker+ลบในคำสั่งเดียว) บนตาราง `cache_entries` (สร้างเองด้วย `CREATE TABLE IF NOT EXISTS`); TTL = คอลัมน์ `expires_at` + `PurgeExpired` ที่ main เรียกทุก 10 นาที | `NewCacher(db *sql.DB)`, `ms.SetCacher`, `ms.Cacher()` | LIVE (auth session) | `backend/pkg/microservice/cacher.go:18-72,204,385,398`; `backend/main.go:90-112` |
| Context | `IContext` abstraction ของ handler — implementation เดียวคือ `HTTPContext` | `NewHTTPContext` | LIVE | `backend/pkg/microservice/context.go:12-30`; `backend/pkg/microservice/context_http.go:21` |
| AuthService + live authorization | session/access/refresh ใน `cache_entries` (prefix `auth-`, `xapikey-`, `refresh-`, `session-`, `session-revoked-`, `refresh-used-`), access ≤ 15 นาที, session ≤ 12 ชม., สิทธิ์อ่านสดจาก `users`/`holding_members`/`holdings`/`companies`/`branches` | `NewAuthService` (4 จุด: `backend/main.go:118`, `backend/internal/authentication/authentication_http.go:70`, `backend/internal/shop/shop_http.go:32`, `backend/internal/goapi/bootstrap.go:307`), `NewAuthServicePrefix` (test), `NewLegacyAuthServicePrefix` + flag `allowLegacyToken` (ไม่มีผู้เรียก — DEAD), `MWFuncMixShop` (`backend/main.go:148`) | LIVE | `backend/pkg/microservice/auth.go:19-45,74-78,87-131`; `backend/pkg/microservice/live_authorization.go:24-121` — รายละเอียดดู [03-auth-tenancy.md](03-auth-tenancy.md) |
| `JwtService` | JWT HS256 + prefix `auth-` | `NewJwtService` | DEAD (มีแค่ `jwt_test.go` เรียก) | `backend/pkg/microservice/jwt.go:28-37` |
| Websocket pool | upgrade + เก็บ conn ตาม id | `Websocket`, `WebsocketClose`, `WebsocketCount` | DEAD (ไม่มีผู้เรียกนอกแพ็กเกจ) | `backend/pkg/microservice/microservice.go:334-361`; `backend/pkg/microservice/microservice_pool.go:13` |
| Liveness | `/healthz` ตรวจ `Cacher.Healthcheck()` แล้ว `TestConnect()` ของ persister PostgreSQL ทุกตัวที่เปิดไว้ | `RegisterLivenessProbeEndpoint` (`backend/main.go:149`) | LIVE | `backend/pkg/microservice/microservice_liveness.go:9-52` |
| Metrics/Tracing | Prometheus namespace `smlcloudplatform` | `HttpUsePrometheus` (`backend/main.go:114`) LIVE / `HttpUseJaeger` DEAD (ไม่มีผู้เรียก) | — | `backend/pkg/microservice/microservice.go:372-382`; `backend/pkg/microservice/microservice_metric.go:5` |
| File / Image persister | `NewFilePersister()` คืน S3/MinIO client (`NewPersisterR2`) ใช้จริงใน `backend/internal/media/media_services.go:33`; `PersisterFile` (local disk) และ `PersisterImage`/`IPersisterImage` ไม่มีผู้เรียก | `NewFilePersister` | LIVE (S3/MinIO) / local-disk + image DEAD | `backend/pkg/microservice/persister_factory.go:4-6`; `backend/pkg/microservice/persister_file_r2.go:30,51-68`; `backend/pkg/microservice/persister_file.go`; `backend/pkg/microservice/persister_image.go` |
| storage_name | normalize ชื่อ table/field เป็น lowercase snake_case | `NormalizeStorageName`, `NormalizeStorageFieldName`, `NormalizePostgresTableName` | DEAD (ไม่มีผู้เรียกนอกไฟล์) | `backend/pkg/microservice/storage_name.go:122-160` |
| util | `NewUUID` (ksuid) ใช้ใน `auth.go` สร้าง session/token id; `escapeName`, `randString` ไม่มีผู้เรียก | `NewUUID` | LIVE (NewUUID) / DEAD (ที่เหลือ) | `backend/pkg/microservice/util.go:15-62`; `backend/pkg/microservice/auth.go:452,467,494,691-692` |
| models | `Pageable{Query,Page,Limit,Sorts}`, `PageableStep{Query,Skip,Limit,Sorts}`, `UserInfo{Username,HoldingCode,BusinessCode,Role,UID,SessionUID,...}` | — | LIVE | `backend/pkg/microservice/models/pageable.go:3-27`; `backend/pkg/microservice/models/user_info.go` |

### 2.1 พฤติกรรมตอน boot ที่ต้องรู้
1. `CheckReadyToStart` ตรวจอย่างเดียว: ถ้า `POSTGRES_HOST != ""` สร้าง `Persister` แล้ว `TestConnect()`; ล้มเหลว → `NewMicroservice` คืน error และ main `panic` (`backend/pkg/microservice/microservice.go:115-131`; `backend/main.go:84-87`)
2. Cacher ถูกสร้างและ `SetCacher` **หลัง** `NewMicroservice` (`backend/main.go:90-98`) — ต่อฐานกลาง `bcai_projection` ผ่าน `mypg.PgSqlFastConnect`; ต่อไม่ได้ → `panic`

## 3. pkg/* อื่น ๆ
`backend/pkg/` เหลือ 5 แพ็กเกจ: `apperr`, `memorycache`, `microservice`, `textguard`, `validator` (แพ็กเกจอื่นถูกลบพร้อมกันเมื่อ 2026-09-23)

| pkg | API หลัก | ผู้ใช้ | สถานะ | อ้างอิง |
|---|---|---|---|---|
| apperr | sentinel `ErrNotFound/ErrDuplicate/ErrValidation/...` + `AppError{Code json:"errorcode", ThaiMsg json:"message_th", HTTPStatus}`; `Respond/RespondErr` เขียน JSON ผ่าน `IContext`; ข้อความที่ผู้ใช้เห็นต้องผ่าน `language.Text(key, lang)` ไม่ใช่ `err.Error()` ดิบ (ตัวอย่าง `localizedAppError`/`shopUserAppError` ใน `backend/internal/shop/shopuser_http.go:260-337`) — จอตั้งค่าแปลง response เป็นประโยคผู้ใช้ด้วย `frontend/src/components/system-settings/user-facing-error.ts` | `apperr.Respond*` 175 จุด (ไม่นับ test); `Middleware()` 0 จุด (มีแค่ในคอมเมนต์ `middleware.go:15`) | LIVE / `Middleware()` DEAD | `backend/pkg/apperr/codes.go:10-88`; `backend/pkg/apperr/errors.go:15-147`; `backend/pkg/apperr/respond.go:18-31`; `backend/pkg/apperr/middleware.go:16` |
| memorycache | go-cache default 5 นาที / sweep 10 นาที | `backend/pkg/microservice/auth.go:54,106,134` เท่านั้น | LIVE | `backend/pkg/memorycache/memorycache.go:19-36` |
| validator | go-playground v10 ใช้ชื่อ field จาก json tag; คืน error แรกตัวเดียวเป็นภาษาอังกฤษ | ตั้งเป็น `e.Validator` ใน `NewMicroservice` (`backend/pkg/microservice/microservice.go:75`) | LIVE | `backend/pkg/validator/validator.go:18-56` |
| textguard | `NULField(v any) string` หา path แรกที่มีอักขระ `U+0000` (PostgreSQL TEXT/JSONB ปฏิเสธด้วย SQLSTATE 22021 — มักมาจากข้อความ paste จาก PDF) + `Message(...)` สร้างข้อความผู้ใช้จาก key ใน languages.tsv | `backend/internal/organization/businesstype/{businesstype_http.go,business_type_rows.go}`, `backend/internal/organization/rolepermission/{permission_set_rows.go,role_permission_http.go}`, `backend/internal/shop/shopuser_http.go` | LIVE | `backend/pkg/textguard/textguard.go:14,21`; `backend/pkg/textguard/textguard_test.go` |

## 4. internal/config — ชื่อ env ทั้งหมด (ค่าจริงมาจาก bootstrap.json + custom_config.json, ไม่โหลด .env: `backend/internal/config/config.go:49-50`)
bootstrap.json มี 4 หมวด `postgresql`, `service`, `integrations`, `storage` (`backend/internal/setupconfig/loader.go:11-16`) map เป็น env ผ่าน `configMapping` (:20); `backend/internal/config/` มีแค่ `config.go`, `config_postgresql.go`, `config_http.go`, `config_logger.go`, `config_file_storage.go`, `data_environment.go` (+ test)

| interface | env keys (key name เท่านั้น) | default/พฤติกรรม | อ้างอิง |
|---|---|---|---|
| `IConfig` | `MODE` (ถ้าว่างจะ `Setenv("MODE","development")`), `SERVICE_NAME`, `PATH_PREFIX`, `HTTP_CORS`, `JWT_SECRET_KEY` (ไม่มี fallback — token จริงไม่ใช่ JWT อยู่ดี), `GOOGLE_CLIENT_ID` (มี default เพราะเป็นค่า public) | test ยืนยัน JWT ไม่มี fallback (`backend/internal/config/config_security_test.go:8`) | `backend/internal/config/config.go:9-22,53-111` |
| data environment | `BC_ENV` > `APP_ENV` > `RUN_ENV` > `ENVIRONMENT` > `MODE` → normalize เป็น `dev/uat/pro` | ไม่ตั้ง = dev; ฟังก์ชันเลือก environment ไม่มีผู้เรียกนอกแพ็กเกจแล้ว — ใช้จริงแค่ค่าคงที่ `DataEnvironmentDev` ที่ dev-login เทียบกับ `BC_ENV` (`backend/internal/authentication/authentication_http.go:103,115`) | `backend/internal/config/data_environment.go:8-39` |
| `IPersisterConfig` (PG) | `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB_NAME` (default `postgres`), `POSTGRES_USERNAME`, `POSTGRES_PASSWORD`, `POSTGRES_SSL_MODE` (default `disable`), `POSTGRES_TIMEZONE` (default `UTC` — DB เก็บ UTC), `POSTGRES_LOGGER_LEVEL` | DSN ประกอบใน `backend/pkg/microservice/persister.go:75-87` | `backend/internal/config/config_postgresql.go:23-66` |
| `IHttpConfig` | `SERVICE_PORT` > `HTTP_PORT` (default 8080); `PATH_PREFIX` > `HTTP_PATH_PREFIX`; `HTTP_IGNORE_LOG_URLS` (default `/healthz`); `HTTP_CORS` | — | `backend/internal/config/config_http.go:20-53` |
| `ILoggerConfig` | `LOG_LEVEL` (default info), `LOG_ENCODER`, `MODE` (development → devMode) | devMode เขียน `logs/mainapi_debug.log` และล้างไฟล์ทุกครั้งที่ start | `backend/internal/config/config_logger.go:15-25`; `backend/internal/logger/logger.go:255-271` |
| Storage / S3 | `STORAGE_DATA_PATH`, `STORAGE_DATA_URI` (local-disk persister DEAD); S3: `S3_ACCOUNT_ID/S3_ENDPOINT/S3_ACCESS_KEY_ID/S3_SECRET_ACCESS_KEY/S3_BUCKET_NAME/S3_REGION` (fallback ชื่อ `R2_*`) | — | `backend/internal/config/config_file_storage.go:9-15`; `backend/pkg/microservice/persister_file_r2.go:51-68` |

- Auth env เฉพาะ mainapi (อ่าน `os.Getenv` ตรง ไม่ผ่าน `IConfig`/`configMapping`): `BCAI_DEMO_LOGIN_ENABLED`, `BCAI_DEMO_USERNAME` (`backend/internal/demo/demo.go:16,21`), `BC_ENV`, `BCAI_DEV_LOGIN_ENABLED`, `BCAI_DEV_LOGIN_USER_UID`, `BCAI_DEV_LOGIN_SECRET` (`backend/internal/authentication/authentication_http.go:103-109`), `RELOAD_CONFIG_SECRET` (`backend/main.go:140`)

## 5. internal/utils
| ไฟล์ | API | สถานะ | อ้างอิง |
|---|---|---|---|
| request_param.go | `GetPageable`, `GetPageableStep`, `GetPageParam` clamp page 1..2e9, limit 1..100000, default limit 20 | `GetPageable`/`GetPageableStep` LIVE | `backend/internal/utils/request_param.go:9-18,25-58,129-150` |
| random.go | `NewGUID` (ksuid — ค่า `guidfixed`), `NewID` (xid), `NewUUID`, `RandNumber` | `NewGUID`/`RandNumber` LIVE; `NewID`/`NewUUID` ไม่มีผู้เรียกนอกแพ็กเกจ | `backend/internal/utils/random.go:41-61` |
| password.go | `HashPassword` = argon2id (m=19 MiB, t=2, p=1); `CheckHashPassword` รองรับ hash `$2...` (bcrypt) เดิม | LIVE | `backend/internal/utils/password.go:16-36,38-63` |
| business_code.go / holding_code.go | `NormalizeBusinessCode`; `NormalizeHoldingCode` regex `^[a-z][a-z0-9_-]{2,29}$` | LIVE | `backend/internal/utils/business_code.go:6`; `backend/internal/utils/holding_code.go:9-11` |
| auth.go | `NormalizeUsername/Phonenumber/Email/Name` | Username/Phonenumber/Email LIVE; `NormalizeName` ไม่มีผู้เรียกนอกแพ็กเกจ | `backend/internal/utils/auth.go:7-24` |
| hash.go, loadkey.go | `FastHash`, `LoadKey`, `LoadFile` | DEAD (ไม่มีผู้เรียกนอกแพ็กเกจ) | `backend/internal/utils/hash.go:5`; `backend/internal/utils/loadkey.go:10,21` |

## 6. internal/models (shape ร่วมที่ยังใช้ข้าม module)
- `Identity{HoldingCode,GuidFixed}`, `HoldingCodeentity`, `DocIdentity{GuidFixed}`, `PartitionIdentity{ParID json:"-"}` — มี gorm tag `primaryKey` (`backend/internal/models/identity.go:3-17`)
- `ActivityDoc` (json:"-" ทุก field) vs `Activity` (json เปิด) (`backend/internal/models/activity.go:5-27`)
- `ApiResponse{Success,Message,DocNo,ID,Data,Pagination,Total}` (`backend/internal/models/api_response.go:5`); `NameX{Code,Name}` + `JSONB []NameX` เป็น driver.Valuer/Scanner สำหรับคอลัมน์ jsonb (`backend/internal/models/name.go:33-38,55-67`; `backend/internal/models/postgresql.go:9-27`); `PaginationData` (`backend/internal/models/pagination.go:11`)
- ไม่มีผู้เรียกนอกแพ็กเกจ (DEAD): `Trans`/`TransItemDetail` (`transactionitem.go`), `SearchFilter` (`search_filter.go`), `Datetime` (`date_time.go`), `Index` (`index.go`), `BulkImport*` (`bulkimport.go`), `ISODate` (`isodate.go`), `XSort` (`xsort.go`), `organization.go` (คอมเมนต์ทั้งไฟล์)

## 7. กับดัก / ข้อควรระวัง (ตรวจจากโค้ดจริง 2026-09-25)
1. CORS ซ้อน 2 ชั้น: `NewMicroservice` ใส่ `AllowOrigins:*` แล้ว `HttpUseCors` ใส่อีกชุดจาก `HTTP_CORS` (`backend/pkg/microservice/microservice.go:77-80,384-390`; `backend/main.go:150`)
2. ทุก route ถูก register 2 ครั้ง (legacy + `/v1` alias) ยกเว้น path ที่ขึ้นต้น `/v1/` หรือ `/api/v1/` — นับ route ใน `ms.echo.Routes()` จะเป็น 2 เท่า (`backend/pkg/microservice/microservice_http.go:28-37`)
3. `IPersisterImage.Upload(fh)` (`backend/pkg/microservice/persister_image.go:9`) ไม่ตรงกับ `PersisterImage.Upload(fh, name, ext)` (:22) → struct ไม่ implement interface; compile ผ่านเพราะไม่มีใครใช้ทั้งคู่
4. `validator.Validate` ลงทะเบียน translation ใหม่ทุก request และคืน error ตัวแรกเป็นภาษาอังกฤษ (`backend/pkg/validator/validator.go:35-56`) — ขัดกฎ Thai-first ถ้าโยนตรงถึงหน้าจอ
5. โค้ดตายใน `pkg/microservice`: `PersisterTenant`, `Websocket*`, `HttpUseJaeger`, `JwtService`, `NewLegacyAuthServicePrefix` (`backend/pkg/microservice/auth.go:147`), `MWFuncSession`, `PersisterImage`, `PersisterFile`, `escapeName`/`randString`, `storage_name.go` ทั้งไฟล์ (ดู §2)
6. Test ในแพ็กเกจ: `auth_session_test.go` (PostgreSQL จริงจาก `BC_GL_TEST_POSTGRES_DSN`, ข้ามถ้าไม่ตั้ง), `live_authorization_test.go`, `live_authorization_integration_test.go` (tag `integration`), `persister_transaction_test.go` (sqlmock), `persister_file_r2_test.go`, `jwt_test.go` — ยังไม่ได้รันในรอบนี้

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ
- ยังไม่รัน `go build ./...`/`go vet ./...`/test ในรอบนี้ (งานเอกสารอย่างเดียว)
- ยังไม่ได้วัดต้นทุนของ `cache_entries` ต่อ request (read + write ทุกคำขอ) และยังไม่ยืนยันว่า `PurgeExpired` ทำงานบน production
- คำถามถึงลุงจืด: (1) จะลบโค้ดตายชุด §7 ข้อ 5 ออกจาก `pkg/microservice` เลยไหม (Zero-Bloat) (2) จะลบ `IPersisterImage`/`PersisterImage` ทิ้งหรือแก้ signature ให้ตรง
