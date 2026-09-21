# Handoff 2026-09-20 → Codex: ระบบบัญชีแยกประเภท (GL) สำหรับสำนักงานบัญชี พร้อมระบบภาษีและเอกสารราชการครบวงจร

> **เป้าหมาย:** ออกแบบและพัฒนาระบบบัญชีแยกประเภท (General Ledger - GL) ของ BC Ai Account ให้ **"สำนักงานบัญชีใช้ตัวเดียว ได้เอกสารทางราชการครบถ้วน 100%"** โดยยึดหลัก **Single Point of Entry (คีย์ใบสำคัญรายวันใบเดียว วิ่งไปออกรายงานและไฟล์ยื่นราชการทั้งหมดโดยไม่ต้องคีย์ซ้ำ)**

---

## 1. บริบทและสิ่งที่ดำเนินการแล้ว (Current State)

1. **โครงสร้างฐานข้อมูล GL หลัก (PostgreSQL 18 ใน Production `159.223.43.229`):**
   - ตรวจสอบและรันสคริปต์ DDL ทั้ง 4 ไฟล์จาก `mydocs/datamodels/gl/` ผ่านสมบูรณ์ 100%:
     - `chartofaccount.sql`: ผังบัญชี 5 หมวด, โครงสร้าง 12 ระดับ, Self-referencing FK `(company_code, parent_account_code) ON DELETE RESTRICT`
     - `journalbook.sql`: สมุดรายวัน (JV, PV, RV, SV, UV, OPENING)
     - `period.sql`: งวดบัญชี `fiscal_periods` (1-12, ปี พ.ศ./ค.ศ.)
     - `gltrans.sql`: ใบสำคัญรายวัน `journal_entries` และรายการเดบิต/เครดิต `journal_lines` พร้อมฟิลด์ `period_year`, `period_no`, `status` (1=ร่าง, 2=ปกติ/Posted, 3=ยกเลิก), `posted` (0/1)
2. **การสืบค้นและเทียบเคียงโค้ดต้นแบบ Champ (`D:\project-champ`):**
   - Champ ใช้การผูกภาษีเข้ากับใบสำคัญรายวันโดยตรง:
     - ภาษีมูลค่าเพิ่ม: ใช้คลาส `_MULTITAX_LIST` และไดอะล็อก `CLIBDlgMultiTax` (ปุ่ม `IDC_BUTT_TAX` -> `OnInputTax`)
     - ภาษีหัก ณ ที่จ่าย: ใช้คลาส `_WTAX_LIST` และไดอะล็อก `CLIBDlgARWTax`
     - รายงานภาษีจัดอยู่ในกลุ่ม GL & Financial Statements (`champ-report/BC5REP_GLFS`): `GLRepInputTax` (ภาษีซื้อ), `GLRepOutputTax` (ภาษีขาย), `GLRepSumTax` (สรุป ภ.พ.30)
3. **การตรวจสอบระเบียบราชการไทยล่าสุด (สรรพากร, DBD, พ.ร.บ.การบัญชี):**
   - มาตรา 87 แห่ง ป.รัษฎากร (รายงานภาษีซื้อ ม.87(2), รายงานภาษีขาย ม.87(1)) ตามประกาศอธิบดีฯ ฉบับที่ 89
   - สิทธิยื่นภาษีซื้อย้อนหลังไม่เกิน 6 เดือนตามมาตรา 86/10
   - รูปแบบไฟล์ **RD Prep Format กลาง (UTF-8, Pipe `|` Delimited)** ของกรมสรรพากรสำหรับยื่น ภ.ง.ด.1/3/53 และ ภ.พ.30 ผ่าน e-Filing
   - รูปแบบ **DBD e-Filing (XBRL in Excel Template)** และแบบ ส.บช.3 ตามมาตรฐาน TFRS for NPAEs

---

## 2. สถาปัตยกรรมการคีย์ข้อมูลจุดเดียว (Single Point of Entry)

ในหน้าจอบันทึกใบสำคัญรายวัน (Journal Voucher) ผู้ใช้สามารถลงรายการ Debit/Credit บัญชี พร้อมปุ่มกดหรือ Auto-Popup เมื่อเลือกผังบัญชีที่เกี่ยวข้องกับภาษี:

