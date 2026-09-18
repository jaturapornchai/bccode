# ADR 2026-09-19 — ตัดเมนู/ระบบที่เกิน Champ ออก, เติมของที่ Champ มี, เรียงเมนูตาม menuconfig.xml, เปิดใช้ 2 ภาษา

**สถานะ:** ใช้งานแล้ว (commit รอบเดียวกับ ADR นี้)
**ผู้ตัดสิน:** ลุงจืด (คำสั่ง 2026-09-19: "ลบระบบและเมนูออกให้เหลือใกล้เคียง Champ / Champ มีแต่ระบบใหม่ไม่มีให้เอาเข้ามา / จัดเมนูให้เล่นเหมือน Champ / ตอนนี้ใช้แค่ 2 ภาษา")
**ต่อยอดจาก:** [2026-09-19-champ-parity-no-bloat-rule.md](2026-09-19-champ-parity-no-bloat-rule.md), [2026-09-11-champ-menu-structure-alignment.md](2026-09-11-champ-menu-structure-alignment.md)

## บริบท

ตรวจ `D:\project-champ\champ\champ\menuconfigxml\menuconfig.xml` (UTF-16, 10 โมดูล, 490 รายการ — รายงานทั้งหมดอยู่ใต้ popup `disable="yes"`) เทียบกับ `frontend/src/lib/menu-data.ts` (208 รายการ) พบ:

1. เมนูที่ Champ ไม่มี 47 รายการ (dashboard, เอกสารประจำ, คลังเอกสาร, สลิป, POS, ล็อต/FIFO, นำเข้าไฟล์, deferred tax, cash flow ฯลฯ)
2. ฟีเจอร์ในจอที่ AI ก่อนหน้าเพิ่มเอง 2026-09-17→19 (6 commits `8921a923`…`45462fae`) เป็น **calculator ฝั่ง browser ทั้งหมด ไม่มี backend** — ผิดทั้งกฎ Champ parity และกฎ Backend-First
3. ของที่ Champ มีแต่ระบบใหม่ขาด 33 รายการ (ส่วนใหญ่รายงาน + 3 รายการ GL ที่ handoff ก่อนหน้าตัดผิดโดยอ้าง Champ)
4. handoff 2026-09-19 ระบุผิดว่า `goods-inspection`, `low-stock-alert`, `sale-reservation-flow` เกิน Champ — ทั้งสามมีใน Champ (ใบตรวจรับสินค้า, รายงานยอดคงเหลือที่ถึงจุดสั่งซื้อ, Flow ใบสั่งจองสินค้า)

## การตัดสินใจ

### A. ลบเมนู 47 รายการ (id ใน `menu-data.ts`)

| โมดูล | ลบ |
|---|---|
| PO | procurement-dashboard, import-documents, recurring-expense, document-vault, inter-company-inbox, deposit-refund |
| BILL | recurring-invoice, tax-invoice, combined-receipt, credit-note, return-deposit, sales-by-customer, sales-by-channel |
| AP | creditor-group, import-partner, payment-voucher, combined-payment |
| AR | debtor-group, sale-invoice |
| CASH-BANK | petty-cash, director-advance, credit-card-expense, cash-drawer, bank-payment-file, slip-in, slip-out, bank-statement, bank-reconcile |
| IC | import-product, import-product-file, import-product-image, fifo-cost-layers, stock-lot, cost-adjustment, stock-balance-location, expiring-stock-alert, stock-lot-movement, audit-data, rebuild-products, rebuild-product-balance |
| FA | asset-purchase, asset-disposal |
| VAT | purchase-tax-invoice-register, unreceived-tax-invoice, deferred-tax |
| GL | cash-flow, project-pnl |

ย้ายโมดูล: `expense-record` (บันทึกจ่ายเงินอื่นๆ) และ `expense-summary-report` (เปลี่ยนป้ายเป็น "รายงานการจ่ายเงินอื่นๆ") จาก PO → CASH-BANK ตาม Champ Bank/Cash; `purchase-price-comparison` ย้ายไปกลุ่มรายงาน

### B. ลบระบบ/ไฟล์ที่ Champ ไม่มี (frontend calculator)

- GL Reports: ตรวจสุขภาพบัญชี, AI Audit Copilot, เปรียบเทียบปีก่อน/แนวโน้ม 12 เดือน, CFO Dashboard (`gl-health-audit-modal`, `ai-audit-guard-modal`, `gl-comparative-view`, `cfo-dashboard-view` + lib `gl-health-check`, `ai-audit-guard`, `cfo-financial-health`, `gl-comparative-report`)
- GL Journals: แม่แบบบันทึกด่วน 8 กลุ่มธุรกิจ (`journal-fast-templates-dialog`, `thai-accounting-business-patterns`) — **คงไว้:** Excel paste, Tab/Enter speed entry, auto-balance, ตรวจเดบิต=เครดิต
- Tax Workbench: กระทบยอด GL vs ภ.พ.30 / WHT, ส่งออก RD Prep e-Filing, e-Tax XML (`thai-vat-reconciliation`, `thai-wht-reconciliation`, `thai-tax-export`, `thai-etax`) — **คงไว้:** แบบยื่น ภ.พ.30/36, ภ.ง.ด.2/3/53, 50 ทวิ
- Banking Workbench ทั้งชุด (`banking-workbench`, `cheque-management-view`, `petty-cash-view`, `bank-reconciliation`)
- `inventory-costing` (FIFO/Average ฝั่ง browser), `fixed-assets-engine` (ค่าเสื่อมฝั่ง browser — ดูผลตัดสินในรายงานคอมมิต), `report/stock-card-screen` + `inventory-stock-card` (dead code ไม่ถูก import), widget ภาพรวมใน `dashboard-home` (+205 บรรทัด) ย้อนกลับเป็นเวอร์ชัน `45462fae`

