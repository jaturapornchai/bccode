# ธุรกรรมสต็อกและเครื่องคำนวณต้นทุน (Stock transactions & cost engines)
> ตรวจล่าสุด: 2026-09-15 — §1, §5, §9 เขียนใหม่หลังเปลี่ยนมาใช้ `stockengine` + `stock_ledger` และลบเครื่องคิดต้นทุนเดิมทิ้ง
> (เดิมตรวจ 2026-09-07 commit d93a210d) — ทุกข้อเท็จจริงอ้าง path:line

## 1. สรุปสั้น (TL;DR)

- เอกสารสต็อก 7 ชนิด (โอน/รับ/คืน/เบิก/ปรับปรุง/ยอดยกมา/รายละเอียดยอดยกมา) เขียนลง MongoDB ผ่าน HTTP module ใต้ `/transaction/stock-*` (`backend/main.go:432-440`) แล้ว publish Kafka topic `when-<module>-created|updated|deleted` (เช่น `backend/internal/transaction/stockadjustment/config/stockadjustment_messagequeue_config.go:4-9`)
- มี **เครื่องคำนวณต้นทุน 2 ชุด** ในโค้ด ทำงานจริงชุดเดียว:
  1. **goapi `stockengine` (LIVE)** — ต้นทุนถัวเฉลี่ยถ่วงน้ำหนักแยกตามคลัง เขียนตาราง `stock_ledger` + `stock_period_balance` (`backend/internal/goapi/process/stockengine/calc.go`, `store.go`) ทำงานเบื้องหลังผ่านคิว `stock_dirty` ไม่คิดสดในตัวรับเอกสาร (ดู §5 และ ADR `docs/kms/decisions/2026-09-15-stock-cost-engine-v2.md`)
  2. **goapi `inventory/costing` (DORMANT)** — engine movingaverage/fifo/lifo/fefo/standard เขียนตาราง `inventorycostlayers` ฯลฯ ผ่าน `/goapi/api/inventory/*` เท่านั้น ไม่มี consumer/frontend เรียก (ดู §6)
  - เครื่องคิดต้นทุนเดิม (`process-stock-calc-cost.go`, `processstockcost`, `processstocklot`, `stockcalculationstate*`, `distributedlocks`, สำเนาบน ClickHouse) **ถูกลบออกจากรีโปแล้ว 2026-09-15** ส่วน legacy `internal/stockprocess` + `pkg/stockcalculator` ยังอยู่ในรีโปแต่ตายอยู่ (ดู §4)
- คิว `stockwaitprocess`/`docwaitprocess` ยังถูก insert อยู่แต่ **ไม่มีใครใช้แล้ว** — คิวจริงของการคิดต้นทุนคือ `stock_dirty` (ดู §5.4)

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

## 5. เส้นทาง B — goapi Kafka consumer → `docdetail` → คิว `stock_dirty` → `stock_ledger` (LIVE)

### 5.1 การเปิดใช้และหลักฐานว่ารันอยู่
- goapi เริ่ม consumer เมื่อ `ENABLE_KAFKA == "true"` (`backend/internal/goapi/bootstrap.go:157-160`); ค่านี้ map จากคีย์ `service.enablekafka` ใน bootstrap.json (`backend/internal/goapi/setupconfig/loader.go:77`) ไม่ใช่ env ของ container (docker inspect ไม่มี `ENABLE_KAFKA`)
- log mainapi 2026-09-05 23:27:40: `bootstrap.go:159 เริ่มต้น Kafka consumers` และ `kafka.go:1008 เริ่มต้น Kafka consumers เรียบร้อย`; Kafka มี consumer group `goapi-stocktransfer|stockreceiveproduct|stockpickupproduct|stockreturnproduct|stockadjustment|stockbalance-consumer-v1` (suffix จาก `getVersionedGroupID` `backend/internal/goapi/handlers/kafka.go:19-22`)
- topic ที่ฟัง: `when-stocktransfer-*` (:599-623), `when-stockreceiveproduct-*` (:637-661), `when-stockpickupproduct-*` (:675-699), `when-stockreturnproduct-*` (:713-737), `when-stockadjustment-*` (:751-775), `when-stockbalance-*` (:789-813), `when-warehouse-*` (:230-255) — ทั้งหมดใน `backend/internal/goapi/handlers/kafka.go`; bridge อยู่ `backend/internal/goapi/handlers/kafka_bridge.go:412-475`