### 2.1 ตาราง `vat_records` (ภาษีมูลค่าเพิ่ม)
เมื่อมีรายการภาษีซื้อ หรือ ภาษีขาย (หรือกดปุ่ม [ภาษี VAT]):
* `company_code`, `doc_no`, `line_no`: เชื่อมโยงกับ `journal_entries`
* `tax_type`: `1` = ภาษีซื้อ (Input VAT), `2` = ภาษีขาย (Output VAT)
* `tax_invoice_no`: เลขที่ใบกำกับภาษี
* `tax_invoice_date`: วันที่ในใบกำกับภาษี
* `tax_period_year`, `tax_period_month`: งวดภาษีที่นำไปยื่น (ภาษีซื้อยื่นล่าช้าได้ไม่เกิน 6 เดือนตาม ม.86/10)
* `partner_tax_id`: เลขประจำตัวผู้เสียภาษี 13 หลัก ของผู้ขาย/ผู้ซื้อ
* `partner_branch_no`: ลำดับสาขา 5 หลัก (`00000` = สำนักงานใหญ่)
* `partner_name`: ชื่อสถานประกอบการ/ผู้ขาย/ผู้ซื้อ
* `base_amount`: มูลค่าสินค้าหรือบริการก่อนภาษี (ฐานภาษี) `NUMERIC(16,2)`
* `vat_rate`: อัตราภาษี (ปกติ 7.00 หรือ 0.00) `NUMERIC(5,2)`
* `vat_amount`: จำนวนเงินภาษีมูลค่าเพิ่ม `NUMERIC(16,2)`
* `exempt_amount`: จำนวนเงินที่ได้รับยกเว้นภาษี (ถ้ามี) `NUMERIC(16,2)`
* `claim_status`: `1` = ขอคืนได้ปกติ, `2` = ภาษีซื้อต้องห้าม/ไม่ขอคืน (นำไปรวมเป็นต้นทุนสินทรัพย์หรือค่าใช้จ่าย)
* `remark`: หมายเหตุ

### 2.2 ตาราง `wht_records` (ภาษีเงินได้หัก ณ ที่จ่าย 50 ทวิ)
เมื่อมีการจ่ายเงินที่มีการหักภาษี ณ ที่จ่าย หรือรับเงินที่ถูกหัก (หรือกดปุ่ม [หัก ณ ที่จ่าย WHT]):
* `company_code`, `doc_no`, `line_no`: เชื่อมโยงกับ `journal_entries`
* `wht_direction`: `1` = เราหักเขา (ภ.ง.ด.1/2/3/53 ต้องนำส่งสรรพากร), `2` = เขาหักเรา (เครดิตภาษีสิ้นปี)
* `wht_cert_no`: เลขที่หนังสือรับรอง 50 ทวิ (Running หรือกรอกตามเล่ม)
* `payment_date`: วันที่จ่ายเงิน
* `form_type`: `PND1` (ค่าจ้าง/เงินเดือน), `PND2` (ดอกเบี้ย/ปันผล), `PND3` (บุคคลธรรมดา), `PND53` (นิติบุคคล)
* `payee_tax_id`: เลขประจำตัวผู้เสียภาษี 13 หลัก
* `payee_branch_no`: สาขา 5 หลัก (`00000` = สำนักงานใหญ่)
* `payee_name`: ชื่อผู้ถูกหักภาษี
* `payee_address`: ที่อยู่เต็มผู้ถูกหักภาษี
* `income_tax_type`: รหัส/ประเภทเงินได้พึงประเมิน (เช่น มาตรา 40(1), 40(2), ค่าบริการ, ค่าเช่า, ค่าจ้างทำของ, ค่าขนส่ง, ค่าโฆษณา)
* `income_description`: คำอธิบายประเภทเงินได้
* `wht_rate`: อัตราภาษีหัก ณ ที่จ่าย (เช่น 1.00%, 2.00%, 3.00%, 5.00%)
* `base_amount`: จำนวนเงินที่จ่าย `NUMERIC(16,2)`
* `tax_amount`: ภาษีที่หักและนำส่ง `NUMERIC(16,2)`
* `condition_type`: เงื่อนไขการหัก: `1` = หัก ณ ที่จ่าย, `2` = ออกให้ตลอดไป, `3` = ออกให้ครั้งเดียว

---

## 3. เอกสารทางราชการและไฟล์ยื่นที่ต้องผลิต (Statutory Deliverables)

