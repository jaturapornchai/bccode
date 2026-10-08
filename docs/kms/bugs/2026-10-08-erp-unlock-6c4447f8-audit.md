---
date: 2026-10-08
severity: high
component: [backend, frontend]
tags: [bc-account, erp, transaction, permission, report, champ-parity, audit]
fixed: partial
---

# Symptom

ตรวจ commit `6c4447f8` ("ปลดล็อกทุกจอทำงานจริง", deploy แล้ว r20261008-gl-crud-1) ตาม handoff ข้อ [1] UAT / [2] ฟิลเตอร์รายงาน / [3] สิทธิ์ผู้ใช้ — สำรวจแบบอ่านอย่างเดียว ยังไม่แก้โค้ด ผลคือข้ออ้าง "ปลดล็อกครบ" ใช้ได้เพียงบางส่วน

หลักฐานตรวจซ้ำเองจากซอร์ส: เส้นทางตาย 20 เส้น (script นับ `apiPath` มี `/` หรือไม่มี flag หรือไม่อยู่ใน allowlist = 20 จาก 69), กลุ่ม `apiPath` ที่ใช้ร่วมกัน, `DeleteDoc` args, flag ซ้ำใน `models.go`, พจนานุกรม flag ขัดกับ stock engine. ข้ออื่นผ่านการตรวจซ้ำโดย agent รุ่น opus อีกชุดเท่านั้น (ระบุไว้ในแต่ละข้อ)

## Root Cause

### A. ห่วงโซ่เมนู → จอ → BFF → backend ไม่ครบ
- **20 เส้นทางตาย (ยืนยันเอง):** `apiPath` มี `/` (`bank/depositrecord`, `chequereceive/chequedeposit`, `chequepayment/...`, `bank/saledebitnote` ฯลฯ — `frontend/src/lib/erp-transaction.ts:173,390,651`) แต่ backend ลงทะเบียนแค่ `/transaction/:module` ชั้นเดียว (`backend/internal/transaction/httpapi/http.go:34-38`) และ `ModuleTransFlags` ไม่มี key → 404 แล้วจอขึ้นข้อความ `module_not_available`
- **เอกสารเข้า flag ผิด (ยืนยันเอง):** กลุ่ม `apiPath` ที่ใช้ร่วมกัน purchase×10, paid×8, paidadvance×4, receivableother×3, stock-adjustment×3, chequereceive/chequedeposit×7, bank/banktransferrecord×5 ฯลฯ backend เลือก flag จาก URL segment → ทุกเส้นที่ร่วม `purchase`(12) ถูกเก็บเป็นใบซื้อ + ลงบัญชีซื้อ + ขยับสต็อกเข้า (`service.go:351,369,405,414`); `paid`(28) ลงบัญชีรับเงินและตัดลูกหนี้ (opus ตรวจซ้ำว่าบางเส้นในกลุ่มเป็นเจ้าของ flag นั้นจริง จึงไม่ใช่ทั้ง 42 เส้นที่ผิด)
- **key ซ้ำ flag ใน `models.go` (ยืนยันเอง):** 34 = quotation/receiveadvance, 62 = stock-transfer/chequepayment/chequeissued, 28 = paid/temporaryreceipt → list ตาม flag ปนประเภท
- **พจนานุกรม flag ขัดกัน 2 ชุดบนตาราง `docdetail` เดียวกัน (ยืนยันเอง):** `backend/internal/goapi/process/stockengine/calc.go` `transFlagDirection` (54=ยกมาเข้า, 56=เบิกออก, 60=รับสำเร็จรูป, 72=โอนออก, 310=รับพาเชียล) ไม่ตรงกับ `backend/internal/transaction/httpapi/models.go:6-51` (56=สินค้ายกมา, 60=ธนาคาร, 62=โอน, 72=รับสำเร็จรูป, 310=ปรับปรุง) — `service.go` ยัง hard-code เลข flag เพื่อตัดสิน GL/สต็อก (`:351-414`, `:1526-1528`) และรายงาน goapi hard-code 44/48/36/38 (`sales_report.go:190`, `debt_report.go:176,296`)
- **แมปที่ commit เพิ่มส่วนใหญ่ไม่ถูกใช้ (opus ตรวจซ้ำ):** `models.go` เพิ่ม 21 key ตรงกับ config แค่ 3; BFF `ALLOWED_MODULES` เพิ่ม 24 รายการไม่มี config ใช้เลย; 14 รายการ (rfq, purchasepartial, accrualreceive, landedcost, stockcount, goods-inspection ฯลฯ) ไม่มี flag → จะได้ 400 ถ้าถูกเรียก
- **จอสินค้า 3 จอไม่มี backend (opus ตรวจซ้ำ):** `main-menu-screen.tsx:2231,2243,2255` เรนเดอร์ ProductScreen/ProductBarcodeScreen/ProductSetScreen ที่เรียก `/api/product` (`product-screen.tsx:662`) แต่ backend ไม่มี route นี้; เดิมถูก guard ด้วย `RETIRED_BACKEND_ROUTES` (`menu-screen-status.ts:35-38`) ที่ commit นี้เลิกใช้
- **`pr_approval` ดูพร้อมแต่ตาย:** `erp-operations.ts:208-210,250` เรียก `/api/goapi/api/approval/pr-status/pending` ซึ่ง BFF allowlist และ backend ไม่มี
- **badge หายหมด:** `menu-pending-badge.tsx:11-13` คืน `null` ทุกเส้นทาง แต่ `isMenuScreenPending` ยัง true 104/204 เส้นทาง และ shortcut ยังกรองด้วยสถานะนั้น (`dashboard-home.tsx:52,56`, `manage-shortcuts-screen.tsx:108,112`)

