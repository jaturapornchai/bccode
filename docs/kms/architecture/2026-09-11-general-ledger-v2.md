# บัญชีแยกประเภทใหม่สำหรับการอัปเกรด Champ

เปิดใช้งานครั้งแรก 2026-09-11 (35 เมนู) — **ตั้งแต่ 2026-09-23 GL รันบน PostgreSQL ล้วน แบบ synchronous ACID transaction เท่านั้น** (`backend/internal/generalledger/httpapi/http.go:41`) หลังถอด MongoDB/Kafka/Redis/ClickHouse ออกจากระบบตาม ADR [ถอด MongoDB/Kafka/Redis/ClickHouse](../decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md) — ห้ามเพิ่มกลับ

## วัตถุประสงค์และหลักฐาน

ใช้ลำดับงาน Champ เป็นหลัก: ข้อมูลหลักและยอดยกมา → สมุดรายวัน → ตรวจ/ผ่านรายการ → รายงาน → ปิดงบและสิ้นปี โดยเพิ่มการทำงานบนเว็บ การตรวจสิทธิ์บริษัท/สาขา จำนวนเงินแม่นยำ และประวัติที่ตรวจสอบได้

- Champ `D:/project-champ/champ/champ/menuconfig.xml:552` เป็นที่มาของลำดับเมนู; `glfrmjournal.cpp:478` และ `:766` ตรวจเอกสารซ้ำ/เดบิตเครดิต; `glfrmcloseperiod.cpp:237` ใช้บัญชีกำไรขาดทุนและกำไรสะสมแยกกัน
- `GLReProcess.cpp:130` สร้างบัญชีจากเอกสารต้นทาง ส่วน `GLDlgReprocess.cpp:196` คำนวณยอดผ่านรายการใหม่ เป็นคนละกระบวนการ
- `YearEndProcess_Page1.cpp:559` มีการลบประวัติในระบบเดิม ระบบใหม่นี้เก็บประวัติและสร้างยอดยกมาในปีถัดไป
- เมนู GL อยู่ใน `MENU_SECTIONS` หมวด `gl` (`frontend/src/lib/menu-data.ts:587`); จุดเข้าหน้าจอ `frontend/src/app/gl/general-ledger-screen.tsx:24`

## วิธีใช้งาน

1. เปิด **ผังบัญชี** สร้างบัญชีคุม/บัญชีลงรายการ พร้อมหมวดและด้านบัญชี ระบุบัญชีเงินสดเพื่อใช้รายงานกระแสเงินสด
2. เมนู **ปีบัญชีและบัญชีปิดปี** (`/gl/fiscal-years`) กำหนดช่วงวัน สกุลเงิน ทศนิยม และบัญชีทุนสองบัญชีสำหรับปิดกำไรขาดทุน/กำไรสะสม จากนั้นสร้างงวดใน **กำหนดงวดบัญชี** (`/gl/periodlock`)
3. บันทึกยอดยกมาในวันแรกของปี หรือบันทึกใน **สมุดรายวัน** (`/gl/journals`) ตามสมุดที่ผู้ใช้กำหนดเองใน **กำหนดสมุดรายวัน** (`/gl/journal-books`) ใส่จำนวนเงินเป็นเลขทศนิยมไม่ใส่จุลภาค เดบิตและเครดิตต้องเท่ากันก่อนบันทึก
4. ตรวจร่างและผ่านรายการ รายงานนับเฉพาะรายการที่ผ่านแล้ว แก้จำนวนเงินหรือลบรายการที่ผ่านแล้วไม่ได้ การกลับรายการสร้างเอกสารใหม่ที่สลับเดบิตเครดิตและเก็บต้นฉบับ
5. ปิดงบ: ตรวจยอดถึงวันที่เลือก สร้างฉบับร่างแยกสาขา ตรวจและผ่านรายการ แล้วล็อกงวดเอง ทุกแผนก/โครงการคงเดิม
6. สิ้นปี: ผ่านรายการร่างและปิดรายได้/ค่าใช้จ่ายให้หมด เลือกปีถัดไปที่ต่อเนื่อง สกุลเงินและทศนิยมตรงกัน ระบบปิดปีเดิมพร้อมสร้างร่างยอดยกมา ถ้าทุกยอดเป็นศูนย์จะปิดปีโดยไม่มีเอกสารเปล่า ยอดยกมาที่ระบบสร้างแก้/ลบไม่ได้ ให้ผ่านรายการและปรับปรุงด้วยรายการแยก
7. รายงานเลือกปีและช่วงวัน กรองบัญชี/สาขา/แผนก/โครงการ ส่งออก CSV ได้ครบทุกหน้า (`frontend/src/app/gl/gl-reports.tsx:394`)

