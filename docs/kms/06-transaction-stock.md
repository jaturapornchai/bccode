# ธุรกรรมสต็อกและเครื่องคำนวณต้นทุน (Stock transactions & cost engines)
> ตรวจล่าสุด: 2026-09-07 (commit d93a210d) — ผู้เขียน: AI reader; ทุกข้อเท็จจริงอ้าง path:line · ตรวจซ้ำโดย fact-checker

## 1. สรุปสั้น (TL;DR)

- เอกสารสต็อก 7 ชนิด (โอน/รับ/คืน/เบิก/ปรับปรุง/ยอดยกมา/รายละเอียดยอดยกมา) เขียนลง MongoDB ผ่าน HTTP module ใต้ `/transaction/stock-*` (`backend/main.go:432-440`) แล้ว publish Kafka topic `when-<module>-created|updated|deleted` (เช่น `backend/internal/transaction/stockadjustment/config/stockadjustment_messagequeue_config.go:4-9`)
- มี **เครื่องคำนวณต้นทุน 3 ชุด** ในโค้ด แต่ที่ทำงานจริงตอนนี้มีชุดเดียว:
  1. **goapi `process-stock` (LIVE)** — ต้นทุนถัวเฉลี่ยถ่วงน้ำหนักบน PostgreSQL ต่อ holding เขียนตาราง `processstockcost` (`backend/internal/goapi/process/process-stock/process-stock-calc-cost.go:392-456`) ถูกเรียกอัตโนมัติจาก goapi Kafka consumer ทุกครั้งที่มีเอกสาร (`backend/internal/goapi/handlers/kafka/stock_adjustment.go:148`)
  2. **legacy `internal/stockprocess` + `pkg/stockcalculator` (DEAD)** — ทำงานเฉพาะ `DEV_API_MODE` = "" หรือ "1" (`backend/main.go:654,673`) และ SQL อ้างตาราง `stock_transaction` ที่ไม่มีอยู่จริง (ดู §4)
  3. **goapi `inventory/costing` (DORMANT)** — engine movingaverage/fifo/lifo/fefo/standard เขียนตาราง `inventorycostlayers` ฯลฯ ผ่าน `/goapi/api/inventory/*` เท่านั้น ไม่มี consumer/frontend เรียก (ดู §6)
- คิว `stockwaitprocess` ถูก insert ทุกเอกสารแต่ **ไม่มีใคร drain** — ตัว drain (`ProcessStockCostAll`) เรียกได้จาก rebuild ที่ handler ไม่ได้ลงทะเบียน (ดู §5.3); runtime local DB `test` ค้าง 8 แถว

## 2. โหมดรันและจุดลงทะเบียน (backend/main.go)

| gate | บรรทัด | สิ่งที่ลงทะเบียนฝั่งสต็อก |
|---|---|---|
| `devApiMode == "" \|\| "2"` (HTTP + goapi) | `backend/main.go:256` | `warehouse.NewWarehouseHttp` (:401), stock HTTP 7 module (:432-440), `stockbalanceimport.NewStockBalanceImportHttp` (:520), goapi routes `/goapi/*` (:574) |
| `devApiMode == "3"` (migration) | `backend/main.go:585` | AutoMigrate ตาราง PG ของ stock consumer (:618-622), `warehouse.MigrationDatabase` (:638), `stockbalance_consumer.MigrationDatabase` (:640) |
| `devApiMode == "" \|\| "1"` (legacy Kafka consumer) | `backend/main.go:654` | `stockprocess.NewStockProcessConsumer` (:673), stock consumer 6 ตัว (:676-681), `warehouse.InitWarehouseConsumer` (:703) |

ค่าจริงที่ใช้: local container `mainapi` มี `DEV_API_MODE=2` (docker inspect 2026-09-07; `backend/docker-compose.yml:80`, `backend/Dockerfile:41`) → legacy consumer **ไม่รัน** บนเครื่องนี้; prod compose มี service `worker` ที่ตั้ง `DEV_API_MODE: "1"` (`deploy/account/compose.yml:272-276`) และ `migrate` ตั้ง "3" (`deploy/account/compose.yml:209-212`) — runtime prod ยังไม่ตรวจ

