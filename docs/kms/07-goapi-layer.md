# ชั้น goapi — HTTP routes, connection managers และ stock engine (PostgreSQL ล้วน)

> ตรวจล่าสุด: 2026-09-25 — ผู้เขียน: AI reader; ทุกข้อเท็จจริงอ้าง path:line ของโค้ดปัจจุบัน
> MongoDB/Kafka/Redis/ClickHouse ถูกถอดออก 2026-09-23 — ดู [ADR](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md); ห้ามเพิ่มกลับ

## 1. ภาพรวมและจุดเกาะกับ mainapi
- goapi ไม่ใช่ binary แยก: `backend/main.go:168-179` เรียก `goapi.New()` → `Init()` → mount ที่ `ms.Echo().Group("/goapi")` → `RegisterMiddleware` → `RegisterRoutes(goapiGroup, "/goapi", ms.Cacher(), controlDB)` และ `defer Shutdown()` — mount เสมอ ยกเว้น `Init()` คืน error ซึ่งแค่ log แล้วข้าม (`backend/main.go:169-170`)
- `/goapi/*` และ `/api/language/*` อยู่ใน bypass list ของ auth กลาง mainapi (`backend/main.go:133-134`) เพราะ goapi มี auth middleware ของตัวเอง (§3); alias สาธารณะ `GET /api/language/:lang` ลงทะเบียนตรงที่ root echo ด้วย (`backend/main.go:178`)
- Version string `1.1.1121` (`backend/internal/goapi/bootstrap.go:34`) คืนที่ `GET /goapi/`, `/goapi/version`, `/goapi/api/health` (`bootstrap.go:311-329`)

## 2. ลำดับ `Init()` (`bootstrap.go:45-127`)
| ขั้น | ทำอะไร | อ้างอิง |
|---|---|---|
| 1 | `setupconfig.LoadBootstrapConfig()` | `bootstrap.go:49` |
| 2 | `handlers.InitR2Client()` (S3/MinIO; fail = Warn) | `bootstrap.go:52-54` |
| 3 | `mydb.InitManagerPool` จาก env `POSTGRES_*`; struct `mydb.DatabaseManagerConfig` มี 5 ฟิลด์ PostgreSQL | `bootstrap.go:58-66`, `mydb/manager_pool.go:19-25` |
| 4 | ขอ manager `"postgres"` → global `*sql.DB` + provider closure (`myglobal.SetGlobalDatabaseConnection/Provider`) → `mypostgres.InitQueueSchema` (ตาราง `queues`/`deadletterqueue` — ไม่มีผู้ใช้ ดู §6) | `bootstrap.go:70-98` |
| 5 | goroutine: `StartConnectionHealthChecker` (ทุก 5 นาที) + `CleanupPreparedStatements` ทุก 30 นาที | `bootstrap.go:103-110`, `handlers/health.go:14-32` |
| 6 | ตั้ง callback `stockengine.AfterRecalculate` (อัปเดตยอดคงเหลือสินค้า) แล้ว `stockengine.StartWorkers(...)` ถ้า `WorkerEnabledFromEnv()` (env `BCAI_STOCK_WORKER` ไม่ใช่ `0`/`false`) | `bootstrap.go:116-123`, `process/stockengine/manager.go:71-74` |
- `Shutdown()` ปิด `mydb.CloseAllManagers`, `myPg.CleanupPreparedStatements`, `myPg.CloseAllPools` (`bootstrap.go:407-413`)

