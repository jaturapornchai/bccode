# ชั้น goapi — HTTP routes, Kafka consumers, workers, และ connection managers
> ตรวจล่าสุด: 2026-09-07 (commit d93a210d) — ผู้เขียน: AI reader; ทุกข้อเท็จจริงอ้าง path:line · ตรวจซ้ำโดย fact-checker

## 1. ภาพรวมและจุดเกาะกับ mainapi
- goapi ไม่ใช่ binary แยก: `backend/main.go:568-576` สร้าง `goapi.New()` → `Init()` → mount ที่ `ms.Echo().Group("/goapi")` → `RegisterMiddleware` → `RegisterRoutes(goapiGroup, "/goapi", pst)` และ `defer Shutdown()`; ทำเฉพาะเมื่อ `DEV_API_MODE == "" || "2"` (`backend/main.go:257`). โหมด `"3"` = migration (`backend/main.go:585`) ไม่ mount goapi
- `/goapi/*` และ `/api/language/*` อยู่ใน bypass list ของ auth กลาง mainapi (`backend/main.go:296-297`) เพราะ goapi มี auth middleware ของตัวเอง (§3); alias สาธารณะ `GET /api/language/:lang` ลงทะเบียนตรงที่ root echo (`backend/main.go:580`)
- ถ้า `Init()` คืน error → log "goapi routes disabled" (`backend/main.go:570`) แต่โค้ด `Init()` ปัจจุบัน return nil เสมอ (`backend/internal/goapi/bootstrap.go:165`) — ความล้มเหลวทุกขั้นเป็นแค่ log
- Version string `1.1.1121` (`bootstrap.go:42`) คืนที่ `GET /goapi/`, `/goapi/version`, `/goapi/api/health` (`bootstrap.go:356-375`)

## 2. ลำดับ `Init()` (`bootstrap.go:55-166`)
| ขั้น | ทำอะไร | อ้างอิง |
|---|---|---|
| 1 | `setupconfig.LoadBootstrapConfig()` แล้ว `serviceConfig.NewServiceConfig()` | `bootstrap.go:59-61` |
| 2-3 | `myglobal.SafeMongoConnectFast()` + `handlers.InitMongoAtlas()` (Mongo 2 connection; fail = Warn) | `bootstrap.go:65-70` |
| 4 | `lineoa.Init` + `approval.Init` ด้วย atlas client (เฉพาะเมื่อ connect สำเร็จ) | `bootstrap.go:73-79` |
| 5 | `handlers.InitR2Client()` (S3/MinIO; fail = Warn) | `bootstrap.go:82-84` |
| 6 | `mydb.InitManagerPool` จาก env `POSTGRES_*` + `CLICKHOUSE_*` (key name เท่านั้น) | `bootstrap.go:88-99` |
| 7 | ขอ manager `"postgres"` → global `*sql.DB` + provider closure → `mypostgres.InitQueueSchema` (ตาราง queues/deadletterqueue/distributedlocks) | `bootstrap.go:104-131`, `mypostgres/init_schema.go:11`, `mypostgres/schema.sql:8,28,45` |
| 8 | `workers.NewWorkerManager(0)` → `Start()` (จำนวน worker auto §7) | `bootstrap.go:136-143` |
| 9 | goroutine: `StartConnectionHealthChecker` (ทุก 5 นาที), `StartBackgroundTask` (ทุก 10 นาที), `CleanupPreparedStatements` ทุก 30 นาที | `bootstrap.go:146-154`, `handlers/health.go:28,53` |
| 10 | ถ้า env `ENABLE_KAFKA=="true"` → `handlers.StartConsumers()` (§5) | `bootstrap.go:157-162`, key map ที่ `setupconfig/loader.go:78` |
- `Shutdown()` หยุด worker, ปิด Mongo ทั้งสอง, `mydb.CloseAllManagers`, `myclickhouse.CloseClickHouseConnection` (no-op stub), `myPg.CloseAllPools` (`bootstrap.go:577-589`)

## 3. Middleware และ auth ของ goapi
- `RegisterMiddleware`: Recover (panic → JSON 500 พร้อม `err.Error()` ดิบ), Secure headers (CSP `default-src 'self'`), tiered rate limiter, timeout 30s (`bootstrap.go:169-206`)
- Rate limit ต่อ IP ตาม suffix ของ path: `/api/health`,`/health`=1000/min; `/tokenize`,`/transliterate`=200 (ไม่มี route นี้จริง); `/upload`=50; `/api/product/search/unified`=500; อื่น ๆ 100 burst 10 (`bootstrap.go:595-672`)
- Auth: `createGoAPIAuthMiddleware` อ่าน `Authorization: Bearer` → `authService.AuthenticateAccessToken` (Redis token ผ่าน `microservice.NewAuthService(authCacher, 15m, 8h, finders...)`; `authCacher` = `microservice.NewCacher(...)` ซึ่งเป็น go-redis client `backend/pkg/microservice/cacher.go:14,95`) → ต้องมี `HoldingCode` ไม่ว่าง → อ่าน `holdingcode|tenantid|database|dbname` จาก query หรือ JSON body (รวม nested `body`) แล้วถ้าไม่ตรง token → 403 (`bootstrap.go:208-259`, `:281-341`, `:350-353`); body ถูกอ่านแล้วใส่คืนด้วย `io.NopCloser` (`bootstrap.go:297`) — มี unit test `bootstrap_test.go:13-47`
- Group: `g` = public (ไม่ auth), `authGroup` = ผ่าน middleware ข้างบน (`bootstrap.go:353`)