## 3. Inventory ของ module (Mongo = source of truth)

| module | หน้าที่ | entry/route | Mongo collection | สถานะ | อ้างอิง |
|---|---|---|---|---|---|
| stocktransfer | ใบโอนสินค้า (transflag 72) | `/transaction/stock-transfer[...]` | `transactionstocktransfer` | LIVE (route) / ข้อมูล 0 doc | `backend/internal/transaction/stocktransfer/stocktransfer_http.go:50-57`, `models/stocktransfer.go:10` |
| stockreceiveproduct | ใบรับสินค้าเข้าคลัง (60) | `/transaction/stock-receive-product[...]` + `/bulk` | `transactionstockreceiveproduct` | LIVE / 0 doc | `stockreceiveproduct_http.go:47-56`, `models/stockreceiveproduct.go:10` |
| stockreturnproduct | ใบคืนสินค้าเข้าคลัง (58) | `/transaction/stock-return-product[...]` | `transactionstockreturnproduct` | LIVE / 0 doc | `stockreturnproduct_http.go:50-59`, `models/stockreturnproduct.go:10` |
| stockpickupproduct | ใบเบิกสินค้า (56) — path สะกด `stock-prickup-product` | `/transaction/stock-prickup-product[...]` | `transactionstockpickupproduct` | LIVE / 0 doc | `stockpickupproduct_http.go:50-59`, `models/stockpickupproduct.go:10` |
| stockadjustment | ปรับปรุงสต็อก (66 เพิ่ม / 68 ลด) | `/transaction/stock-adjustment[...]` | `transactionstockadjustment` | LIVE / 0 doc | `stockadjustment_http.go:50-59`, `models/stockadjustment.go:10` |
| stockbalance | ยอดยกมา (54) | `/transaction/stock-balance[...]` | `transactionstockbalance` | LIVE / 0 doc | `stockbalance_http.go:65-74`, `models/stockbalance.go:10` |
| stockbalancedetail | รายการย่อยยอดยกมา | `/transaction/stock-balance-detail[...]` | `transactionstockbalancedetails` | LIVE / 0 doc | `stockbalancedetail_http.go:49-56`, `models/stockbalancedetail_detail.go:10` |
| stockbalanceimport | นำเข้า Excel ยอดยกมา | `/stockbalanceimport[...]` | อ่าน/เขียนผ่าน **ClickHouse** persister | BROKEN-ON-USE (ดู §8) | `backend/internal/stockbalanceimport/stockbalanceimport_http.go:39,74-82` |
| warehouse | คลัง/ที่เก็บ/ชั้นวาง | `/warehouse[/:id]`, `/warehouse/tree`, `/warehouse/:warehouseguid/location[/:locationguid]`, `.../location/:locationguid/bin[/:binguid]` | `warehouse` 24, `warehouselocation` 23, `warehousebin` 11 doc | LIVE (frontend ใช้จริง) | `backend/internal/warehouse/warehouse_http.go:66-85`, `models/warehouse.go:13`, `location.go:9`, `bin.go:9` |

หลักฐาน runtime (mongosh `appdb` 2026-09-07): collection ธุรกรรมสต็อกทั้ง 7 ตัว `countDocuments` = 0; warehouse 24/23/11. Frontend มีเฉพาะจอ warehouse (`frontend/src/app/system-settings/warehouse-tree-view.tsx:506,704,757,820`) ส่วนเมนู dashboard แค่ยิง list เพื่อทำ badge (`frontend/src/app/menu/dashboard-home.tsx:39-40`; log mainapi มีบรรทัด `stock-transfer/list`/`stock-adjustment/list` ซ้ำ 168 ครั้ง ล่าสุด 2026-09-06T23:46Z — คาบเวลาการยิง ยังไม่ตรวจ ไม่พบ `setInterval` ใน `dashboard-home.tsx`) — **ยังไม่มีจอสร้าง/แก้เอกสารสต็อก** (`git ls-files frontend/src/app | grep -i stock` เจอเฉพาะ `menu/tab-product-stock.tsx` ซึ่งเป็น tab ในฟอร์มสินค้า ไม่เรียก API สต็อก)