### C. เติมของที่ Champ มี 33 รายการ

- PO: `purchase-reduce-debt` บันทึกลดหนี้, `rfq-price-table` บันทึกใบตารางราคารวม
- AP รายงาน: เคลื่อนไหวเจ้าหนี้, สถานะเจ้าหนี้, ใบส่งของค้างจ่ายชำระ, การจ่ายเงินประจำวัน
- AR รายงาน: เคลื่อนไหวลูกหนี้, สถานะลูกหนี้, ใบส่งของค้างชำระ, ตรวจสอบยอดวงเงิน
- CASH-BANK รายงาน: เช็ครับ, เช็คจ่าย, บัตรเครดิต, Bank Statement, เคลื่อนไหวเงินสด, เคลื่อนไหวเงินสดย่อย, สมุดจ่ายเงินประจำเดือน
- IC รายงาน: ถึงจุดสูงสุด, สินค้าไม่เคลื่อนไหว, ผลต่างตรวจนับ, ค้างรับ, ค้างส่ง, เคลื่อนไหว Serial Number
- FA รายงาน: ค่าเสื่อมเป็นเดือน, ค่าเสื่อมเป็นปี, ค่าเสื่อมตาม ภ.ง.ด.50, สินทรัพย์ที่ถูกขายพร้อมกำไรขั้นต้น
- VAT: รายงานสรุปยอดภาษี
- GL: `gl-account-groups` บันทึกกลุ่มผังบัญชี (คืน), `gl-reprocess` ประมวลผลข้อมูลบัญชีใหม่ (คืน), `gl-journal-books` กำหนดสมุดรายวัน, `gl-daily-report` รายงานข้อมูลรายวัน, `budget-comparison-report` รายงานเปรียบเทียบงบประมาณ; `gl-opening-balance` เปลี่ยนป้ายเป็นคำ Champ "บันทึกยอดสะสมประจำปี" (ชื่อเดิมเป็น alias ค้นหา) แทนการเพิ่ม `gl-annual-accumulated` ซ้ำ

รายการที่ **ยังไม่มีจอ/API** แสดงเป็น "ยังไม่พร้อม" ผ่าน `isMenuScreenPending` — ห้ามแต่ง column/API code ให้จนกว่า backend จะมีจริง (NO MAGIC). รายงานย่อยของ Champ ที่แยก "-ตามวันที่/แผนก/กลุ่ม/โครงการ/เจ้าหนี้" ไม่ทำเป็นเมนูแยก — เป็น filter ในรายงานเดียว (กฎ "ยุบรวมให้กระชับ")

### D. เรียงเมนูตาม Champ

แต่ละโมดูลเรียงกลุ่ม/รายการตามลำดับใน `menuconfig.xml` (งานประจำ → popup ย่อย → เงินมัดจำ → ต้นทุน → รายงาน; CHQSYS มาก่อน Bank/Cash) กลุ่มใหม่ `bill-orders`, `bank-reports`; ยุบ `po-approvals`, `bill-approvals`, `bill-delivery`, `bill-adjustments`, `ap-payment`, `ar-adjustments`, `ar-payment`, `ic-requests`

### E. ใช้ 2 ภาษา

`ACTIVE_LANGUAGE_CODES = ["th","en"]` (`frontend/src/lib/i18n.ts`) เป็นค่าเริ่มต้นของ `LanguageDialog`; โครง 12 ภาษาคงไว้ (กฎใน `AGENTS.md`)

### ข้อยกเว้นที่ **เก็บไว้** แม้ Champ ไม่มี (มีคำสั่ง/กฎรองรับ)

`etax-invoice` (กฎ Enhance & Modernize 2026-09-15), แบบยื่นภาษี `vat-pp30/pp36/pnd2/pnd3/pnd53/wht-certificate` (ข้อบังคับสรรพากร), ข้อมูลหลักสินค้า `product-unit/product-group/brand/warehouse/barcode` (Champ มีในจอข้อมูลหลัก ไม่อยู่ใน menuconfig), `gl-fiscal-years` (ประมวลผลสิ้นปีต้องใช้), `withholding-tax-deduction` (ต้นทาง 50 ทวิ), `profit-loss`/`balance-sheet` (งบจาก "รูปแบบงบการเงิน"), `thai-document-print-modal` + `thai-baht-text` (Champ พิมพ์เอกสาร)

## ผลกระทบ

- เมนู 208 → 194; ไฟล์ frontend ลบ ~20 ไฟล์ / ~8,000 บรรทัด (ดู `git show --stat` ของคอมมิต)
- `menu-before-champ-upgrade.json` (fixture สิทธิ์) สร้างใหม่เป็น 194 รายการ — permission code ของ 47 เมนูที่ลบหายไปโดยตั้งใจ (pre-launch, ข้อมูล disposable)
- `languages.tsv` +33 แถว (th จริง, en จริง, ภาษาอื่น = en placeholder), แก้ th/en 11 แถวตามคำ Champ
- Production deploy รอบเดียวกัน (frontend only)

## ทางเลือกที่ปัดตก

- เก็บฟีเจอร์ calculator ไว้แบบซ่อน (flag) — ขัดกฎ disposable/no-dual-path 2026-09-15
- ทำจอ/API ให้รายงานใหม่ 33 รายการทันที — ต้องออกแบบ backend จริงก่อน (Backend-First) ไม่แต่งเอง
- ลด `LanguageCode` เหลือ 2 — กระทบ locales/tsv/type ทั้งระบบ ไม่คุ้มเมื่อจะเปิดกลับภายหลัง