## 4. ตาราง route ทั้งหมด (`RegisterRoutes`, `bootstrap.go:350-574`) — ทั้งหมด LIVE
prefix จริง = `/goapi` + path; คอลัมน์ auth: **P** = public `g`, **A** = `authGroup`
| path (method) | handler | auth | อ้างอิง |
|---|---|---|---|
| `/`, `/version`, `/api/health` (GET) | inline closures (version/health JSON) | P | `bootstrap.go:356-375` |
| `/api/health/kafka`, `/background`, `/queue`, `/queue/:holdingcode`, `/database`, `/system` (GET) | `handlers.KafkaHealthHandler`, `BackgroundTaskStatusHandler`, `QueueStatusHandler`, `QueueShopStatusHandler`, `DatabaseHealthHandler`, `SystemHealthHandler` | P | `bootstrap.go:377-382`, `handlers/health.go:105-249` |
| `/api/products/:itemcode/costing-config` (GET,PUT), `/api/products/:itemcode/cost-layers` (GET) | `inventory.GetCostingConfig`, `UpdateCostingConfig`, `GetCostLayers` | A | `bootstrap.go:385-386`, `inventory/handler.go:17-21` |
| `/api/inventory/receipt|issue|transfer|adjustment|sales-return|purchase-return` (POST) | `inventory.ProcessReceipt`…`ProcessPurchaseReturn` (PG per-holding ผ่าน `myPg.PgSqlFastConnect(holdingCode)`) | A | `inventory/handler.go:24-29,39` |
| `/api/reports/inventory-valuation`, `/api/reports/stock-card/:itemcode` (GET) | `inventory.GetInventoryValuation`, `GetStockCard` | A | `inventory/handler.go:32-33` |
| `/processstockcalccost`, `/api/stockcost/query|summary|check` (POST) | `ProcessStockCalcCostHandler`, `ProcessStockCostHandler`, `ProcessStockCostSummaryHandler`, `ProcessStockCostCheckHandler` | A | `bootstrap.go:389-392` |
| `/api/transaction/calculate|quick-calc|validate-payment|purchase-history` (POST) | `TransactionCalculatorHandler`, `QuickCalculatorHandler`, `ValidatePaymentHandler`, `PurchaseHistoryHandler` | A | `bootstrap.go:395-398` |
| `/api/report/sales/by-document|summary` (POST) | `SalesReportByDocumentHandler`, `SalesReportSummaryHandler` | A | `bootstrap.go:401-402` |
| `/api/product/search`, `/barcode`, `/barcode/list`, `/search/unified` (POST) | `ProductSearchHandler`, `ProductBarcodeSearchHandler`, `BarcodeListHandler`, `UnifiedProductSearchHandler` | A | `bootstrap.go:405-407,414` |
| `/api/search/aliases` (GET,POST), `/api/search/aliases/:id` (DELETE) | `SearchAliasListHandler`, `SearchAliasCreateHandler`, `SearchAliasDeleteHandler` | A | `bootstrap.go:408-410` |
| `/api/process/product-balance` (POST), `/api/product/cache/stats` (GET), `/api/product/cache/clear` (POST) | `ProductBalanceUpdateHandler`, `ProductCacheStatsHandler`, `ProductCacheClearHandler` | A | `bootstrap.go:411-413` |
| `/product/v2/item/get|update-listing|readiness`, `/product/v2/tier/init|update` (POST) | `ProductV2ItemGetHandler`, `ProductV2ItemUpdateListingHandler` (whitelist ฟิลด์ E24 `handlers/product_v2_whitelist.go:3-15`), `ProductV2ItemReadinessHandler`, `ProductV2TierInitHandler`, `ProductV2TierUpdateHandler` | A | `bootstrap.go:417-421` |
| `/api/stock-report/barcodes|warehouses` (POST) | `StockReportBarcodesHandler`, `StockReportWarehousesHandler` | A | `bootstrap.go:424-425` |
| `/api/lineoa/configs|config|config/save|employees|employee/add|employee/remove|employee/link`, `/api/user/lineoa/link|callback|profile` (POST) | `lineoa.GetConfigsHandler`…`GetUserProfileHandler` (10 ตัว) | A | `bootstrap.go:428-437` |
| `/api/lineoa/webhook`, `/webhook/lineoa` (POST) | `lineoa.WebhookHandler` | P | `bootstrap.go:438-439` |
| `/api/approval/po-settings|po-setting|po-setting/save|po-setting/delete` (POST) | `approval.GetPOApprovalSettingsHandler`, `GetPOApprovalSettingHandler`, `SavePOApprovalSettingHandler`, `DeletePOApprovalSettingHandler` | A | `bootstrap.go:442-445` |
| `/api/approval/po-status/get|batch|submit|approve|reject|withdraw|pending|rejected` (POST) | `GetPOApprovalStatusHandler`, `GetBatchPOApprovalStatusHandler`, `SubmitPOApprovalHandler`, `ApprovePOHandler`, `RejectPOHandler`, `WithdrawPOHandler`, `GetPendingApprovalsHandler`, `GetRejectedPOListHandler` | A | `bootstrap.go:446-453` |
| `/api/approval/notification/check|send|process|logs|mark-opened|send-real|resend`, `/api/approval/timeline` (POST) | `CheckNotificationSentHandler`, `SendApprovalNotificationHandler`, `ProcessPendingNotificationsHandler`, `GetNotificationLogsHandler`, `MarkNotificationOpenedHandler`, `GetApprovalTimelineHandler`, `SendRealApprovalNotificationHandler`, `ResendApprovalNotificationHandler` | A | `bootstrap.go:454-461` |
| `/api/approval/smtp-status|lineoa-config-status|lineoa-configs` (GET) | `GetSMTPStatusHandler`, `GetLineOAConfigStatusHandler`, `ListAllLineOAConfigsHandler` | A | `bootstrap.go:462-464` |
| `/api/approval/action` (GET,POST), `/api/approval/token-info` (GET) | `approval.ApproveViaTokenHandler`, `GetApprovalTokenInfoHandler` (อนุมัติผ่าน token ในอีเมล/LINE) | P | `bootstrap.go:465-467` |
| `/api/approval/po-details`, `/api/approval/liff-approve` (POST) | `GetPODetailsForLIFFHandler`, `LiffApproveHandler` | A | `bootstrap.go:468-469` |
| `/api/approval/pr-*` (12 routes) และ `/api/approval/rfq-*` (12 routes) | alias ของ handler PO ชุดเดียวกัน (frontend ส่ง `purchasetypecode` = PR/RFQ) | A | `bootstrap.go:473-499` |
| `/api/datahistory`, `/api/datahistory/po` (GET) | `datahistory.GetHistoryHandler`, `GetPOHistoryHandler` | A | `bootstrap.go:502-503` |
| `/api/purchase-order/manual-close` (POST) | `handlers.ManualClosePOHandler` (มี `updateClickHouseManualClose` เรียก stub → ได้ error "clickhouse is disabled" → `logger.Error` แล้ว `return` โดยไม่กระทบผลลัพธ์ของ handler) | A | `bootstrap.go:506`, `handlers/purchase_order_close.go:27,139,166-170` |
| `/s3/file/*` (GET) | `S3FileProxyHandler` | A | `bootstrap.go:509` |
| `/upload` (POST), `/upload/init|chunk|merge` (POST), `/upload/status/:uploadID` (GET), `/upload/cancel/:uploadID` (DELETE) | `FileUploadHandler`, `InitChunkedUploadHandler`, `UploadChunkHandler`, `MergeChunksHandler`, `GetUploadStatusHandler`, `CancelUploadHandler` | A | `bootstrap.go:512-517` |
| `/api/language/:lang`, `/api/address/thailand` (GET) | `GetLanguageHandler`, `GetThailandAddressHandler` | P | `bootstrap.go:520-521` |
| `/image/upload|list|get|info|delete|promptpayverify`, `/video/upload` (POST) | `ImageUploadHandler`, `VideoUploadHandler`, `ImageListHandler`, `ImageGetHandler`, `ImageInfoHandler`, `ImageDeleteHandler`, `ImageVerifyHandler` | A | `bootstrap.go:524-530` |
| `/api/attachment/upload|list|delete` (POST), `/api/attachment/download/:id` (GET) | `AttachmentUploadHandler`, `AttachmentListHandler`, `AttachmentDeleteHandler`, `AttachmentDownloadHandler` | A | `bootstrap.go:533-536` |
| `/xlsx/product/start` (POST) | `dataimport.StartProductPrepareHandler` | A | `bootstrap.go:539` |
| `/api/v1/chatbot/chat-gemini`, `/analyze-document` (POST) | `aichat.ChatGemini`, `aichat.AnalyzeDocument` | A | `bootstrap.go:542-544` |
| `/api/v1/kb/health|view` (GET), `/list|upload|delete|update-status|update-allday|update-schedule|query` (POST) | `knowledgebase.Health`…`Query` (RAGFlow) | A | `bootstrap.go:547`, `handlers/knowledgebase/handler.go:481-490` |
| `/api/aichat/v1/models` (GET), `/api/aichat/result/:id` (GET) | `aichat.OpenAIGatewayListModels`, `aichat.GetAIChatResult` | A | `bootstrap.go:550-554` |
| `/api/v1/ai-provider/list|save|delete|models|status|question-history|complaint` (POST) | `aichat.ListAIProviders`, `SaveAIProvider`, `DeleteAIProvider`, `ListAIModels`, `AIProviderStatus`, `ListQuestionHistory`, `SubmitComplaint` | A | `bootstrap.go:557-564` |
| `/api/v1/unified/query` (POST), `/health`, `/cache/stats` (GET) | `unified.UnifiedAPIServer.ProcessUnifiedQuery`, `HealthCheck`, `GetCacheStats` | A | `bootstrap.go:567-571` |

