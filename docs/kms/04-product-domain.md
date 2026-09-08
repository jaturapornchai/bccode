# โดเมนสินค้า (Product / Barcode / Unit / BOM) — Mongo SoT, outbox, projection PG, import, listing v2, จอ frontend
> ตรวจล่าสุด: 2026-09-07 (commit d93a210d) — ผู้เขียน: AI reader; ทุกข้อเท็จจริงอ้าง path:line · ตรวจซ้ำโดย fact-checker

## 1. ภาพรวม 1 นาที

- **MongoDB `appdb` คือ source of truth** ของสินค้า: collection `products` (backend/internal/product/product/models/product.go:11), `productbarcodes` (backend/internal/product/productbarcode/models/product_barcode.go:13), `units` (backend/internal/product/unit/models/unit.go:9), `productgroups` (backend/internal/product/productgroup/models/productgroup.go:9), `productcategories` (backend/internal/product/productcategory/models/productcategory.go:9), `productbarcodeboms` (backend/internal/product/bom/models/bom.go:10), `productbarcodespricehistory` (backend/internal/product/productbarcode/models/product_price_history.go:10), `outboxevents` (backend/internal/product/product/outbox/outbox.go:62)
- **PostgreSQL ต่อ holding** (ชื่อ DB = holding code: `mypg.PgSqlFastConnect(holding)` → `PgSqlFastConnectV2` → `mydb.GetGlobalConnectionFromPool` → `pool.GetConnection(holdingCode)` backend/internal/goapi/mypg/fast_utils.go:39-42; mypg/compatibility.go:11-13; mydb/manager_pool.go:210-212, DSN `dbname=%s` backend/internal/goapi/mydb/connection_pool_manager.go:72-81) เป็นแค่ **projection** ตาราง `product` / `productbarcode` (DDL backend/internal/goapi/process/build/create-database.go:42,141) + `product_projection_fences` (backend/internal/goapi/handlers/kafka/projection_reconcile.go:82-88) เติมโดย goapi Kafka consumer
- **เส้นทางเขียนหลัก**: HTTP `mainapi` → validate → Mongo transaction + แถว `outboxevents` → worker ส่ง Kafka → goapi consumer reconcile PG (อ่าน Mongo PRIMARY ซ้ำ ไม่เชื่อ payload) — รายละเอียด §4
- **ClickHouse ถูกถอด container 2026-09-06** แต่โค้ดยังสร้าง persister ตอน boot ใน 3 module ของโดเมนนี้ (§7) — mainapi ตัวที่รันอยู่ start ตั้งแต่ 2026-09-05T23:27Z (`docker inspect mainapi`) ก่อนถอด → **ยังไม่ตรวจว่า restart แล้วขึ้นไหม**

## 2. Inventory module backend (`backend/internal/product/**` + ที่เกี่ยวข้อง)

ทุก module ลงทะเบียนใน `httpServices` ของ backend/main.go (mode `""`/`"2"` backend/main.go:256) — บรรทัดอ้างอิงด้านล่าง