## 3. Middleware และ auth ของ goapi
- `RegisterMiddleware`: Recover (panic → JSON 500), Secure headers (CSP `default-src 'self'`), tiered rate limiter, timeout 30s (`bootstrap.go:130-167`)
- Rate limit ต่อ IP จับด้วย suffix ของ path — ค่า `Rate` ของ echo คือ **คำขอต่อวินาที** (burst ในวงเล็บ): `/api/health`,`/health` 1000 (100); `/tokenize`,`/transliterate` 200 (20) — ไม่มี route สองเส้นนี้แล้ว เป็นค่าค้าง; `/upload` (suffix นี้รวม `/image/upload`, `/video/upload`) 50 (10); `/api/product/search/unified` 500 (50); อื่น ๆ 100 (10) (`bootstrap.go:419-497`)
- Auth: `createGoAPIAuthMiddleware` อ่าน `Authorization: Bearer` → `authService.AuthenticateAccessToken` → ต้องมี `HoldingCode` ไม่ว่าง → อ่าน `holdingcode|tenantid|database|dbname` จาก query หรือ JSON body (รวม nested `body`) ถ้าไม่ตรง token → 403 (`bootstrap.go:169-206`, `:237-286`); body ถูกอ่านแล้วใส่คืนด้วย `io.NopCloser` (`bootstrap.go:253`) — unit test `bootstrap_test.go:14-48`
- session/token อยู่ในตาราง PostgreSQL `cache_entries` ของฐานกลาง (`backend/pkg/microservice/cacher.go:16,46`; cacher สร้างจาก `controlDB` ที่ `backend/main.go:90-94`); `authService` สร้างใน `RegisterRoutes` ด้วย `microservice.NewAuthService(cacher, 15m, 8h, finders...)` (`bootstrap.go:307`) โดย `controlDB` เป็น `AuthorizationFinder` → `AuthenticateAccessToken` (`pkg/microservice/auth.go:920`) ตรวจสิทธิ์สดทุกคำขอผ่าน `liveAuthorization.Authorize` (`auth.go:950`, `pkg/microservice/live_authorization.go:46`)
- ข้อผิดพลาดของด่านตรวจสิทธิ์คืน `{success:false, code, message}` โดย `message` แปลตามภาษา (`?lang` หรือ `Accept-Language`) ผ่าน `language.Text` (`bootstrap.go:210-216`, ทดสอบที่ `bootstrap_test.go:123-172`)
- Group: `g` = public (ไม่ auth), `authGroup` = ผ่าน middleware ข้างบน (`bootstrap.go:308`)

## 4. ตาราง route ทั้งหมด (`RegisterRoutes`, `bootstrap.go:306-404`)
prefix จริง = `/goapi` + path; คอลัมน์ auth: **P** = public `g`, **A** = `authGroup`
| path (method) | handler | auth | อ้างอิง |
|---|---|---|---|
| `/`, `/version`, `/api/health` (GET) | inline closures (version/health JSON) | P | `bootstrap.go:311-329` |
| `/api/health/queue`, `/queue/:holdingcode`, `/database`, `/system` (GET) | `QueueStatusHandler`, `QueueShopStatusHandler` คืนข้อความคงที่ ไม่ query อะไร; `DatabaseHealthHandler`, `SystemHealthHandler` คืน `mypg.GetConnectionStats()` ซึ่ง**ว่างเสมอ** เพราะ map `connectionPools` (`mypg/fast_utils.go:27`) ถูกเติมเฉพาะ `PgSqlFastConnectLegacy` ที่ไม่มีผู้เรียก (`:79`) | P | `bootstrap.go:332-335`, `handlers/health.go:36-87` |
| `/api/products/:itemcode/costing-config` (GET,PUT), `/api/products/:itemcode/cost-layers` (GET), `/api/inventory/receipt|issue|transfer|adjustment|sales-return|purchase-return` (POST), `/api/reports/inventory-valuation`, `/api/reports/stock-card/:itemcode` (GET) | `inventory.RegisterRoutes` ทั้งชุด (PG per-holding) | A | `bootstrap.go:338-339`, `inventory/handler.go:15-34` |
| `/processstockcalccost`, `/api/stockcost/query|summary|check` (POST) | `ProcessStockCalcCostHandler` (คำนวณทันทีด้วย `stockengine.Recalculate` ไม่ผ่านคิว `handlers/process_stock_calc_cost.go:175-203`), `ProcessStockCostHandler`, `ProcessStockCostSummaryHandler`, `ProcessStockCostCheckHandler` | A | `bootstrap.go:342-345` |
| `/api/transaction/calculate|quick-calc|validate-payment|purchase-history` (POST) | `TransactionCalculatorHandler`, `QuickCalculatorHandler`, `ValidatePaymentHandler`, `PurchaseHistoryHandler` | A | `bootstrap.go:348-351` |
| `/api/report/sales/by-document|summary`, `/api/report/tax/vat-register|wht|wht/certificate`, `/api/report/tax/form/catalog|schema|prefill|compute|pdf|rdfile|save|list|load|delete`, `/api/report/debt/query` (POST) | `SalesReportByDocumentHandler`, `SalesReportSummaryHandler`, `TaxVatRegisterHandler`, `TaxWithholdingHandler`, `WhtCertificateHandler`, `TaxFormCatalogHandler`…`TaxFormDeleteHandler` (แบบยื่นภาษี ภ.ง.ด./ภ.พ./ภ.ธ. ดึงยอดจาก GL แล้วพิมพ์ PDF/ไฟล์ .txt), `DebtReportHandler` | A | `bootstrap.go:354-370` |
| `/api/product/search`, `/barcode` (POST), `/api/search/aliases` (GET,POST), `/api/search/aliases/:id` (DELETE), `/api/process/product-balance`, `/api/process/queue-status` (POST), `/api/product/cache/stats` (GET), `/api/product/cache/clear`, `/api/product/search/unified` (POST) | `ProductSearchHandler`, `ProductBarcodeSearchHandler` (`handlers/product_cache.go:218,439`), `SearchAliasListHandler`, `SearchAliasCreateHandler`, `SearchAliasDeleteHandler`, `ProductBalanceUpdateHandler`, `StockQueueStatusHandler` (อ่านคิว `stock_dirty` ผ่าน `stockengine.LoadQueueStatus` — `process/stockengine/dirty.go:219`), `ProductCacheStatsHandler`, `ProductCacheClearHandler`, `UnifiedProductSearchHandler` | A | `bootstrap.go:373-382` |
| `/api/stock-report/barcodes|warehouses` (POST) | `StockReportBarcodesHandler`, `StockReportWarehousesHandler` | A | `bootstrap.go:385-386` |
| `/s3/file/*` (GET), `/upload`, `/upload/init|chunk|merge` (POST), `/upload/status/:uploadID` (GET), `/upload/cancel/:uploadID` (DELETE), `/image/upload`, `/video/upload` (POST) | `S3FileProxyHandler`, `FileUploadHandler`, `InitChunkedUploadHandler`, `UploadChunkHandler`, `MergeChunksHandler`, `GetUploadStatusHandler`, `CancelUploadHandler`, `ImageUploadHandler`, `VideoUploadHandler` | A | `bootstrap.go:389-397` |
| `/api/language/:lang`, `/api/address/thailand` (GET) | `GetLanguageHandler`, `GetThailandAddressHandler` | P | `bootstrap.go:400-401` |