## 5. Handler ที่มีโค้ดแต่ไม่ได้ลงทะเบียน (DEAD)
วิธีตรวจ: รวบรวม `func X(c echo.Context) error` ทุกตัวใน `backend/internal/goapi/**` (190 บรรทัด = 184 ชื่อไม่ซ้ำ; ชื่อ `Test*Consumer`/`GetConsumerStatus` ซ้ำกันใน 2 ไฟล์) เทียบกับชื่อที่ปรากฏใน `bootstrap.go`, `inventory/handler.go`, `knowledgebase/handler.go`, `backend/main.go` → **54 ชื่อ**ไม่ถูกอ้าง (ตรวจซ้ำ 2026-09-07; ไม่นับ helper ภายใน 3 ตัวที่มี signature เดียวกันแต่ไม่ใช่ route handler: `checkConnection`, `productV2NotFound`, `validateDeployToken`); 39 path ในนั้นถูกล็อกไว้เป็น forbidden โดย `bootstrap_test.go:49-105` (slice `forbidden` ที่ `:53-100`)
| กลุ่ม | handler (DEAD) | ไฟล์ |
|---|---|---|
| raw SQL / Mongo passthrough | `PgSelectHandler`, `PgExecHandler`, `PgGetDocHandler`, `MongoGetDataHandler`, `MongoAtlasGet/Update/DeleteHandler` | `handlers/database.go:127,306`, `handlers/getdoc.go:20,134`, `handlers/mongodb_atlas.go:245-546` |
| report/result/PDF | `ReportGetHandler`, `ReportPostHandler`, `RebuildProgressSSEHandler`, `ResultFromQueryHandler`, `ResultGetHandler`, `ResultToPDFHandler`, `GenPDFHandler`, `PdfHistoryGet/List`, `PdfReprintHandler` | `handlers/reports.go:24`, `handlers/commands.go:31,154`, `handlers/result-handlers.go:198-711`, `handlers/genpdf_handler.go:29-723` |
| ClickHouse | `ClickHouseQueryHandler`, `ClickHouseSelectHandler`, `ClickHouseMultiQueryHandler` | `handlers/clickhouse.go:14,76,191` |
| ops/migration/deploy | `CopyMongoUatToDevHandler`, `PreviewCopyMongoHandler`, `ListSourceShopsHandler`, `MigrateCurrencyColumnsHandler`, `BackfillCurrencyDataHandler`, `MigrateClickHouseSoftDeleteHandler`, `DeployBackend/Frontend/StatusHandler`, `Setup*Handler` (7 ตัว) | `handlers/mongo_copy_uat_to_dev.go:227-610`, `handlers/migrate_currency.go:19-280`, `handlers/deploy_handler.go:55-280`, `handlers/setup_config_handler.go:101-715` |
| test/status ของ Kafka | `Test*Consumer` (5 ตัวใน `handlers/kafka.go:1014-1163`, 8 ตัวใน `handlers/kafka/api.go:11-258`), `GetConsumerStatus` | ตามซ้าย |
| อื่น ๆ | `dataimport.GetProductPrepareStatusHandler`, `GetProductPrepareResultHandler` (เริ่ม import ได้แต่ไม่มี route ดูผล), `approval.SendTestEmailHandler`, `GetEmailStatusHandler`, `TestLinePushHandler`, `lineoa.TestHandler`, `aichat.TestAIProvider`, `datahistory.MigrateHistoryHandler` | `dataimport/xlsx_product.go:292,345`, `handlers/approval/notification.go:1477,1442,1723`, `handlers/lineoa/handlers.go:248`, `handlers/aichat/ai_provider_config.go:218`, `handlers/datahistory/migrate.go:24` |
- `handlers/process_consumer.go` (146 บรรทัด) ทั้งไฟล์เป็น comment นอกจาก `package handlers` (import block ที่ถูก comment เริ่ม `handlers/process_consumer.go:3`; `grep -vc '^\s*//|^\s*$|^package'` = 0) — ไม่มีโค้ดจริง