หมายเหตุ scope: `internal/transaction/documentwarehouse` และ `internal/transaction/warehouse` **ไม่มีในรีโป** (`git ls-files` ว่าง); ที่มีคือ `internal/documentwarehouse/documentimage` = อัปโหลดรูปเอกสาร (`backend/internal/documentwarehouse/documentimage/documentimage_http.go:71-75`) ไม่เกี่ยวกับสต็อก

## 4. เส้นทาง A — legacy consumer → PG `stocktransaction` → `stockprocess` (DEAD)

ลำดับตามโค้ด (ตัวอย่าง stockadjustment; module อื่นโครงเดียวกัน):
1. consumer ฟัง topic `when-stockadjustment-*` ใน group `TRANSACTION_CONSUMER_GROUP` (default `transaction-consumer-group-01`) — `backend/internal/transaction/transactionconsumer/stockadjustment/stock_adjustment_transaction_consumer.go:181-196`
2. `ConsumeOnCreateOrUpdate` → upsert ตารางเอกสาร แล้วแปลงเป็น `StockTransaction` (`CalcFlag` = -1 เมื่อ transflag 68) — `stock_adjustment_transaction_consumer.go:68-94`, `stock_adjustment_stock_phaser.go:10-20`
3. `StockTransactionConsumerService.Upsert` เขียน GORM model `StockTransaction`/`StockTransactionDetail` (ชื่อตาราง `stocktransaction`, `stocktransactiondetail` — `backend/internal/transaction/models/stock_transaction_postgres.go:48-49,88-89`) แล้ว publish `StockProcessRequest{holdingcode,barcode}` ไป topic `when-stock-process-created` — `backend/internal/transaction/transactionconsumer/stocktransaction/stock_transaction_consumer_service.go:39-81`, `backend/internal/stockprocess/config/queue_config.go:4-5`
4. `StockProcessConsumer` (group `CONSUMER_GROUP_ID` default `consumer-stockprocess-group-01`) เรียก `CalculatorStock` — `backend/internal/stockprocess/stockprocess_consumer.go:45-54,66`
5. `CalculatorStock` อ่าน movement ด้วย SQL `FROM stock_transaction ... JOIN stock_transaction_detail` — `backend/internal/stockprocess/repositories/stockprocess_pg_repository.go:29-47` — แล้วใช้ `pkg/stockcalculator` คำนวณ **moving average** ทศนิยม 2 ตำแหน่ง hard-code (`backend/internal/stockprocess/stockcalculator.go:49`) และ update กลับ `stock_transaction_detail` + `productbarcode.balanceqty/balanceamount/averagecost` (`stockcalculator.go:146-165`, `stockprocess_pg_repository.go:68-70`)

ทำไมถือว่า DEAD:
- **Schema drift**: AutoMigrate สร้าง `stocktransaction`/`stocktransactiondetail` (`backend/internal/transaction/transactionconsumer/migration.go:9-13`) แต่ SQL ของ stockprocess อ้าง `stock_transaction`/`stock_transaction_detail` (`stockprocess_pg_repository.go:43-44,68,85`) → แม้เปิดโหมด 1 ก็ query ไม่เจอตาราง; ไม่มีไฟล์อื่นในรีโปอ้างชื่อ `stock_transaction` (rg 2026-09-07)
- runtime local: ไม่มีตาราง `stocktransaction` หรือ `stock_transaction` ในทุก DB (appdb, bc001, bctest01, demo, postgres, test, uat260810a, qa23995213 — `to_regclass` เป็น null); Kafka ไม่มี topic `when-stock-process-created` และไม่มี consumer group `transaction-consumer-group-01`/`consumer-stockprocess-group-01` (`kafka-topics --list`, `kafka-consumer-groups --list` 2026-09-07)
- ผู้เรียก `NewStockProcessMessageQueueRepository` มีอีกที่คือ `backend/internal/systemadmin/productadmin/productadmin_service.go:31` (ยังไม่ตรวจว่า endpoint ไหนเรียกและถูกลงทะเบียนหรือไม่)