| module | หน้าที่ | entry / route | สถานะ | อ้างอิง |
|---|---|---|---|---|
| product | สินค้าหลัก (รหัส/ชื่อ/หน่วยฐาน/ชั้นลงขาย `listing`) | `GET/POST /product`, `POST /product/resync`, `GET/PUT/DELETE /product/:guid` | LIVE (outbox) | backend/internal/product/product/product_http.go:70-76; main.go:371 |
| productbarcode | บาร์โค้ด = ตัวเลือกที่ขายได้ 1 doc, ราคา, รูป/วิดีโอ, BOM view, price history | `/product/barcode*` 24 route (bulk, import, list, by-code, master, xsort, branch, business-type, units, groups, export, bom/:barcode, price-history, import-refbarcode) | LIVE (outbox `CommitBarcode` สำหรับ Create/Update/Delete บริษัท — productbarcode_http_service.go:387,602,1021) ยกเว้น `/product/barcode2` = DEAD-CH | backend/internal/product/productbarcode/productbarcode_http.go:98-126; main.go:373 |
| unit | หน่วยนับ (holding-scope: model มีเฉพาะ `HoldingCodeentity` unit/models/unit.go:29) | `/unit*` 12 route | LIVE (legacy DB-before-MQ goroutine) | backend/internal/product/unit/unit_http.go:49-60; main.go:366 |
| productgroup | กลุ่มสินค้า | `/product/group*` 11 route | LIVE (legacy MQ) | backend/internal/product/productgroup/productgroup_http.go:47-58; main.go:376 |
| productcategory | หมวดสินค้า | `/product/category*` 9 route | LIVE (ไม่มี MQ config ใน git ls-files) | backend/internal/product/productcategory/productcategory_http.go:45-54; main.go:372 |
| bom | สูตร/ชุดสินค้า (Mongo + PG legacy) | `/product/bom*` 6 route | LIVE (legacy MQ `when-bom-*`) | backend/internal/product/bom/bom_http.go:52-57; main.go:544 |
| option / optionpattern / color | ตัวเลือกสินค้า (choices), แพทเทิร์น, สี | `/option*`, `/optionpattern*`, `/color*` | LIVE | backend/internal/product/option/inventoryoption_http.go:36-41; optionpattern/optionpattern_http.go:39-45; color/color_http.go:39-46; main.go:365-368 |
| promotion / ordertype | โปรโมชัน, ประเภทออเดอร์ | `/product/promotion*`, `/product/order-type*` | LIVE | backend/internal/product/promotion/promotion_http.go:43-52; ordertype/ordertype_http.go:46-55; main.go:501-503 |
| eorder | หน้าร้าน e-order อ่านสินค้า/หมวดแบบ public | `GET /e-order/category`, `GET/POST /e-order/product-barcode`, shop-info ฯลฯ (อยู่ใน `publicPath`) | LIVE (สร้าง CH persister ตอน boot) | backend/internal/product/eorder/eorder_http.go:160-172,64; main.go:282-284,498 |
| productimport | นำเข้าสินค้าจาก Excel ผ่าน staging | `/productimport*` 12 route | DEAD-CH (staging = ClickHouse) | backend/internal/productimport/productimport_http.go:129-143; main.go:521 |
| goapi product read | ค้นหาสินค้า/บาร์โค้ดฝั่ง goapi | `POST /goapi/api/product/search` (PG+cache), `/api/product/barcode/list` (Mongo), `/api/product/search/unified` (PG) | LIVE; UI เรียกเฉพาะ `barcode/list` (grep `product/search` ใน frontend/src = 0) | backend/internal/goapi/bootstrap.go:405-414; handlers/product_cache.go:264,353-365; handlers/product_search.go:99,161; handlers/barcode_list.go:143 |
| goapi listing v2 | ชั้นลงขาย (read-only ชั้นบัญชี, `$set` + `__v`) | `POST /goapi/product/v2/item/{get,update-listing,readiness}`, `/tier/{init,update}` | LIVE เฟส 1; UI ยังไม่เรียก | backend/internal/goapi/bootstrap.go:417-421; handlers/product_v2_update.go:130 |

Model สำคัญ: `ProductDoc.Listing *ProductListing` (product.go:102), package weight/dimension (product.go:94-97), `ProductBarcode.Listing *ProductBarcodeListing` (product_barcode.go:83,227), รูป `imageuri`/`imageurithumb`/`images[]`/`videos[]` (product_barcode.go:61-64,82), รูปย่อชั้นลงขาย `urithumb` (product_barcode.go:223), `MarketplaceProducts *[]MarketplaceProductMap` (product_barcode.go:79,467), ตัวเลือก `ProductOption.Choices` (backend/internal/product/productbarcode/models/product_option.go:11). Unique index Mongo: partial unique `uniq_products_active_holdingcode_code` (deletedat=null, code>"") และ `..._businesscode_code` (backend/internal/product/product/repositories/product_mongo_repository.go:79-100). ลบสินค้า = soft delete (`pst.SoftDelete` product_mongo_repository.go:159-160)

