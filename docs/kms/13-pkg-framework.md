# แพ็กเกจ framework กลางของ backend: pkg/microservice, pkg/*, internal/config, utils, models, repositories
> ตรวจล่าสุด: 2026-09-07 (commit d93a210d) — ผู้เขียน: AI reader; ทุกข้อเท็จจริงอ้าง path:line · ตรวจซ้ำโดย fact-checker

## 1. ภาพรวม
- module ชื่อ `smlcloudplatform` บน Go 1.26 (backend/go.mod:1-3) — ทุก path ในบทความนี้ relative จาก repo root, ตัวเลข "ผู้ใช้" = จำนวนบรรทัดที่ `rg -F` เจอในไฟล์ non-test นอกแพ็กเกจนั้นเอง (นับ 2026-09-07, เป็นค่าประมาณ)
- `main.go` โหลด bootstrap.json → env ก่อน แล้วค่อยสร้าง config + Microservice (backend/main.go:236, 248-249); `NewMicroservice` ถูกเรียกจาก main.go + 19 ไฟล์ใน `cmd/*` (รวม 20 จุด) แต่ image `mainapi` ของ stack local build จาก `backend/main.go` เท่านั้น (backend/docker-compose.local.yml:106 → backend/Dockerfile.local:18; backend/Dockerfile:21 เช่นกัน) — `Dockerfile-member` (cmd/member/main.go, บรรทัด 20) และ `Dockerfile.goapi` (./cmd/goapi, บรรทัด 15) มีอยู่ แต่ยังไม่ตรวจว่าถูกใช้ deploy จริง
- Microservice ตัวเดียวถือ echo, pool ของ persister/cacher/producer, websocket pool, logger (backend/pkg/microservice/microservice.go:51-76) และมี `Echo()` ให้ goapi mount route ตรง ๆ (microservice.go:643; `ms.Echo()` ถูกเรียก 12 จุด, goapi import แพ็กเกจนี้เพียง 1 ไฟล์)