## 6. Kafka consumers (`handlers.StartConsumers`, `handlers/kafka.go:26-1009`)
- เงื่อนไข: ต้องมี env `KAFKA_SERVER_URL` (`kafka.go:28-32`); group id = base + `-` + `KAFKA_CONSUMER_GROUP_VERSION` (default `v1`) (`kafka.go:19-23`, `config/config.go:118-120`)
- ทุก topic เปิด goroutine แยก 1 ตัวเรียก `consumers.ConsumeMessage(server, topic, group, 0, fn)` และ fn ส่งต่อไป `handlers.Call*Consumer` → `kafka.OnConsumeMessage*` ใน `handlers/kafka/*.go` (`handlers/kafka_bridge.go:195-580`)
| topic family (created/updated → handler; deleted → handler) | group (ก่อนต่อ -v1) | อ้างอิง kafka.go |
|---|---|---|
| when-saleinvoice-* → `CallSaleInvoiceConsumer` / `CallSaleInvoiceDeleteConsumer` | goapi-saleinvoice-consumer | `:41-68` |
| when-saleinvoicereturn-* → `CallSaleReturnConsumer` / `CallSaleReturnDeleteConsumer` | goapi-saleinvoicereturn-consumer | `:79-107` |
| when-product-barcode-* → `CallInventoryConsumer` / `CallInventoryDeleteConsumer` | biapi-inventory-consumer | `:118-147` |
| when-product-* → `CallProductConsumer` / `CallProductDeleteConsumer` | biapi-product-consumer | `:158-179` |
| when-product-barcode-bulk-* → `CallInventoryBulkConsumer` / `CallInventoryBulkDeleteConsumer` | biapi-inventory-bulk-consumer | `:190-219` |
| when-warehouse-* → `CallWarehouseConsumer` / `CallWarehouseDeleteConsumer` | biapi-warehouse-consumer | `:230-258` |
| when-saleorder-* → `CallSaleOrderConsumer` / `CallSaleOrderDeleteConsumer` | goapi-saleorder-consumer | `:269-298` |
| when-purchase-* → `CallPurchaseConsumer` / `CallPurchaseDeleteConsumer` | goapi-purchase-consumer | `:309-338` |
| when-purchaseorder-* → `CallPurchaseOrderConsumer` / `CallPurchaseOrderDeleteConsumer` | goapi-purchaseorder-consumer | `:349-376` |
| when-purchaserequisition-* และ -bulk-* → `CallPurchaseRequisitionConsumer` / `CallPurchaseRequisitionDeleteConsumer` | goapi-purchaserequisition-consumer | `:387-444` |
| when-rfq-* และ -bulk-* → `CallRFQConsumer` / `CallRFQDeleteConsumer` | goapi-rfq-consumer | `:455-512` |
| when-purchasepartial-* → `CallPurchasePartialConsumer` / `CallPurchasePartialDeleteConsumer` | goapi-purchasepartial-consumer | `:523-550` |
| when-purchasereturn-* → `CallPurchaseReturnConsumer` / `CallPurchaseReturnDeleteConsumer` | goapi-purchasereturn-consumer | `:561-588` |
| when-stocktransfer-* → `CallStockTransferConsumer` / `CallStockTransferDeleteConsumer` | goapi-stocktransfer-consumer | `:599-626` |
| when-stockreceiveproduct-* → `CallStockReceiveProductConsumer` / `...DeleteConsumer` | goapi-stockreceiveproduct-consumer | `:637-664` |
| when-stockpickupproduct-* → `CallStockPickupProductCreateOrUpdateConsumer` / `...DeleteConsumer` | goapi-stockpickupproduct-consumer | `:675-702` |
| when-stockreturnproduct-* → `CallStockReturnProductCreateOrUpdateConsumer` / `...DeleteConsumer` | goapi-stockreturnproduct-consumer | `:713-740` |
| when-stockadjustment-* → `CallStockAdjustmentCreateOrUpdateConsumer` / `...DeleteConsumer` | goapi-stockadjustment-consumer | `:751-778` |
| when-stockbalance-* → `CallStockBalanceCreateOrUpdateConsumer` / `...DeleteConsumer` | goapi-stockbalance-consumer | `:789-816` |
| when-creditor-* → `CallCreditorConsumer` / `CallCreditorDeleteConsumer`; -bulk-created/-bulk-deleted → `CallCreditorBulkConsumer` / `CallCreditorBulkDeleteConsumer` | goapi-creditor-consumer / goapi-creditor-bulk-consumer | `:827-875` |
| when-customer-* → `CallCustomerConsumer` / `CallCustomerDeleteConsumer` | goapi-customer-consumer | `:886-910` |
| when-employee-* → `CallEmployeeConsumer` / `CallEmployeeDeleteConsumer` | goapi-employee-consumer | `:921-945` |
| when-debtor-* → `CallDebtorConsumer` / `CallDebtorDeleteConsumer`; -bulk-created/-bulk-deleted → `CallDebtorBulkConsumer` / `CallDebtorBulkDeleteConsumer` | goapi-debtor-consumer / goapi-debtor-bulk-consumer | `:956-1004` |
- หลักฐาน runtime (local, read-only): `docker exec kafka kafka-consumer-groups --list` เห็นครบ 25 group ลงท้าย `-v1` ตรงตาราง; `docker logs mainapi` มี "GoAPI: ✅ Kafka consumers initialized" (2026-09-05 23:27:40)
- **`handlers/kafka/manager.go` = DEAD**: `kafka.StartConsumers()` (`manager.go:10-48` + helper `start*Consumers` `:51-112`; ใช้ค่าคงที่ `CONSUMER_GROUP_*` ที่ไม่มี version suffix `handlers/kafka/constants.go:86-92`, log บอกเอง "Total consumer groups: 6" `manager.go:47`) ถูกเรียกจาก `StartKafkaConsumersWithActualImplementation` เท่านั้น (`handlers/kafka_bridge.go:584-586`) ซึ่งไม่มีผู้เรียกใน repo; `handlers/kafka/exports.go:7` re-export ไว้เฉย ๆ
- Consumer แต่ละตัวเรียก `build.DatabaseChecker(holdingCode,false)` เพื่อสร้าง PG database/ตารางของ holding อัตโนมัติครั้งแรก (`process/build/create-database.go:2019-2031`; caller เช่น `handlers/utils.go:112` และ `handlers/kafka/*.go` 10 จุดใน 9 ไฟล์)