## 3. Kafka topics ของโดเมนนี้ — ใครส่ง ใครรับ

ชื่อ topic ทั้งหมดอยู่ใน `*_messagequeue_config.go` (created/updated/deleted + bulk-* อย่างละ 6):

| prefix | ประกาศที่ | ผู้ส่ง | ผู้รับ (consumer group) |
|---|---|---|---|
| `when-product-{created,updated,deleted}` | backend/internal/product/product/config/product_messagequeue_config.go:4-9 | outbox worker ของ product (product_http_service.go:310,378,423,469-478) | goapi `biapi-product-consumer-<ver>` (handlers/kafka.go:158-179) → `CallProductConsumer` (handlers/kafka_bridge.go:284) → `OnConsumeMessageProductCreateOrUpdate` (handlers/kafka/product.go:37) → `consumeProductSignal` (projection_reconcile.go:279) |
| `when-product-bulk-*` | เดียวกัน :7-9 | ยังไม่ตรวจว่ามีใครส่ง (ไม่มีใน `kafka-topics --list` ของ local) | ไม่มี consumer ใน kafka.go |
| `when-product-barcode-{created,updated,deleted}` | productbarcode/config/productbarcode_messagequeue_config.go:4-6 | outbox worker ของ barcode (productbarcode_outbox.go:14-18) + legacy `mqRepo` ใน SaveInBatch (productbarcode_http_service.go:2129-2136) | goapi `biapi-inventory-consumer-<ver>` (kafka.go:118-147); legacy mainapi group `<CONSUMER_GROUP_ID>-projection` (env `CONSUMER_GROUP_ID` default `consumer-productbarcode-group-01` productbarcode_consumer.go:45; pkg/microservice/barcode_projection_consumer.go:19; productbarcode_consumer.go:94-97) |
| `when-product-barcode-bulk-{created,updated,deleted}` | เดียวกัน :7-9 | product update/resync ส่ง `bulk-updated` พร้อม barcode ที่ผูก (product_http_service.go:478); SaveInBatch ส่ง bulk create/update ผ่าน generic `KafkaRepository.CreateInBatch/UpdateInBatch` (backend/internal/repositories/kafka_repository.go:60-71) | goapi `biapi-inventory-bulk-consumer-<ver>` (kafka.go:190-219); legacy รับเฉพาะ bulk-created (productbarcode_consumer.go:95) |
| `when-product-unit-*` | unit/config/unit_messagequeue_config.go:4-9 | unit service goroutine หลังเขียน Mongo (unit_http_service.go:218-221) และ outbox barcode เมื่อสร้าง unit อัตโนมัติ (productbarcode_outbox.go:17-18) | legacy barcode consumer รับเฉพาะ `unit-updated` → อัปเดต snapshot หน่วยใน `productbarcodes` (productbarcode_consumer.go:101; productbarcode_consumer_service.go:96-98) |
| `when-product-group-*`, `when-product-type-*`, `when-product-order-type-*` | productgroup/config:4-9; producttype/config:4-9; ordertype/config:4-9 | productgroup service (productgroup_http_service.go:153,218,295) | legacy barcode consumer รับเฉพาะ `*-updated` (productbarcode_consumer.go:99-102) |
| `when-bom-*` | bom/config/bom_message_queue_config.go:4-9 | bom service `repoMq` (bom_http_service.go:330,354,681) | `bom.InitBOMConsumer` group `PRODUCT_BOM_CONSUMER_GROUP` (bom_consumer.go:45-62) → PG `productbarcodeboms` (bom/models/bom_postgres.go:64) |

