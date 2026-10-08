# ธุรกรรมซื้อ–ขาย (Transaction: Sales & Purchase) — module `/transaction/:module` บน PostgreSQL

> ตรวจล่าสุด: 2026-10-08 — module เดิม (พึ่ง MongoDB/Kafka) ถูกลบใน ADR [`decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md`](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md) แล้วสร้างใหม่บน PostgreSQL อย่างเดียวที่ `backend/internal/transaction/httpapi/` (commit 04341468, f3bd2465, 6c4447f8)

## 1. สถานะปัจจุบัน

- route: `GET /transaction/:module/list`, `GET|PUT|DELETE /transaction/:module/:id`, `POST /transaction/:module` (`httpapi/http.go`) — CRUD ลงตาราง `doc`/`docdetail`/`docref` ของฐาน holding นั้น กรองด้วยบริษัท (`businesscode`) + `transflag`; DELETE = soft delete (`isdelete`)
- **ทะเบียนประเภทเอกสารเดียว:** `backend/internal/doctype/doctype.go` (63 ประเภท: key, transflag, คำนำหน้าเลขที่, ผลต่อสต็อก None/In/Out/ByLine) — `/transaction/:module`, เครื่องคำนวณสต็อก (`myglobal.TransFlagsToProcess = doctype.StockFlags()`) และรายงานใช้ร่วมกัน ห้ามสร้างตารางแมป flag ที่อื่น; ไม่มีสองประเภทใช้ flag เดียวกัน
- ตารางสร้างเองครั้งแรกที่ใช้ (`httpapi/schema.go` `ensureSchema` เรียกจาก `Http.scope`): `doc`, `docdetail`, `docref`, `stock_ledger`, `stock_period_balance` — ฐาน holding ใหม่ใช้งานได้ทันทีโดยไม่ต้องรัน DDL มือ
- เลขที่เอกสาร: ส่ง `docno` ว่าง = backend ออกให้ `PREFIX+YY(พ.ศ.)+MM-NNNN` (`NextDocNo`) ; คำนำหน้าคงที่ตามทะเบียน (Champ ให้ผู้ใช้ตั้งเองต่อเอกสารที่ `BCSysTRFormat.DocFormat` — ยังไม่ทำ)
- header เก็บ `custname`, `description`, `remark` แยกคอลัมน์ (ไม่ใช้ `cancelreason` แทนหมายเหตุอีก) ; ตอบ `id` แถวใหม่กลับตอนสร้าง
- `mypg.DeleteDocPgSql`/`DeleteDocPgSqlTx` (`backend/internal/goapi/mypg/doc.go:15-23`) เป็นฟังก์ชันลบของ goapi เดิม ไม่ใช่ทางที่ module นี้ใช้

## 2. โค้ด goapi ที่ยังค้าง (อ่านตารางที่ไม่มีใครเติม)

- `POST /goapi/api/transaction/{calculate,quick-calc,validate-payment,purchase-history}` (`backend/internal/goapi/bootstrap.go:348-351`, handler `handlers/transaction_calculator.go`, `handlers/purchase_history_handler.go`) — BFF `/api/goapi/[...goPath]` ไม่เปิด path เหล่านี้ (มีเทสต์ยืนยันว่า `api/transaction/calculate` ถูกปฏิเสธ 404: `frontend/src/app/api/goapi/[...goPath]/route.test.ts:52-60`)
- `POST /goapi/api/report/sales/by-document` (`bootstrap.go:354`, `handlers/sales_report.go`) อ่าน `stock_ledger` + `doc` + `debtor` + `productbarcode` — **ยังเปิดใน BFF** (`frontend/src/app/api/goapi/[...goPath]/route.ts:31`) และรายงาน `/report/salesreportbydocument`, `/report/reportgrossprofitbydocument` ถูกตั้งเป็น "API พร้อม" (`API_READY_REPORTS`, `frontend/src/lib/erp-reports.ts:530-533`, เรียกที่บรรทัด 851) ทั้งที่ไม่มีอะไรเติมตารางต้นทางแล้ว

## 3. Frontend: ERP_MODULE_CONFIGS ↔ ทะเบียน backend