### 5.2 ขั้นตอนต่อเอกสาร (ตัวอย่าง `ProcessStockAdjustmentDocument`)
1. ลบเอกสารเดิมด้วย `mypg.DeleteDocPgSql` แล้ว insert `doc`/`docdetail` (ClickHouse step เป็น no-op) — `backend/internal/goapi/handlers/kafka/stock_adjustment.go`, transfer: `stock_transfer.go`
   `DeleteDocPgSql` สั่งคิดต้นทุนใหม่ด้วย **วันที่เดิม** ของเอกสารก่อนลบเสมอ (`backend/internal/goapi/mypg/doc.go`, `stockengine.MarkDocumentDirty`) เอกสารที่แก้แล้วย้ายวันข้ามงวดจึงไม่ทิ้งยอดค้างในงวดเก่า
2. `ProcessDocumentStockCalculation` → `EnqueueStockRecalculation` ฝากงานเข้าคิว `stock_dirty` หนึ่งแถวต่อสินค้าหนึ่งตัว พร้อมวันที่เก่าสุดที่ถูกกระทบ แล้วส่งสัญญาณ `pg_notify('stock_dirty', …)` ปลุก worker (`backend/internal/goapi/handlers/kafka/utils.go`, `stock_engine_hook.go`, `stockengine/dirty.go`) — บรรทัดที่ไม่กระทบสต็อกถูกข้าม ไม่ทำให้ทั้งใบล้มเหลว
3. worker ของ holding นั้นจอง งานแล้วคำนวณ (`stockengine/worker.go` → `Recalculate`) เสร็จแล้วเรียก `AfterRecalculate` เพื่ออัปเดตยอดคงเหลือในตาราง product (`backend/internal/goapi/bootstrap.go` ผูกไว้ตอนเริ่มระบบ) — ยอดบนจอสินค้าจึงมาจากสมุดสต็อกชุดเดียวกับรายงาน
4. เอกสารที่ถูกลบ (`DeleteDocumentFromDatabases` ตั้ง `doc.isdelete = true`) ก็สั่งคิดใหม่เช่นกัน และ `LoadMovements` กรอง `iscancel`/`isdelete` ออก
transflag ต่อชนิด: 72 โอน, 60 รับ, 56 เบิก, 58 คืน, 66/68 ปรับเพิ่ม/ลด, 54 ยอดยกมา (`backend/internal/goapi/handlers/kafka/constants.go`)

### 5.3 ตัวคำนวณ `stockengine` (`process/stockengine/`)
- **ลำดับการคิดคงที่เสมอ**: `docdatetime, behindindex, docno, linenumber` ทั้งใน SQL และในหน่วยความจำ (`store.go` `calcOrderBy`, `calc.go` `SortMovements`) คิดซ้ำกี่ครั้งก็ได้ผลเดิม
- **ขอบเขตงาน**: คิดใหม่ตั้งแต่ต้นงวดของวันที่ที่ถูกกระทบ โดยอ่านยอดยกมาจาก `stock_period_balance` งวดล่าสุดที่ `periodkey <` งวดปัจจุบัน (`LoadOpening`); ถ้าไม่มียอดยกมาเลยแต่สินค้ามีประวัติเก่ากว่านั้น จะถอยไปคิดตั้งแต่เอกสารใบแรก (`earliestMovement`)
- **ทิศทาง**: `transFlagDirection` กำหนดทิศทางต่อประเภทเอกสาร ยกเว้นใบโอนคลัง (72) ที่อ่านจาก `calcflag` ของบรรทัด เพราะใบเดียวมีสองขา ขาเข้าปลายทางรับของด้วยต้นทุนของขาออก (`DirectionOfMovement`, `transferValue`)
- **สูตร = ถัวเฉลี่ยถ่วงน้ำหนักต่อคลัง**: ขาออกตัดด้วย average ปัจจุบันของคลังนั้น ขาเข้าใช้ `sumamount` (fallback = จำนวนตามหน่วยนับในเอกสาร × ราคา) — ราคาในเอกสารเป็นราคาต่อหน่วยนับในเอกสาร ไม่ใช่ต่อหน่วยฐาน; ยอดที่เกือบศูนย์ถูกบังคับเป็นศูนย์; นโยบายสต็อกติดลบ 4 แบบอยู่ที่ `applyPolicy`
- **เขียนผลในทรานแซกชันเดียว** พร้อมจอง `pg_try_advisory_xact_lock` ต่อ (บริษัท, สินค้า): ลบ ledger และยอดปลายงวดตั้งแต่จุดที่คิดใหม่ → COPY แถวใหม่ → upsert ยอดปลายงวด → ลบงานในคิวเฉพาะที่ `markedat` เก่ากว่าเวลาที่อ่านข้อมูล (`Persist`) เอกสารที่เข้ามาระหว่างคำนวณจึงยังค้างคิวไว้คิดรอบหน้า
- **งานล้มเหลว**: ปล่อยกลับเข้าคิวพร้อมเวลาพักตามจำนวนครั้ง (30 วิ × attempts) ครบ 5 ครั้งย้ายเข้า `stock_dead_letter` (`dirty.go`)