สูตรใน `pkg/stockcalculator` (ใช้ได้ทั่วไป ทดสอบใน `backend/pkg/stockcalculator/stockcaculator_test.go`): `ApplyStock` เพิ่ม qty/amount แล้ว average = amount/qty (`backend/pkg/stockcalculator/stockcalculator.go:51-72`); `ReduceStock` ตัดด้วย average ปัจจุบัน และล้าง balanceAmount เป็น 0 เมื่อ qty เหลือ 0 (`:98-117`); `ApplyCost`/`ReduceCost` ปรับเฉพาะมูลค่า (transflag 866/868 — `backend/internal/stockprocess/stockcalculator.go:76-78,115-117`); `AverageCostCalc` ใช้ค่าสัมบูรณ์ทั้งเศษและส่วน (`stockcalculator.go:160-178`)

## 5. เส้นทาง B — goapi Kafka consumer → `docdetail` → `processstockcost` (LIVE)

### 5.1 การเปิดใช้และหลักฐานว่ารันอยู่
- goapi เริ่ม consumer เมื่อ `ENABLE_KAFKA == "true"` (`backend/internal/goapi/bootstrap.go:157-160`); ค่านี้ map จากคีย์ `service.enablekafka` ใน bootstrap.json (`backend/internal/goapi/setupconfig/loader.go:78`) ไม่ใช่ env ของ container (docker inspect ไม่มี `ENABLE_KAFKA`)
- log mainapi 2026-09-05 23:27:40: `bootstrap.go:159 เริ่มต้น Kafka consumers` และ `kafka.go:1008 เริ่มต้น Kafka consumers เรียบร้อย`; Kafka มี consumer group `goapi-stocktransfer|stockreceiveproduct|stockpickupproduct|stockreturnproduct|stockadjustment|stockbalance-consumer-v1` (suffix จาก `getVersionedGroupID` `backend/internal/goapi/handlers/kafka.go:19-22`)
- topic ที่ฟัง: `when-stocktransfer-*` (:599-623), `when-stockreceiveproduct-*` (:637-661), `when-stockpickupproduct-*` (:675-699), `when-stockreturnproduct-*` (:713-737), `when-stockadjustment-*` (:751-775), `when-stockbalance-*` (:789-813), `when-warehouse-*` (:230-255) — ทั้งหมดใน `backend/internal/goapi/handlers/kafka.go`; bridge อยู่ `backend/internal/goapi/handlers/kafka_bridge.go:412-475`

### 5.2 ขั้นตอนต่อเอกสาร (ตัวอย่าง `ProcessStockAdjustmentDocument`)
1. ลบเอกสารเดิมด้วย `mypg.DeleteDocPgSql` แล้ว insert `doc`/`docdetail` (ClickHouse step เป็น no-op: `backend/internal/goapi/handlers/kafka/utils.go:375-378`) — `backend/internal/goapi/handlers/kafka/stock_adjustment.go:112,131-146`, transfer: `stock_transfer.go:72-101`
2. `ProcessDocumentStockCalculation` → incremental=true, minimalLog=true (`kafka/utils.go:161-164`) → ต่อ itemcode เรียก `processstock.ProductCalcCostIncrementalCompany` ด้วยทศนิยมจาก `myglobal.ConfigSystem.Stock{Qty,Amount,Cost}Point` (`kafka/utils.go:209-212,227-242`); error ถูก log แล้วเดินต่อ (`stock_adjustment.go:148-152`)
3. หลังคำนวณ → `ProcessProductBalanceUpdateByItemsAsync` อัปเดตยอดคงเหลือ/ค้างรับ/ค้างส่งในตาราง product แบบ async (`kafka/utils.go:266-271`, `backend/internal/goapi/process/process-stock/process-product-balance-update.go:171`)
4. `mypg.AddToDocWaitProcessQueues` insert `docwaitprocess` + `stockwaitprocess` (`stock_adjustment.go:156`, `backend/internal/goapi/mypg/doc.go:47-70`; ผู้เรียก 16 จุดรวมซื้อ/ขาย เช่น `kafka/purchase.go:134`, `kafka/sale_invoice.go:127`)
transflag ต่อชนิด: 72 โอน, 60 รับ, 56 เบิก, 58 คืน, 66/68 ปรับเพิ่ม/ลด, 54 ยอดยกมา (`backend/internal/goapi/handlers/kafka/constants.go:117-123`)