## 2. pkg/microservice — Inventory
| module | หน้าที่ | entry | สถานะ | อ้างอิง |
|---|---|---|---|---|
| Microservice core | สร้าง echo + validator + CORS `*`, อ่าน `MODE`, test connection ทุก backend ตอน boot | `NewMicroservice`, `CheckReadyToStart`, `Start`, `Stop`, `Cleanup` | LIVE | microservice.go:80-132, 134-200, 233-330 (Start :233, Stop :296, Cleanup :304) |
| HTTP wrapper | GET/POST/PUT/PATCH/DELETE ห่อ echo + สร้าง alias `/v1/...` อัตโนมัติ, listen `0.0.0.0:<port>`, shutdown 10s | `ms.GET(...)` ฯลฯ, `startHTTP` | LIVE | microservice_http.go:16-39, 141-160, 162-166 |
| Consumer (librdkafka) | `Consume(servers, topic, group, timeout, h)` → goroutine ต่อ topic; barcode topic แยกไป worker kafka-go | `Consume`, `ms.RegisterConsumer(` (39 จุด non-test ทั้ง repo, 27 จุดใน main.go), `ConsumeFromBegining` (0 จุด) | LIVE เฉพาะ `DEV_API_MODE` "" / "1" (main.go:654-726) | microservice_consumer.go:15-61, 101-115 |
| barcode_projection_consumer | reader kafka-go group `<group>-projection`, `FirstOffset`, commit หลัง handler สำเร็จ, retry ทุก 2s, recover panic ต่อ message | `consumeBarcodeProjection` (เรียกจาก `Consume` เมื่อ `projection.IsBarcodeTopic`) | LIVE | barcode_projection_consumer.go:19, 23-55; internal/product/projection/consumer.go:45 |
| Persister (PostgreSQL/gorm) | CRUD/Where*/Raw/AutoMigrate/Transaction; cache ต่อ `cfg.Host()`; `PersisterTenant` cache ต่อ `host/dbName` + AutoMigrate `tenantModels` + `DBCheckHook` | `Persister`, `PersisterTenant`, `RegisterTenantModel`, `DBCheckHook` | LIVE (PersisterTenant ใช้ 1 จุด internal/shop/shop_http.go:399; hook ตั้งที่ internal/migration/postgres_migration.go:16-22) | persister.go:22-46, 55-66, 104-131; microservice.go:343-401 |
| PersisterMongo | Find/FindPage/Aggregate/SoftDelete*/Transaction/CreateIndex; collection มาจาก `CollectionName()` ของ model เท่านั้น | `MongoPersister(` (190 จุด), `NewPersisterMongo(` (25 จุด) | LIVE — source of truth | persister_mongo.go:22-54, 101-157, 159-167, 704-743 |
| PersisterClickHouse | Count/Select/Exec/Create/CreateInBatch ผ่าน clickhouse-go v2 | `ClickHousePersister` (6 ไฟล์ legacy: productimport, productbarcode http+consumer, eorder, stockbalanceimport, reportqueryc) | STUBBED ระดับ goapi / container ถูกถอด 2026-09-06 | persister_clickhouse.go:13-20, 47-70; microservice.go:414-428 |
| Cacher (Redis) | KV/Hash/List/BitField + retry ping; TLS min 1.2 (ไม่มี PubSub ใน `ICacher`) | `.Cacher(` (143 จุด), `NewCacher` | LIVE (auth session อยู่ที่นี่) | cacher.go:18-79, 95-130, 132-162, 1615-1631 |
| Producer (librdkafka) | `SendMessage` แบบ sync รอ delivery report; timeout default 43,200,000 ms (12 ชม.); `NewProducerWithTimeout` สำหรับ outbox | `.Producer(` (90 จุด), `NewProducer(` (18 จุด) | LIVE | producer.go:13-54, 76-122, 125-133 |
| MQ admin | CreateTopic/CreateTopicR ผ่าน AdminClient | `NewMQ` (36 จุด) | LIVE | mq.go:14-18, 44-53 |
| Context | `IContext` abstraction ให้ handler ใช้ร่วมกันทั้ง HTTP/Consumer/AsyncTask; `UserInfo()` อ่านจาก echo key `"UserInfo"` | `NewHTTPContext`, `NewConsumerContext` | LIVE | context.go:11-33; context_http.go:70-77, 94-99; context_consumer.go:31-35, 91-94 |
| AuthService + live authorization | session/access/refresh ใน Redis (prefix `auth-`, `xapikey-`, `refresh-`, `session-`, `session-revoked-`, `refresh-used-`), access ≤ 15 นาที, session ≤ 12 ชม., สิทธิ์จริงอ่านสดจาก Mongo (`users`, `shopusers`, `shops`, `organizationcompanies`, `organizationbranches`) | `NewAuthService` (15 จุด), `MWFuncWithRedisMixShop` (main.go:342) | LIVE | auth.go:19-36, 38-45, 75-79, 87-113; live_authorization.go:24-26, 39-89, 100-130 |
| JwtService | JWT HS256 (jwt.go:271, 285) + RS256 (:247) + Redis prefix `auth-` (:34) | `NewJwtService` | DEAD (ผู้เรียก 3 จุดเป็น comment ทั้งหมด: cmd/authenticationservice/main.go:20, cmd/app/main.go:130, cmd/ws/main.go:20) | jwt.go:28-37 |
| AsyncTask | AsyncPOST/AsyncPUT ผ่าน Kafka + cache | `AsyncPOST`, `AsyncPUT` | DEAD (0 ผู้เรียก) | microservice_asynctask.go:111-136 |
| Websocket pool | upgrade + เก็บ conn ตาม id | `Websocket`, `WebsocketClose` | LIVE เฉพาะ vfgl journal (internal/vfgl/journal/journal_ws.go:75, 214, 321) | microservice.go:486-512 |
| Liveness | `/healthz` ตรวจ cacher ทุกตัว + mongo ทุกตัว | `RegisterLivenessProbeEndpoint` (main.go:343, 658) | LIVE | microservice_liveness.go:38-51, 65-76 |
| Metrics/Tracing | Prometheus namespace `smlcloudplatform`; Jaeger | `HttpUsePrometheus` (main.go:254), `HttpUseJaeger` | LIVE / Jaeger DEAD ใน `mainapi` (ไม่ถูกเรียกใน backend/main.go; เรียกเฉพาะ cmd/ws/main.go:35, cmd/app/main.go:181 เป็น comment) | microservice.go:524-534 |
| Elk / OpenSearch persister | index/create/update/delete | `ElkPersister`, `SearchPersister` | DEAD (0 ผู้เรียก) | persister_elk.go:15-21; persister_opensearch.go:18-24; microservice.go:430-462 |
| File / Image persister | `NewFilePersister()` คืน R2/S3 client เสมอ; `PersisterFile` (local disk) ไม่มีผู้เรียก | `NewFilePersister` (main.go:349), `NewPersisterImage` (main.go:350) | LIVE (S3/MinIO) / local-disk DEAD | persister_factory.go:4-6; persister_file_r2.go:51-55, 68; persister_file.go:33-38; persister_image.go:16-43 |
| storage_name | normalize ชื่อ collection/table เป็น lowercase ไม่มี `_` + alias map legacy | `NormalizeStorageName`, `NormalizeMongoCollectionName` ฯลฯ | DEAD (0 ผู้เรียกนอกแพ็กเกจ รวม test; `PersisterMongo.getCollectionName` ไม่เรียกใช้) — cmd/storage_name_migration มีสำเนา lowercase ของตัวเอง (cmd/storage_name_migration/main.go:459-467) และ main_test.go:5, 23 ทดสอบสำเนานั้น ไม่ใช่แพ็กเกจนี้ | storage_name.go:122-168; persister_mongo.go:159-167 |
| util | `NewUUID` (ksuid), `escapeName` | — | DEAD ในทางปฏิบัติ: `escapeName` ถูกเรียกเฉพาะจาก AsyncPOST/AsyncPUT ที่ DEAD (microservice_asynctask.go:15, 50); `NewUUID` 0 ผู้เรียกนอกแพ็กเกจ | util.go:20-57, 59-62 |
| models | `Pageable{Query,Page,Limit,Sorts}`, `PageableStep{Skip,...}`, `UserInfo{Username,HoldingCode,BusinessCode,Role,UID,SessionUID,...}` | — | LIVE | pkg/microservice/models/pageable.go:3-27; user_info.go:3-16 |