## สัญญาจำนวนเงินและการบันทึก

- JSON ใช้ decimal string; Go ใช้ `shopspring/decimal`; PostgreSQL ใช้ `NUMERIC(38,8)` สำหรับเดบิต/เครดิตแต่ละบรรทัด (`backend/internal/generalledger/amount.go:1`, `backend/internal/generalledger/schema.sql:24`)
- ข้อมูลเข้าไม่เกิน 26 หลักหน้าจุดและ 8 ตำแหน่ง (`amountPattern` ใน `amount.go:15`); ปีบัญชีกำหนด scale 0–8 ไม่ปัดเศษให้อัตโนมัติ จำนวนเงินเกิน scale ถูกปฏิเสธ (`amount.go:63-71`)
- รหัสเหตุการณ์ (`gl_events.id`) เป็น `<company>-<sequence 12 หลัก>`; ลำดับ (`sequence`) นับต่อบริษัทใน `gl_projection_state` และล็อกด้วย `pg_advisory_xact_lock` ทุกธุรกรรมกันชนกัน (`backend/internal/generalledger/postgres_store.go:160-174`, `postgres.go:83-89`)
- `requestid` เดิมพร้อมข้อมูลเดิมคืนผลเดิม (idempotency ตรวจจาก `gl_events.payload->>'requestid'` เทียบ hash คำสั่ง) ถ้าข้อมูลต่างกันให้ปฏิเสธ; เอกสารใช้ version กันเขียนทับพร้อมกัน (`postgres_store.go:106-125`)
- ปีที่มีรายการหรืองบประมาณ/ประมาณการแล้วห้ามเปลี่ยนสกุลเงิน/ทศนิยม ปีปิดแล้วห้ามแก้ ลบ หรือย้ายปี (`postgres_guards.go:250`)

## การประมวลผลและสิทธิ์

