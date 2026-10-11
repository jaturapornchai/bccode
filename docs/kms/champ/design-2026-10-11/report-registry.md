# Design: Report Registry + จอรายงานกลาง (audit E1–E6, ORD-12)

วันที่ 2026-10-11 · สถานะ: ข้อเสนอออกแบบ ยังไม่มีโค้ด
ขอบเขต: รายงานบัญชีมาก่อน (GL, ลูกหนี้, เจ้าหนี้, ธนาคาร-เงินสด, ภาษี) ทะเบียนสินค้าทำทีหลัง

ฐานการตัดสินใจของลุงจืด (2026-10-11):
- ชื่อรายงานของ Champ ทุกตัวต้องอยู่ในเมนู
- ทุกเมนูรายงานเปิดจอรายงานกลาง**จอเดียว** ซึ่งขับด้วย registry ฝั่ง backend
- จอที่มีแต่ UI ยังไม่มีของจริง ให้แสดง "รอพัฒนา"

## 1. Champ ทำอย่างไร (หลักฐาน)

- **ทางเข้ารายงานในโปรแกรมจริง = submenu รายงานของแต่ละโมดูล** (แก้ 2026-10-11 หลังตรวจภาพจอ — ภาพจอชนะ `.rc`/`menuconfig.xml`)
  - `bcspec/เมนูโปรแกรมทั้งหมด.jpg` (build 5.5 Pro #20260709-12): แถบเมนู 13 หัว ไม่มีหัว "รายงาน" หรือศูนย์รายงาน
  - `bcspec/หน้าจอเมนูบัญชี/เมนู-ประกอบ.jpg`: ท้ายเมนูบัญชีมี "รายงานรายวัน >", "รายงานงบการเงิน >", "รายงานภาษี >"; `หน้าจอเมนูลูกหนี้/เมนู-ประกอบ.jpg`: "รายงานรายวัน >", "รายงานเชิงวิเคราะห์ >"
  - เหตุที่ซอร์สขัดกับภาพ: `menuconfig.xml` ถูกโหลดจากโฟลเดอร์ exe ตอนรัน (`champ/champ/XMLMenuConfig.cpp:34-38`) จึงไม่ใช่ไฟล์ใน source tree; ตัวสร้างเมนูข้าม `disable="yes"` (`XMLMenuConfig.cpp:150-153,179`)
  - ข้ออ้างเดิม "ศูนย์รายงาน 5007 เป็นทางเข้าเดียว" (`10-reports.md` §0 ข้อ 1, §1.2) **ผิดกับโปรแกรมจริง** — ต้องแก้ `10-reports.md` ใน change ที่ลงมือ R6 (ไม่แก้ในเอกสารนี้)
  - ผลต่อ BC: ชื่อรายงานทุกตัวอยู่ใน submenu รายงานของโมดูลนั้นตามลำดับ/ตัวคั่น/ถ้อยคำในภาพ; ทุกรายการเปิดจอรายงานกลางจอเดียว; **ห้ามเพิ่มหัวเมนู "รายงาน"/"ศูนย์รายงาน" ระดับบนสุด**
- **จำนวนรายงาน:** ยึดโปรแกรมจริง = PDF ตัวอย่างใน `bcspec/Report_*` 418 ไฟล์ − แบบฟอร์ม "เมนู - พิมพ์ภาษี…" 3 ไฟล์ = **415 รายงาน** (จัดซื้อ 122 / จัดขาย 59 / ลูกหนี้ 60 / เจ้าหนี้ 34 / สินค้า 44 / เช็ค 28 / สินทรัพย์ 22 / ธนาคาร 21 / บัญชี 25) ตรงกับ `11-menu-parity-audit-2026-10-09.md:26-38`
  - `menuconfig.xml` (366 รายการ, `10-reports.md` §1.3) ใช้เป็นตัวเทียบรหัสเมนูเท่านั้น; ชื่อที่มีฝั่งเดียวต้องทำตารางกระทบยอด (R0) ก่อนลงทะเบียน
- **กรอบการทำงานร่วม** (`10-reports.md` §3):
  - เปิด dialog เงื่อนไขก่อนทุกรายงาน กด Cancel = ไม่เปิดรายงาน
  - คีย์ลัด: F11 ประมวลผล / Esc / F6 ตัวเลือก / F7 Test SQL / F9 ตัวกรองผู้ใช้ .urf
  - หัวรายงาน 4 บรรทัด: บริษัท / ชื่อรายงาน + หน้า n/N / Printed By + Date / คำอธิบายเงื่อนไข
  - ยอดรวมท้ายหน้า (`m_bShowPageSummery`) + ยอดรวมทั้งสิ้น
  - Export .xls/.txt สิทธิ์ print/export/change แยกกัน (`BCS_GET_PRINTREPORT` / `ISEXPORTREPORT` / `CHANGEREPORT`) — **เป็นสิทธิ์รายผู้ใช้/กลุ่ม ไม่ใช่รายรายงาน**: โหลดจาก `bcrolepermission2` (`champ/champ/BC5Account.cpp:6159-6214`: `bCanPrintReport`, `bCanChangeReportOption`, `bCanExportReportData`), ตอบผ่าน `champ-lib/BC5FRMWRK/BCSWinApp.cpp:2709-2725,2769-2774`; จอ "กำหนดสิทธิ์การใช้งาน › เฉพาะ" ข้อ 1 "พิมพ์รายงานได้", ข้อ 6 "แก้ไขรายงานได้" (`bcspec/หน้าจอเมนูไฟล์/กำหนดสิทธิ์การใช้งาน-เฉพาะ.jpg`)
  - Drill-through ด้วยดับเบิลคลิก (`REPComm_InsertCommand`)
  - เก็บรูปแบบคอลัมน์ไว้ใน .rcf/.rex
  - AmountPoint ค่าเริ่มต้น = 2
- **หลายมุมมอง ใช้ View เดียว:** มุมมอง "ตามวันที่/ลูกค้า/แผนก/โครงการ" ใช้ View เดียว ต่างกันที่ `m_nSortBy` (`ChmRepAnaPurchaseView.cpp:105-140`) ซึ่งผู้ใช้เลือกจาก combo ใน dialog (`ChmRepAnaPurchaseDlg.cpp:95`) จึงใช้ engine เดียว + group-by ได้ตรง Champ; เมนูแต่ละมุมมองเป็นแค่ค่าเริ่มต้น (ล็อกเฉพาะ preset ที่เปลี่ยนตัวรายงาน เช่น ลูกหนี้/เจ้าหนี้, ซื้อ/ขาย — §3)
- **เงื่อนไขเฉพาะ GL:**
  - 5532/5533 ใช้ docno from/to + เรียง docno/date (`GLRepTransDlg.cpp:167-246`, อ้างที่ `contracts.go:67-70`)
  - ช่วงงบประมาณ (`GLRepBudgetDlg.cpp:244-251`)
  - ช่วงรหัส (`GLRepDeaprtmentDlg.cpp:87-98`)
- **ห้ามลอกจาก Champ** (§3.12):
  - ต่อ SQL ด้วย string / MakeUpper ทั้ง SQL
  - รวมยอดด้วย double
  - ดึงชื่อคู่ค้า N+1 / โหลดทุกแถวเข้าหน่วยความจำ
  - map ข้อความ→id ชนกัน
  - F7 Test SQL ที่ทุกคนเห็น
  - ตัวเลือกที่ปิดถาวรแต่ยังแสดง

### 1.1 ขั้นตอนประมวลผลรายงานของ Champ ที่ BC ทำตาม (กฎ 2026-10-11 "ประมวลผล/workflow ยึด Champ")
รายงานเป็นงานอ่านอย่างเดียว — Champ ไม่เขียนตารางบัญชีใดตอนออกรายงาน (เขียนแค่ไฟล์ .rcf/.rex ของเครื่อง) จึงไม่มีขั้น "ย้อนกลับ"

| ขั้น | Champ ทำอะไร (file:line) | BC ทำ | ต่างจาก Champ? |
|---|---|---|---|
| 1. เปิดรายงาน | `OnCreate` เรียก `OnFilterReport` ก่อนสร้างหน้าต่าง; dialog ไม่ OK = ไม่เปิด (`champ-lib/BOLEDBREPORT/BaseReportView.cpp:2309-2310`) | แผงเงื่อนไขก่อนรันเสมอ, Esc = ปิด ไม่รัน | ไม่ต่าง |
| 2. ประมวลผล | `RunProcess(CBCRepProcessDlg*)` วนเอกสารทีละใบ + แถบความคืบหน้า (`champ-report/BC5REP_AR/APRepAgingView.cpp:243`) | SQL ชุดเดียวใน `REPEATABLE READ READ ONLY` | ต่างเชิงเทคนิค ผลเท่ากัน (ไม่ N+1, ไม่ double) |
| 3. สถานะ/flag ที่นับ (ลูกหนี้/เจ้าหนี้) | เอกสารหนี้ `IsCancel = 0`, `IsCompleteSave = 1`, ยกเว้นใบรับสินค้ายังไม่ตั้งหนี้ `GRIRBillStatus <> 1`, `BillType` 1/3; รวมยอดยกมา `BCAPINVBALANCE` (`champ-report/BC5REP_AR/APRepAgingDlg.cpp:188-214`) | เอกสารใน `gl_subledger_documents` ที่ใบสำคัญผ่านรายการแล้ว ณ วันที่รายงาน และไม่ถูกกลับรายการ (`debt_report.go` `buildAgingQuery` CTE `posted_journals`) | **Q2 ตัดสินแล้ว 2026-10-11 = ทำตาม Champ**: นับทุกใบที่บันทึกแล้วและไม่ยกเลิก ไม่ดูสถานะผ่านบัญชี |
| 4. ตัดชำระ ณ วันที่ | หักยอดจ่ายที่ `IsCancel = 0 AND TransState = 0 AND DocDate <= ณ วันที่` ทีละใบ (`APRepAgingView.cpp:378-379`) | หัก `gl_subledger_settlements` ที่ `settlement_date <= ณ วันที่` และยังไม่ถอน | ต่างเชิงเทคนิค ผลเท่ากัน |
| 5. ยอดยกมา + ยอดสะสม | เคลื่อนไหวลูกหนี้: ยอดยกมา `dBeginBalance` แล้วบวกลบทีละแถว (`champ-report/BC5REP_AR/ARRepArTransMovementView.cpp:287-329`, ใบแจ้งหนี้ `posstatus = 0 and iscancel = 0`); บัญชีแยกประเภท: ยอดยกมา = `docdate < วันเริ่ม` (`champ-report/BC5REP_GLFS/GLRepGL2Dlg.cpp:434,442`) | `sum() OVER (...)` ก่อนตัดหน้า, ยอดยกมาจากชุดเดียวกัน | ไม่ต่าง (decimal แทน double) |
| 6. ตัวเลือกสถานะผ่านรายการ (GL) | combo "เฉพาะที่ผ่านรายการ / เฉพาะที่ยังไม่ผ่าน / ทั้งหมด" ค่าเริ่มต้น = ทั้งหมด (`GLRepGL2Dlg.cpp:215-222,428-436,617-618`); ลำดับตัวเลือกต่างกันรายรายงาน (`09-gl.md:497`) | **ยังไม่มี** — `gl_lines` เก็บเฉพาะที่ผ่านรายการ; ต้องเพิ่มพารามิเตอร์ `post_status` อ่านใบสำคัญร่างจาก `gl_records` | ทำตาม Champ (R2) |
| 7. ยอดรวมท้ายหน้า | ทุก `m_nRowPerPage − m_nMaxRowOfPageSummery` แถว เรียก `OnAddPageSummery` (`BaseReportView.cpp:558-564`); แถวต่อหน้าคิดจากความสูงกระดาษ (`:1550`); แต่ละรายงาน override เอง (723 view ใน `champ-report`) | backend แบ่ง "หน้าพิมพ์" ตาม `rows_per_page` ที่จอส่งมา แล้วคิดยอดท้ายหน้าเป็น decimal | ต่างเชิงเทคนิค ผลเท่ากัน |
| 8. สิทธิ์พิมพ์/ส่งออก + ข้อความเตือน | ไม่มีสิทธิ์พิมพ์ → "Can Not Print Report !!!" (`BaseReportView.cpp:736-745`); ไม่มีสิทธิ์ส่งออก → "Can not Export Data !!!" (`:2406-2415`); สิทธิ์รายผู้ใช้ (`BC5Account.cpp:6159-6214`) | สิทธิ์ `report:print` / `report:export` ระดับผู้ใช้ (ไม่ผูกรายเมนู) + ข้อความไทยผ่าน language key | ไม่ต่าง (ข้อความแปลเป็นไทย) |
| 9. ธนาคาร (5164) | รวมรายการสมุดบัญชีแต่ละประเภท `IsCancel = 0` เรียงตาม `MyDocOrder` (`champ-report/BC5REP_CASH/CASHRepBankStatementView.cpp:503-523`) | `gl_subledger_bank_lines` + `gl_lines.is_cash` | ต้องเทียบลำดับประเภทเอกสารตอนทำ R9 |
| 10. คำนวณยอดใหม่ | Champ เก็บยอดคงเหลือไว้ในตาราง จึงมีเมนู "คำนวณยอดคงเหลือใหม่" (ลูกหนี้/เจ้าหนี้) และ "คำนวณยอดผ่านรายการใหม่" (บัญชี) (`bcspec/หน้าจอเมนูลูกหนี้/คำนวณยอดคงเหลือใหม่-*.jpg`, `หน้าจอเมนูบัญชี/คำนวณยอดผ่านรายการใหม่.jpg`; GL เขียน `BCPeriodBalance` `09-gl.md:427`) | รายงานคิดยอดสดจาก `gl_lines`/subledger ใน snapshot เดียว ไม่พึ่งยอดเก็บ | ต่างเชิงเทคนิค ผลเท่ากัน — เมนูคำนวณใหม่ไม่อยู่ในงานนี้ |
| 11. F7 Test SQL | แสดง SQL ให้ผู้ใช้ (`10-reports.md` §3.12) | **ไม่ทำ** | ต่างโดยเจตนา (ความปลอดภัย) |

### เมนู Champ ที่ใช้ใน wave 1
- **GL** (`09-gl.md:31-77`):
  - 5532 รายงานข้อมูลรายวัน (`CGLRepTransView`) / 5533 แบบสรุป
  - 5534 บัญชีแยกประเภท (`CGLRepGL2View`) / 5535 แบบสรุป
  - 5506 งบทดลอง / 5507 งบทดลอง 2 / 5499 กระดาษทำการ
  - 5530 เปรียบเทียบงบประมาณ / 5505 รูปแบบงบการเงิน
  - (เพิ่มหลังตรวจทาน) engine GL มีอยู่แล้วและตรง Champ: 5525 รายละเอียดงบประมาณ / 5520 ผังบัญชี / 5531 สมุดรายวัน / 5538 แผนก / 5521 โครงการ (`generalledger/reports.go:89-123`; PDF ใน `bcspec/Report_ระบบบัญชี`)
- **ภาษี:**
  - 5536 ภาษีซื้อ / 5537 ภาษีขาย / 5517 ขายในประเทศ / 5518 ขายต่างประเทศ / 5539 สรุปยอดภาษี
  - 5522/5523 ยกเลิกข้ามงวด = **มุมมองของรายงานภาษีซื้อ/ขายเดิม** (`view=reversed_later`, `tax_report.go:45,300`) ไม่ใช่นิยามหรือจอใหม่ (ADR 2026-09-19 ยุบเป็นแท็บ, `menu-data.ts:156`)
  - 5541 ภ.ง.ด.3 / 5543 ภ.ง.ด.53 / 5540 ถูกหัก ณ ที่จ่าย — 5540/5541 และ 5522/5523 ไม่มี PDF ตัวอย่างใน `bcspec/Report_ระบบบัญชี` แต่เปิดใช้ใน `menuconfig.xml:608-613` (ไม่มี `disable="yes"`) ซึ่งโปรแกรมโหลดตอนรัน → ถือว่าอยู่ในโปรแกรมจริง ทำตาม; รูปแบบคอลัมน์อ้าง view ใน `champ-report/BC5REP_GLFS` (`GLRepInputTaxView.cpp:305-308` จัดการ `IsCancel==2` จาก POS)
  - อ้างอิงทางการ: RD-CODE-87, RD-VAT-P89 ข้อ 5, RD-VAT-P202 (`docs/kms/21-thai-tax-form-references.md:63-65,175-180`); คงการแยกสถานประกอบการ `tax_report.go:199-206` และคำเตือนรวมทุกสาขา `:259-262`
- **AR** (`04-ar.md:24-34`):
  - 5074-5079 ใบส่งของค้างชำระ (6 มุมมอง)
  - 5081-5086 ถึงกำหนด / 5087-5092 เกินกำหนด
  - 5093-5095 อายุลูกหนี้ / 5096-5097 เคลื่อนไหว / 5098-5100 สถานะ
  - 5138 รับชำระหนี้ประจำวัน
  - 5065 ตรวจสอบยอดวงเงิน = **จอ ไม่ใช่รายงาน** (`bcspec/หน้าจอเมนูลูกหนี้/ตรวจสอบวงเงินเครดิต.jpg`) ต้องมีวงเงินอนุมัติ+เช็ครับค้าง ซึ่ง BC ยังไม่มี → **รอพัฒนา** (มติลุงจืด 2026-10-11 ข้อ facades→รอพัฒนา) ไม่อยู่ใน wave 1
- **AP** (`03-ap.md:1104-1157`):
  - 5008/5020/5023/5024 จ่ายเงินประจำวัน / 5025-5028 ค้างจ่ายชำระ
  - 5016 อายุเจ้าหนี้ (`CAPRepAgingView`) / 5017 เคลื่อนไหว / 5018 แบบละเอียด / 5021 สถานะ
- **ธนาคาร-เงินสด** (`06-cash-bank.md:196-222`):
  - 5164 Bank Statement (`CCASHRepBankStatementView`)
  - 5152 เคลื่อนไหวเงินสด: คอลัมน์ วันที่ / เลขที่ / คำอธิบาย / รับ / จ่าย / คงเหลือ, dialog 35009
- **เช็ค** (`05-cheque.md:39-42`, `:318-378`): 5380-5416 **ไม่อยู่ใน wave 1**

> ต้องตรวจกับ `menuconfig.xml` ก่อนลงมือ R6/R10 ว่า id แต่ละตัวในช่วง 5074-5079 และ 5025-5028 ตรงกับมุมมองไหน

## 2. สถานะปัจจุบันของ BC

### Backend
- **binary เดียว** (`backend/main.go:159-166`): ms ทำงานกับ `fahttp`/`glhttp`/`txhttp` + goapi echo `/goapi/*`
- **IContext** มี `ResponseWriter()`/`EchoContext()` (`backend/pkg/microservice/context.go:20-30`) ส่ง binary ได้
- **GL engine** `(*Postgres).Report(ctx, scope, name, ReportQuery)` (`generalledger/reports.go:62-137`):
  - ทำงานใน RepeatableRead read-only
  - ตัวเลขเป็น string (`contracts.go:87-101`)
  - ป้ายคอลัมน์ไทยถูกเขียนตายตัว (`reports.go:230-237`)
- **สิทธิ์:**
  - GL ใช้ `reportScreens` (`httpapi/http.go:95`) + `allowed(p, screen, action)` ซึ่งตรวจ `p[screen]` และ `p[screen+":"+action]` (`http.go:97-105`)
  - `ResolveSessionAccess`/`.Allows` (`httpapi/scope.go:146-165`)
  - goapi tax ใช้ `taxCallerAllowed` (`tax_report.go:546`) ตรวจแค่ membership ไม่ตรวจรายจอ
- **ลูกหนี้/เจ้าหนี้มี payload 2 แบบในตารางเดียว `gl_subledger_documents`:**
  - ERP posting เขียน `{doc_no, doc_date, journal_id, total, description, due_date}` + คอลัมน์ `ledger/partner_code/control_account_code/amount/side` ไม่สร้าง allocation (`transaction/httpapi/posting.go:409-423`)
  - GL เขียน `document_no/document_date/ledger/balance_side` + allocation (`subledger_types.go:50`)
  - `generalledger/subledger_queries.go:46,48,156,158` กรอง `payload->>'document_date'` และ `payload->>'ledger'` + นับจาก allocation → **เอกสาร ERP มองไม่เห็น**
  - ฐานที่รวมสองแบบได้ถูกต้องมีแล้วที่ `debt_report.go` `buildAgingQuery` (`:202-262`: posted ณ วันที่, allocation, settlement, `COALESCE(doc_date, document_date)`, ฝั่งตามคอลัมน์ `side` หรือ `balance_side`)
- **goapi `/api/report/*`** (`goapi/bootstrap.go:353-372`): sales/*, tax/*, debt/query
- **ข้อมูลที่มีใน PG แล้ว:**
  - `gl_lines`: numeric(38,8), `is_cash`, dims, index period/account/dimension (`generalledger/schema.sql:7-29`)
  - `gl_records`: journal payload รวม VAT/WHT
  - `gl_subledger_*` (`subledger.sql`, `subledger_types.go:18-126`)
- **ยังไม่มี:**
  - ทะเบียนเช็ค: doc ไม่มี chequeno/duedate/bank (`transaction/httpapi/schema.go:13-74`)
  - doctype เช็คมีแค่ `chequereceived` 61 (`doctype/doctype.go:60-110`)
  - วงเงินเครดิตใน `SubledgerPartner` (`subledger_types.go:18-34`)
- **Export:**
  - `go.mod` ไม่มี xlsx; มี pdfcpu, shopspring/decimal, go-sqlmock, go-text/typesetting
  - `internal/pdftext`: NewOverlay/NewPage/Text/Fit/Check/Bytes (`pdftext.go:189`) ใช้เฉพาะแบบฟอร์มภาษีและ 50 ทวิ

### Frontend
- **route:** `menu-data.ts` มี `/report/*` 70 route + `/gl/workingpaper`
- **dispatch** ใน `main-menu-screen.tsx`:
  1. vatsummary (`:2305`)
  2. isThaiTaxRoute → TaxFilingWorkbench (`:2320`)
  3. isErpReportRoute → ErpReportViewer (`:2324`)
  4. GLReports (`gl/gl-reports.tsx:579`)
  - (เลขบรรทัดตาม working tree 2026-10-11 ซึ่งมีงานอื่นแก้ค้างอยู่ — ตรวจใหม่ก่อนลงมือ)
- **erp-reports.ts:**
  - มี 44 configs (`:72`) เปิดใช้ API ใน `API_READY_REPORTS` (`:834`) รวม `ar_credit_limit` (`:552`, `:848`) ซึ่งเป็น facade
  - client ตั้ง `limit: 1000` (`:1299`, `:1338`); ป้าย `th:` ฝัง inline 379 จุด
- **ระบบพิมพ์:** `gl/gl-print.tsx` (`useGLPrint` `:66-77`, `ReportPrintHeading:100`, `ReportPrintConditions:114`) ดีที่สุด; `gl-reports.tsx` มีหัวตารางสองชั้น (`:82-90`) และ `GLReportPrint` (`:479-501`) — ใช้สองไฟล์นี้เป็นฐานจอกลาง

### ปัญหา Rule 20 / ความถูกต้อง
> เลขบรรทัดตาม working tree 2026-10-11 (มีงานอื่นแก้ `debt_report.go`, `customer_labels.go`, `sales_report.go` ค้างไม่ commit) — ตรวจใหม่ก่อนลงมือ

1. ~~`sales_report.go:384,448` float64~~ — working tree แก้เป็น decimal แล้ว (งานอื่น ยังไม่ commit); R-test ต้องยืนยันซ้ำ
2. `frontend/src/app/report/erp-report-viewer.tsx:313` บวกยอดด้วย `Number(row[col.key])`
3. รั่วข้ามบริษัท `company = '01'`: `debt_report.go` และ `customer_labels.go:115` (HEAD) — working tree แก้เป็น `d.company = $1` + `errReportCompanyRequired` แล้ว (งานอื่น ยังไม่ commit); **ต้องเพิ่ม Go test สแกน handler ทุกไฟล์ห้ามมี `company = '01'`** กันกลับมา
4. `debt_report.go:560,575,618,632,652,664,680,695,738,752` ใช้ `payload->>'doc_date'` อย่างเดียว
   - แต่ GL บันทึกเป็น `document_date` (`subledger_types.go:50`); งาน aging COALESCE ทั้งสองตัว (`:216-217`)
   - ผล: เอกสารจาก GL หลุดตัวกรองวันที่; กลับกัน `subledger_queries.go` มองไม่เห็นเอกสาร ERP (ดู Backend ข้างบน)
5. `debt_report.go:565` มีข้อความไทย 'ซื้อเชื่อ'/'จ่ายชำระหนี้' ฝังใน SQL
6. ~~`< toDate`~~ — ตรวจแล้ว `:461` บวก 1 วันก่อนเทียบ = รวมวันสุดท้าย ไม่ใช่บั๊ก
7. ap_movement คำนวณ `balanceamount = d.amount` แทนยอดสะสม; ap_status รวมตาม `side` ปนความหมายฝั่ง ERP กับ `balance_side` ของ GL
8. client ดึงได้สูงสุด 1000 แถว → ยอดรวมและ export ไม่ครบ
9. export:
   - `gl-reports.tsx:706` ส่งออกเฉพาะแถวที่โหลด
   - CSV ไม่ escape
   - `erp-report-viewer.tsx:394` เรียก `window.print()` เปล่า
10. ป้ายและ error ไทยฝังในโค้ด: `reports.go:124,230-237`, `contracts.go:111,113`, `company_context.go:31-56`, `subledger_queries.go:74,87,99,102,152,164,179`
11. `debt_report.go:513-514` `rows.Scan` error แล้ว `continue` → แถวหายเงียบ ยอดรวมผิดโดยไม่มีใครรู้ ต้องคืน error
12. `debt_report.go:769-791` ar_credit_limit คืน `0 AS creditday` ไม่มีวงเงิน = facade (ดู 5065)
13. GL ไม่มีตัวเลือก "ผ่านรายการ/ยังไม่ผ่าน/ทั้งหมด" แบบ Champ (`contracts.go:44-79` ไม่มีฟิลด์นี้)

## 3. สถาปัตยกรรม

### แพ็กเกจใหม่ `backend/internal/report`
- `registry.go`, `params.go`, `totals.go`
- `defs_{gl,tax,debt,cash}.go`
- `source_gl.go`: adapter เรียก `generalledger.(*Postgres).Report` แบบ in-process
- `source_tax.go`: VatRegisterRecords / VatCrossPeriodCancellations / WHT
- `source_debt.go`: **ฐานเดียว** ของลูกหนี้/เจ้าหนี้ ยกจาก `buildAgingQuery` (`debt_report.go:202-262`) ใช้กับ ค้างชำระ/ถึงกำหนด/เกินกำหนด/อายุ/เคลื่อนไหว/สถานะ/รับ-จ่ายประจำวัน
  - อ่านคอลัมน์ตาราง (`ledger`, `partner_code`, `amount`, `side`) + `COALESCE(payload->>'doc_date', payload->>'document_date')`; ไม่แก้ `posting.go` ในงานนี้
  - ลบ query เดิม movement/status/outstanding/credit_limit ใน change เดียวกัน
- `source_cash.go`
- `export_csv.go` (xlsx เมื่ออนุมัติ Q3) — **ไม่มี PDF ฝั่ง server**: `pdftext` เป็นแค่ overlay แบบฟอร์ม ไม่มี layout ตาราง; พิมพ์ผ่าน browser ด้วย `gl-print.tsx`
- `paging.go`: แบ่งหน้าพิมพ์ + ยอดรวมท้ายหน้าตาม Champ (§1.1 ขั้น 7)
- `httpapi/http.go`
- `main.go` เพิ่มบรรทัดเดียว `reporthttp.NewHttp(ms, cfg)`

Definition เขียนเป็นโค้ด Go ไม่เก็บใน DB:
```go
type Definition struct { Code, Module, Family, Status /*ready|planned*/, Route string; Variants []Variant; Params []Param; Columns []Column; Sorts, Groups []Option; Totals TotalsSpec; Drill *Drill; Exports []string; Landscape bool; Source Source }
type Variant struct { MenuID string; ChampMenuID int; NameKey string; Preset map[string]string }
type Param struct { Key, Type, LabelKey string; Required bool; Default string; Enum []string }
type Column struct { Key, LabelKey, Type /*text|date|amount|qty|int*/, GroupKey string; Scale int32 }
```

### กฎการ run
- **scope:**
  - company มาจาก session เท่านั้น
  - ตัวกรองสาขาต้องอยู่ใน `access_scopes`
- **สิทธิ์ (แก้หลังตรวจทาน):**
  - `run` รับ `menu_id` เท่านั้น; server หา definition + variant จาก `menu_id` แล้ว **ล็อก preset ที่กำหนดตัวรายงาน** (เช่น ledger ar/ap, ทิศ VAT ซื้อ/ขาย, view) — client ส่งค่าทับ preset ที่ล็อก → 403 `REPORT_VARIANT_FORBIDDEN`
  - ตัวเลือกที่ Champ ให้ผู้ใช้เลือกใน dialog (เช่น combo เรียงลำดับ `ChmRepAnaPurchaseDlg.cpp:95` → `ChmRepAnaPurchaseView.cpp:105-140`) ยังเลือกได้
  - ดู: `Allows(variant.MenuID, "")` ราย menu
  - พิมพ์/ส่งออก: สิทธิ์ **ระดับผู้ใช้** `report:print` / `report:export` ตาม Champ (`BC5Account.cpp:6159-6214`, จอ "กำหนดสิทธิ์การใช้งาน-เฉพาะ.jpg") ไม่ผูกราย menu
  - `registry` คืนเฉพาะ menu_id ที่ผู้ใช้มีสิทธิ์ดู
- **ตัวเลือกสถานะผ่านรายการ (GL):** param `post_status` = `posted|unposted|all` ค่าเริ่มต้นตาม Champ รายรายงาน (§1.1 ขั้น 6); unposted อ่านใบสำคัญร่างจาก `gl_records`
- **validate ก่อนเรียก engine:** registry ตรวจ param ครบ/ชนิดถูกก่อน; error จาก engine map เป็น key (`report_error_*`) ไม่ส่งข้อความไทยของ engine ออกไป
- **transaction:** `REPEATABLE READ READ ONLY` + `SET LOCAL statement_timeout`
  - แถวและยอดรวมอยู่ใน snapshot เดียวกัน
- **paging:**
  - `LIMIT/OFFSET` + `ORDER BY` แบบมี tiebreak
  - limit 200 (สูงสุด 1000)
  - `count(*)` เป็น total_rows
- **ยอดสะสม:** ใช้ `sum() OVER (...)` ใน subquery ก่อนตัดหน้า
- **ยอดรวม:**
  - grand total = `SUM()` numeric ทั้งชุด
  - page/group total ใช้ decimal ใน Go
  - ปัด half-up ครั้งเดียวที่ scale ของคอลัมน์
  - ห้าม float
- **SQL:**
  - `$n` ทั้งหมด
  - sort/group ใช้ whitelist ที่ map เป็น SQL คงที่
  - ห้ามเงื่อนไขตามรหัสบัญชี ใช้ `is_cash` / `account_type` / bank `gl_account_code`
- **ภาษา:** label key ทับป้ายไทยของ GL engine (`reports.go:237`) และ `subledger_queries.go:152,164`; แปลด้วย `language.Text(key, lang)`
- **หน้าพิมพ์ + ยอดท้ายหน้า:** client ส่ง `rows_per_page` (คิดจากขนาดกระดาษ ตาม `BaseReportView.cpp:1550`); server คืน `page_totals[]` เป็น decimal string ต่อหน้าพิมพ์ (`:558-564`)
- **ปัดเศษ:** รวมค่าที่เก็บจริง (numeric) แล้วปัด half-up ครั้งเดียวที่ scale 2 (Champ `AmountPoint` = 2); ห้ามรวมค่าที่ปัดแล้ว
- **ข้อจำกัดแถว:** หน้าจอ 200/หน้า (สูงสุด 1000); export/พิมพ์ดึงครบทั้งชุดใน snapshot เดียว สูงสุด 100,000 แถว เกิน → error key ให้แคบเงื่อนไข
- **ต่างจาก Champ โดยเจตนา:** SQL ชุดเดียว+decimal (Champ N+1+double), ไม่มี F7 Test SQL, อ่านใน snapshot เดียว, คิดยอดสด (ไม่ต้องคำนวณยอดใหม่) — ดู §1.1

### Schema
Wave 1 ไม่ต้องสร้างตารางใหม่ มีแค่ index (ต้อง `EXPLAIN` ก่อน):
```sql
CREATE INDEX IF NOT EXISTS gl_lines_cash_idx ON gl_lines(company, entry_date, journal_id, line_no) WHERE is_cash;
CREATE INDEX IF NOT EXISTS gl_subledger_bank_lines_account ON gl_subledger_bank_lines(company, bank_account_code, journal_id, line_number);
CREATE INDEX IF NOT EXISTS gl_subledger_statement_account ON gl_subledger_statements(company, bank_account_code, transaction_date, id);
CREATE INDEX IF NOT EXISTS gl_subledger_settlement_date ON gl_subledger_settlements(company, settlement_date);
```
Wave 2 (Champ F9 บันทึกเงื่อนไข/.rcf — parity ไม่ต้องรออนุมัติ แต่ทำหลัง wave 1):
```sql
CREATE TABLE IF NOT EXISTS report_user_prefs (company text NOT NULL, user_uid text NOT NULL, report_code text NOT NULL, pref_name text NOT NULL DEFAULT '', payload jsonb NOT NULL, version bigint NOT NULL DEFAULT 1, updated_at timestamptz NOT NULL DEFAULT now(), updated_by text NOT NULL, PRIMARY KEY (company, user_uid, report_code, pref_name));
```

## 4. REST API

ใช้ ms `/report/v1/*` + BFF ใหม่ `frontend/src/app/api/report/[...reportPath]/route.ts`:
- allowlist registry|run|export
- JSON ผ่าน `proxyMainApiJson`
- binary ใช้แบบเดียวกับ `proxyFile` ใน `api/goapi/[...goPath]/route.ts:134-160`
- ส่ง `Accept-Language` ต่อ

### `GET /report/v1/registry?module=gl`
คืน `{success, data:{version, reports:[{code, module, family, status, route, variants:[{menu_id, champ_menu_id, name, preset}], params:[{key, type, required, label, default}], columns:[{key, label, type, scale, group}], sorts, groups, totals:{grand, page}, drill, exports, landscape}]}}`
- แสดงเฉพาะ variant ที่ผู้ใช้มีสิทธิ์
- รายงาน planned ก็ส่งมาด้วย

### `POST /report/v1/run`
request: `{menu_id, params:{date_from, date_to, account_from, post_status, ...}, sort, group, page, limit, rows_per_page}` — **ไม่รับ `code`**: server หา definition จาก `menu_id` และล็อก preset ของ variant (§3)

response: `{success, data:{columns, rows:[{_kind: detail|opening|group_total|day_total|control_total, ...string values}], totals:{debit:"98765.43"}, page_totals:[{print_page, debit:"..."}], total_rows, page, limit, header:{company_name, branch_name, title, conditions:[{label, value}], printed_by, printed_at}, warnings}}`
- โหมดพิมพ์ (`print=true`) ต้องมีสิทธิ์ `report:print` ระดับผู้ใช้ คืนทุกแถว + `page_totals` ทุกหน้าพิมพ์

### `POST /report/v1/export`
- body เหมือน run + `format: csv` (`xlsx` เมื่ออนุมัติ Q3) — **ไม่มี `pdf`** (พิมพ์ผ่าน browser)
- ต้องมีสิทธิ์ `report:export` ระดับผู้ใช้
- ตอบเป็น binary พร้อม `Content-Disposition` (ASCII + `filename*` UTF-8)
- export ทุกแถวฝั่ง server; CSV แบบ stream, escape ตาม RFC 4180, ตัวเลขเป็น decimal string

### Error
รูปแบบ `{success:false, code, message: language.Text(key, lang), field}`

| code | HTTP | language key |
|---|---|---|
| REPORT_UNKNOWN | 404 | report_error_unknown |
| REPORT_NOT_READY | 409 | report_error_not_ready |
| REPORT_FORBIDDEN | 403 | report_error_forbidden |
| REPORT_VARIANT_FORBIDDEN | 403 | report_error_variant_forbidden |
| REPORT_PRINT_FORBIDDEN | 403 | report_error_print_forbidden (Champ "Can Not Print Report !!!" `BaseReportView.cpp:736-745`) |
| REPORT_EXPORT_FORBIDDEN | 403 | report_error_export_forbidden (Champ "Can not Export Data !!!" `:2406-2415`) |
| REPORT_COMPANY_REQUIRED | 400 | report_error_company_required |
| REPORT_PARAM_REQUIRED | 422 | report_error_param_required |
| REPORT_PARAM_INVALID | 422 | report_error_param_invalid |
| REPORT_DATE_RANGE_INVALID | 422 | report_error_date_range_invalid |
| REPORT_SORT_INVALID / REPORT_GROUP_INVALID | 422 | report_error_option_invalid |
| REPORT_LIMIT_INVALID | 422 | report_error_limit_invalid |
| REPORT_EXPORT_FORMAT_UNAVAILABLE | 422 | report_error_export_format_unavailable |
| REPORT_EXPORT_TOO_LARGE | 422 | report_error_export_too_large |
| REPORT_ENGINE_FAILED | 500 | report_error_engine (แทนข้อความไทยของ engine) |
| REPORT_PROJECTION_PENDING | 409 | report_error_projection_pending |
| REPORT_TIMEOUT | 504 | report_error_timeout |
| REPORT_INTERNAL | 500 | report_error_internal |

## 5. Frontend

- **`frontend/src/lib/report-registry.ts`:**
  - `fetchRegistry` (cache ต่อ session)
  - `registryRouteCode(route)`
  - ~~adapter ไป ErpReportViewer~~ (ตัดหลังตรวจทาน — ErpReportViewer บวกเลขเอง/`window.print()` เปล่า)
- **`frontend/src/app/report/report-center.tsx`** (ชื่อไฟล์ภายใน ไม่ใช่ชื่อเมนู) — **สร้างจาก `gl/gl-reports.tsx` + `gl/gl-print.tsx`** (หัวตารางสองชั้น `:82-90`, `GLReportPrint` `:479-501`, `useGLPrint` `:66-77`):
  - เปิดแผงเงื่อนไขก่อน: Enter/F11 = รัน, Esc = ปิด, ไม่รันเอง (Champ `BaseReportView.cpp:2309-2310`)
  - preset ตาม variant (แสดงแต่แก้ไม่ได้ถ้าล็อก)
  - paging จาก server
  - ยอดรวมและยอดท้ายหน้าใช้ `totals`/`page_totals` ของ backend เท่านั้น ไม่บวกเลขเอง
  - พิมพ์ผ่าน browser ด้วย `ReportPrintHeading/Conditions` (`gl-print.tsx:100,114`) + แทรกยอดท้ายหน้าตาม `page_totals`
  - export CSV ผ่าน API; ปุ่มพิมพ์/ส่งออกซ่อนเมื่อไม่มีสิทธิ์ระดับผู้ใช้
  - drill-through
  - prop `language` + `useBackendLanguage`
- **`main-menu-screen.tsx`:** ถ้า route อยู่ใน registry → ReportCenter
  - เช็คก่อน `isErpReportRoute` (`:2324`)
  - route เดิมไม่เปลี่ยน
- **ทางเข้า = เมนูย่อยรายงานของแต่ละระบบตาม Champ** (§1): ห้ามเพิ่มเมนูบนสุด "รายงาน"/"ศูนย์รายงาน"
- **ห้าม dual path:** ย้ายรายงานไหนแล้ว ให้ลบของเดิมใน change เดียวกัน
  - config ใน `ERP_REPORT_CONFIGS` / `API_READY_REPORTS` / case ใน `fetchErpReportData`
  - endpoint goapi เดิม
  - ตัวเลือกใน GLReports และกิ่ง dispatch ใน `main-menu-screen.tsx` ที่ไม่มีใครใช้แล้ว
  - `ar_credit_limit` ออกจาก `API_READY_REPORTS` (`erp-reports.ts:848`) + ลบ query `debt_report.go:769-791`; เมนู 5065 → รอพัฒนา
- **เมนู:**
  - id เดิมคงไว้
  - variant ใหม่ใส่ `tx()` ใน `menu-data.ts` และ submenu `REPORTS_DAILY` / `REPORTS_ANALYSIS` ใน `menu-bars.ts`
  - ชื่อรายงาน Champ ที่เหลือใส่เป็น planned ("รอพัฒนา") ผ่าน menu-screen-status (R10)

### menu id ที่ต้องต่อใน wave 1
- **id เดิม:**
  - GL: gl-daily-report, general-ledger, trial-balance, working-paper
  - ภาษี: vat-buy, vat-sale, vat-summary-report, withholding-tax-report, withholding-tax-received
  - AR: ar-aging, ar-movement, ar-status, ar-outstanding, payment-daily
  - AP: ap-aging, ap-movement, ap-status, ap-outstanding, ap-daily-payment
  - ธนาคาร-เงินสด: bank-statement-report, cash-movement-report
  - alias: gl-ar-outstanding, gl-ap-outstanding
- **id ใหม่:**
  - gl-daily-summary-report (5533), general-ledger-summary (5535), trial-balance-2 (5507)
  - ar-movement-detail (5097), ap-movement-detail (5018)
  - wht-pnd3-report (5541), wht-pnd53-report (5543)
  - gl-budget-detail (5525), gl-chart-of-accounts-report (5520), gl-journal-books-report (5531), gl-department-report (5538), gl-project-report (5521) — ถ้ามี id เดิมใน `menu-data.ts` แล้วใช้ของเดิม
  - vat-cancel-cross-period-buy/sale (5522/5523) = variant ของ vat_purchase/vat_sale preset `view=reversed_later` (ไม่ใช่ definition ใหม่)

## 6. Wave 1: 22 นิยาม ครอบคลุมประมาณ 40 เมนู Champ (ปรับหลังตรวจทาน)

ข้อมูลอยู่ใน PG แล้ว ไม่พึ่งทะเบียนสินค้า; ไฟล์ PDF อยู่ใต้ `D:/project-champ/bcspec/`
- #10-#19 ทุกตัวใช้ **ฐานเดียว** ใน `source_debt.go` (ยกจาก `buildAgingQuery`) ที่เห็นทั้งเอกสาร GL และ ERP — ห้ามแต่ละรายงานเขียน SQL อ่าน payload เอง
- GL #1-#4c มีตัวเลือกสถานะผ่านรายการตาม Champ (§1.1 ขั้น 6)

| # | code | เมนู Champ | PDF ตัวอย่าง | แหล่งข้อมูล | route เดิม |
|---|---|---|---|---|---|
| 1 | gl_journal(+summary) | 5532, 5533 | Report_ระบบบัญชี/รายงานข้อมูลรายงาน.pdf, รายงานข้อมูลรายวันแบบสรุป.pdf | gl_lines/gl_records, GL gljournal | /report/gljournal |
| 2 | gl_ledger(+summary) | 5534, 5535 | รายงานบัญชีแยกประเภท.pdf, ...แบบสรุป.pdf | GL ledger | /report/ledger |
| 3 | gl_trial_balance(+gross) | 5506, 5507 | รายงานงบทดลอง.pdf, รายงานงบทดลอง2.pdf | GL trialbalance | /report/trialbalance |
| 4 | gl_working_paper | 5499 | รายงานกระดาษทำการ.pdf | GL workingpaper | /gl/workingpaper |
| 4a | gl_budget_comparison / gl_budget_detail | 5530, 5525 | รายงานเปรียบเทียบงบประมาณ.pdf, รายงานรายละเอียดงบประมาณ.pdf | GL budgetcomparison / budgetdetail | /report/budgetcomparison |
| 4b | gl_master_lists | 5520, 5531, 5538, 5521 | รายงานรายละเอียดผังบัญชี.pdf, ...สมุดรายวัน.pdf, ...แผนก.pdf, ...โครงการ.pdf | GL chartofaccounts / journalbooks / departments / projects | (ตรวจ menu-data) |
| 4c | gl_statement | 5505 | — (template งบ) | GL statement | (ตรวจ menu-data) |
| 5 | vat_purchase (+variant 5522 `view=reversed_later`) | 5536, 5522 | รายงานภาษีซื้อ.pdf | gl_records VAT (VatRegisterRecords, subledger_vat.go:398-500); reversed_later `tax_report.go:45,300` | /report/reportvatbuy |
| 6 | vat_sale (+variant 5523 `view=reversed_later`) | 5537, 5523 (+5517/5518 ถ้า engine แยกในประเทศ/ต่างประเทศได้ — ต้องตรวจ) | รายงานภาษีขาย.pdf, ...ในประเทศ.pdf, ...ต่างประเทศ.pdf | เหมือน #5 | /report/reportvatsale |
| 7 | vat_summary | 5539 | รายงานสรุปยอดภาษี.pdf | tax_vat_summary.go:73 | /report/vatsummary |
| ~~8~~ | ~~vat_cancel_cross_period~~ | ตัดหลังตรวจทาน — รวมเป็น variant ของ #5/#6 | | | |
| 9 | wht_register | 5541, 5543, 5540 | รายงานภาษีหัก ณ ที่จ่าย(ภ.ง.ด.53).pdf | WHT ใน gl_records (tax_report.go:590) | /report/wht-reports, /report/whtreceived |
| 10 | ar_outstanding | 5074-5079 | Report_ระบบลูกหนี้/รายงานใบส่งของค้างชำระ-ตามลูกค้า.pdf, -ตามวันที่.pdf | gl_subledger_documents ลบ settlements | /report/aroutstanding, /report/ar-outstanding |
| 11 | ar_aging | 5093-5095 | รายงานอายุลูกหนี้-ตามลูกหนี้.pdf | documents/settlements as-of; ช่วงอายุ 1-7/8-14/15-21/22-28/>28 (erp-reports.ts:43) | /report/araging |
| 12 | ar_movement(+detail) | 5096, 5097 | รายงานเคลื่อนไหวลูกหนี้.pdf | documents + settlements + ยอดยกมา | /report/armovement |
| 13 | ar_status | 5098-5100 | รายงานสถานะลูกหนี้-ตามลูกหนี้.pdf | partners + documents | /report/arstatus |
| 14 | ar_daily_receipt | 5138 | รายงานการรับชำระหนี้ประจำวัน.pdf | settlements ar | /report/reportdedebipaymentdaily |
| 15 | ap_outstanding | 5025-5028 | Report_ระบบเจ้าหนี้/รายงานใบส่งของค้างจ่ายชำระ-ตามเจ้าหนี้.pdf, -ตามวันที่.pdf | เหมือน #10 ฝั่ง ap | /report/apoutstanding, /report/ap-outstanding |
| 16 | ap_aging | 5016 | รายงานอายุเจ้าหนี้.pdf | เหมือน #11 | /report/apaging |
| 17 | ap_movement(+detail) | 5017, 5018 | รายงานเคลื่อนไหวเจ้าหนี้.pdf | เหมือน #12 | /report/apmovement |
| 18 | ap_status | 5021 | รายงานสถานะเจ้าหนี้.pdf | เหมือน #13 | /report/apstatus |
| 19 | ap_daily_payment | 5008, 5020, 5023, 5024 | รายงานการจ่ายเงินประจำวัน-ตามวันที่.pdf, -ตามเจ้าหนี้.pdf | settlements ap | /report/apdailypayment |
| 20 | cash_bank_statement + cash_movement | 5164, 5152 | Report_ระบบธนาคาร เงินสด/รายงาน Back Statement.pdf, รายงานการเคลื่อนไหวเงินสด.pdf | bank_accounts / bank_lines / statements; gl_lines is_cash | /report/bankstatement, /report/cashmovement |

**ยังไม่อยู่ใน wave 1:**
- เช็ค/บัตร (5380-5416): ยังไม่มีทะเบียนเช็ค (รอ ORD-06/11)
- 5065: เป็นจอ ไม่มีวงเงิน → รอพัฒนา
- 5081-5092: ทำ wave 2 ใช้ source #10 + due_date
- GL รายละเอียดสาขา/ฝ่าย/งาน/หน่วยงาน/การจัดสรร, แยกประเภทตามแหล่งที่มา (มี PDF ใน `bcspec/Report_ระบบบัญชี` แต่ engine ยังไม่มี): wave 2
- ภ.พ.30/36 และแบบ ภ.ง.ด.: เป็นแบบยื่น คงไว้ใน TaxFilingWorkbench
- รายงานที่มีเฉพาะ BC (pnl, balancesheet, annual-balances, daily-check, cashflow, cashflowforecast, project-pnl/dimensionpnl/projectsummary, dashboard/executivesummary/financialgraphs, allocate): ไม่ย้ายเข้า registry จนลุงจืดตอบ Q3 (§11)

**ภาษี:**
- registry ไม่เปลี่ยนคอลัมน์หรือสูตร ใช้ engine เดิมที่อ้างใน `docs/kms/21-thai-tax-form-references.md`: RD-CODE-82, RD-P80-2542, **RD-CODE-87, RD-VAT-P89 ข้อ 5, RD-VAT-P202** (`:63-65`), แท็บยกเลิกข้ามงวด+รายงานแยกสถานประกอบการ (`:175-180`)
- คงตัวกรองสาขา/สถานประกอบการ (`tax_report.go:199-206`) และคำเตือนเมื่อรวมทุกสถานประกอบการ (`:259-262`) ในทุก variant ภาษี

## 7. ผัง route ของ BC 70 ตัว → registry

| กลุ่ม (บรรทัดใน menu-data.ts) | route → สถานะ |
|---|---|
| GL 9 (`:660-671`) | gljournal / ledger / trialbalance / budgetcomparison → wave 1<br>pnl / balancesheet → รอ Q3 (รายงานเฉพาะ BC)<br>ar-outstanding / ap-outstanding → รอ Q3 (ถ้าคงไว้ = ชี้ไปรายงานเดียวกับ ar/ap_outstanding)<br>bank-unmatched → รอ Q3 (ส่วนเสริม BC) |
| ภาษี 13 (`:680-699`) | reportvatbuy / reportvatsale / vatsummary / wht-reports / whtreceived → wave 1<br>vatpp30 / vatpp36 / vatpnd2 / vatpnd3 / vatpnd53 / taxforms / whtcertificate → ไม่เข้า registry (แบบยื่น)<br>whtdeduction → ต้องตรวจก่อนว่าคืออะไร |
| AR 6 (`:375-380`) | 5 route → wave 1<br>arcreditlimit (`:380`) → รอพัฒนา (ลบ facade) |
| AP 5 (`:328-332`) | ทั้งหมด → wave 1 |
| เช็ค/ธนาคาร 8 (`:472-479`) | bankstatement / cashmovement → wave 1<br>expensesummary / pettycashmovement / monthlypaymentbook → wave 2<br>chequereceived / chequeissued / creditcard → planned |
| FA 5 (`:608-612`) | wave 2 |
| ซื้อ 3 + ขาย 7 | wave 3 (หลังแก้ float) |
| IC 14 (`:555-568`) | หลังทำทะเบียนสินค้า |

- ชื่อรายงาน Champ ที่ยังไม่มี → เมนู planned (R10) ตามตารางเทียบ R0 (PDF 415 + จอ ↔ `menuconfig.xml`) ไม่ใช้ตัวเลขประมาณ

## 8. Language keys

ครบ 13 คอลัมน์ th/en จริง ภาษาอื่นใส่อังกฤษ
- prefix ใหม่ไม่ชนของเดิม: report_center_ / report_col_ / report_param_ / report_error_ / report_export_ / report_status_ = 0 แถว
- key ของเมนูเดิมมีอยู่แล้ว: general_ledger, trial_balance, gl_daily_report, ar_aging, cash_movement_report

**จอ:**
- report_center_title
- report_center_run, report_center_close, report_center_conditions
- report_center_print, report_center_export_csv (ตัด `report_center_export_pdf` หลังตรวจทาน)
- report_center_no_data, report_center_loading
- report_center_page_total, report_center_grand_total
- report_center_printed_by, report_center_printed_at, report_center_page_of, report_center_rows_count
- report_status_planned

**พารามิเตอร์:**
- report_param_date_from, report_param_date_to, report_param_as_of, report_param_fiscal_year
- report_param_account_from/to, report_param_partner_from/to, report_param_doc_from/to
- report_param_branch, report_param_department, report_param_project, report_param_book
- report_param_sort, report_param_group
- report_param_summary, report_param_hide_zero, report_param_hide_opening, report_param_day_totals, report_param_gross
- report_param_aging_ranges, report_param_bank_account, report_param_direction, report_param_form
- report_param_post_status, report_post_status_posted, report_post_status_unposted, report_post_status_all (ถ้อยคำไทยตาม combo Champ `GLRepGL2Dlg.cpp:215-222`)

**คอลัมน์:**
- report_col_date, report_col_doc_no, report_col_description
- report_col_account_code, report_col_account_name
- report_col_debit, report_col_credit, report_col_balance, report_col_opening, report_col_closing
- report_col_partner_code, report_col_partner_name, report_col_tax_id, report_col_tax_branch
- report_col_base_amount, report_col_vat_amount, report_col_wht_rate, report_col_wht_amount
- report_col_due_date, report_col_amount, report_col_settled, report_col_remaining
- report_col_age_bucket, report_col_overdue_days
- report_col_receipt, report_col_payment, report_col_bank_reference
- report_col_kind_credit_purchase, report_col_kind_payment

**error:** 18 key ตาม §4 (เพิ่ม `report_error_variant_forbidden`, `report_error_print_forbidden`, `report_error_engine`)
- ป้าย/error ไทยใน `subledger_queries.go:74,87,99,102,152,164,179` และ `reports.go:124,237` ย้ายเป็น key ในงานเดียวกับที่ registry ห่อ engine นั้น

**เมนู:**
- gl_daily_summary_report, general_ledger_summary, trial_balance_2
- ar_movement_detail, ap_movement_detail
- wht_pnd3_report, wht_pnd53_report
- vat_cancel_cross_period_buy, vat_cancel_cross_period_sale (variant ของภาษีซื้อ/ขาย)
- key ของ 5525/5520/5531/5538/5521 ถ้ายังไม่มี (ถ้อยคำตามชื่อ PDF ใน `bcspec/Report_ระบบบัญชี`)

## 9. Tests

### Go (sqlmock)
- **registry_test:**
  - code และ menu_id ไม่ซ้ำ
  - ทุก key มีอยู่ใน languages.tsv ครบ 13 คอลัมน์
  - amount ต้องมี scale
  - sort/group อยู่ใน whitelist
  - golden `testdata/registry.json`
- **params_test:**
  - วันที่ผิด, from>to, enum นอก whitelist → 422
  - `account_from="1'; DROP TABLE gl_lines;--"` ต้องไม่โผล่ใน SQL
- **source_debt_test:**
  - มี `d.company = $1` และ**ไม่มี** `'01'`
  - COALESCE document_date/doc_date
  - to-date แบบ inclusive
  - ยอดสะสมแบบ window
  - scan error → คืน error ไม่ข้ามแถว
- **company_scope_guard_test** (สแกนซอร์ส `backend/internal/goapi/handlers/*.go` + `backend/internal/report/*.go`): ห้ามมี `company = '01'` / `OR $1 = ''`
- **integration PG (AR/AP):** ใส่เอกสาร GL 1 ใบ (`document_date`, allocation) + เอกสาร ERP 1 ใบ (`doc_date`, ไม่มี allocation) ของลูกค้าเดียวกัน → ทุกรายงาน #10-#19 เห็นทั้งสองใบ, ยอดคงเหลือ = amount − settlement ณ วันที่, ยอดรวมเท่ากับยอดบัญชีคุม
- **paging_test (ยอดท้ายหน้า):** `rows_per_page` 3 ค่า → ผลรวม `page_totals` = grand total ทุกครั้ง, ยอดยกไปหน้าถัดไปถูก, ค่า decimal string
- **post_status_test (GL):** posted/unposted/all ให้ชุดแถวต่างกันถูก, ค่าเริ่มต้นตาม Champ
- **totals_test:**
  - 0.1+0.2 = "0.30"
  - ขอบการปัด
  - ค่าติดลบ
  - ชุดว่าง → "0.00"
  - grand total = ทั้งชุด
  - TB เดบิต = เครดิต
- **export_csv_test:** มี BOM, escape ถูก, กัน formula injection เฉพาะคอลัมน์ข้อความ
- ~~export_pdf_test~~ (ตัด — ไม่มี PDF ฝั่ง server)
- **httpapi_test:**
  - ไม่มีสิทธิ์ดู menu_id → 403
  - ส่ง preset ทับค่าที่ล็อก (เช่น menu ภาษีซื้อแต่ส่ง direction=sale, menu ลูกหนี้แต่ส่ง ledger=ap) → 403 `REPORT_VARIANT_FORBIDDEN`
  - ไม่มีสิทธิ์พิมพ์/ส่งออกระดับผู้ใช้ → 403 print/export
  - registry ไม่คืน menu_id ที่ไม่มีสิทธิ์
  - error จาก engine ไม่มีข้อความไทยดิบเมื่อขอ en
  - planned → 409
  - ไม่รู้จัก → 404
  - en ได้ข้อความ en
- **integration PG:** เทียบกับ `SELECT sum(debit),sum(credit) FROM gl_lines WHERE company=$1 AND entry_date BETWEEN $2 AND $3`

### Vitest
- registry golden ↔ menu-data: id, route, key ภาษา ตรงกันและ dispatch ถูก
- report-center:
  - เปิดแผงเงื่อนไขก่อน
  - Esc / F11 ทำงาน
  - ห้ามใช้ Number/parseFloat กับ amount
  - export เรียก API
- BFF route: allowlist, binary, Accept-Language

### UAT (Demo → rungrueng → 01 → 00000, ใช้ psql ในฐานของ holding)
1. **Create:** journal 1,234.50 → `SELECT sum(debit),sum(credit) FROM gl_lines WHERE company='01' AND doc_no='<no>'`
2. **Read:** ตรวจเทียบกับ psql
   - gl_journal / ledger ยอดสะสม / cash_movement / TB เดบิต = เครดิต
3. **Update:** ทำรายการกลับ → ต้องได้แถวใหม่ ไม่ใช่แก้แถวเดิม
4. **AR/AP:**
   - ตั้งหนี้แล้วรับชำระบางส่วน
   - remaining = amount − sum(settlements)
   - daily receipt ตรง
5. **แยกบริษัท:** ข้อมูลบริษัท 01 ต้องไม่โผล่ในบริษัทอื่น
6. **Export/พิมพ์:** CSV ยอดตรงกับบนจอ; พิมพ์ผ่าน browser มียอดท้ายหน้า; payload เป็น string
6a. **AR/AP สองแหล่ง:** บันทึกขายเชื่อจากจอ ERP 1 ใบ + ตั้งหนี้จาก GL 1 ใบ → ค้างชำระ/อายุ/เคลื่อนไหว/สถานะ เห็นทั้งสองใบ เทียบ psql
7. **Edge:** ช่วงว่าง, `"=SUM(A1)",`, ดับเบิลคลิก, ภาษา en
8. **ล้างข้อมูล:** ลบตาม id เท่านั้น

## 10. ความเสี่ยง
- **ชนกับ workflow อื่น:** workflow อื่นกำลังแก้ debt_report.go, erp-reports.ts, erp-report-viewer.tsx, route.ts, menu-screen-status
  - R6/R8/R10 ต้องรอหลังเขาส่งงาน
  - R1-R5 เป็นไฟล์ใหม่ + main.go แค่ 1 บรรทัด
- **dual engine:** ถ้าไม่ลบของเดิมใน change เดียวกัน ยอดจะออกมาสองแบบ
- **แก้รั่ว '01':** ยอดที่ผู้ใช้เห็นจะเปลี่ยน (ถูกต้อง แต่ต้องแจ้ง); ตัวแก้อยู่ใน working tree ของงานอื่นยังไม่ commit — ถ้างานนั้นถูกทิ้ง รั่วกลับมา → guard test (§9) ต้องเข้าก่อน R8
- **สิทธิ์ variant ใหม่:** role ที่ไม่ใช่ `*` ต้องเพิ่มสิทธิ์เอง; สิทธิ์พิมพ์/ส่งออกระดับผู้ใช้เป็นของใหม่ใน BC → role เดิมต้องตั้งค่าเริ่มต้น (owner = ได้)
- **ฐาน AR/AP เดียว:** ยอดค้างที่ผู้ใช้เคยเห็นจะเปลี่ยน เพราะเอกสาร ERP/GL ที่เคยหลุดจะโผล่; ลุงจืดเลือก "นับตาม Champ" (Q2 2026-10-11) ยอดจึงไม่ผูกกับบัญชีคุม — ผลต่างกับบัญชีคุมต้องอธิบายได้จากเอกสารที่ยังไม่ผ่านบัญชี
- **ตัวเลือกสถานะผ่านรายการ (GL):** อ่านใบสำคัญร่างจาก `gl_records` เป็น payload JSON — ต้องมี test ว่ายอดร่างตรงกับยอดหลังผ่านรายการ
- **ยอดท้ายหน้า:** แถวต่อหน้าจริงขึ้นกับ browser/ฟอนต์; ถ้า CSS พิมพ์ตัดหน้าไม่ตรง `rows_per_page` ยอดท้ายหน้าจะไม่ตรงกับหน้ากระดาษ → ต้องบังคับ page-break ตามกลุ่มแถวที่ server แบ่ง
- **ประสิทธิภาพ:** ledger ทั้งปีหนัก ต้องใช้ timeout / stream / เพดานแถว / index
- **ป้ายของ GL engine:** ถ้า engine เปลี่ยนคอลัมน์ ต้องมี test จับ
- **เมนูรายงานจำนวนมาก:** ผู้ใช้ 40+ จะหายาก ต้องจัด submenu ตามจอ Champ จริง (ไม่ใช่เมนูกลางเดียว) + ช่องค้นหา
- **ตัวเลขจำนวนรายงาน:** PDF 415 ฉบับ vs menuconfig 366 รายการ ต้องทำตารางเทียบ R0 ก่อน; ห้ามเขียน "ครบ 100%"

## 11. Open decisions (ฉบับสุดท้ายหลังตรวจทาน — ไม่มีข้อใดขวางการเริ่มงาน)
| # | คำถามถึงลุงจืด | แนะนำ | ทางเลือก | หลักฐาน Champ |
|---|---|---|---|---|
| Q2 **[ตัดสินแล้ว 2026-10-11: ทำตาม Champ]** | รายงานลูกหนี้/เจ้าหนี้ นับเอกสารที่ยังไม่ผ่านบัญชีด้วยไหม | นับเฉพาะที่ผ่านบัญชีแล้ว (เหมือนจออายุหนี้ตอนนี้) ยอดจะตรงบัญชีคุม | ทำแบบ Champ: นับทุกใบที่บันทึกแล้วและไม่ยกเลิก | `APRepAgingDlg.cpp:188-214`, `APRepAgingView.cpp:378-379` ไม่ดูสถานะผ่านบัญชี |
| Q3 | รายงาน/เมนูที่ BC มีแต่ Champ ไม่มี (งบกำไรขาดทุน-งบดุลแบบสำเร็จ, งบกระแสเงินสด, แดชบอร์ด, ลูกหนี้-เจ้าหนี้ค้างในเมนูบัญชี, รายการธนาคารที่ยังไม่จับคู่) | เก็บไว้ ให้ชี้ไปรายงานชุดเดียวกัน และรายงานเป็นส่วนเสริม | ถอดออกให้เหลือเท่า Champ | ไม่มีใน `menuconfig.xml` / `bcspec/Report_ระบบบัญชี` |
| Q1 | เพิ่มตัวเขียนไฟล์ Excel ไหม | เพิ่ม เพราะ Champ ส่งออกได้ ระหว่างรอใช้ CSV | ส่งออก CSV อย่างเดียว | `BCSWinApp.cpp:2709-2725` สิทธิ์ "Export ข้อมูลของรายงานได้" |

ข้อที่ตัดออก (Champ หรือกฎตอบแล้ว): ที่ตั้ง API (ตัดสินใน §3), PDF ฝั่ง server + ยอดท้ายหน้า (Champ พิมพ์+ยอดท้ายหน้า `BaseReportView.cpp:558-564`), สิทธิ์ (Champ: ดูรายเมนู พิมพ์/ส่งออกรายผู้ใช้), F9/.rcf (parity wave 2, F7 ไม่ทำ), alias (รวมใน Q3), เมนู 366 (R0 ตารางเทียบ), การปัดเศษ (§3 half-up ครั้งเดียว scale 2), เพดานแถว (§3), ชื่อที่มีแค่ใน menuconfig (โปรแกรมโหลด menuconfig ตอนรัน `XMLMenuConfig.cpp:34-38` → ทำตาม)

## 12. ส่วนเสริมเล็ก (ต้องรายงาน)
- CSV BOM
- เวลา "ข้อมูล ณ"
- `_kind` ของแถว
- ไม่มีเมนูหรือจอใหม่นอกเหนือ Champ (ไม่มีเมนูบนสุด "รายงาน")
- ~~ปุ่ม PDF ฝั่ง server~~ (ตัด)

## 13. Work items (ปรับหลังตรวจทาน)
- **R0** ตารางเทียบรายงาน Champ: PDF 415 + จอ bcspec ↔ `menuconfig.xml` 366 ↔ route BC → กำหนด submenu และรายการ planned (ทำก่อน R6/R10)
- **R1** core registry + `run` รับ `menu_id` ล็อก preset + สิทธิ์ดูราย menu / พิมพ์-ส่งออกรายผู้ใช้ + error keys + `paging.go` ยอดท้ายหน้า + main.go
- **R2** GL source + นิยาม #1-#4c (รวม 5530/5525/5520/5531/5538/5521/5505) + param สถานะผ่านรายการ + ป้าย/error เป็น key
- **R3** CSV export (BOM, RFC 4180, กันสูตร, stream)
- ~~R4 PDF ฝั่ง server~~ (ตัด) → พิมพ์ผ่าน browser ใน R6
- **R5** BFF `/api/report` + `report-registry.ts`
- **R6** จอกลางสร้างจาก `gl-reports.tsx`/`gl-print.tsx` + dispatch + เมนูย่อยตาม Champ + ลบตัวเลือกเดิมใน change เดียว + แก้ `10-reports.md` §0/§1.2
- **R7** ภาษี #5-#9: 5522/5523 เป็น variant `view=reversed_later`, คงตัวกรองสถานประกอบการ+คำเตือน, อ้าง RD-CODE-87/RD-VAT-P89/RD-VAT-P202
- **R8** AR/AP ฐานเดียวจาก `buildAgingQuery` + #10-#19 + scan error คืน error + ลบ query เดิม/endpoint `/api/report/debt/query`/config ERP + ลบ facade 5065 → รอพัฒนา
- **R9** ธนาคาร-เงินสด 5164/5152 + index
- **R10** เมนู planned "รอพัฒนา" ตาม R0
- **R11** Rule 20: ลบ `Number()` ใน erp-report-viewer (sales_report ยืนยันว่า working tree แก้แล้ว)
- **R12** tests ตาม §9 (guard `'01'`, GL+ERP integration, variant forbidden, ยอดท้ายหน้า, สถานะผ่านรายการ)
- **R13** xlsx (หลัง Q1)

## ปรับปรุงหลังตรวจทาน (2026-10-11)
1. ไม่รับ — "ไม่พบไฟล์เอกสาร": ไฟล์มีอยู่แล้วและตรวจทานทั้งฉบับ
2. รับ — ทางเข้าเดียว "ศูนย์รายงาน" ผิด: Champ เข้าทางเมนูย่อยรายงานของแต่ละระบบ (จอ `เมนูโปรแกรมทั้งหมด.jpg`, `เมนู-ประกอบ.jpg`) → แก้ §1, §5; ห้ามเมนูบนสุดใหม่
3. รับ — จำนวน 366 vs 415: PDF 418 − แบบฟอร์ม 3 = 415 → เพิ่ม R0 ตารางเทียบ, ตัดตัวเลขประมาณ
4. รับ (ปรับวิธี) — payload ลูกหนี้/เจ้าหนี้ 2 แบบ: ใช้ฐานเดียวจาก `buildAgingQuery` ไม่แก้ `posting.go` ตอนนี้ + integration test GL+ERP
5. รับ — รั่ว '01' ใน `customer_labels.go`, scan `continue`, เลขบรรทัดเลื่อน: บันทึกว่า working tree แก้รั่วแล้ว + guard test + scan error ต้อง fail; ข้อ to-date ปิด (`:461` inclusive)
6. รับ — ป้าย/error ไทยจาก engine (`reports.go`, `subledger_queries.go`): key ทับป้าย + validate ก่อนเรียก engine + map error เป็น key
7. รับ — จอรายงานซ้อนสองชุด/GL แค่ 4 นิยาม: จอกลางสร้างจาก gl-reports/gl-print, เพิ่มนิยาม GL ที่ตรง Champ, รายงานเฉพาะ BC → Q3
8. รับ — ข้ามสิทธิ์ด้วย variant: `run` รับ `menu_id` + ล็อก preset + 403 `REPORT_VARIANT_FORBIDDEN`; พิมพ์/ส่งออกเป็นสิทธิ์รายผู้ใช้ตาม Champ
9. รับ — PDF ฝั่ง server เกินขอบเขต: ตัด R4/ปุ่ม PDF, พิมพ์ผ่าน browser, ยอดท้ายหน้าคิดที่ server ตาม `BaseReportView.cpp:558-564`
10. รับ — อ้างอิงภาษีไม่ครบ: เพิ่ม RD-CODE-87/RD-VAT-P89 ข้อ 5/RD-VAT-P202, คงตัวกรองสถานประกอบการ, 5522/5523 เป็น variant ไม่ใช่นิยามใหม่
11. รับ — 5065 เป็น facade: ย้ายเป็นรอพัฒนา ลบ `ar_credit_limit` ออกจาก API_READY และลบ query
- เพิ่มเอง (กฎยึด Champ): §1.1 ขั้นตอนประมวลผลพร้อม file:line, ตัวเลือกสถานะผ่านรายการของ GL, สิทธิ์พิมพ์/ส่งออกรายผู้ใช้, ระบุข้อต่างโดยเจตนา (SQL ชุดเดียว+decimal, ไม่มี F7, snapshot, ยอดสด, ฐาน AR/AP นับทุกใบที่บันทึกตาม Q2)