### 5.3 ตัวคำนวณ `productCalcCostWithOptions` (`process-stock-calc-cost.go`)
- lock ต่อ `stock:calc:<holding>:<business>:<item>` ผ่าน PG distributed lock 5 นาที (`:86-100`); incremental ใช้ checksum จาก `docdetail` เก็บใน `stockcalculationstate_company` (`incremental.go:38-60,302-304`); ถ้าไม่เปลี่ยนข้าม (`process-stock-calc-cost.go:105-119`)
- ลบ `processstockcost`/`processstocklot` ของ item ก่อนเสมอ (delete-then-insert = idempotency; UPSERT ที่ minimalLog สัญญาไว้ไม่เคยทำ — comment `:41-46,123-141`)
- อ่าน `docdetail` เฉพาะ `transflag IN myglobal.TransFlagsToProcess` = {54,12,310,48,60,58,66,44,16,56,68,72} (`backend/internal/goapi/myglobal/global.go:32`, `process-stock-calc-cost.go:143-148,323-328`) เรียงตาม `docdatetime, linenumber, calcflag DESC, docno`
- **สูตร = ต้นทุนถัวเฉลี่ยถ่วงน้ำหนักแบบต่อเนื่อง**: ออก (44,56,72,20,21,62) ตัดด้วย average ปัจจุบัน; เข้า (12,310,60,61,54,66,48) ใช้ `sumamount` (fallback qty×priceexcludevat); 16 รับคืนใช้ sumamount ไม่งั้น average; 866/868 ปรับเฉพาะมูลค่า; เมื่อ balanceQty ≈ 0 บังคับ balanceAmount = 0; average ใหม่ = balanceAmount/balanceQty เฉพาะเมื่อ balanceQty > 0 (`:376-456`); **58 (คืนเข้าคลัง) และ 68 (ปรับลด) ไม่มี case เฉพาะ** → ตกไป `default` ใช้ `calcAmount = sumamount` และ qty ตามเครื่องหมายใน `docdetail.totalqty` ตรง ๆ (`:424-427`, qty มาจาก `:373-379` ไม่ได้คูณ calcflag)
- **FIFO/lot ไม่ทำงาน**: บล็อก `addLotFunc`/`processLotFunc` (มี `calcType == 1` FIFO) ถูก comment ทั้งก้อน `:163-219` → slice `lots` ว่างเสมอ, ตาราง `processstocklot` จึงไม่มีข้อมูล (local `test`: `processstocklot` = 0 ขณะ `processstockcost` = 9)
- ผลลัพธ์เขียน `processstockcost` ด้วย COPY (`:292`; คอลัมน์ `businesscode, docdatetime, docno, ..., averagecost, calcamount, balanceqty, balanceamount, unitcost, guid` — `backend/internal/goapi/process/build/create-database.go:705-731`); ส่วน ClickHouse ยิงเป็น goroutine แล้ว fail เงียบเพราะ `ClickHouseFastConnect` คืน `clickhouse is disabled` (`process-stock-calc-cost.go:574-586`, `backend/internal/goapi/myclickhouse/utils.go:234-236`)