## 7. กลไกใต้ consumer: `mykafkaconsumer`, projection, retry, DLQ
- Reader (segmentio kafka-go): `StartOffset=FirstOffset`, `MinBytes=1`, `MaxBytes=10MB`, `MaxWait=1s`, `SessionTimeout/RebalanceTimeout=30s`, `HeartbeatInterval=3s` (`mykafkaconsumer/kafka_consumer.go:425-444`); `CommitInterval` = 0 (sync) สำหรับ product topics, 5s async สำหรับที่เหลือ (`mykafkaconsumer/product_projection.go:25-30`)
- เส้นทาง A (product/barcode 9 topics, `product_projection.go:14-23`): `ConsumeProjectionMessages` ประมวลผลใน goroutine reader เอง, commit หลัง handler สำเร็จ; message โครงสร้างเสียถูก ack แล้ว log ตำแหน่ง ไม่ log payload (`product_projection.go:36-69`)
- เส้นทาง B (topic อื่นทั้งหมด): reader loop แยก `holdingcode` จาก payload (`kafka_consumer.go:548`) แล้วโยนเข้า channel ต่อ shop (buffer 500, 3 worker/shop, สูงสุด 1000 shop LRU, job timeout 30s) (`kafka_consumer.go:69-84,133-180,191-242`); ถ้า channel เต็มรอ 5s แล้ว block (`kafka_consumer.go:347-356`) — loop ใช้ `reader.ReadMessage` (`kafka_consumer.go:298`, auto-commit ของ kafka-go) + `CommitInterval` 5s → offset ถูก commit แบบ async โดยไม่รอผล handler → ล้มเหลว = log แล้วข้าม (at-most-once); เมื่อ shop worker ครบ 1000 จะ `evictLRUWorker` (`kafka_consumer.go:153-160,515`)
- Consume loop reconnect: backoff exponential, `maxRetryAttempts=3` แล้วพัก 30s วนใหม่ไม่รู้จบ (`kafka_consumer.go:243-271`)
- `SafeConsumerWrapper` (`handlers/kafka/utils.go:37-110`): recover panic → `mydlq.QuickSendToDLQ`; retry ผ่าน `myretry.WithRetry` MaxRetries 3, 1s→30s x2 เฉพาะ error ที่มีคำว่า connection/timeout/temporary/unavailable; หมด retry → `mydlq.SendToDLQ`
- **DLQ = log เท่านั้น**: `globalDLQHandler` default = `NewLogDLQHandler()` (`mydlq/dlq.go:131`) ซึ่งพิมพ์ 500 ตัวอักษรแรกลง log (`mydlq/dlq.go:84-101`); `FileDLQHandler` (`dlq.go:33-78`) มีแต่ไม่มีใครเรียก `SetDLQHandler` (`dlq.go:134`) (ผู้เรียก `mydlq.*` มีแค่ `handlers/kafka/utils.go:48,101`); ตาราง PG `deadletterqueue` (`mypostgres/schema.sql:28`) ใช้โดย queue ของ workers (§8) ไม่ใช่ DLQ ของ Kafka
- `myretry` ใช้เฉพาะจาก `handlers/kafka/utils.go` (`grep myretry` → 2 ไฟล์)