### B. Strict Parity: premise ของ handoff หลายข้อไม่ตรงกับ Champ
- **Champ ไม่มี transflag:** grep C++/SQL ทั้งหมดไม่พบ; Champ ใช้ตารางแยกต่อเอกสาร + ตัวแบ่ง เช่น `BCAPInvoice.GRIRBillStatus` 0/1/2 (`POFrmPurchase.cpp:1309`, `POFrmGoodsReceive.cpp:1009`, `POFrmInvoiceReceive.cpp:1050`), `BCStkRefund.BillType` (`POFrmStkReturn.cpp:135,1109`); เลขเดียวที่เป็นรหัสคือ `BCGLSource.GLSource` (`SQLSERVER_Script.sql:15365-15422`) เฉพาะเอกสารที่ลง GL. เลข transflag ของ BC จึงเทียบกับ Champ 1:1 ไม่ได้ (สอดคล้อง `docs/kms/05` §เลขที่เอกสาร: transflag ไม่จำเป็นต้องคงของเดิม)
- **คำนำหน้าเลขเอกสาร Champ เป็นข้อมูลผู้ใช้:** `BCSysTRFormat.DocFormat` ต่อ TransKey (`SQLSERVER_Script.sql:156,171`, รูปแบบ `YYMM-@@#####` ที่ `BC5Account.rc:12816`, `BRunningEdit.cpp:131`) — `ModulePrefixes` (`models.go:54-100`) ตรวจกับ Champ ไม่ได้; ใบกำกับภาษีมี `TaxFormat` แยก (`ChmFrmBilling.cpp:245`)
- **คลาส Champ บางแถวในตารางของ agent อ้างคลาสที่ถูก comment ออก:** คลาสที่ active คือ `CChmFrmGoodsPurchase` (MainFrm.cpp:1773), `CChmFrmQuotation` (:2094), `CARFrmBcpaybill1View` (:2120), `CChmFrmGoodsReceive` (:3127)
- **[2] ฟิลเตอร์รายงาน:** `ErpReportConfig` (`erp-reports.ts:19-27`) ไม่มีแนวคิดฟิลเตอร์; ฟิลเตอร์จริงอยู่ที่ viewer + BFF `route.ts:83-85` + SQL ใน handler Go 3 ชุด (≥6 ชั้น). Champ ให้ฟิลเตอร์เป็นช่วงจาก-ถึงเฉพาะรายงาน (เช่น `APRepAgingDlg.cpp:172`, `APRepMovementDlg.cpp:58`, `IC2RepItemMovementDlg.cpp:91,189`); แผนก/โครงการเป็นรายงานย่อยหรือการจัดกลุ่ม ไม่ใช่ฟิลเตอร์กลาง (`APRepPayByDateView.cpp:145-149`) และไม่มีใน handler ERP ทั้ง 3 ชุด → ชุดเดียวทุกรายงานขัด Strict Parity
- **[3] สิทธิ์:** ชื่อ `CChmFrmUserPermission` ไม่มีใน Champ; โมเดล Champ คือสิทธิ์ต่อเมนูต่อ role (`bcrolepermission`: ismake/isdelete/isinsert/isedit — `SQLSERVER_Script.sql:14945-14956`) + option flag ~25 ตัว + ขอบเขตคลังต่อ role, SA/ADMIN ข้ามได้ (`BCSWinApp.cpp:891`). `MenuItem` ใน `menu-data.ts:12-17` ไม่มีฟิลด์สิทธิ์ — สิทธิ์อยู่ใน `backend/internal/organization/rolepermission/*` + `frontend/manual/permissiondefinition.json`