### 5.4 คิว `stockwaitprocess` — ใครเติม / ใคร drain
- เติม: ทุกเอกสารผ่าน `AddToDocWaitProcessQueues` (§5.2 ข้อ 4), rebuild (`backend/internal/goapi/process/build/build-doc.go:480-490`), `TruncateProcessTables` (`backend/internal/goapi/process/utils.go:19-35`)
- drain: **เฉพาะ** `ProcessStockCostAllWithCallback` — `SELECT DISTINCT itemcode FROM stockwaitprocess` แล้วลบทีละ batch (`product-calc-cost.go:67,280-300`) — ผู้เรียกมีแค่ `build.DatabaseRebuild` (`create-database.go:2065`) และ `CalcStockCostAll` (:2370) ผ่าน wrapper `ProcessStockCostAll` (`product-calc-cost.go:49-51`), `DatabaseRebuildWithProgress` (:2523), `CalcStockCostAllWithProgress` (:2546) และ comment ใน `handlers/process_consumer.go:107`
- ฟังก์ชัน build เหล่านั้นถูกเรียกจาก `ReportPostHandler` command `rebuild` (`backend/internal/goapi/handlers/commands.go:154,197-226`) ซึ่ง **ไม่ได้ลงทะเบียน route ใน `bootstrap.go`** (rg `ReportPostHandler` เจอเฉพาะนิยาม) → ไม่มีทาง HTTP/consumer ใด drain คิวนี้ (ตรงกับ `docs/handoff/HANDOFF-2026-09-06.md:59`)
- runtime local `test`: `stockwaitprocess` 8 แถว, `docwaitprocess` 8, `docdetail` 9 → คิวโตไปเรื่อยโดยไม่มีผลต่อความถูกต้องของ `processstockcost` (เพราะ consumer คำนวณทันทีอยู่แล้ว) แต่เป็นขยะสะสม

### 5.5 route goapi ที่เกี่ยวกับต้นทุน (ทั้งหมด LIVE แต่ไม่มี frontend เรียก — rg `frontend/src` 2026-09-07 ไม่พบ)
| route | หน้าที่ | อ้างอิง |
|---|---|---|
| `POST /goapi/processstockcalccost` | คำนวณต้นทุนรายการที่ระบุ (holdingcode จาก body, incremental เลือกได้) | `backend/internal/goapi/bootstrap.go:389`, `handlers/process_stock_calc_cost.go:43-119,195-230` |
| `POST /goapi/api/stockcost/query|summary|check` | อ่าน `public.processstockcost` | `bootstrap.go:390-392`, `handlers/process_stock_cost.go:37,156,264,316,355,394` |
| `POST /goapi/api/process/product-balance` | `ProcessProductBalanceUpdate` ทั้ง holding | `bootstrap.go:411`, `handlers/product_balance_update.go:29,40` |
| `POST /goapi/api/health/queue[/:holdingcode]` | สถานะคิว worker pool | `bootstrap.go:379-380` |
worker pool ที่ busy-poll ตาราง `queues` ทุก 100 ms (`backend/internal/goapi/workers/doc_processor.go:232-245`, สร้างที่ `bootstrap.go:136`) **ไม่ได้** เกี่ยวกับ `stockwaitprocess` (คนละคิว — worker เรียก `QueueManager.GetActiveShops` ที่ query ตาราง `queues` `backend/internal/goapi/mypostgres/init_schema.go:110,133`; runtime local: `queues` มี 0 แถวใน bc001/bctest01/demo/postgres/test และไม่มีตารางนี้ใน appdb/qa23995213/uat260810a)

## 6. goapi `inventory/costing` — engine หลายวิธี (DORMANT)