หมายเหตุสำคัญ
- **goapi consumer ทำงานเฉพาะเมื่อ `ENABLE_KAFKA=true`** (backend/internal/goapi/bootstrap.go:157-160) และ group ต่อท้าย `KAFKA_CONSUMER_GROUP_VERSION` (default `v1`, backend/internal/goapi/config/config.go:119; handlers/kafka.go:19-22)
- **legacy consumer ของ mainapi ทำงานใน mode `""` และ `"1"`** (backend/main.go:653) — ไม่ใช่เฉพาะ `"1"`; prod แยก service `worker` `DEV_API_MODE: "1"` (deploy/account/compose.yml:275) และ `migrate` `"3"` (compose.yml:212)
- barcode topic ถูกส่งเข้า loop ack-after-success (`projection.Consume`) ทั้ง 2 ฝั่ง: goapi ผ่าน `mykafkaconsumer.isProductProjectionTopic` → `projection.Consume` (backend/internal/goapi/mykafkaconsumer/product_projection.go:14-19,36-37), legacy ผ่าน `ms.Consume` selector `projection.IsBarcodeTopic` (backend/pkg/microservice/microservice_consumer.go:55-58; backend/internal/product/projection/consumer.go:23-52)
- Topic ที่มีจริงบน local kafka (`docker exec kafka kafka-topics --list`, 2026-09-07): product ×3 (created/updated/deleted — ไม่มี bulk), product-barcode ×6, product-unit ×3, product-group ×3, bom ×3

## 4. Transactional outbox → projection PG (เส้นทางที่ออกแบบใหม่ 2026-09-05)

1. **Commit**: `outbox.Store.Commit/CommitBarcode` เปิด Mongo transaction, อ่าน version ล่าสุด, mutate, insert `outboxevents` (`aggregatetype=productprojection`) (backend/internal/product/product/outbox/outbox.go:21,100-114; transaction+อ่าน version ล่าสุด :122-125, `InsertOne` version+1 :146-148); retry `TransientTransactionError` สูงสุด 5 ครั้ง (outbox.go:159-175). Aggregate key `product:`/`barcode:` + SHA-256 (outbox.go:94-96,106-107; backend/internal/product/projection/metadata.go:19-23)
2. **Worker**: `Run` ticker 1 วินาที, `Dispatch` ดึง head สูงสุด 100 aggregate ต่อรอบ, lease 60 วิ ต่ออายุทุก 20 วิ (timeout 5 วิ), backoff 1–60 วิ (outbox.go:22,184-198,238-253,306,323-324). **มี worker 2 ตัวบน collection เดียว** (product_http.go:52-58 และ productbarcode_http.go:78-84 — แต่ละตัวมี producer timeout 30 วิ ของตัวเอง)
3. **Consumer goapi**: parse → `projection.ErrRejected` ถ้า JSON/identity ใช้ไม่ได้ (ack แล้วข้าม — backend/internal/product/projection/consumer.go:34-38; identity check handlers/kafka/product.go:25-31) → `projectionRuntime(holding)` เปิด PG ต่อ holding + สร้าง DB ถ้าไม่มี (`build.DatabaseChecker`) (projection_reconcile.go:270-278,279-294) → advisory lock company shared + row exclusive, เกิน 256 แถวใช้ company exclusive (projection_reconcile.go:112-133) → อ่าน Mongo PRIMARY (`products`/`productbarcodes`, projection_reconcile.go:31-56) → upsert/delete PG + บันทึก fence `(holding_code,businesscode,itemcode,aggregateuid,version,deleted)` (projection_reconcile.go:158,190)
4. **Consumer legacy (mainapi)**: `UpSert/Delete` ไม่เชื่อ payload เช่นกัน → `ReconcileInCompany` ใต้ lock เดียวกัน, omit `balanceqty/balanceamount/averagecost` (productbarcode_consumer_service.go:116-128,188-205; repositories/productbarcode_pg_repository.go:166-192) — **เขียน PG จาก `cfg.PersisterConfig()` = DB เดียวตาม `POSTGRES_DB_NAME`** (backend/internal/config/config_postgresql.go:31-32) ไม่ใช่ PG ต่อ holding; `appdb` บน local **ไม่มีตาราง productbarcode** (psql information_schema 2026-09-07) → ยังไม่ตรวจว่า legacy consumer เขียนลง DB ไหนจริง
5. **Resync**: `POST /product/resync` = rebuild PG แบบ synchronous (`build.ProcessProductRebuildCompany` ล็อก company exclusive + อ่าน Mongo PRIMARY, backend/internal/goapi/process/build/build-product.go:50,78,84) แล้ว queue snapshot สินค้าที่ active ผ่าน outbox; ตอบ `{rebuilt,queued,published:0}` (product_http.go:337-358; product_http_service.go:434-465). frontend proxy: `POST /api/product/resync` (frontend/src/app/api/product/[[...productPath]]/route.ts:51)
6. cache: goapi `productCache` ถูก clear ทุกครั้งที่ consumer ประมวลผล product/barcode (backend/internal/goapi/handlers/kafka_bridge.go:245-251 barcode, :284-291 product)