### 3.1 กรมสรรพากร (Revenue Department)
1. **รายงานภาษีซื้อ (Input Tax Report)** (ม.87(2) แห่ง ป.รัษฎากร): เรียงตามงวดภาษี/วันที่ แสดงเลข 13 หลัก, สาขา, ยอดฐานภาษี, ยอดภาษี พร้อมสรุปยอดรวมประจำเดือน
2. **รายงานภาษีขาย (Output Tax Report)** (ม.87(1) แห่ง ป.รัษฎากร): แสดงเลขที่ใบกำกับ, ลูกค้า, เลข 13 หลัก, สาขา, ฐานภาษี, ยอด 0%, ยอดยกเว้น, VAT รวม
3. **ใบสรุปแบบ ภ.พ.30**: สรุปยอดภาษีขายหักภาษีซื้อ คำนวณยอดชำระ/ยอดขอคืน พร้อมใบแนบสรุปยอดรายสาขา
4. **หนังสือรับรองการหักภาษี ณ ที่จ่าย (มาตรา 50 ทวิ)**: ฟอร์มพิมพ์มาตรฐาน PDF และกระดาษต่อเนื่อง (ฉบับที่ 1 ให้ผู้ถูกหัก, ฉบับที่ 2 เป็นหลักฐาน)
5. **ใบแนบ ภ.ง.ด.1, ภ.ง.ด.2, ภ.ง.ด.3, ภ.ง.ด.53**: พร้อมใบปะหน้าสรุปยอดนำส่งประจำเดือน
6. **Export Text File สำหรับโปรแกรม RD Prep (e-Filing)**:
   - ฟอร์แมต Text File (.txt) เข้ารหัส UTF-8 คั่นฟิลด์ด้วย Pipe (`|`)
   - แถวที่ 1 = Header ข้อมูลผู้มีหน้าที่หักและสรุปยอด
   - แถวถัดไป = Detail ใบแนบแต่ละรายการ
   - ตัวเลขทศนิยม 2 ตำแหน่ง (`0.00`) ห้ามมีเครื่องหมายคอมม่าหรืออักขระพิเศษ `* + / \ ! $ % # & @ ' "`
   - ผู้ใช้นำไฟล์เข้า RD Prep แปลงเป็น `.rdx` เพื่อยื่น e-Filing ได้ทันที
7. **รายงานเตรียมตัวเลข ภ.ง.ด.51 (ครึ่งปี) & สรุปเครดิตภาษีถูกหักเพื่อยื่น ภ.ง.ด.50 (สิ้นปี)**

### 3.2 กรมพัฒนาธุรกิจการค้า (DBD)
1. **งบแสดงฐานะการเงิน (Balance Sheet)** (มาตรฐาน TFRS for NPAEs)
2. **งบกำไรขาดทุน (Income Statement)** (จำแนกตามหน้าที่ / ลักษณะ)
3. **งบแสดงการเปลี่ยนแปลงส่วนของเจ้าของ (Statement of Changes in Equity)**
4. **หมายเหตุประกอบงบการเงิน (Notes to Financial Statements)**
5. **แบบ ส.บช.3 (แบบนำส่งงบการเงิน)** พร้อมข้อมูลผู้ทำบัญชี/ผู้สอบบัญชี และรหัสธุรกิจ (TSIC)
6. **Export DBD e-Filing Excel Template (.xlsx)**: สร้างไฟล์ตาม Template XBRL in Excel ของ DBD เพื่อ Upload ส่งงบการเงินออนไลน์

### 3.3 พ.ร.บ.การบัญชี พ.ศ. 2543 (เอกสารที่สำนักงานบัญชีต้องจัดพิมพ์เก็บ 5 ปี)
1. **สมุดรายวันทั่วไป (General Journal)** และสมุดรายวันเฉพาะ (ซื้อ, ขาย, รับ, จ่าย)
2. **สมุดบัญชีแยกประเภททั่วไป (General Ledger)**
3. **งบทดลอง (Trial Balance)** ประจำเดือนและประจำปี
4. **กระดาษทำการปิดบัญชี (Working Papers 6 ช่อง / 8 ช่อง)**

---

## 4. ร่าง DDL สำหรับ `vat_records` และ `wht_records` (PostgreSQL 18)

*(หมายเหตุ: โฟลเดอร์ `mydocs/` เป็น Strict Read-Only สำหรับ AI ตาม Rule 18 ให้ลุงจืดเป็นผู้พิจารณานำสคริปต์นี้ไปใส่ใน `mydocs/datamodels/gl/`)*