### C. ช่องโหว่/บั๊กที่เจอระหว่างทาง (opus ตรวจซ้ำ ยกเว้นที่ระบุ)
- **`/transaction/:module` ไม่ตรวจสิทธิ์เลย** (`http.go:42-52` ตรวจแค่ session/company; BFF ตรวจ bearer + allowlist) ขณะที่ GL ตรวจ (`generalledger/httpapi/permissions.go`); `ErpCrudWorkbench` ไม่เคารพสิทธิ์ create/update/delete ใน UI; ฟิลด์ `post_to_gl` กำหนดจาก client แล้ว backend เขียนใบสำคัญสถานะ posted ลง GL โดยไม่ผ่านสิทธิ์ GL; fixed-asset API (รวม asset-post-gl) ก็ไม่ตรวจ
- **`DeleteDoc` (ยืนยันเอง `service.go:1646`):** `args[:len(args)-2]` ส่ง bind param ไม่ครบ ผิดพลาดถูกกลืน (`_ =`) → `docNo` ว่าง ไม่ลบ `stock_ledger`; ทั้งฟังก์ชัน (`:1628-1664`) ไม่กลับรายการ GL → ลบเอกสารขาย/ซื้อ/รับเงินแล้วใบสำคัญยังอยู่
- **แก้เอกสารเขียนทับใบสำคัญที่ posted:** `ON CONFLICT (company, kind, id) DO UPDATE SET version=2, payload=$4` (`service.go:507-510`) + DELETE/INSERT `gl_lines` — ขัดกฎ append-only/กลับรายการ
- **Auto-GL hard-code ผัง + float:** บัญชี 1121/1131/1141/1151/2111/2131/4111/5201 (`service.go:458-479`) ขัดกฎ "ห้ามตัดสินตามรหัสบัญชี"; ปีบัญชี fallback `"2569"` (`:437,590,744`); จำนวนเงิน `float64` (`models.go:109-113,131`) ขัด Rule 20; ความผิดพลาดของ auto-GL/สต็อกแค่ `fmt.Printf` แล้ว commit ต่อ (`:351-421`)
- **แจ้ง "บันทึกสำเร็จ" ทั้งที่ล้มเหลว:** BFF ส่ง `userErrorStatusOk` (`erp-transaction/[...erpPath]/route.ts:133,173,202`) → proxy แปลง 4xx ที่มี `code` เป็น HTTP 200 + `success:false` (`workspace-api.ts:96-97`) แต่ `saveErpTransaction`/`deleteErpTransaction` เช็คแค่ `res.ok` (ยังไม่ได้ทดสอบจริงบนจอ)
- **รายงาน:** (1) รายงานหนี้ 8 query ใช้ `d.company = $1 OR d.company = '01' OR $1 = ''` (`debt_report.go:53,183,272,390`) → บริษัท 02 เห็นเอกสารบริษัท 01; (2) มูลค่าสต็อกอ่าน `stock_period_balance` ทุกงวดโดยไม่เลือกงวดล่าสุด (`inventory/service.go:449-454`) → นับซ้ำรายเดือน; (3) `ar_credit_limit` ไม่มีวงเงินเลย (`debt_report.go:381-388`) ขณะที่ Champ มี 13 คอลัมน์ (`ARReportDebtLimitView.cpp:80-93`); (4) ปุ่มช่วงวันที่ของ stock_balance_*/ap_status/ar_status/ar_credit_limit ไม่มีผล (backend ไม่ใช้วันที่); (5) รายงานเคลื่อนไหวสต็อกไม่มี LIMIT/OFFSET (`service.go:262-292`)