- ฝั่ง frontend เรียก goapi แบบทั่วไปผ่าน BFF `frontend/src/app/api/goapi/[...goPath]/route.ts` ที่มี allowlist เต็ม path: GET `api/reports/inventory-valuation` (`:14`) และ POST เฉพาะ `api/report/tax/*`, `api/report/debt/query`, `api/report/sales/by-document` (`:16-32`) — route อื่นในตารางเรียกผ่าน BFF ตัวนี้ไม่ได้; รูป/วิดีโอ/S3 ผ่าน `frontend/src/lib/image-upload-proxy.ts`
- route ของโมดูลที่ไม่มีในโค้ดแล้ว (LINE OA, อนุมัติ PO/PR/RFQ, AI chat/knowledge base, product v2, attachment, xlsx import ฯลฯ) ไม่มีไฟล์ handler เหลือใน `backend/internal/goapi/handlers/` (53 ไฟล์ `.go` รวม test)

## 5. Handler ที่มีโค้ดแต่ไม่ได้ลงทะเบียน (DEAD)
- `PgSelectHandler`, `PgExecHandler` (`handlers/database.go:127,306`) — raw SQL passthrough ยังอยู่ในไฟล์แต่ไม่ถูก `RegisterRoutes` เรียก และถูกล็อกไว้ใน `bootstrap_test.go:67-68` (ตรวจว่า `/goapi/get`, `/goapi/exec` ต้องไม่ถูก register) — เป็น handler เดียวใน `handlers/` ที่ไม่ได้ลงทะเบียน
- `inventory.CreateTablesHandler` (`inventory/handler.go:277`) ไม่ได้ลงทะเบียน (`/goapi/api/inventory/create-tables` อยู่ใน forbidden list) และเป็นผู้เรียก `CreateInventoryCostingTables` (`inventory/database.go:12`) เพียงรายเดียว (`inventory/handler.go:288`)
- `TestGoAPIRouteSurfaceExcludesOperationalEndpoints` (`bootstrap_test.go:50-105`) มี forbidden อีก ~30 path (report/result/pdf, rebuild progress, data passthrough/copy, deploy, image list/verify, test endpoints) — เป็น regression guard กันไม่ให้ route เหล่านี้กลับมา; นอกจาก `/get`, `/exec` และ `api/inventory/create-tables` แล้ว handler ของ path เหล่านี้ไม่มีในโค้ด
- `backend/internal/goapi/gemini/client.go` และ `backend/internal/goapi/ragflow/*.go` (4 ไฟล์) — client ไป Gemini/RAGFlow ยังอยู่ แต่ไม่มี Go source ใด import (`grep -rl 'goapi/gemini\|goapi/ragflow' backend --include=*.go` ว่าง)
- `mypg/query_results.go` (`CreateResultTableIfNotExists`, `InsertQueryResult*`, `GetQueryResults`, `DeleteQueryResults`) ไม่มีผู้เรียก