- routes ใต้ `/goapi/api`: `GET|PUT /products/:itemcode/costing-config`, `GET /products/:itemcode/cost-layers`, `POST /inventory/receipt|issue|transfer|adjustment|sales-return|purchase-return`, `GET /reports/inventory-valuation`, `GET /reports/stock-card/:itemcode` (`backend/internal/goapi/inventory/handler.go:15-34`, mount ที่ `bootstrap.go:384-386`); holdingcode รับจาก query param (`handler.go:49,68,206`) ต่อ DB ด้วย `PgSqlFastConnect(holdingCode)` (`handler.go:39`)
- engine เลือกตาม `productcostingconfig.costingmethod` default `movingaverage` (`backend/internal/goapi/inventory/service.go:352-366`); ที่ implement: `MovingAverageEngine` (สูตร new_avg = (old_qty×old_avg + qty×cost)/(old_qty+qty) — `costing/moving_average.go:11-33`), `FIFOEngine`, `LIFOEngine`, `FEFOEngine`, `StandardEngine` (`costing/engine.go:37-53`); ค่าคงที่ `periodicaverage`/`lot` มีชื่อแต่ไม่มี engine (`inventory/models/models.go:10-15` เทียบ `engine.go:37-53`)
- ตาราง: `productcostingconfig`, `inventorycostlayers`, `inventorystockbalances`, `marketplacestockbalances`, `marketplacedimensionprices`, `inventorycosttransactions`, `inventoryvariances`, `inventorylandedcosts`, `inventorylandedcostallocations`, `inventoryaccountingperiods` (`backend/internal/goapi/inventory/database.go:52-236`) สร้างโดย `CreateInventoryCostingTables` ซึ่งถูกเรียกจาก `process/build/create-database.go` (import :13) และ handler `create-tables` ที่ **ไม่ได้ผูก route** (`CreateTablesHandler` `handler.go:277-288` ไม่มีใน `RegisterRoutes`; ตั้งใจถอด — `TestGoAPIRouteSurfaceExcludesOperationalEndpoints` ใส่ `POST /goapi/api/inventory/create-tables` ในรายการ forbidden และ fail ถ้ายัง register `bootstrap_test.go:83,101-103`)
- ไม่มีใครเรียกจาก Kafka consumer/`process-stock` (rg `goapi/inventory` เจอเฉพาะ `bootstrap.go:23` และ `build/create-database.go:13`); runtime local: `inventorycostlayers`/`inventorystockbalances`/`productcostingconfig` = 0 แถวใน bc001, demo, test → ระบบต้นทุนคู่ขนานที่ยังไม่ถูกใช้ และ **ไม่ sync** กับ `processstockcost`

## 7. Warehouse projection (สองทางเขียน PG คนละตาราง)

- legacy consumer (โหมด 1): `warehouse.InitWarehouseConsumer` → GORM `WarehousePg` ตาราง `warehouse` (`backend/main.go:703`, `backend/internal/warehouse/models/warehouse_pg.go:23-24`)
- goapi consumer (LIVE): `when-warehouse-*` → ลบ/insert `ic_warehouse` + `ic_shelf` (`backend/internal/goapi/handlers/kafka/warehouse.go:82-142`, group `biapi-warehouse-consumer-v1`)
- runtime local `test`: `ic_warehouse` 3, `warehouse` 7 แถว; `demo`: `ic_warehouse` 1 — ตาราง `warehouse` ใน `test` มาจากไหน ยังไม่ตรวจ (โหมด 1 ไม่รันบนเครื่องนี้ อาจเป็น migration/rebuild เก่า)

## 8. stockbalanceimport พึ่ง ClickHouse (BROKEN-ON-USE)

- `NewStockBalanceImportHttp` ขอ `ms.ClickHousePersister(cfg.ClickHouseConfig())` (`backend/internal/stockbalanceimport/stockbalanceimport_http.go:39`) และ repository ทั้งหมด (`All/Meta/List/Create/Update`) ใช้ `IPersisterClickHouse` (`repositories/stockbalanceimport_clickhouse_repository.go:28,47-216`)
- constructor เรียก `getClient()` → `clickhouse.Open` และ **panic ถ้า Open คืน error** (`backend/pkg/microservice/persister_clickhouse.go:39-42,52-64`); container `mainapi` ที่ลงทะเบียน route นี้ยังรันอยู่ (DEV_API_MODE=2) จึงสรุปได้ว่า Open ไม่ล้มตอน start — พฤติกรรม lazy-dial ของไลบรารี clickhouse-go ยังไม่ตรวจในรีโป; ClickHouse container ถูกถอด 2026-09-06 → คาดว่า `/stockbalanceimport/*` ทุก route ล้มเมื่อเรียกจริง (**ยังไม่ทดสอบ**); `SaveTask` ปลายทางเขียน stockbalance/stockbalancedetail ผ่าน service เดียวกับ §3 (`stockbalanceimport_http.go:41-60`, `services/stockbalanceimport_service.go:252`)

