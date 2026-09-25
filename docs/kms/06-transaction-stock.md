# ธุรกรรมสต็อกและเครื่องคำนวณต้นทุน (Stock transactions & cost engines)

> ตรวจล่าสุด: 2026-09-25 — เขียนใหม่ทั้งหมดหลัง ADR ถอด MongoDB/Kafka/Redis/ClickHouse (2026-09-23); ทางเข้าเอกสารสต็อกเดิมถูกลบทั้งหมด เหลือแต่ engine ต้นทุนฝั่ง goapi ที่ไม่มีข้อมูลป้อน

## 1. สิ่งที่ถูกลบทั้งหมด

- HTTP module สต็อกเดิมของ mainapi (`stocktransfer`, `stockreceiveproduct`, `stockreturnproduct`, `stockpickupproduct`, `stockadjustment`, `stockbalance`, `stockbalancedetail`, `stockbalanceimport`) — ไม่มีในรีโปแล้ว (`backend/internal/transaction/` ไม่มีไฟล์ใดใน git)
- `backend/internal/warehouse/` (คลัง/ที่เก็บ/ชั้นวาง) ไม่มีแล้วทั้งโฟลเดอร์
- `backend/internal/stockprocess/` (consumer ถัวเฉลี่ยรุ่นเก่า) และ `backend/pkg/stockcalculator` ไม่มีในรีโปแล้ว (บนดิสก์อาจเหลือโฟลเดอร์ `logs/` ที่ถูก `.gitignore:54` ข้าม)
- การลบเป็นส่วนหนึ่งของ ADR [`decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md`](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md)

## 2. สิ่งที่ยังอยู่ในโค้ดแต่ไม่มีทางป้อนข้อมูลใหม่

- **`backend/internal/goapi/process/stockengine/`** — คำนวณต้นทุนถัวเฉลี่ยถ่วงน้ำหนักแยกตามคลัง (ADR [`decisions/2026-09-15-stock-cost-engine-v2.md`](decisions/2026-09-15-stock-cost-engine-v2.md)) อ่าน `doc`/`docdetail` เขียน `stock_ledger`/`stock_period_balance` คิวงาน `stock_dirty`
  - wiring ใน `backend/internal/goapi/bootstrap.go`: บรรทัด 116-120 ผูก `stockengine.AfterRecalculate` → `processstock.ProcessProductBalanceUpdateByItems`, บรรทัด 121-122 เริ่ม `stockengine.StartWorkers` (เปิดโดยค่าเริ่มต้น เว้นแต่ env `BCAI_STOCK_WORKER=0|false` — `stockengine/manager.go:71-74`)
  - route เรียกด้วยมือ: `POST /goapi/processstockcalccost` (บรรทัด 342), `/goapi/api/stockcost/{query,summary,check}` (343-345), `/goapi/api/process/queue-status` (379), `/goapi/api/stock-report/{barcodes,warehouses}` (385-386) — frontend ไม่มีจอเรียกและ BFF `/api/goapi/[...goPath]` ไม่เปิด path เหล่านี้
  - ผู้เรียก `stockengine.MarkDocumentDirty` เหลือแค่ใน `backend/internal/goapi/mypg/doc.go` (ฟังก์ชันลบเอกสารที่ไม่มีผู้เรียกแล้ว) → คิว `stock_dirty` ไม่มีใครเติมงานใหม่; และไม่มีโค้ดนอกเทสต์สร้างตาราง `doc`/`docdetail`/`stock_ledger`/`stock_dirty`/`stock_period_balance` (DDL มีแค่ใน `stockengine/store_integration_test.go`)
- **`backend/internal/goapi/process/process-stock/`** (package `processstock`) — คำนวณยอดคงเหลือ/ค้างรับ/ค้างส่งจาก `stock_ledger` + `docdetail` แล้ว UPDATE ลงตาราง `product` (`process-product-balance-update.go`); เรียกจาก `AfterRecalculate` ข้างบน และ `POST /goapi/api/process/product-balance` (`bootstrap.go:378`)
- **`backend/internal/goapi/inventory/`** — engine ต้นทุนหลายวิธี (ตาราง `productcostingconfig`, `inventorycostlayers`, `inventorystockbalances`, `inventorycosttransactions` ฯลฯ สร้างเองตอนเรียกครั้งแรก `inventory/database.go:12`, `handler.go:288`) ลงทะเบียนที่ `bootstrap.go:339` ใต้ `/goapi/api/*` (`inventory/handler.go:15-35`) — เป็นโค้ด PostgreSQL ล้วน ไม่ sync กับ `stock_ledger`
  - **มีจอเรียกอยู่ 1 จุด**: รายงาน `/report/stockbalanceitem` (`stock_balance_item` อยู่ใน `API_READY_REPORTS`, `frontend/src/lib/erp-reports.ts:530-533`) เรียก `GET /api/goapi/api/reports/inventory-valuation` (`erp-reports.ts:915`; BFF allowlist `frontend/src/app/api/goapi/[...goPath]/route.ts:14`); route เอกสารรับ/จ่าย/โอน (`/inventory/receipt|issue|transfer|adjustment|...`) ไม่มีผู้เรียก จึงไม่มีอะไรเติมยอดให้รายงานนี้

## 3. Frontend ยังอยู่แต่ "รอพัฒนา"

- เอกสารสต็อกอยู่ใน registry `ERP_MODULE_CONFIGS` ของ `frontend/src/lib/erp-transaction.ts` (domain `inventory` 11 รายการ เช่น `/transaction/stockbalance`, `/transaction/stockreceiveproduct`, `/transaction/stockpickupproduct`, `/transaction/stockreturnproduct`, `/transaction/stocktransfer`, `/transaction/adjust`, `/transaction/stockcount`) — ดูรายละเอียด registry และ BFF `/api/erp-transaction/[...erpPath]` ใน [`05-transaction-sales-purchase.md`](05-transaction-sales-purchase.md) §3; ทุก route ถูก `isMenuBackendRetired()` ตัดสินว่า "ยังไม่พร้อม"
- หน้าคลัง `frontend/src/app/system-settings/warehouse-tree-view.tsx` ยังอยู่ แต่ตั้งค่า `basePath: "/warehouse"` (`frontend/src/lib/system-setting-screens.ts:1288`) ซึ่งไม่อยู่ใน `POSTGRES_SETTING_BASE_PATHS` (`frontend/src/lib/menu-screen-status.ts:46-49`) จึง "ยังไม่พร้อม" เช่นกัน (ไม่มี backend `/warehouse` แล้ว)

## 4. ยังไม่ตัดสินใจ

- `stockengine`/`process-stock`/`stock_ledger` ออกแบบมารับเอกสารจาก module ที่ถูกลบไปแล้ว — ยังไม่มีคำสั่งว่าจะออกแบบทางเข้าใหม่ (เอกสารสต็อกบน PostgreSQL โดยตรง) เมื่อไร หรือยุบ engine แล้วออกแบบใหม่
- `goapi/inventory` จะเก็บหรือลบ และรายงาน `/report/stockbalanceitem` ควรคงสถานะ "API พร้อม" หรือไม่ ยังไม่มีคำตอบ
- ห้ามเดา schema/endpoint ใหม่เอง — ต้องถามลุงจืดและเทียบกับ `D:\project-champ` ก่อนออกแบบตามกฎ "ยึด Champ เป็นต้นแบบ" ใน `AGENTS.md`