### 2.1 พฤติกรรมตอน boot ที่ต้องรู้ (CheckReadyToStart)
1. Mongo: ถ้า `MongodbURI() != ""` จะ `TestConnect` แล้ว cache persister ไว้ทันที (microservice.go:139-150)
2. Kafka: ถ้า `MQConfig().URI() != ""` จะ **ส่ง message จริงไป topic `TEST-CONNECT`** ทุกครั้งที่ boot (microservice.go:153-166) — หลักฐาน runtime: `docker exec kafka kafka-topics --list` บน stack local มี topic `TEST-CONNECT` อยู่จริง
3. Redis: `Healthcheck()` = ping สูงสุด 5 รอบ × 250 ms (cacher.go:1615-1631) ถ้าล้มเหลว process ไม่ start
4. PostgreSQL: ถ้า `POSTGRES_HOST != ""` → `NewPersister` ซึ่ง **panic** ถ้าเชื่อมไม่ได้ (persister.go:55-66) ก่อนจะถึง error path ของ CheckReadyToStart (microservice.go:186-196)

## 3. pkg/* อื่น ๆ
| pkg | API หลัก | ผู้ใช้ | สถานะ | อ้างอิง |
|---|---|---|---|---|
| apperr | sentinel `ErrNotFound/ErrDuplicate/ErrValidation/...` + `AppError{Code json:"errorcode", ThaiMsg, HTTPStatus}`; `Respond/RespondErr` เขียน JSON ผ่าน `IContext`; `Middleware()` แปลง error → JSON | `apperr.Respond*` 155 จุด; `Middleware()` 0 จุด (มีแค่ใน doc comment middleware.go:15) | LIVE / Middleware DEAD | backend/pkg/apperr/codes.go:10-46 (sentinel), 51-83 (constructor); errors.go:15-29 (`AppError`: Code :17, ThaiMsg :21, HTTPStatus :23), 118-138 (`Response`/`ToResponse`); respond.go:18-31; middleware.go:16-59 |
| bcrypt | `HashPassword` cost 12, `CheckPasswordHash` | 0 ไฟล์ import | DEAD (ระบบใช้ `internal/utils/password.go` แทน) | backend/pkg/bcrypt/bcrypt.go:7-14 |
| calendar | `ICurrentTime.Now()` | 0 | DEAD | backend/pkg/calendar/currenttime.go:5-9 |
| memorycache | go-cache default 5 นาที / sweep 10 นาที | pkg/microservice/auth.go เท่านั้น (internal/authentication/repositories/authentication_mongo_cache_repository.go:145 ใช้ `ttlcache` คนละ lib) | LIVE | backend/pkg/memorycache/memorycache.go:19-36 |
| round | `Round(val, places)` ปัดครึ่งขึ้น (สมมาตรกับค่าลบ) | 21 จุด ทั้งหมดใน pkg/stockcalculator/stockcalculator.go (importer non-test ไฟล์เดียว) | LIVE | backend/pkg/round/round.go:7-30 |
| serializer | ไฟล์ถูก comment ทั้งหมด (go-json) | มีเฉพาะ serializer_test.go | DEAD | backend/pkg/serializer/serializer.go:1-13 |
| stockcalculator | ต้นทุนเฉลี่ยถ่วงน้ำหนัก: `NewStockCalculator(holding, barcode, digit, qty0, amount0)`, `ApplyStock/ReduceStock/...`; digit ≤ 0 → 2; NaN → 0 | internal/stockprocess/stockcalculator.go | LIVE (เฉพาะ legacy stock consumer) | backend/pkg/stockcalculator/stockcalculator.go:8-49, 51-73 |
| tokenize | mapkha ตัดคำไทย, dict จาก env `DICT_PATH_TH` | 0 ไฟล์ import | DEAD (utils/search และ SearchRepository เรียก mapkha เอง) | backend/pkg/tokenize/tokenize.go:19-25, 47-70 |
| validator | go-playground v10 ใช้ชื่อ field จาก json tag; คืน error แรกตัวเดียวเป็นภาษาอังกฤษ | ตั้งเป็น `e.Validator` ใน NewMicroservice (microservice.go:88) | LIVE | backend/pkg/validator/validator.go:18-32, 35-56 |