## 8. Workers (`workers/doc_processor.go`, `workers/db_log_cleaner.go`)
- `NewWorkerManager(0)` → `OptimalWorkerCount(4,32,2.0)` = NumCPU×2 หนีบ 4–32 (`doc_processor.go:48-68`); เครื่อง local ได้ 24 (`docker logs mainapi`: "Worker Manager เริ่มทำงานแล้ว - Workers: 24", log ชี้ `doc_processor.go:120`)
- ทุก worker วน `for { select … default: }` ถ้าไม่มี shop/ไม่มีงาน `time.Sleep(100ms)` แล้ว query `queues` ใหม่ (`doc_processor.go:207-289`) + `monitorActiveShops` ทุก 5s (`:160-163`) → busy-poll 24 goroutine ต่อ DB กลางแม้คิวว่าง; local `SELECT count(*) FROM queues` ใน db `demo` = 0
- Pop ใช้ `UPDATE … WHERE status='pending' … FOR UPDATE SKIP LOCKED` (`mypostgres/queue.go:100-112`); requeue/`AddToDeadLetterQueue` เปลี่ยน status เป็น pending/failed (`queue.go:143-175`)
- งานที่รองรับ: transflag `"6"` → `process.ProcessPurchaseOrderStatus`, `"12"` → `ProcessSaleInvoiceStatus`; `"44"/"48"/"54"` (creditor/customer/debtor) เป็น `return nil // TODO: implement` (`doc_processor.go:357-410`, switch ที่ `:359-371`, TODO ที่ `:394,400,406`; `process/process_status.go:11-56`)
- ผู้ผลิตงานเข้าคิว (ตรวจแล้ว 2026-09-07): `AddToQueue` ถูกเรียกจาก `handlers/kafka/queue_helper.go:57,94` เท่านั้น และ `AddDocToProcessQueue` มีผู้เรียกจุดเดียวคือ consumer purchase order (`handlers/kafka/purchase_order.go:155`, transflag 6); `AddDocToProcessQueueWithPriority` ไม่มีผู้เรียก
- `mypostgres/lock.go` มี distributed lock บนตาราง `distributedlocks` พร้อม fallback in-memory เมื่อไม่มีตาราง (`lock.go:36-130`)
- `DBLogCleaner` (ตั้ง TTL ClickHouse ทุก ≥1 ชม.) ไม่มีผู้เรียก `NewDBLogCleaner` นอกไฟล์ตัวเอง → DEAD (`db_log_cleaner.go:25-117`)