## Fix

**แก้แล้ว (2026-10-08, ลุงจืดอนุมัติข้อ [1],[2] — ทดสอบ CRUD ด้วย):**

1. **ทะเบียนประเภทเอกสารเดียว** `backend/internal/doctype/doctype.go` (63 ประเภท: key, transflag, คำนำหน้า, ผลต่อสต็อก None/In/Out/ByLine) แทน `ModuleTransFlags`/`ModulePrefixes`/`PostToGL` ที่ซ้ำและขัดกัน ; `/transaction/:module`, `stockengine`, `myglobal.TransFlagsToProcess` และรายงานขาย/หนี้/สต็อกอ่านจากทะเบียนนี้ทั้งหมด — flag ไม่ซ้ำข้ามประเภท
2. **เส้นทางตาย 20 เส้น/เอกสารเข้า flag ผิด** — `apiPath` ทุกตัวเป็น key ชั้นเดียวของทะเบียน ; BFF allowlist (`erp-transaction-modules.ts`) เท่ากับทะเบียนเป๊ะ ล็อกด้วย vitest ที่อ่าน `doctype.go`
3. **ลงบัญชี/ตัดสต็อกเฉพาะ 8 ประเภทที่ service.go ลงอยู่แล้ว (ยังไม่ได้เทียบ Champ ทีละประเภท)** — ที่เหลือบันทึกเอกสารอย่างเดียว และจอบอกผู้ใช้ด้วย `erp_store_only_notice` ; ไม่มีเอกสาร "ซื้อ" ปลอมจากเส้นทางที่ใช้ `purchase` ร่วมกันอีก
4. **ตาราง doc/docdetail/docref/stock_ledger/stock_period_balance ไม่มี DDL ในโค้ด** (พบระหว่างทำ UAT) — prod `rungrueng` ไม่มีตารางเหล่านี้เลย (มีแต่ฐาน `demo`) ; เพิ่ม `httpapi/schema.go` `ensureSchema` สร้างครั้งแรกที่ใช้ (advisory lock) ก่อน deploy ฝั่ง backend ต้องพร้อมนี้
5. **Create→Read ทำ description/custname หาย, หมายเหตุไปอยู่ `cancelreason`, ไม่ตอบ `id`, แก้เอกสารที่ไม่มี/ลบแล้วตอบสำเร็จเงียบ ๆ** — เพิ่มคอลัมน์ custname/description/remark, ตอบ `id`, UPDATE ที่ไม่พบแถว = 404
6. **เลขที่เอกสาร** — ส่ง `docno` ว่าง backend ออกให้ (`NextDocNo`) ; จอแสดง placeholder `docno_auto_placeholder`
7. **จอสินค้า/บาร์โค้ด/ชุดสินค้า/ประวัติราคา** คืน guard `isRetiredBackendScreen` (backend ไม่มี `/api/product`) ; `APPROVAL_KINDS` ว่างเพราะไม่มี `/api/approval/*` ; รายงานหนี้/ขาย/สต็อกที่อ่าน flag ตามทะเบียนแก้ตามแล้ว (มีเทสต์ flag ทั้ง 3 จุด)