หลักฐาน runtime local (อ่านอย่างเดียว 2026-09-07): `outboxevents` productprojection = PUBLISHED 7 / PENDING 0; Mongo `demo` active products 45 = PG `demo.product` 45 ✓; Mongo `demo` active productbarcodes **45 ≠ PG `demo.productbarcode` 54** (ยังไม่วินิจฉัย — docs/handoff/HANDOFF-2026-09-06.md:63 ระบุแถวค้าง 9 แถวใน C03); `product_projection_fences` 2 แถว; PG `bc001` product 20/productbarcode 20, `test` 49/45

ช่องโหว่ที่ README ยอมรับเอง (backend/internal/product/product/outbox/README.md): at-least-once; unit เดี่ยว/legacy batch ยัง DB-before-MQ; ไม่มี DLQ/alert; ClickHouse ไม่ได้ทดสอบ. Test/CI: `.github/workflows/ci.yml:82-122` (outbox + barcode batch integration), `:138` (`backend-projection-kafka-integration` ใช้ backend/.ci/projection.compose.yml services mongo/mongo-init/postgres/kafka/tests :3-52) — Actions ถูก billing-lock ตามที่ lead แจ้ง (ยังไม่ตรวจเอง)

## 5. Price history, รูปภาพ, BOM

- **Price history**: บันทึกใน `productbarcodespricehistory` เมื่อ Create/Update บาร์โค้ดผ่าน `priceHistorySvc.RecordPriceChange` (productbarcode_http_service.go:453,670; services/product_price_history_service.go:39-60) อ่านผ่าน `GET /product/barcode/price-history[/ :barcode]` (productbarcode_http.go:124-125) → frontend `/api/product-price-history/**` และ `/api/product-barcode/price-history/[barcode]` (frontend/src/app/api/product-price-history/[[...historyPath]]/route.ts:35; api/product-barcode/price-history/[barcode]/route.ts:28)
- **รูป/วิดีโอ**: frontend `/api/product-barcode/image|video` → goapi `POST /goapi/image/upload` / `/video/upload` (frontend/src/app/api/product-barcode/image/route.ts:7; video/route.ts:9; backend/internal/goapi/bootstrap.go:524-525) → S3/MinIO ผ่าน env `S3_*` (fallback `R2_*`) และสร้าง WebP thumbnail เป็น object แยก (backend/internal/goapi/handlers/image_r2.go:75-79,356-392) — Mongo เก็บ URI เท่านั้น (product_barcode.go:61-82)
- **BOM**: Mongo `productbarcodeboms` + view ผ่าน `GET /product/barcode/bom/:barcode` (productbarcode_http.go:123); consumer legacy เขียน PG `productbarcodeboms` (bom_consumer.go:23; bom_migration.go:11-13) จาก `cfg.PersisterConfig()` เช่นกัน

## 6. Frontend (Next.js) — จอและ proxy

กติกา: browser เรียก `/api/**` ของ Next เท่านั้น ไม่ยิง backend ตรง (frontend/src/lib/product-barcode/api.ts:5)