## 6. Stock engine (`process/stockengine/`) และคิวที่ไม่มีผู้ใช้
- ต้นทุนสินค้าต่อ item คำนวณโดย PostgreSQL LISTEN/NOTIFY worker: `Worker.Run` (`worker.go:60-99`) รอสัญญาณจาก `pq.NewListener` บน channel `stock_dirty` (`notifyChannel`, `dirty.go:26`) หรือ poll ทุก 15 วินาที (`fallbackInterval`, `worker.go:19`) → `DrainOnce` (`worker.go:101-112`) → `ClaimDirty` (`FOR UPDATE SKIP LOCKED`, `dirty.go:105-122`) → `Recalculate` (`store.go:367`) → `AfterRecalculate` callback ที่ผูกใน `bootstrap.go:116-120`
- Error handling (`worker.go:114-145`): ล็อกชน (`ErrLockBusy`, `:127`) → `ReleaseDirty` คืนคิวทันที; error อื่นครบ `MaxAttempts` = 5 (`dirty.go:20`, ตรวจที่ `worker.go:136`) → `MoveToDeadLetter` (`dirty.go:183`, ตาราง `stock_dead_letter`)
- `StartWorkers(ctx, transFlags)` แค่เปิด manager (`manager.go:36-48`); worker ต่อ holding เกิดจาก `EnsureWorker(holdingCode, db, dsn)` (`manager.go:78`) เท่านั้น — **ปัจจุบันไม่มีโค้ดใดเรียก `EnsureWorker`** และไม่มีผู้เรียก `MarkDirty`/`MarkDocumentDirty` (`dirty.go:42,70`) นอกจาก `mypg.DeleteDocPgSql*` (`mypg/doc.go:15-32`) ซึ่งก็ไม่มีผู้เรียก (ผู้เรียกเดิมถูกลบใน commit `550d5489` วันที่ 2026-09-23) → worker เบื้องหลังไม่เคยเริ่มทำงาน; การคำนวณต้นทุนตอนนี้เกิดได้ทางเดียวคือ `/processstockcalccost` (§4)
- ไม่พบโค้ดนอก test ที่สร้างตาราง `doc`, `docdetail`, `stock_ledger`, `stock_dirty`, `stock_period_balance` ที่ stock engine และรายงานสต็อกอ่าน (`CREATE TABLE` ของตารางเหล่านี้มีแค่ใน `process/stockengine/store_integration_test.go:57-68`)
- `mypostgres/queue.go` (`QueueManager`: `AddToQueue`, `PopFromQueue`, `AddToDeadLetterQueue`, `GetActiveShops` ฯลฯ) ไม่มีผู้เรียก `NewQueueManager` เลย แต่ `InitQueueSchema` (`mypostgres/init_schema.go:11`) ยังสร้างตาราง `queues`/`deadletterqueue` ทุกครั้งที่บูต; `process/process_status.go:11-56` (`ProcessPurchaseOrderStatus`, `ProcessSaleInvoiceStatus`, `ProcessCreditorStatus`, `ProcessCustomerStatus`, `ProcessDebtorStatus`) เป็น placeholder `TODO` ไม่มีผู้เรียก