## 9. Connection managers: `mydb` / `mypg` / `mypostgres`
| ชิ้น | หน้าที่ | ค่าคงที่สำคัญ | อ้างอิง |
|---|---|---|---|
| `mydb.ManagerPool` | 1 `DatabaseManager` ต่อ holdingcode; PG database และ ClickHouse database ชื่อ = holdingcode (`manager_pool.go:105,112`); สร้างใหม่เมื่อไม่ healthy (`:77-84`) | PG pool 50/15/1h | `mydb/manager_pool.go:70-122`, `mydb/unified_database_manager.go:169-171` |
| `mydb.CircuitBreaker` | ห่อ `QueryPostgreSQL/ExecPostgreSQL/…ClickHouse` แยก breaker PG/CH | resetTimeout 10s, half-open 3 calls | `mydb/circuit_breaker.go:62-67`, `unified_database_manager.go:70-71,266,323,351,379` |
| `mydb.ConnectionPoolManager` | pool ต่อชื่อ database (มี `PeriodicMaintenance`) | 75/25/30m | `mydb/connection_pool_manager.go:40-42,48,223` |
| `mydb.GetBrandDB` | pool เล็กสำหรับ brand DB | 5/2 | `mydb/brand_db.go:20,50-51` |
| `mypg.PgSqlFastConnect` / `ConnectOptimized` / `Connect` | เส้นทางที่ handler ส่วนใหญ่ใช้ (เช่น inventory) + prepared-statement cache LRU + `CleanupPreparedStatements` | 25/7/10m | `mypg/fast_utils.go:39,133,168-170,341`, `mypg/utils.go:57-59` |
| `myglobal` global DB | connection กลาง (database `postgres`) สำหรับ queue/worker ผ่าน provider closure | — | `bootstrap.go:104-118` |
- ข้อสังเกต: มี pool 3 ชั้นซ้อนกันคนละค่า (50/75/25 max conns) ต่อ holding — ผลรวมต่อ PG server ไม่ได้ควบคุมจากที่เดียว (ยังไม่ตรวจว่า pool ไหนถูกใช้จริงต่อ holding บน .202)

## 10. `process/**` และ ClickHouse
- `process/build` (ripgrep ข้าม dir นี้เพราะ `.ignore` — ต้องใช้ `rg --no-ignore`): สร้างตาราง PG ต่อ holding (`TableProductCreate`…`TableSearchAliasesCreate`, `create-database.go:28-1814`), `DatabaseRebuildAll/DatabaseChecker/DatabaseRebuild` (`:1837,2019,2040`), rebuild จาก Mongo (`ProcessProductRebuildAll`, `ProcessBarcodePostgresRebuildAll`, `ProcessCreditorRebuildAll`, `ProcessDebtorRebuildAll`, `ProcessErpUserRebuildAll`, `DocRebuildAllFromMongo`) และ `RebuildJob` progress (`rebuild_progress.go:40-149`) ซึ่งเสิร์ฟผ่าน SSE handler ที่ DEAD (§5)
- `process/process-doc`: สถานะเอกสารซื้อ (`ProcessDocPurchase*`, `PurchaseStatusByDocNo`) (`process-doc/*.go`); `process/process-stock`: ต้นทุน/ยอดคงเหลือ (`ProductCalcCost*`, `ProcessProductBalanceUpdate*`, incremental checksum, stock validation) (`process-stock/*.go`); `process/report-stock`: รายงานยอดคงเหลือ + `language_dict.go`
- ClickHouse: `ClickHouseFastConnect`/`CreateClickHouseConnection` คืน error "clickhouse is disabled" เสมอ (`myclickhouse/utils.go:234-241`); ไฟล์ที่ยังเรียก `myclickhouse.*` (grep 2026-09-07): `handlers/clickhouse.go`, `handlers/clond-clickhouse-database.go` (ฟังก์ชัน `CloneClickHouseDatabase*` ไม่มีผู้เรียก), `handlers/migrate_currency.go` (DEAD ทั้งสาม), `handlers/purchase_order_close.go:167` (LIVE), `process/build/*` 5 ไฟล์ (`build-barcode-product.go`, `build-creditors.go`, `build-debtos.go`, `build-doc.go`, `create-database.go`), `process/process-stock/*` 2 ไฟล์ (`process-stock-calc-cost.go`, `product-calc-cost.go`), `workers/db_log_cleaner.go:97`, `mydb/unified_database_manager.go`, `setupconfig/loader.go`, และ `bootstrap.go:585` (`CloseClickHouseConnection` ตอน Shutdown); ใน log mainapi ปัจจุบันยังไม่พบข้อความ "clickhouse is disabled" (0 บรรทัด)