## 4. internal/config — ชื่อ env ทั้งหมด (ค่าจริงมาจาก bootstrap.json + custom_config.json, ไม่โหลด .env: backend/internal/config/config.go:58-59; backend/main.go:234-236)
| interface | env keys (key name เท่านั้น) | default/พฤติกรรม | อ้างอิง |
|---|---|---|---|
| `IConfig` | `MODE` (ถ้าว่างจะ `Setenv("MODE","development")`), `SERVICE_NAME`, `PATH_PREFIX`, `TOPIC_NAME`, `HTTP_CORS`, `JWT_SECRET_KEY` (ไม่มี fallback), `LINE_CLIENT_ID`, `GOOGLE_CLIENT_ID` | test ยืนยันว่า JWT ไม่มี fallback ใน source | config.go:62-65 (MODE), 72, 83, 99, 103, 117, 147, 154; config_security_test.go:8 |
| data environment | `BC_ENV` > `APP_ENV` > `RUN_ENV` > `ENVIRONMENT` > `MODE` → normalize เป็น `dev/uat/pro` | ไม่ตั้ง = dev; `MODE=development` ก็นับเป็น dev | data_environment.go:14-39 |
| `IPersisterMongoConfig` | URI: dev=`MONGODB_DEV_URI`→`MONGODB_URI`, uat=`MONGODB_UAT_URI`, pro=`MONGODB_PRO_URI`→`MONGODB_PRODUCTION_URI`; DB: dev=`MONGODB_DEV_DB`/`MONGODB_DEV_DATABASE`/`MONGO_DB_NAME`/`MONGODB_DB` (default `bcaiclouddb`), uat=`MONGODB_UAT_DB`/`..._DATABASE`, pro=`MONGODB_PRO_DB`/`MONGODB_PRO_DATABASE`/`MONGODB_PRODUCTION_DB`; `MONGODB_DEBUG=true` เปิด command monitor | getter legacy `MONGODB_PROTOCAL/SERVER/PORT/USERNAME/PASSWORD/SSL/TLS_CA_FILE` ยังอยู่แต่ `MongodbURI()` ไม่ใช้ (test config_mongodb_test.go:10) | data_environment.go:41-64; config_mongodb.go:18-20, 22-67 |
| `IPersisterConfig` (PG) | `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB_NAME` (default `postgres`), `POSTGRES_USERNAME`, `POSTGRES_PASSWORD`, `POSTGRES_SSL_MODE` (default `disable`), `POSTGRES_TIMEZONE` (default `UTC`), `POSTGRES_LOGGER_LEVEL` | DSN ประกอบใน persister.go:75-87 | config_postgresql.go:23-66 |
| `IPersisterClickHouseConfig` | `CH_SERVER_ADDRESS`, `CH_DATABASE_NAME`, `CH_USERNAME`, `CH_PASSWORD` | `ServerAddress()` คืน slice 1 ตัวเสมอ (ไม่ split comma) | config_clickhouse.go:16-31 |
| `ICacherConfig` | `REDIS_CACHE_URI`, `REDIS_CACHE_PASSWORD`, `REDIS_CACHE_USERNAME`, `REDIS_CACHE_TLS_ENABLE` | `DB()` คืน 0 ตายตัว; pool settings เป็นค่า default ในโค้ด | config_cacher.go:35-66 |
| `IMQConfig` | `ENABLE_KAFKA` (=`false` → config ว่าง → ข้าม Kafka ทั้งหมด), `KAFKA_SERVER_URL`, `KAFKA_SECURITY_PROTOCOL`, `KAFKA_SSL_CA_FILE`, `KAFKA_SSL_KEY_FILE`, `KAFKA_SSL_CERT_FILE` | — | config_mq.go:21-40 |
| `IHttpConfig` | `SERVICE_PORT` > `HTTP_PORT` (default 8080); `PATH_PREFIX` > `HTTP_PATH_PREFIX`; `HTTP_IGNORE_LOG_URLS` (default `/healthz`); `HTTP_CORS` | — | config_http.go:20-53 |
| `ILoggerConfig` | `LOG_LEVEL` (default info), `LOG_ENCODER`, `MODE` (development → devMode) | devMode เขียน `logs/mainapi_debug.log` และล้างไฟล์ทุกครั้งที่ start | config_logger.go:15-25; backend/internal/logger/logger.go:86-97, 292-308 |
| Elk / OpenSearch | `ELK_ADDRESS`, `ELK_USERNAME`, `ELK_PASSWORD`; `OPEN_SEARCH_ADDRESS`, `OPEN_SEARCH_USERNAME`, `OPEN_SEARCH_PASSWORD` | **default ของ ELK ใน source เป็น host LAN + username + password literal** (ไม่คัดลอกค่า) — แพ็กเกจ DEAD แต่ค่ายังอยู่ใน git | config_elk.go:18-27; config_opensearch.go:18-27 |
| Storage / อื่น ๆ | `STORAGE_DATA_PATH`, `STORAGE_DATA_URI` (local-disk persister DEAD); `PRODUCT_GROUP_SERVICE_PRODUCT_HOST` (`ProductHost()` 0 ผู้เรียก); S3: `S3_ACCOUNT_ID/S3_ENDPOINT/S3_ACCESS_KEY_ID/S3_SECRET_ACCESS_KEY/S3_BUCKET_NAME/S3_REGION` (fallback `R2_*`) | — | config_file_storage.go:9-15; config_product_group_service.go:15-17; persister_file_r2.go:51-55, 68 |