## 7. Connection managers: `mydb` / `mypg`
| ชิ้น | หน้าที่ | ค่าคงที่สำคัญ | อ้างอิง |
|---|---|---|---|
| `mydb.ManagerPool` | 1 `DatabaseManager` ต่อ holdingcode (ชื่อฐาน PostgreSQL = holdingcode); สร้างใหม่เมื่อไม่ healthy | PG pool 50/15/1h | `mydb/manager_pool.go:60-109`, `mydb/unified_database_manager.go:148-150` |
| `mydb.CircuitBreaker` | ห่อ query PostgreSQL ใน `DatabaseManager` (`pgCircuitBreaker`, `unified_database_manager.go:58,219,277`) | maxFailures 5, resetTimeout 10s, half-open 3 calls | `mydb/circuit_breaker.go:62-67` |
| `mypg.PgSqlFastConnect` / `Connect` | ทางเข้าที่ handler ใช้ (เช่น inventory, tax) → `PgSqlFastConnectV2` → `mydb.GetGlobalConnectionFromPool` คือ pool เดียวกับ `ManagerPool` | ตาม `ManagerPool` | `mypg/fast_utils.go:39-43`, `mypg/utils.go:21-25`, `mypg/compatibility.go:11-13`, `mydb/manager_pool.go:194-197` |
| prepared-statement cache (LRU) | `CleanupPreparedStatements` ถูกเรียกทุก 30 นาที (§2) | สูงสุด 1000 statement, idle 1 ชม. | `mypg/fast_utils.go:33-34,418` |
| `myglobal` global DB | connection กลาง (database `postgres`) ที่ใช้สร้างตาราง queue ใน §6 | — | `bootstrap.go:70-98` |
- โค้ด pool ที่ไม่มีผู้เรียก: `mydb.ConnectionPoolManager` (config 75 conns, ใช้แค่ใน `NewDatabasePoolOptimizer` ที่ไม่มีผู้เรียก — `mydb/connection_pool_manager.go:40,249-252`), `mydb.GetBrandDB` (`mydb/brand_db.go:20`), `mypg.PgSqlFastConnectLegacy`/`ConnectLegacy`/`ConnectAdminDatabase` (`mypg/fast_utils.go:46`, `mypg/utils.go:29`, `mypg/compatibility.go:52`) — ค่า 25/7/10m ใน `mypg/fast_utils.go:190-192` และ `mypg/utils.go:57-59` อยู่ในเส้นทางเหล่านี้ ไม่ใช่เส้นทางหลัก

## 8. ภาษาไทย: `language` package และที่อยู่
- `language/language.go`: โหลด dictionary จาก `backend/assets/language/languages.tsv`, `Normalize/Text/Dictionary/Load` (`language.go:48,69,81,102`) → เสิร์ฟผ่าน `GetLanguageHandler` (`handlers/language_handler.go:14-31`) ทั้ง `/goapi/api/language/:lang` และ alias root (`backend/main.go:178`)
- `GetThailandAddressHandler` (`handlers/address_handler.go:25`) เสิร์ฟไฟล์ JSON ที่อยู่ไทยจาก asset ในเครื่อง (env `THAILAND_ADDRESS_DATA_PATH` หรือ path สำรองใน `thailandAddressDataPathCandidates` `address_handler.go:52`) ไม่ได้อ่านฐานข้อมูล
- ตัดคำไทย: `UnifiedProductSearchHandler` เรียก `tokenizeKeyword` (`handlers/product_search.go:151,381`) ซึ่งยิงไป `THAI_NLP_URL` (default `http://thai-nlp:8890`, `handlers/thai_nlp_handlers.go:55-60`) พร้อม circuit 60 วินาที; `deploy/account/compose.yml` ไม่มี service `thai-nlp` → บน prod จะตกไป `fallbackTokenize` (`product_search.go:440`)

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ
1. ยังไม่ตรวจว่ามีผู้เรียก route นอก allowlist ของ BFF (§4) ทางอื่นหรือไม่ — เช่น stock/transaction/product search/aliases
2. ตารางที่ inventory และ stock engine อ่าน (§5–§6) ไม่มีโค้ดสร้างในระบบปัจจุบัน — ยังไม่ตรวจว่าบนฐานของ holding ที่ใช้งานอยู่มีตารางเหล่านี้หรือไม่
3. โค้ดที่ไม่มีผู้เรียก (§5–§7: `gemini`/`ragflow`, `mypostgres/queue.go`, `process/process_status.go`, stock worker, pool legacy) รอการตัดสินใจว่าจะลบหรือกลับมาใช้เมื่อสร้างโมดูลบน PostgreSQL ใหม่