- `frontend/src/lib/erp-transaction.ts` มี `ERP_MODULE_CONFIGS` (64 config) แต่ละตัวมี `apiPath` = key ชั้นเดียวของทะเบียน `doctype.go` (ไม่มี `/` ; ที่ใช้ร่วมกันมีเฉพาะ `stock-count` ซึ่ง `/transaction/stockcount` กับ `/inventory/count-sheet` เปิดเอกสารชนิดเดียวกัน) ; ไม่มี `defaultDocPrefix` แล้ว เพราะ backend ออกเลขที่ให้
- `frontend/src/lib/erp-transaction-modules.ts` (`ERP_TRANSACTION_MODULES`) = รายการ key 63 ตัวเรียงเหมือน `doctype.go` เป๊ะ และเป็น allowlist ของ BFF `frontend/src/app/api/erp-transaction/[...erpPath]/route.ts` ; `erp-transaction.test.ts` เทียบสองฝั่งโดยอ่าน `doctype.go` จริง (เพิ่ม/ลบประเภทต้องแก้ทะเบียนก่อน แล้วทดสอบจะบังคับให้แก้ฝั่ง frontend ตาม) — **จึง deploy frontend อย่างเดียวไม่ได้ ต้อง `--all` พร้อม mainapi**
- ประเภทที่ลงบัญชี/สต็อกอัตโนมัติตอนบันทึกมี 8 ตัว (`ERP_AUTO_POSTING_MODULES` ใน `erp-transaction.ts` ตรงกับ `service.go` SaveDoc): GL = purchase, sale-invoice, paid, pay, sale-invoice-return, purchase-return, expense ; `stock_ledger` = purchase, sale-invoice, sale-invoice-return, purchase-return, stock-pickup-product ; ประเภทอื่น "บันทึกเอกสารอย่างเดียว" และจอแสดงข้อความ `erp_store_only_notice`
- เมนูที่ไม่มี config (5 route: ใบกำกับภาษี/e-Tax, ภาษีหัก ณ ที่จ่าย, ปรับปรุงใบรับสินค้า, วงเงินเงินสดย่อย, ผู้ติดต่อธนาคาร — Champ ไม่มีเอกสารจริงหรือเป็น master) ตกเป็น "ยังไม่พร้อม" ผ่าน fallback ของ `isMenuScreenPending` ; จอสินค้า/บาร์โค้ด/ชุดสินค้า/ประวัติราคา (`RETIRED_BACKEND_ROUTES`) ยังถูกกันด้วย `isRetiredBackendScreen` เพราะ backend ไม่มี `/api/product`
- ที่เกี่ยวข้อง: `frontend/src/lib/erp-operations.ts` (`OPERATIONS_CONFIGS`, `isOperationsRoute()`) เป็น registry งานอนุมัติ/ยกเลิกเอกสาร/ประกอบ-แยกชุดสินค้า/serial/ปรับราคา — `APPROVAL_KINDS` ว่าง เพราะ goapi ไม่มี route `/api/approval/*` ; โค้ดดึง/อนุมัติรายการรออนุมัติยังอยู่แต่ไม่ทำงานจนกว่าจะมี endpoint จริง

## 4. ยังไม่ทำ / ยังไม่ตัดสินใจ

- ตรวจสิทธิ์รายเมนูที่ `/transaction/:module` (ตอนนี้ตรวจแค่ session + บริษัท) และ UI ยังไม่เคารพสิทธิ์ create/update/delete — ดู [`bugs/2026-10-08-erp-unlock-6c4447f8-audit.md`](bugs/2026-10-08-erp-unlock-6c4447f8-audit.md)
- แก้/ลบเอกสารที่ลงบัญชีหรือตัดสต็อกแล้วยังไม่กลับรายการ GL/`stock_ledger` (แก้ไม่ลงบัญชีซ้ำ ลบไม่ย้อน) ; auto-GL ยัง hard-code ผังบัญชีและใช้ `float64`
- ประเภท ByLine (`stock-transfer`, `stock-adjustment`) ยังไม่เขียน `calcflag` ลง `docdetail` จึงไม่เคลื่อนสต็อก ; `apdebitnote`/`apcreditnote`/`ardebitnote` ผลต่อสต็อกรอลุงจืดตัดสิน
- เลขที่เอกสารต่อบริษัทแบบ Champ (`BCSysTRFormat.DocFormat`) ยังไม่ทำ ; ใบเช็ค/บัตรเครดิตเป็นเอกสารบันทึกอย่างเดียว ยังไม่มีวงจรสถานะแบบ Champ
- โค้ด goapi ใน §2 และสถานะ "API พร้อม" ของรายงานขายจะเก็บ/ปิด/ลบอย่างไรยังไม่มีคำตอบ — ถามลุงจืดก่อนแตะ