bootstrap → env: `internal/setupconfig/loader.go` map section `mongodb/mongodbdev/mongodbuat/mongodbpro/mongodbproduction/postgresql/clickhouse/service/integrations/storage/kafka` ไปยัง env ด้านบน (loader.go:12-24, 29-61, 83) และตั้ง default `REDIS_CACHE_URI=redis:6379`, `KAFKA_SERVER_URL=kafka:9092` (loader.go:267-270). หลักฐาน runtime (ตรวจซ้ำ 2026-09-07): `docker exec mainapi env` มี key กลุ่ม `BCAI_*`, `BC_ENV`, `DEV_API_MODE`, `FIREBASE_PROJECT_ID`, `GOOGLE_CLIENT_ID`, `GO_ENV`, `KAFKA_SERVER_URL`, `LOG_LEVEL`, `R2_*`/`S3_*`, `SERVICE_PORT`, `STORAGE_ALLOW_PRESIGNED_URL`, `TZ` — ไม่มี `MONGODB_*`/`POSTGRES_*`/`REDIS_*` เลย → Mongo/PG/Redis ไม่ได้มาจาก env ของ container แต่มาจากไฟล์ bootstrap

## 5. internal/utils
| ไฟล์ | API | ผู้ใช้ | หมายเหตุ | อ้างอิง |
|---|---|---|---|---|
| request_param.go | `GetPageable` (172), `GetPageableStep` (112), `GetPageParam` clamp page 1..2e9, limit 1..100000, default limit 20 | LIVE | — | backend/internal/utils/request_param.go:9-18, 25-58, 129-150 |
| requestfilter/ | `GenerateFilters(getParam, []FilterRequest{Param,Field,Type})` type string/int/float64/boolean/date/rangeDate | 140 จุด | LIVE | requestfilter/request_filter.go:17-32, 157-199 |
| importdata/ | generic `FilterDuplicate`, `PreparePayloadData`, `UpdateOnDuplicate` | 319 จุด | LIVE (bulk import ทุก module) | importdata/importdata.go:3-19, 21-28, 50-58 |
| random.go | `NewGUID` (ksuid, 281 จุด — ค่า `guidfixed`), `NewID` (xid), `NewUUID` | LIVE | — | random.go:49-61 |
| password.go | `HashPassword` = argon2id (m=19 MiB, t=2, p=1); `CheckHashPassword` รองรับ hash `$2...` (bcrypt) เดิม | 7 / 1 จุด | LIVE | password.go:16-36, 38-63 |
| business_code.go / holding_code.go | `NormalizeBusinessCode` upper+ตัดช่องว่าง (76 จุด); `NormalizeHoldingCode` regex `^[a-z][a-z0-9]{2,29}$` (7 จุด) | LIVE | — | business_code.go:6-8; holding_code.go:9-19 |
| auth.go | `NormalizeUsername/Phonenumber/Email`; `HasPermissionShop*` | Normalize LIVE; `HasPermissionShop` 0 ผู้เรียก | DEAD บางส่วน | auth.go:14-34, 36-43 |
| checksum/hash.go | `Sum(val)` | 3 จุด | **บั๊ก**: ใช้ `sha1.New().Sum(data)` โดยไม่ `Write` → ผลคือ hex(json + sha1("")) ไม่ใช่ hash ของข้อมูล | checksum/hash.go:9-22 |
| search/search.go | `CreateTextFilter` + mapkha dict `./tdict-std.txt` (relative cwd) | 3 ไฟล์ (internal/member/member_repository.go, internal/repositories/search_repository.go, internal/shop/shopuser_repository.go) | LIVE; SearchRepository ทำซ้ำ logic เดียวกัน | search/search.go:14-33; backend/internal/repositories/search_repository.go:328-339 |
| hash.go, loadkey.go, xreflect.go, mogoutil/ | `FastHash` (1), `LoadKey` (0), `GetReflectTagValue/ReplaceByTemplate` (0), `AggregatePageDecode` (3) | — | loadkey + xreflect DEAD | hash.go:5; loadkey.go:21; xreflect.go:26, 43; mogoutil/aggregate.go:8 |