### 5.4 คิวงานคิดต้นทุน `stock_dirty` (และคิวเก่าที่ยังเหลือ)
- **คิวจริง** = `stock_dirty` (PK `businesscode, itemcode`) เอกสารหลายใบที่แตะสินค้าเดียวกันถูกยุบเป็นแถวเดียว โดย `fromdate` เก็บวันที่เก่าสุดที่กระทบ (`LEAST`), `enqueuedat` = เวลาที่เข้าคิวครั้งแรก (ใช้เรียงคิว จึงไม่มีสินค้าที่ถูกดันท้ายคิวไม่รู้จบ), `markedat` = เวลาที่ถูกแตะล่าสุด
- worker หนึ่งตัวต่อหนึ่ง holding เปิดอัตโนมัติเมื่อมีงานเข้ามาครั้งแรก (`stockengine/manager.go` `EnsureWorker`) ปิดทั้งหมดได้ด้วย `BCAI_STOCK_WORKER=0`
- `stockwaitprocess`/`docwaitprocess` ยังถูก insert อยู่ (`AddToDocWaitProcessQueues`) แต่ไม่มีใครอ่านแล้ว — เป็นขยะสะสมที่ยังไม่ได้ถอด (ค้างอยู่ในรายการงานที่ต้องตัดสินใจ)

### 5.5 route goapi ที่เกี่ยวกับต้นทุน
| route | หน้าที่ | อ้างอิง |
|---|---|---|
| `POST /goapi/processstockcalccost` | สั่งคิดต้นทุนของสินค้าที่ระบุใหม่ทันที (รอผล) | `backend/internal/goapi/bootstrap.go`, `handlers/process_stock_calc_cost.go` |
| `POST /goapi/api/process/queue-status` | จำนวนงานค้างจริงของบริษัทที่ล็อกอินอยู่ | `handlers/stock_queue_status.go` |
| `POST /goapi/api/stockcost/query|summary|check` | อ่าน `public.stock_ledger` (กรองด้วย businesscode จาก JWT) | `handlers/process_stock_cost.go` |
| `POST /goapi/api/process/product-balance` | `ProcessProductBalanceUpdate` ทั้ง holding | `bootstrap.go:411`, `handlers/product_balance_update.go:29,40` |
| `POST /goapi/api/health/queue[/:holdingcode]` | สถานะคิว worker pool | `bootstrap.go:379-380` |
worker pool ที่ busy-poll ตาราง `queues` ทุก 100 ms (`backend/internal/goapi/workers/doc_processor.go:232-245`, สร้างที่ `bootstrap.go:136`) **ไม่ได้** เกี่ยวกับ `stockwaitprocess` (คนละคิว — worker เรียก `QueueManager.GetActiveShops` ที่ query ตาราง `queues` `backend/internal/goapi/mypostgres/init_schema.go:110,133`; runtime local: `queues` มี 0 แถวใน bc001/bctest01/demo/postgres/test และไม่มีตารางนี้ใน appdb/qa23995213/uat260810a)

## 6. goapi `inventory/costing` — engine หลายวิธี (DORMANT)