## 11. AI: `aichat` / `aiprovider` / `gemini` / `myollama` / `ragflow`
| module | หน้าที่ | สถานะ | อ้างอิง |
|---|---|---|---|
| `handlers/aichat` | `ChatGemini` = คำถาม → `GenerateQueryFromQuestion` (LLM สร้าง SQL) → `ExecuteQuery` บน PG holding → `GenerateAnswerFromData`; `AnalyzeDocument`; per-shop provider config ใน Mongo; result store; complaint | LIVE (§4) | `aichat/handler_gemini.go:14,53,79`, `query_generator.go:23`, `query_executor.go:21,87`, `handler_document.go:44,105` |
| `aiprovider` | abstraction `GetProvider()` (fallback chain), `GetShopProviders(holdingCode)`, Gemini + OpenAI-compatible; env key: `AI_PROVIDER` (`fallback.go:95`), `GEMINI_API_KEY` (`fallback.go:119,232`), `GEMINI_MODEL` (`fallback.go:246`) + key/model ต่อ provider อ่านจาก `c.envKey`/`c.envModel` ใน `buildProvider` (`fallback.go:132-140`) | LIVE (ใช้โดย aichat, `lineoa/chatbot.go:217`, `unified/unified_server.go:77`) | `aiprovider/provider.go:37-59`, `shop_providers.go:191-311`, `fallback.go:92-140` |
| `gemini/client.go` | client REST Gemini ดิบ | ใช้ผ่าน `aiprovider/gemini.go` เท่านั้น | `gemini/client.go:82` |
| `myollama/client.go` | `GenerateEmbeddings` | DEAD — ไม่มี import ที่ไหน | `myollama/client.go:43,91` |
| `ragflow/*` | dataset/document/retrieval client (`RAGFLOW_BASE_URL`, `RAGFLOW_API_KEY`) | LIVE ผ่าน `/api/v1/kb/*` | `ragflow/client.go:46`, `knowledgebase/handler.go:481-490` |
| `handlers/unified` + `cache/unified` | `ProcessUnifiedQuery` + cache manager | LIVE | `bootstrap.go:567-571`, `cache/unified/cache_manager.go` |

## 12. ภาษาไทย / tokenizer / dataimport
- `language/language.go`: โหลด dictionary จากไฟล์, `Normalize/Text/Dictionary/Load` (`language.go:48-141`) → เสิร์ฟผ่าน `GetLanguageHandler` (LIVE ทั้ง `/goapi/api/language/:lang` และ alias root `backend/main.go:580`)
- `mythaitokenizer/client.go` (HTTP client ไป tokenizer service) DEAD — ไม่มี caller; ส่วน `handlers/thai_nlp_handlers.go` เป็นแค่ helper `getThaiNLPURL()` (env `THAI_NLP_URL`, default `http://thai-nlp:8890`) + circuit flag ที่ `ProductSearchHandler` ใช้ (`thai_nlp_handlers.go:26-62`, `handlers/product_search.go:383-387`); ยังไม่ตรวจว่ามี container `thai-nlp` ใน compose หรือไม่
- `dataimport/xlsx_product.go`: `StartProductPrepareHandler` (LIVE) ดาวน์โหลด xlsx → parse แถวสินค้า → เทียบ Mongo (`loadMongoProducts`, `compareWithMongo`) เก็บ session ใน memory; `GetProductPrepareStatus/Result` ไม่มี route (§5) → frontend ดูผลไม่ได้ผ่าน goapi; `dataimport/tokenizer.go` = Trie + Levenshtein ใช้ภายใน

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ
1. **ยังไม่ตรวจ** เนื้อใน `OnConsumeMessage*` (50 ฟังก์ชันใน 22 ไฟล์ จากทั้งหมด 36 ไฟล์ใน `handlers/kafka/`) ว่าตัวไหนเขียน PG สำเร็จจริงหลัง schema drift (lead ยืนยันแล้วว่า debtor/creditor/erp_user พัง) — บทความนี้ยืนยันแค่ topic→handler wiring
2. **ยังไม่ตรวจ** ว่า frontend เรียก route ใดบ้างใน §4 จริง (บาง route เช่น `/image/promptpayverify`, `/api/v1/unified/*`, `/api/aichat/v1/models` อาจไม่มีผู้ใช้) — ต้อง grep `frontend/src` แยก
3. **ยังไม่ตรวจ** ค่า `THAI_NLP_URL`/container `thai-nlp` และ RAGFlow/Gemini บน .202 ว่าตั้งอยู่หรือไม่ (ดูได้แค่ key name ตามกฎ)
4. `manager.go`, `db_log_cleaner.go`, `myollama`, `mythaitokenizer`, `process_consumer.go`, `Test*Consumer`, และ handler DEAD 54 ตัว — คำถามถึงลุงจืด: ลบทิ้งได้เลยไหม (R2 ตามกฎ DB disposable) หรือมีแผนเปิด route กลับ (`bootstrap_test.go` จะต้องแก้ตาม)
5. คำถาม: worker pool (§8) มี producer จริงเพียงจุดเดียวคือ consumer `when-purchaseorder-*` (`handlers/kafka/purchase_order.go:155` → `queue_helper.go:57`) — **ยังไม่ตรวจ** ว่าใน runtime มีแถวเข้าคิวจริงหรือไม่ (local `queues` = 0 ทั้ง db `demo` และ `postgres`); ถ้างาน PO status ย้ายไปทำ inline ได้ ควรปิด `NewWorkerManager` ทั้งก้อนเพื่อหยุด busy-poll
6. Auth ของ goapi พึ่ง `authorizationFinders` (`pst`) ที่ส่งจาก `backend/main.go:574` — ยังไม่ตรวจว่า finder นั้นเช็ค role/permission ต่อ route หรือแค่ตรวจ session+holding (โค้ด middleware เช็คแค่ holding ตรงกัน `bootstrap.go:243-249`)