## 6. internal/models (shape ร่วมของทุก document)
- `Identity{HoldingCode,GuidFixed}` (135), `HoldingCodeentity` (163), `DocIdentity{GuidFixed}` (182), `PartitionIdentity{ParID json:"-"}` (135) — มี gorm tag `primaryKey` ทั้งคู่ (backend/internal/models/identity.go:3-17)
- `ActivityDoc` (json:"-" ทุก field, bson `createdby/createdat/updatedby/updatedat/deletedby/deletedat`) 143 จุด vs `Activity` (json เปิด) (activity.go:5-27); soft delete ทั้งระบบพึ่ง `deletedat` (search_repository.go:28-31)
- `ApiResponse{Success,Message,DocNo,ID,Data,Pagination,Total}` ~1,739 บรรทัดที่อ้าง `models.ApiResponse` (api_response.go:5-13); `NameX{Code,Name}` ~991 บรรทัด (`[]models.NameX` 205) + `JSONB []NameX` สำหรับ PG (name.go:33-36; postgresql.go:9-20)
- DEAD: `Trans/TransItemDetail` (transactionitem.go:5-53, 0 ผู้เรียก), `SearchFilter` (search_filter.go:3, 0), `Datetime` (date_time.go:8, 0), `organization.go` (comment ทั้งไฟล์), `models/other/error.go` (package ชื่อ `adminrest`, 0 importer) — ส่วน `Index` (index.go:3) ยังมีผู้ใช้ 1 จุด internal/member/models/member.go:87 ไม่นับ DEAD