- routes ใต้ `/goapi/api`: `GET|PUT /products/:itemcode/costing-config`, `GET /products/:itemcode/cost-layers`, `POST /inventory/receipt|issue|transfer|adjustment|sales-return|purchase-return`, `GET /reports/inventory-valuation`, `GET /reports/stock-card/:itemcode` (`backend/internal/goapi/inventory/handler.go:15-34`, mount ที่ `bootstrap.go:384-386`); holdingcode รับจาก query param (`handler.go:49,68,206`) ต่อ DB ด้วย `PgSqlFastConnect(holdingCode)` (`handler.go:39`)
- engine เลือกตาม `productcostingconfig.costingmethod` default `movingaverage` (`backend/internal/goapi/inventory/service.go:352-366`); ที่ implement: `MovingAverageEngine` (สูตร new_avg = (old_qty×old_avg + qty×cost)/(old_qty+qty) — `costing/moving_average.go:11-33`), `FIFOEngine`, `LIFOEngine`, `FEFOEngine`, `StandardEngine` (`costing/engine.go:37-53`); ค่าคงที่ `periodicaverage`/`lot` มีชื่อแต่ไม่มี engine (`inventory/models/models.go:10-15` เทียบ `engine.go:37-53`)
- ตาราง: `productcostingconfig`, `inventorycostlayers`, `inventorystockbalances`, `marketplacestockbalances`, `marketplacedimensionprices`, `inventorycosttransactions`, `inventoryvariances`, `inventorylandedcosts`, `inventorylandedcostallocations`, `inventoryaccountingperiods` (`backend/internal/goapi/inventory/database.go:52-236`) สร้างโดย `CreateInventoryCostingTables` ซึ่งถูกเรียกจาก `process/build/create-database.go` (import :13) และ handler `create-tables` ที่ **ไม่ได้ผูก route** (`CreateTablesHandler` `handler.go:277-288` ไม่มีใน `RegisterRoutes`; ตั้งใจถอด — `TestGoAPIRouteSurfaceExcludesOperationalEndpoints` ใส่ `POST /goapi/api/inventory/create-tables` ในรายการ forbidden และ fail ถ้ายัง register `bootstrap_test.go:83,101-103`)
- ไม่มีใครเรียกจาก Kafka consumer/`stockengine` (rg `goapi/inventory` เจอเฉพาะ `bootstrap.go:23` และ `build/create-database.go:13`); runtime local: `inventorycostlayers`/`inventorystockbalances`/`productcostingconfig` = 0 แถวใน bc001, demo, test → ระบบต้นทุนคู่ขนานที่ยังไม่ถูกใช้ และ **ไม่ sync** กับ `stock_ledger`

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
| goapi Kafka consumer + `stockengine`/`stock_ledger` | LIVE | ตัวรับเอกสารเปิดครบ 17 group; รายงานทุกตัวอ่าน `stock_ledger` |
| `stockwaitprocess`/`docwaitprocess` | LIVE-เติม / ไม่มีผู้อ่าน | คิวจริงคือ `stock_dirty`; ยังไม่ได้ถอดการ insert ออก |
| FIFO lot | ไม่มีในระบบ | โค้ดและตาราง `processstocklot` ถูกลบ 2026-09-15 (ถ้าต้องการ FIFO ต้องออกแบบใหม่บน `stock_ledger`) |
| legacy `stockprocess` + `pkg/stockcalculator` | DEAD | โหมด 1 ไม่รัน local; ชื่อตาราง SQL ไม่ตรง AutoMigrate |
| goapi `inventory/costing` | DORMANT | route มี ไม่มีผู้เรียก ตาราง 0 แถว |
| ClickHouse ในเส้นทางสต็อก | STUBBED | `myclickhouse/utils.go:234-236`, `kafka/utils.go:375-378` |
| stockbalanceimport | BROKEN-ON-USE | พึ่ง ClickHouse persister |

## ช่องว่าง / สิ่งที่ยังไม่ตรวจ

1. **ยังไม่ทดสอบ end-to-end บนเครื่องนี้**: POST เอกสารสต็อกจริง → ดู `stock_ledger` เปลี่ยน (Mongo ว่างทั้ง 7 collection) — ที่พิสูจน์แล้วคือเทสต์ที่รันกับ PostgreSQL 18 จริง (`go test -tags integration ./internal/goapi/process/stockengine/`)
2. prod (`worker` โหมด 1 ตาม `deploy/account/compose.yml:272-276`) จะเจอ error `relation "stock_transaction" does not exist` จาก `stockprocess` จริงหรือไม่ — ยังไม่เห็น log prod
3. `systemadmin/productadmin` ที่ publish `when-stock-process-created` (`productadmin_service.go:31`) ถูกลงทะเบียน route ไหม ใครใช้ — ยังไม่ตรวจ
4. ทศนิยม `StockQtyPoint/StockAmountPoint/StockCostPoint` ค่าจริงต่อ holding — ยังไม่ตรวจ (`stock_ledger` เก็บเป็น NUMERIC(18,8) จึงไม่ตัดทศนิยมทิ้งเหมือนของเดิม)
5. `report-stock` package (`backend/internal/goapi/process/report-stock/*`) และคำสั่ง `reportproductstockmovement` ฯลฯ ใน `commands.go` เข้าถึงได้จาก route ไหน (ถ้าอิง `ReportPostHandler` ก็ตายไปด้วย) — ยังไม่ตรวจ
6. (ปิดแล้ว — `bootstrap_test.go:83,101-103` ยืนยันว่า `create-tables` ต้อง**ไม่**ถูก register; ดู §6)
7. คำถามถึงลุงจืด: (ก) `inventory/costing` ที่เลือกวิธีคิดต้นทุนได้ จะเก็บไว้หรือลบทิ้ง (ตอนนี้ไม่มีใครเรียกและไม่ sync กับ `stock_ledger`); (ข) ต้องการ FIFO จริงไหม ถ้าต้องการต้องออกแบบใหม่บนสมุดสต็อก; (ค) legacy `stockprocess` + `pkg/stockcalculator` + consumer โหมด 1 ฝั่งสต็อก ลบได้หรือไม่; (ง) `stockwaitprocess`/`docwaitprocess` จะเลิก insert เลยไหม