## 9. ตารางสรุปสถานะ

| ส่วน | สถานะ | เหตุผลหลัก |
|---|---|---|
| HTTP `/transaction/stock-*` + Mongo | LIVE แต่ไม่มีข้อมูล/ไม่มีจอ | route ลงทะเบียน `main.go:432-440`; Mongo 0 doc; frontend ยิงแค่ list |
| goapi Kafka consumer + `processstockcost` | LIVE | consumer group `-v1` มีจริง; `test.processstockcost` = 9 |
| `stockwaitprocess` | LIVE-เติม / DEAD-drain | drainer เข้าถึงได้เฉพาะ `ReportPostHandler` ที่ไม่ได้ register |
| FIFO lot (`processstocklot`) | DEAD (comment out) | `process-stock-calc-cost.go:163-219` |
| legacy `stockprocess` + `pkg/stockcalculator` | DEAD | โหมด 1 ไม่รัน local; ชื่อตาราง SQL ไม่ตรง AutoMigrate |
| goapi `inventory/costing` | DORMANT | route มี ไม่มีผู้เรียก ตาราง 0 แถว |
| ClickHouse ในเส้นทางสต็อก | STUBBED | `myclickhouse/utils.go:234-236`, `kafka/utils.go:375-378` |
| stockbalanceimport | BROKEN-ON-USE | พึ่ง ClickHouse persister |

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ

1. **ยังไม่ทดสอบ end-to-end บนเครื่องนี้**: POST เอกสารสต็อกจริง → ดู `processstockcost` เปลี่ยน (Mongo ว่างทั้ง 7 collection; แถว 9 ใน `test.docdetail` = transflag 12 ×5 + 44 ×4 คือซื้อ/ขายล้วน (`constants.go:108,111`) — `processstockcost` 9 แถวบนเครื่องนี้จึง**ยังไม่เคย**ผ่านเอกสารสต็อกเลย)
2. prod (`worker` โหมด 1 ตาม `deploy/account/compose.yml:272-276`) จะเจอ error `relation "stock_transaction" does not exist` จาก `stockprocess` จริงหรือไม่ — ยังไม่เห็น log prod
3. `systemadmin/productadmin` ที่ publish `when-stock-process-created` (`productadmin_service.go:31`) ถูกลงทะเบียน route ไหม ใครใช้ — ยังไม่ตรวจ
4. transflag 20/21/61/62/310/866/868 ในสูตร §5.3 มาจากเอกสารชนิดใดในระบบนี้ (มีเฉพาะ 12/44/54/56/58/60/66/68/72 ใน `TransFlagsToProcess` ที่ทับซ้อน) — ไม่พบนิยามใน `constants.go:117-123`; ต้องถามลุงจืด
5. ทศนิยม `StockQtyPoint/StockAmountPoint/StockCostPoint` ค่าจริงต่อ holding และผลต่อ NUMERIC(18,2) ของ `averagecost` — ยังไม่ตรวจ
6. `report-stock` package (`backend/internal/goapi/process/report-stock/*`) และคำสั่ง `reportproductstockmovement` ฯลฯ ใน `commands.go:333-473` เข้าถึงได้จาก route ไหน (ถ้าอิง `ReportPostHandler` ก็ตายไปด้วย) — ยังไม่ตรวจ
7. (ปิดแล้ว — `bootstrap_test.go:83,101-103` ยืนยันว่า `create-tables` ต้อง**ไม่**ถูก register; ดู §6)
8. คำถามถึงลุงจืด: (ก) จะเก็บ engine ไหนเป็นทางการ — `process-stock` (average) หรือ `inventory/costing` (เลือกวิธีได้)? (ข) ต้องการ FIFO จริงไหม ถ้าไม่ ควรลบ `processstocklot` + โค้ด lot ที่ comment ไว้; (ค) legacy `stockprocess` + `pkg/stockcalculator` + consumer โหมด 1 ฝั่งสต็อก ลบได้หรือไม่; (ง) `stockwaitprocess`/`docwaitprocess` จะเลิก insert หรือทำ drainer จริง