**UAT (Create→Read→Update→Delete + ตรวจ PostgreSQL ทุก step, Demo→`rungrueng`/01, local `postgres` db `rungrueng`):** cashdeposit ผ่าน ; ประเภท store-only 55 ตัวผ่าน (doc +1, ไม่แตะ GL/สต็อก, แก้แถวเดิม, ลบ = soft delete) ; ประเภท posting 8 ตัวผ่านขั้น CRUD ; ลบข้อมูลทดสอบด้วย id เป้าหมายแล้ว จำนวนแถวกลับ baseline

**ยังไม่แก้ — รอลุงจืดอนุมัติ [3] (รายงานเฉย ๆ):**

- `/transaction/:module` ไม่ตรวจสิทธิ์รายเมนู ; UI ไม่เคารพสิทธิ์ create/update/delete ; ฟิลด์ `post_to_gl` จาก client
- `DeleteDoc` (`service.go:~1646`) bug `args[:len(args)-2]` ไม่ลบ `stock_ledger` และไม่กลับรายการ GL ; แก้เอกสารที่ posted ไม่ลงบัญชีซ้ำและไม่กลับสต็อก — **ยืนยันจาก UAT:** แก้/ลบเอกสาร 8 ประเภท posting ไม่ย้อน GL/สต็อก
- Auto-GL hard-code ผัง/ปีบัญชี fallback `"2569"` ; จำนวนเงิน `float64` (Rule 20) ; ความผิดพลาดของ auto-GL/สต็อกแค่ `fmt.Printf`
- ประเภท ByLine (`stock-transfer`, `stock-adjustment`) ไม่เขียน `calcflag` → ไม่เคลื่อนสต็อก ; ไม่มีตัวกัน stock ติดลบ ; `apdebitnote`/`apcreditnote`/`ardebitnote` ผลต่อสต็อกรอตัดสิน
- `status` draft/approved ไม่ถูกเก็บ ; เลขที่เอกสารต่อบริษัทแบบ Champ (`DocFormat`) ; ใบเช็ค/บัตรเครดิตไม่มีวงจรสถานะแบบ Champ
- รายงานหนี้ 8 query ปนบริษัท (`d.company = '01'`) ; มูลค่าสต็อกไม่เลือกงวดล่าสุด ; ฟิลเตอร์รายงานตาม handoff ข้อ [2] เดิม ; `TestBuildDebtReportQuery` ล้มที่ HEAD
- BFF `userErrorStatusOk` แปลง 4xx เป็น 200 แต่ `saveErpTransaction`/`deleteErpTransaction` เช็คแค่ `res.ok` (ยังไม่ทดสอบบนจอ)
- จอ `/price_history`, `/product_barcode_shelf` แบบ standalone ยังไม่ guard (มีมาก่อน)

## Regression Test

- Go: `backend/internal/doctype/doctype_test.go` (flag/key/prefix ไม่ซ้ำ, StockFlags), `stockengine/calc_test.go`, `httpapi/doctype_crud_test.go` (sqlmock: create→id, round-trip custname/description/remark, update 404), `report_doctype_flags_test.go`, `service_doctype_flags_test.go`, `process_product_balance_flags_test.go`
- vitest: `erp-transaction.test.ts` (apiPath ชั้นเดียว ∈ ทะเบียน, BFF allowlist = `doctype.go`), `menu-screen-status.test.ts`, `erp-operations.test.ts`
- UAT script (นอก repo): `C:/tmp/uat-erp-crud.mjs` — ยังไม่มีเทสต์สำหรับ [3] (สิทธิ์ 403, DeleteDoc ย้อนสต็อก/GL) ให้เพิ่มเมื่อแก้

อ้างอิง: [[2026-09-19-champ-parity-menu-cut]] · `docs/kms/05` · `docs/kms/19-menu-coverage-market-standard.md`