```sql
-- 1. ตารางบันทึกภาษีมูลค่าเพิ่ม (VAT Records)
CREATE TABLE IF NOT EXISTS vat_records (
    company_code        VARCHAR(20) NOT NULL,
    doc_no              VARCHAR(30) NOT NULL,
    line_no             INT NOT NULL,
    tax_type            SMALLINT NOT NULL CHECK (tax_type IN (1, 2)), -- 1=ภาษีซื้อ (Input), 2=ภาษีขาย (Output)
    tax_invoice_no      VARCHAR(50) NOT NULL,
    tax_invoice_date    DATE NOT NULL,
    tax_period_year     INT NOT NULL,
    tax_period_month    SMALLINT NOT NULL CHECK (tax_period_month BETWEEN 1 AND 12),
    partner_tax_id      VARCHAR(20) NOT NULL,
    partner_branch_no   VARCHAR(10) NOT NULL DEFAULT '00000',
    partner_name        VARCHAR(255) NOT NULL,
    base_amount         NUMERIC(16, 2) NOT NULL DEFAULT 0.00,
    vat_rate            NUMERIC(5, 2) NOT NULL DEFAULT 7.00,
    vat_amount          NUMERIC(16, 2) NOT NULL DEFAULT 0.00,
    exempt_amount       NUMERIC(16, 2) NOT NULL DEFAULT 0.00,
    claim_status        SMALLINT NOT NULL DEFAULT 1 CHECK (claim_status IN (1, 2)), -- 1=ขอคืนได้, 2=ภาษีซื้อต้องห้าม
    remark              VARCHAR(255),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (company_code, doc_no, line_no),
    CONSTRAINT fk_vat_journal_entries FOREIGN KEY (company_code, doc_no)
        REFERENCES journal_entries(company_code, doc_no) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vat_records_period 
    ON vat_records (company_code, tax_type, tax_period_year, tax_period_month);
CREATE INDEX IF NOT EXISTS idx_vat_records_partner 
    ON vat_records (company_code, partner_tax_id);

-- 2. ตารางบันทึกภาษีหัก ณ ที่จ่าย (WHT Records - 50 ทวิ)
CREATE TABLE IF NOT EXISTS wht_records (
    company_code        VARCHAR(20) NOT NULL,
    doc_no              VARCHAR(30) NOT NULL,
    line_no             INT NOT NULL,
    wht_direction       SMALLINT NOT NULL CHECK (wht_direction IN (1, 2)), -- 1=เราหักเขา, 2=เขาหักเรา
    wht_cert_no         VARCHAR(50) NOT NULL,
    payment_date        DATE NOT NULL,
    form_type           VARCHAR(10) NOT NULL CHECK (form_type IN ('PND1', 'PND2', 'PND3', 'PND53')),
    payee_tax_id        VARCHAR(20) NOT NULL,
    payee_branch_no     VARCHAR(10) NOT NULL DEFAULT '00000',
    payee_name          VARCHAR(255) NOT NULL,
    payee_address       TEXT NOT NULL,
    income_tax_type     VARCHAR(50) NOT NULL, -- เช่น 40(1), 40(2), ค่าบริการ, ค่าเช่า
    income_description  VARCHAR(255) NOT NULL,
    wht_rate            NUMERIC(5, 2) NOT NULL DEFAULT 3.00,
    base_amount         NUMERIC(16, 2) NOT NULL DEFAULT 0.00,
    tax_amount          NUMERIC(16, 2) NOT NULL DEFAULT 0.00,
    condition_type      SMALLINT NOT NULL DEFAULT 1 CHECK (condition_type IN (1, 2, 3)), -- 1=หัก ณ ที่จ่าย, 2=ออกให้ตลอดไป, 3=ออกให้ครั้งเดียว
    remark              VARCHAR(255),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (company_code, doc_no, line_no),
    CONSTRAINT fk_wht_journal_entries FOREIGN KEY (company_code, doc_no)
        REFERENCES journal_entries(company_code, doc_no) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_wht_records_payment_date 
    ON wht_records (company_code, form_type, payment_date);
CREATE INDEX IF NOT EXISTS idx_wht_records_payee 
    ON wht_records (company_code, payee_tax_id);
```

---

## 5. แผนการดำเนินงานต่อสำหรับ Codex (Action Plan)

1. **Phase 1: Database & Backend API (Go)**
   - เพิ่ม Model และ Repository สำหรับ `vat_records` และ `wht_records` ใน backend Go
   - ทำ Atomic Transaction บันทึก `journal_entries` + `journal_lines` + `vat_records` + `wht_records` พร้อมกัน
2. **Phase 2: Frontend Speed Entry & Modal Dialogs (Next.js)**
   - ปรับปรุงฟอร์มคีย์ใบสำคัญรายวันให้มีปุ่ม [VAT] และ [WHT]
   - Auto-balance และ Auto-suggest อัตราภาษี 7% / 3% เมื่อคีย์ยอด
   - รองรับคีย์บอร์ดนำทาง (Tab / Enter Flow เหมือน Champ)
3. **Phase 3: Statutory Reports & RD Prep Exporter**
   - พัฒนา Query และหน้าจอรายงานภาษีซื้อ, ภาษีขาย, สรุป ภ.พ.30
   - พัฒนาแบบพิมพ์หนังสือรับรอง 50 ทวิ (PDF/Print) และใบแนบ ภ.ง.ด.1/2/3/53
   - พัฒนาตัวส่งออกไฟล์ Text File (`.txt` Pipe-delimited UTF-8) นำเข้า RD Prep
4. **Phase 4: Financial Statements & DBD e-Filing Exporter**
   - พัฒนารายงานงบทดลอง (Trial Balance), งบกำไรขาดทุน, งบแสดงฐานะการเงิน
   - พัฒนาตัวส่งออก DBD e-Filing Excel Template (.xlsx)