- `/api/gl/*` เป็น Next.js BFF ไป `/gl/v2/*` ของ main API; ตรวจ membership/holding/company/branch และ permission ปัจจุบันจากต้นทาง สิทธิ์เข้าเมนูไม่เท่ากับสิทธิ์ create/update/delete (`backend/internal/generalledger/httpapi/permissions.go`, `permissions_test.go`)
- PostgreSQL ฐานตาม holding เก็บทุกทรัพยากรใน `gl_records` (แยก company/kind/id), รายการผ่านบัญชีใน `gl_lines`, ประวัติแบบ append-only ใน `gl_events`, ลำดับต่อบริษัทใน `gl_projection_state` (`backend/internal/generalledger/schema.sql:1-29`)
- งบประมาณ (Champ 5500 `BCGLBudget`) แยกจาก `gl_records`: หัวงบ `gl_budgets` + ยอดรายเดือน `gl_budget_lines` (บัญชี × งวด 1–12, `numeric(18,2)`) ใน `backend/internal/generalledger/budget.sql`; คำสั่ง `resource: "budgets"` = `create`/`update`/`delete` ผ่าน `Execute` (requestid, company lock, audit ใน `gl_events`) + `spread` คำนวณแบ่งยอดทั้งปีเป็น 12 เดือนโดยไม่บันทึก; `GET /gl/v2/budgets[/code]`; รายงาน `budgetcomparison` (Champ 5530) เทียบกับ `gl_lines` ตาม `normal_balance` ของบรรทัด — ADR [2026-09-25-gl-monthly-budget](../decisions/2026-09-25-gl-monthly-budget.md)
- **ทุกคำสั่งเขียนในธุรกรรมเดียว** (`BeginTx` → ล็อกบริษัท → ตรวจ idempotency → mutate `gl_records`/`gl_lines` → เพิ่ม `gl_events` พร้อม sequence ใหม่ → `Commit`) ไม่มี outbox ไม่มี relay ไม่มี broker คั่นกลาง (`backend/internal/generalledger/postgres_store.go:58-229`)
- `Ready()` ไม่ทำอะไรและคืนสำเร็จเสมอ, `ProjectionPending` เป็น `false` เสมอเพราะเขียนสำเร็จ = อ่านได้ทันที (`postgres_store.go:42-44,211`); โค้ด error `GL_PROJECTION_PENDING` (`contracts.go:81`) และ retry ฝั่ง frontend (`frontend/src/lib/general-ledger-api.ts:124`) ยังอยู่เพื่อความเข้ากันได้ของ contract แต่ในโหมดนี้ไม่ควรถูกเรียกใช้แล้ว
- รายงานอ่าน PostgreSQL แบบ repeatable-read; export ตรวจ sequence ทุกหน้าและทุก resource ป้องกันไฟล์รวมข้อมูลต่างเวลาปะปน
- ปิดงบและสิ้นปีอ่านยอดแยกบัญชี/สาขา/แผนก/โครงการภายในธุรกรรมเดียวกัน ไม่บังคับสมดุลต่อแผนก/โครงการเพราะบางบรรทัดต้นฉบับอาจไม่ได้กำหนดมิติ (`processes.go`, `reports_process.go`)
- คำนวณยอดผ่านรายการใหม่ replay จากประวัติ `gl_events` ที่แก้ไม่ได้ (append-only มี trigger ป้องกันแก้/ลบ/truncate)

## ข้อผิดพลาด UI ที่พบจากภาพ UAT

กฎ CSS กลางนอก Tailwind layer กำหนด section กว้าง 100% ทำให้ pane รายการกินพื้นที่ทั้งหมดและบีบฟอร์มบัญชี แก้เฉพาะ GL โดยใช้ div เป็น pane พร้อม min-w-0; เพิ่มตรวจขนาด pane และดูภาพจริง การเช็คว่าไม่มี document overflow อย่างเดียวจับปัญหานี้ไม่ได้ (`frontend/src/app/globals.css:383`, `panel` ใน `frontend/src/app/gl/gl-common.tsx:47`)

## การตั้งค่าและ dependency

Dependency ปัจจุบัน: Go, PostgreSQL, `shopspring/decimal`, `lib/pq`, Next.js/React, BigInt และ Playwright — ไม่มี MongoDB driver, Kafka client หรือ Redis ในระบบนี้แล้ว (ถอดออกจาก `backend/go.mod` ทั้งหมดเมื่อ 2026-09-23)

ต้องมีฐาน PostgreSQL ชื่อ holding ตัวพิมพ์เล็กก่อนเปิด GL; package สร้างตารางและ index เองจาก `schema.sql`, `budget.sql`, `subledger.sql` ไม่เพิ่มรหัสผ่านใน repo ส่วน `BCAI_LOCAL_BACKEND_URL` ใช้กำหนด backend ของ frontend ตอนเริ่ม/build

## ข้อจำกัดที่ยังต้องมีข้อมูลต้นทาง