| จอ | route ใน main menu | ไฟล์ | เรียก API |
|---|---|---|---|
| สินค้า (tab หลายแท็บ: basic/units/classification/stock/bom/media/marketplace/…) | `/product`, `/productextension` | frontend/src/app/menu/product-screen.tsx (3109 บรรทัด) + `tab-product-*.tsx` | `/api/product`, `/api/product/:id`, `/api/product-barcode/…` (product-screen.tsx:543,726,830,1062) |
| ชุดสินค้า | `/productset` | menu/product-set-screen.tsx | `/api/product`, `/api/product-barcode/list` (product-set-screen.tsx:323,375,590,669) |
| บาร์โค้ด | `/productbarcode` | menu/product-barcode-screen.tsx | `/api/product-barcode/…` (product-barcode-screen.tsx:409) ผ่าน lib api.ts:117-300 |
| ชั้นวาง/พิมพ์ป้าย | `/productbarcodeshelf` และหน้า `/product_barcode_shelf` | menu/product-barcode-shelf-screen.tsx; app/product_barcode_shelf/page.tsx:3-11 | `POST /api/product-barcode/list` (shelf-screen.tsx:119) |
| ประวัติราคา | `/pricehistory` และหน้า `/price_history` | menu/product-price-history-screen.tsx; app/price_history/page.tsx:3-11 | `/api/product-barcode/list`, `/api/product-price-history` (:144,220) |

การเลือกจอ: frontend/src/app/menu/main-menu-screen.tsx:2579-2614. Proxy: `/api/product/**` → mainapi `/product*` (api/product/[[...productPath]]/route.ts:33-114); `/api/product-barcode/[[...barcodePath]]` → `/product/barcode*` (route.ts:22-79); `/api/product-barcode/list` → **goapi** `/api/product/barcode/list` ผ่าน `serverGoApiBase()` (list/route.ts:39; frontend/src/lib/backend-url.ts:125) ซึ่งอ่าน Mongo `productbarcodes` ตรง (backend/internal/goapi/handlers/barcode_list.go:143); `/api/product-barcode/master/[master]` whitelist ไป `/product/group`, `/product/type`, `/product/order-type`, `/product-section/business-type`, `/product` (master/[master]/route.ts:19-38); BOM view proxy (bom/[barcode]/route.ts:40). แท็บ marketplace แก้ฟิลด์ `marketplaceproducts` ของบาร์โค้ด (tab-product-marketplace.tsx:163-343) — **ไม่มีจอไหนเรียก `/goapi/product/v2/*`** (grep `product/v2` ใน frontend/src = 0 ไฟล์)

## 7. ClickHouse ในโดเมนนี้ (container ถูกถอด 2026-09-06)

| จุด | พฤติกรรม | อ้างอิง |
|---|---|---|
| `NewProductBarcodeHttp` | สร้าง `ms.ClickHousePersister` ตอน boot; ใช้จริงเฉพาะ `GET /product/barcode2` → `chRepo.Search` | productbarcode_http.go:50; productbarcode_http_service.go:1887; repositories/productbarcode_clickhouse_repository.go:33 |
| `NewEOrderHttp` | สร้าง CH persister ตอน boot; ส่ง `clickHouseRepo` เข้า `NewProductBarcodeHttpService` แต่ handler ของ e-order เรียกเฉพาะ `SearchProductBarcode`/`GetProductBarcodeByBarcodes` (Mongo) — `chRepo` ถูกใช้ที่เดียวคือ `SearchProductBarcode2` ซึ่ง e-order ไม่เรียก | eorder_http.go:64,76,90,265,327,366; productbarcode_http_service.go:1883-1887 |
| `productimport` | staging ทั้งหมดอยู่ตาราง CH `productbarcodeimport` + task status CH; apply เขียน Mongo ผ่าน `CreateProductBarcodeInCompany` (outbox) / `UpdateByIDInCompany` | productimport_http.go:66,94-95; services/productimport_service.go:86,158,2807,3103; repositories/productimport_clickhouse_repository.go:57 |
| legacy barcode consumer | สร้าง CH persister เฉพาะเมื่อ `ServerAddress()` ไม่ว่าง | productbarcode_consumer.go:50-58 |
| persister | `NewPersisterClickHouse` เรียก `getClient()` → `clickhouse.Open` แล้ว `panic` ถ้า error; ไฟล์นี้ไม่มีการเรียก `Ping` (grep 2026-09-07) | backend/pkg/microservice/persister_clickhouse.go:33-45,52 |