## 7. internal/repositories (generic บน IPersisterMongo/IProducer)
| repo | contract | ผู้ใช้ | อ้างอิง |
|---|---|---|---|
| `CrudRepository[T]` | Create/CreateInBatch, Update = `UpdateOne{holdingcode,guidfixed}`, Delete = SoftDelete, Find* กรอง `deletedat $exists false` | 138 | backend/internal/repositories/crud_repository.go:34-38, 72-129, 141-170 |
| `SearchRepository[T]` | Find/FindStep/FindPage(+Filter, +NoHoldingCode)/FindAggregatePage + ตัดคำไทย | 184 | search_repository.go:26-51, 184, 300 |
| `ActivityRepository[TCU,TDEL]` | sync ตาม `lastUpdatedDate`: FindDeleted* กรอง `deletedat $gte`, FindCreatedOrUpdated* กรอง `createdat`/`updatedat $gte` (Page/Step) | 111 | activity_repository.go:16-22, 41, 63-66 |
| `GuidRepository[T]` | FindInItemGuid(s) | 124 | guid_repository.go:10-31 |
| `KafkaRepository[T]` | Create/Update/Delete(+InBatch) → `SendMessage(topic, mqKey, doc)` ตาม `KafkaConfig` 6 topic | 64 | kafka_repository.go:7-40 |
| `CacheRepository` | Save/Get บน ICacher | 1 | cache_repository.go:13-30 |