- **เชื่อมข้อมูลจากเอกสารซื้อขาย**: ยังไม่มี — backend เอกสารขาย/ซื้อเดิมถูกถอดพร้อม MongoDB เมื่อ 2026-09-23; GL รับข้อมูลผ่านสมุดรายวัน/Subledger ของตัวเอง และการนำเข้าใบสำคัญจากระบบภายนอกที่ระบุ `source_system`+`source_record_id` (`backend/internal/generalledger/postgres_source.go`) หากทำทางเชื่อมในอนาคตต้องใช้จำนวนเงินแบบ decimal ห้ามใช้ float
- **ส่งออก XBRL**: ไม่อยู่ในเมนู GL ปัจจุบัน (`/report/xbrl` เหลือเฉพาะใน catalog `frontend/src/lib/erp-reports.ts:498` สถานะรอข้อมูล) ยังไม่สร้างไฟล์สำหรับยื่น ต้องมีแม่แบบและประเภทกิจการ/มาตรฐานที่ตรงกับบริษัท และตรวจด้วยตัวตรวจรับของ DBD ก่อน [คู่มือ DBD](https://efiling.dbd.go.th/efiling-documents/01_ManualFN.pdf) หน้า 6–10 ระบุการดาวน์โหลดแม่แบบแยกกิจการและขั้นตอนสร้างไฟล์
- งบ/กราฟและ cash flow เป็นรายงานจากบัญชีและประเภทเงินสดที่ผู้ใช้กำหนด ไม่ใช่แม่แบบงบที่ผ่านการรับรอง DBD; ประมาณการเงินสดเป็นแผนที่ผู้ใช้บันทึกเอง
- เอกสารไม่เกิน 500 บรรทัด; process ไม่เกิน 10,000 กลุ่มมิติ และต้องสร้างเอกสารที่สมดุลแต่ละสาขาภายในขนาดที่รองรับ; export ไม่เกิน 100,000 รายการ ไม่มีการส่งออกบางส่วนเงียบ ๆ

## ผลตรวจและวิธีทดสอบ

หลักฐานรอบ 2026-09-11 (`docs/evidence/2026-09-11-*`) ทดสอบบนสถาปัตยกรรมที่ถอดไปแล้ว — เก็บเป็นประวัติเท่านั้น ห้ามใช้เป็นสถานะปัจจุบันหรือ how-to

**ทดสอบปัจจุบัน:** integration test ใช้ฐาน PostgreSQL แยกด้วย `BC_GL_TEST_POSTGRES_DSN` (ไม่ตั้ง = ข้าม, `backend/internal/generalledger/postgres_integration_test.go:21-23`) build/test บน Windows ได้

```powershell
# จาก backend; DSN ต้องเป็นฐานทดสอบแยกเท่านั้น
$env:BC_GL_TEST_POSTGRES_DSN='postgres://postgres@127.0.0.1:15443/postgres?sslmode=disable'
go test -tags=integration ./internal/generalledger/...
```

ตัวอย่างยอดยืนยัน: `0.1 + 0.2 = 0.3` และ `0.00000001 + 0.00000002 = 0.00000003`; กำไร `300.50 − 50.30 = 250.20` คงเดิมหลังปิดงบ พร้อม trial balance รายได้/ค่าใช้จ่ายปิดเป็นศูนย์และยอดทุกมิติยกไปปีใหม่ตรงกัน (ตรวจซ้ำใน `amount_precision_audit_test.go`, `postgres_precision_integration_test.go`)

## การย้อนกลับ

Deploy สำรอง PostgreSQL (`pg_dumpall`) + config (`release.env.before`) ก่อนสลับ release ทุกครั้ง (`tools/fast-deploy.py:113-151`) ย้อนรุ่นด้วย release เดิมและ backup ชุดนั้น

## เมนูทั้งหมดและสถานะ

เมนู GL ปัจจุบัน 23 รายการ (กลุ่ม `gl-master`, `gl-journals`, `gl-posting`, `gl-reports` ใน `frontend/src/lib/menu-data.ts:587`) — ชุด 35 เมนูของ 2026-09-11 ถูกปรับตาม Champ เมื่อ 2026-09-19 (ADR [champ-parity-menu-cut](../decisions/2026-09-19-champ-parity-menu-cut.md)); รายงานที่ตัดออกจากเมนู (เช่น cash flow, project P&L, dashboard) ยังมีตัวคำนวณใน `backend/internal/generalledger/reports.go:43-67` เมนูภาษีอยู่หมวดเดียวกันตาม ADR [ภาษีอยู่ในบัญชีแยกประเภท](../decisions/2026-09-23-tax-inside-general-ledger.md)

สถานะเชื่อมแล้วหมายถึงมีหน้าจอและ API ในโค้ดชุดนี้ ไม่ได้หมายถึงได้รับรองรูปแบบงบตามกฎหมาย

| เมนู | เส้นทาง | สถานะ |
|---|---|---|
| รายละเอียดผังบัญชี | `/gl/chartofaccounts` | เชื่อมหน้าจอและ API แล้ว |
| บันทึกกลุ่มผังบัญชี | `/gl/account-groups` | เชื่อมหน้าจอและ API แล้ว |
| บันทึกยอดสะสมประจำปี | `/gl/openingbalance` | เชื่อมหน้าจอและ API แล้ว |
| กำหนดงบประมาณ | `/gl/budget` | API งบรายเดือนพร้อม (`budgets`); จอบันทึกรายเดือนยังไม่ทำ — เมนูขึ้น "ยังไม่พร้อม" (`frontend/src/lib/menu-screen-status.ts:34`) |
| กำหนดรูปแบบการเชื่อม | `/gl/account-mapping` | เชื่อมหน้าจอและ API แล้ว |
| กำหนดกลุ่มบัญชีสินค้า | `/gl/product-account-groups` | เชื่อมหน้าจอและ API แล้ว |
| ปีบัญชีและบัญชีปิดปี | `/gl/fiscal-years` | เชื่อมหน้าจอและ API แล้ว |
| สมุดรายวัน | `/gl/journals` | เชื่อมหน้าจอและ API แล้ว |
| กำหนดสมุดรายวัน | `/gl/journal-books` | เชื่อมหน้าจอและ API แล้ว |
| กำหนดงวดบัญชี | `/gl/periodlock` | เชื่อมหน้าจอและ API แล้ว |
| ปิดงวดบัญชี | `/gl/financialclose` | เชื่อมหน้าจอและ API แล้ว |
| ผ่านรายการและยกเลิก | `/gl/posting` | เชื่อมหน้าจอและ API แล้ว |
| คำนวณยอดผ่านรายการใหม่ | `/gl/recalculate-posted` | เชื่อมหน้าจอและ API แล้ว |
| ประมวลผลข้อมูลบัญชีใหม่ | `/gl/reprocess` | ประมวลผลใหม่จากประวัติใบสำคัญ (`frontend/src/app/gl/gl-processes.tsx:14`) |
| ประมวลผลสิ้นปี | `/gl/year-end` | เชื่อมหน้าจอและ API แล้ว |
| รายงานข้อมูลรายวัน | `/report/gljournal` | เชื่อมหน้าจอและ API แล้ว |
| รายงานบัญชีแยกประเภท | `/report/ledger` | เชื่อมหน้าจอและ API แล้ว |
| รายงานงบทดลอง | `/report/trialbalance` | เชื่อมหน้าจอและ API แล้ว |
| รายงานกระดาษทำการ | `/gl/workingpaper` | เชื่อมหน้าจอและ API แล้ว |
| รูปแบบงบการเงิน | `/gl/statement-designer` | เชื่อมหน้าจอและ API แล้ว |
| งบกำไรขาดทุน | `/report/pnl` | เชื่อมหน้าจอและ API แล้ว |
| งบดุล | `/report/balancesheet` | เชื่อมหน้าจอและ API แล้ว |
| รายงานเปรียบเทียบงบประมาณ | `/report/budgetcomparison` | เชื่อมหน้าจอและ API แล้ว |
