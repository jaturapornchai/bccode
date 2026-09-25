# ธุรกรรมซื้อ–ขาย (Transaction: Sales & Purchase) — backend เดิมถูกลบ 2026-09-23

> ตรวจล่าสุด: 2026-09-25 — เขียนใหม่ทั้งหมด บทความเดิมอธิบายสถาปัตยกรรม module ซื้อ-ขายที่พึ่ง MongoDB และ Kafka ใช้ไม่ได้แล้วเพราะโค้ดทั้งชุดถูกลบจริง ไม่ใช่แค่ปิดการเชื่อมต่อ

## 1. สถานะปัจจุบัน

- module ซื้อ-ขายเดิมของ mainapi (saleinvoice, saleorder, quotation, purchase, purchaseorder, purchasereturn, purchaserequisition, rfq, paid, pay ฯลฯ) **ไม่มีในรีโปแล้ว** — `backend/internal/transaction/` ไม่มีไฟล์ใดใน git (บนดิสก์อาจเหลือแค่โฟลเดอร์ `logs/` ที่ถูก `.gitignore:54` ข้าม) และ `backend/main.go` ไม่ import แพ็กเกจ `internal/transaction`
- การลบเป็นส่วนหนึ่งของ ADR [`decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md`](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md) — ลบโค้ดทั้งชุดแทนการเขียนใหม่ทันที เพราะลุงจืดสั่งให้โฟกัส GL ก่อน
- ตารางเอกสารเดิมใน PG ต่อ holding (`doc`, `docdetail`, `docref`, `docpayment`) **ไม่มีโค้ดนอกเทสต์สร้างตารางหรือ insert แถวแล้ว** (grep `INSERT INTO doc|docdetail|docref|docpayment` ใน `backend/` ไม่พบ); ฟังก์ชันลบเอกสารเดิม `mypg.DeleteDocPgSql`/`DeleteDocPgSqlTx` (`backend/internal/goapi/mypg/doc.go:15-23`) ไม่มีผู้เรียกแล้ว

## 2. โค้ด goapi ที่ยังค้าง (อ่านตารางที่ไม่มีใครเติม)

- `POST /goapi/api/transaction/{calculate,quick-calc,validate-payment,purchase-history}` (`backend/internal/goapi/bootstrap.go:348-351`, handler `handlers/transaction_calculator.go`, `handlers/purchase_history_handler.go`) — BFF `/api/goapi/[...goPath]` ไม่เปิด path เหล่านี้ (มีเทสต์ยืนยันว่า `api/transaction/calculate` ถูกปฏิเสธ 404: `frontend/src/app/api/goapi/[...goPath]/route.test.ts:52-60`)
- `POST /goapi/api/report/sales/by-document` (`bootstrap.go:354`, `handlers/sales_report.go`) อ่าน `stock_ledger` + `doc` + `debtor` + `productbarcode` — **ยังเปิดใน BFF** (`frontend/src/app/api/goapi/[...goPath]/route.ts:31`) และรายงาน `/report/salesreportbydocument`, `/report/reportgrossprofitbydocument` ถูกตั้งเป็น "API พร้อม" (`API_READY_REPORTS`, `frontend/src/lib/erp-reports.ts:530-533`, เรียกที่บรรทัด 851) ทั้งที่ไม่มีอะไรเติมตารางต้นทางแล้ว

## 3. Frontend มี registry รอ backend

- `frontend/src/lib/erp-transaction.ts` มี registry กลาง `ERP_MODULE_CONFIGS` (บรรทัด 56-850, **70 รายการ**) ครอบคลุม domain `inventory`/`sales`/`purchase`/`ar`/`ap`/`cash-bank` — **ยังไม่มี backend รองรับสักตัว**
- ทุก route ใน registry ถูกกันด้วย `isErpTransactionRoute()` (`frontend/src/lib/erp-transaction.ts:856-859`) ใน `isMenuBackendRetired()` (`frontend/src/lib/menu-screen-status.ts:53`) จึงขึ้น "ยังไม่พร้อม" เสมอ
- BFF รวม `frontend/src/app/api/erp-transaction/[...erpPath]/route.ts` มี allowlist `ALLOWED_MODULES` (บรรทัด 7 เป็นต้นไป: `sale-invoice`, `sale-order`, `quotation`, `purchase`, `purchase-order`, `paid`, `pay` ฯลฯ) แล้วส่งต่อไป mainapi `/transaction/<module>` — mainapi **ไม่มี route นี้แล้ว**; ฝั่ง frontend แปลงสถานะ 404 เป็น key `module_not_available` (`frontend/src/lib/erp-transaction.ts:874`)
- ที่เกี่ยวข้อง: `frontend/src/lib/erp-operations.ts` (`OPERATIONS_CONFIGS` บรรทัด 25, `isOperationsRoute()` บรรทัด 185) เป็น registry งานอนุมัติ/ยกเลิกเอกสาร/ประกอบ-แยกชุดสินค้า/serial/ปรับราคา — ถูกกันด้วย `isMenuBackendRetired()` เช่นกัน และ goapi ไม่มี route `/api/approval/*` ที่ไฟล์นี้เรียก

## 4. ยังไม่ตัดสินใจ

- ยังไม่มีกำหนดว่าจะสร้างซื้อ-ขายใหม่บน PostgreSQL เมื่อไร/เลขที่เอกสาร/schema หน้าตาอย่างไร — ห้ามเดา ต้องรอลุงจืดสั่ง (ADR) และเทียบกับ `D:\project-champ` ก่อนออกแบบ
- โค้ด goapi ใน §2 และสถานะ "API พร้อม" ของรายงานขายจะเก็บ/ปิด/ลบอย่างไรยังไม่มีคำตอบ — ถามลุงจืดก่อนแตะ
- เลขที่เอกสาร, transflag และตาราง `doc/docdetail/docref/docpayment` เดิมถือเป็นประวัติศาสตร์ — schema ใหม่ (ถ้าสร้าง) ไม่ผูกพันต้องเหมือนเดิม