**ยังไม่ตรวจ**: `clickhouse.Open` (clickhouse-go v2) dial ตอน Open หรือรอใช้งาน — ถ้า dial ทันที mainapi จะ panic ตอน boot เมื่อไม่มี container; log ที่มีคือ boot เก่าก่อนถอด (`docker logs mainapi` 2026-09-05 23:27 แสดง `CH_SERVER_ADDRESS = clickhouse:9000`)

## 8. Listing API v2 (`docs/kms/architecture/product-listing-api-v2*.md`)

- contract วันที่ 2026-09-03 ตั้งใจให้ชั้นบัญชี read-only, เขียน `$set` เฉพาะ path + `__v` (product-listing-api-v2.md:1-20); handoff ระบุเฟส 1 = model + 5 endpoint (product-listing-api-v2-handoff.md:1-25)
- implement แล้ว: 5 route (bootstrap.go:417-421), อ่าน/เขียน Mongo `products`/`productbarcodes` (handlers/product_v2_types.go:17-18; product_v2_get.go:74,82; product_v2_update.go:92,130) — **ไม่ผ่าน outbox** จึงไม่ยิง Kafka/PG projection; comment ในโค้ดยอมรับว่า save ฝั่งบัญชีไม่เพิ่ม `__v` ทำให้จับ conflict ไม่ครบ (product_v2_update.go:7)
- ชื่อ marketplace ห้ามปรากฏในโค้ดตาม memory `marketplace-api-neutral-naming` แต่ model เดิมยังมี `Marketplace*` struct (product_barcode.go:79,239-283,467) — ยังไม่ตรวจว่าเป็นข้อยกเว้นที่ลุงจืดยอมรับ

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ

1. legacy consumer (`productbarcode`, `bom`) เขียน PG จาก `POSTGRES_DB_NAME` ซึ่ง `appdb` ไม่มีตาราง `productbarcode` — เขียนลง DB ไหน หรือไม่เคยสำเร็จ? (ต้องดู log consumer หรือ env key ใน container — ยังไม่เปิด)
2. mainapi restart โดยไม่มี ClickHouse จะ panic ที่ `NewPersisterClickHouse` หรือไม่ (ขึ้นกับ dial semantics ของ clickhouse-go v2) — ต้องทดลองบน local เท่านั้น (R1 ถ้าแตะ prod)
3. PG `demo.productbarcode` 54 แถว vs Mongo active 45 — ยังไม่เทียบรายแถว; `/product/resync` เคลียร์ได้ตาม HANDOFF แต่ยังไม่รัน
4. `when-product-bulk-*` และ `when-product-type-*` ไม่มีผู้ส่งที่หาเจอในโดเมนนี้ (ยังไม่ grep ทั้ง repo) — ควรตัดหรือคง?
5. outbox worker 2 ตัว (product/barcode) แข่ง claim collection เดียว — README บอกว่า lease กันซ้ำแล้ว แต่ยังไม่วัด throughput/lock contention จริง
6. `productcategory` ไม่มี MQ config ใน git ls-files → ยืนยันว่า category ไม่ต้อง project ไป PG จริงไหม
7. ไม่มี Playwright spec สำหรับสินค้าใน `tests/` (git ls-files tests 2026-09-07: employee/currency/login/menu/user CRUD ใน `tests/uat-crud.spec.ts` — ไม่มีไฟล์ที่ทดสอบ `/product*`) — UAT CRUD+Mongo ของสินค้าอาศัย unit/integration test ฝั่ง Go เท่านั้น
8. คำถามถึงลุงจืด: (ก) ถอด `/product/barcode2`, `/productimport/*`, CH persister ใน eorder/barcode พร้อมกันหรือรอ redesign import? (ข) จอ marketplace ยังเก็บ `marketplaceproducts` ในบาร์โค้ด ในขณะที่ v2 ออกแบบ `channel_listing` แยก — จะ migrate ทางไหน? (ค) PG projection ของสินค้าต้องคงไว้ไหมเมื่อ UI ไม่อ่าน PG (docs/handoff/HANDOFF-2026-09-06.md:75)