## 8. กับดัก / ข้อควรระวัง (ตรวจจากโค้ดจริง)
1. `ServerAddress()` คืน `[]string{addr}` 1 ตัวเสมอ — comma-separated cluster ไม่ทำงาน และ key cache ของ `ClickHousePersister` คือ `strings.Join(addr,"_")` (config_clickhouse.go:16-19; microservice.go:416)
2. `SearchPersister` เขียนผลลง `ms.elkPersisters` แทน `openSearchPersisters` และ `openSearchPersisters` ไม่ถูก init ใน `NewMicroservice` (microservice.go:458, 109-114) — DEAD จึงยังไม่ระเบิด
3. `Persister/MongoPersister/Cacher/Producer` อ่าน map นอก lock แล้วค่อย lock ตอนเขียน (check-then-act race) ต่างจาก `PersisterTenant` ที่ lock ทั้งก้อน (microservice.go:392-401, 403-412, 464-484 vs 365-367)
4. CORS ซ้อน 2 ชั้น: `NewMicroservice` ใส่ `AllowOrigins:*` แล้ว `HttpUseCors` ใส่อีกชุดจาก `HTTP_CORS` (microservice.go:89-92, 536-543; main.go:344)
5. ทุก route ถูก register 2 ครั้ง (legacy + `/v1` alias) ยกเว้น path ที่ขึ้นต้น `/v1/` หรือ `/api/v1/` (microservice_http.go:27-37) — นับ route ใน `ms.echo.Routes()` จะเป็น 2 เท่า
6. `consumeSingle` เจอ error ที่ไม่ใช่ timeout → `ms.Stop()` ปิดทั้ง process (microservice_consumer.go:39-45); consumer librdkafka ใช้ `enable.auto.commit=true` ทุก 500 ms (microservice.go:571-576) = at-least-once ที่อาจหลุด message ถ้า crash ระหว่าง handler; barcode topic เท่านั้นที่ commit หลัง handler (barcode_projection_consumer.go:27)
7. `Producer.SendMessage` เป็น sync และ `p.logger.Error("... %s: %v", ...)` ใช้ `Error` ไม่ใช่ `Errorf` → ข้อความ log ไม่ถูก format (producer.go:103, 114; บรรทัด 118 ใช้ `Debug` กับ format verb แบบเดียวกัน)
8. `getCollectionName` ต้องการ `CollectionName()` เท่านั้น ไม่ผ่าน `NormalizeMongoCollectionName` — ชื่อ collection จึงเป็นอะไรก็ได้ที่ model คืน (persister_mongo.go:159-167)
9. `mongo.NewClient` + `Connect` เป็น API เก่าของ driver v1 และ `Ping(context.TODO())` (persister_mongo.go:122-140); `pst.db != nil` ตรวจนอก mutex (101-107)
10. `validator.Validate` ลงทะเบียน translation ใหม่ทุก request และคืน error ตัวแรกเป็นภาษาอังกฤษ (validator.go:35-56) — ขัดกฎ Thai-first ของ UX ถ้าโยนตรงถึงหน้าจอ
11. `IPersisterImage.Upload(fh)` (persister_image.go:7-10) ไม่ตรงกับ `PersisterImage.Upload(fh, name, ext)` (:22) → struct ไม่ implement interface; main.go ใช้ concrete type จึง compile ผ่าน. เช่นเดียวกับ `ICRUDRepository.CountByKey` ไม่มี ctx แต่ impl มี (crud_repository.go:14 vs :50) และ `ICacheRepository.Save(holdingCode, moduleName)` vs impl `Save(key, value, expire)` (cache_repository.go:8-11 vs :23)
12. `go vet` (รันใน container golang:1.26 + librdkafka-dev; fact-checker รันซ้ำ 2026-09-07 บน `./pkg/... ./internal/repositories/ ./internal/utils/... ./internal/models/... ./internal/config/`) พบ 18 รายการ: `internal/repositories/crud_repository.go:153-154` (bson.E unkeyed 3 จุด) และ `pkg/microservice/persister_mongo_test.go` 15 จุด (unkeyed 14 ที่ :75, 87, 161-163, 167-169, 222-224, 228-230 + unreachable code :295 เพราะ `return errors.New("test")` ที่ :287) — docs/handoff/HANDOFF-RISKS-2026-09-05.md ไม่ได้กล่าวถึงผล vet นี้
13. Test ในแพ็กเกจต้องการ infra จริง: `persister_test.go` (PG), `persister_clickhouse_test.go:60-144` (ClickHouse ที่ถอดไปแล้ว), `cacher_test.go:15` (Redis) — ยังไม่ได้รันในรอบนี้
14. หลักฐานเชิงบวก: `NewProducerWithTimeout` (producer.go:48-54) มี test `producer_timeout_test.go:15`; barcode consumer มี test `barcode_projection_consumer_test.go:5, 20`; R2/MinIO persister มี test `persister_file_r2_test.go:33-59`

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ
- `HttpPreRemoveTrailingSlash` เรียกที่ main.go:345, cmd/app/main.go:179, cmd/ws/main.go:33, cmd/member/main.go:30; `HttpUseJaeger` เรียกเฉพาะ cmd/ws/main.go:35 (ไม่ใช่ `mainapi`) — ยังไม่ตรวจว่า `jaegertracing.New(ms.echo, nil)` (microservice.go:532) อ่าน endpoint จาก env ใดของ lib (ไม่มี env key ในตาราง §4)
- ยังไม่ตรวจว่า `NewMQ` 36 จุดอยู่ใน mode ไหนบ้าง (`escapeName` ตรวจแล้ว → ใช้เฉพาะใน AsyncPOST/PUT ที่ DEAD, ดูตาราง §2)
- ยังไม่รัน unit test ของ pkg/microservice ที่ต้องใช้ PG/Redis/ClickHouse (ข้อ 8.13) — รอ ClickHouse ตัดสินใจก่อน
- ยังไม่ตรวจว่า `internal/goapi` ใช้ persister ของตัวเอง (myclickhouse/mypg) ทั้งหมดหรือแบ่งใช้ `IPersisterMongo` ผ่านไฟล์เดียวที่ import (goapi import pkg/microservice แค่ 1 ไฟล์ non-test คือ backend/internal/goapi/bootstrap.go — ยังไม่เปิดดูว่าใช้ persister ใดจากแพ็กเกจนี้)
- ยังไม่ตรวจว่า bootstrap.json ที่ mount ใน container มี section `clickhouse` อยู่ไหม (ถ้ามี `CH_SERVER_ADDRESS` จะถูก compose ที่ loader.go:287-297 แม้ container ถูกถอด)
- คำถามถึงลุงจืด: (1) จะลบแพ็กเกจ DEAD ชุดนี้เลยไหม — `pkg/bcrypt`, `pkg/calendar`, `pkg/serializer`, `pkg/tokenize`, `JwtService`, `AsyncPOST/PUT`, Elk/OpenSearch persister, local `PersisterFile`, `utils/loadkey.go`, `utils/xreflect.go`, `models/transactionitem.go`, `models/other/error.go` (2) `checksum.Sum` ที่ hash ผิด (ข้อ 8 แถวสุดท้าย §5) มี 3 ผู้เรียก — ต้องแก้หรือ callers ไม่พึ่งค่านี้จริง? (3) default credential ของ ELK ใน config_elk.go:18-27 ควรลบออกจาก source ก่อน go-live ไหม (4) topic `TEST-CONNECT` ที่ถูกสร้างทุก boot ยอมรับได้หรือควรเปลี่ยนเป็น metadata check
