# Report parity R0 (wave 1 GL) — เมนูรายงานย่อยของ "บัญชี" Champ เทียบ BC

> วันที่ 2026-10-11 · จุดประสงค์: R0 ของดีไซน์ `report-registry.md` คลื่นที่ 1 (GL) = ตารางเทียบเมนูรายงานของ Champ กับ menu id/route ของ BC เพื่อให้ R6 จัดเมนูย่อยทั้งสามตามลำดับ Champ
> ไม่ได้รัน Champ หรือ query ฐานข้อมูล; อ่านซอร์ส/ภาพ/ไฟล์ spec อย่างเดียว

## 0. แหล่งข้อมูลและข้อควรระวัง

| แหล่ง | ที่อ่าน |
|---|---|
| Champ menu (UTF-16LE) | `D:\project-champ\champ\champ\menuconfigxml\menuconfig.xml:552-615` (GL; รายวัน 574-593, งบการเงิน 595-600, ภาษี 602-613) |
| Champ menu ในรีซอร์ส | `champ\champ\BC5Account.rc:20956-21034` (เข้ารหัส cp874) · ตัวสร้างเมนู `champ\champ\XMLMenuConfig.cpp:150-153,179` (ข้ามรายการ `disable="yes"`) |
| ภาพ | `bcspec\หน้าจอเมนูบัญชี\เมนู-ประกอบ.jpg` · `bcspec\เมนูโปรแกรมทั้งหมด.jpg` (build 5.5 Pro 20260709-12) |
| PDF ตัวอย่าง | `bcspec\Report_ระบบบัญชี\` 28 ไฟล์ = รายงาน 25 + ใบแนบ/หนังสือรับรองของเมนู 5513 อีก 3 (ไม่อยู่ในสามเมนูย่อย); ข้อความไทยในหัว PDF อ่านเป็นข้อความไม่ได้ จึงใช้ชื่อไฟล์เป็นชื่อบน PDF |
| BC | `frontend/src/lib/menu-data.ts:657-672` (gl-reports), `:684-700` (vat-reports) · `frontend/src/lib/menu-bars.ts:282-295` (บาร์ `gl`) · `docs/kms/champ/09-gl.md:31-77` · `docs/kms/champ/10-reports.md` §1.2-1.3 (ไฟล์อยู่ใต้ `champ/` ไม่ใช่ `docs/kms/10-reports.md`) |

- **ภาพชนะ menuconfig:** ป๊อปอัปทั้งสาม (รายงานรายวัน/งบการเงิน/ภาษี) มี `disable="yes"` (`menuconfig.xml:574,595,602`) ซึ่งตัวสร้างเมนูจะข้าม แต่ `เมนู-ประกอบ.jpg` แสดงทั้งสามเป็นเมนูย่อย ">" บนจอจริง → ถือว่ามีบนจอ (`10-reports.md` §1.2 ที่ว่าซ่อนหมด ขัดกับภาพ — ยังไม่แก้ในงานนี้)
- ไม่มีภาพของ "เนื้อใน" เมนูย่อย → ลำดับใช้ `menuconfig.xml` ซึ่งตรงกับ RC ยกเว้นจุดใน §3 ข้อ 4-5; ข้อความที่แสดงจริงมาจากตาราง resource ตาม `resourceid` (ไฟล์ภาษาไม่อยู่ใน repo) จึงใช้ `text` ใน menuconfig/RC เป็นชื่อ

## 1. Table A — รายการรายงานของ Champ ตามลำดับบนจอ

สถานะ: **W1** = registry wave 1 (GL R2) · **W2** = planned wave 2 · **R7** = ภาษี R7 · **แบบยื่น** = คงใน TaxFilingWorkbench · "ใหม่ `x`" = id ใหม่ที่ R6 เพิ่ม · route ของ id ใหม่ให้ R6 กำหนด

### A1. รายงานรายวัน (`menuconfig.xml:574-593`)

กลุ่มซ้อน "รายงานข้อมูลหลักต่างๆ" (575-585) อยู่บนสุด ก่อนตัวคั่น

| ลำดับ | Champ id | ชื่อในจอ Champ | PDF ตัวอย่าง | BC id ปัจจุบัน | route BC | สถานะ wave 1 |
|---|---|---|---|---|---|---|
| 1 | 5531 | รายงานรายละเอียดสมุดบัญชี | รายงานรายละเอียดสมุดรายวัน.pdf (ชื่อต่าง) | — (ปุ่มพิมพ์ใน `gl-journal-books`) | `/gl/journal-books` | W1 · ใหม่ `gl-journal-books-report` |
| 2 | 5538 | รายงานรายละเอียดแผนก | รายงานรายละเอียดแผนก.pdf | — | — | W1 · ใหม่ `gl-department-report` |
| 3 | 5521 | รายงานรายละเอียดโครงการ | รายงานรายละเอียดโครงการ.pdf | — | — | W1 · ใหม่ `gl-project-report` |
| 4 | 5524 | รายงานรายละเอียดการจัดสรร | รายงานรายละเอียดการจัดสรร.pdf | — | — | W2 |
| 5 | 5525 | รายงานรายละเอียดงบประมาณ | รายงานรายละเอียดงบประมาณ.pdf | — | — | W1 · ใหม่ `gl-budget-detail` |
| 6 | 5526 | รายงานรายละเอียดงาน | รายงานรายละเอียดงาน.pdf | — | — | W2 |
| 7 | 5527 | รายงานรายละเอียดหน่วยงาน | รายงานรายละเอียดหน่วยงาน.pdf | — | — | W2 |
| 8 | 5528 | รายงานรายละเอียดฝ่าย | รายงานรายละเอียดฝ่าย.pdf | — | — | W2 |
| 9 | 5529 | รายงานรายละเอียดสาขา | รายงานรายละเอียดสาขา.pdf | — | — | W2 |
| 10 | 5520 | รายงานรายละเอียดผังบัญชี | รายงานรายละเอียดผังบัญชี.pdf | — (ปุ่มพิมพ์/CSV ใน `chart-of-accounts`) | `/gl/chartofaccounts` | W1 · ใหม่ `gl-chart-of-accounts-report` |
| 11 | 5532 | รายงานข้อมูลรายวัน | รายงานข้อมูลรายงาน.pdf (ชื่อไฟล์น่าจะพิมพ์ผิด) | `gl-daily-report` | `/report/gljournal` | W1 |
| 12 | 5533 | รายงานข้อมูลรายวันแบบสรุป | รายงานข้อมูลรายวันแบบสรุป.pdf | — | — | W1 · ใหม่ `gl-daily-summary-report` |
| 13 | 5534 | รายงานบัญชีแยกประเภท | รายงานบัญชีแยกประเภท.pdf | `general-ledger` | `/report/ledger` | W1 |
| 14 | 5535 | รายงานบัญชีแยกประเภทแบบสรุป | รายงานบัญชีแยกประเภทแบบสรุป.pdf | — | — | W1 · ใหม่ `general-ledger-summary` |
| 15 | 5530 | รายงานเปรียบเที่ยบงบประมาณ (สะกดตาม Champ) | รายงานเปรียบเทียบงบประมาณ.pdf | `budget-comparison-report` | `/report/budgetcomparison` | W1 |

ตัวคั่นอยู่หลังลำดับ 10 และหลังลำดับ 14 (`:586`, `:592`)

### A2. รายงานงบการเงิน (`menuconfig.xml:595-600`)

| ลำดับ | Champ id | ชื่อในจอ Champ | PDF ตัวอย่าง | BC id ปัจจุบัน | route BC | สถานะ wave 1 |
|---|---|---|---|---|---|---|
| 1 | 5506 | รายงานงบทดลอง | รายงานงบทดลอง.pdf | `trial-balance` | `/report/trialbalance` | W1 |
| 2 | 5507 | รายงานงบทดลอง 2 | รายงานงบทดลอง2.pdf | — | — | W1 · ใหม่ `trial-balance-2` |
| 3 | 5508 | รายงานงบทดลอง 3 | — | — | — | W2 (ยังไม่ verify ว่ามีใช้จริง: ON_COMMAND ถูกคอมเมนต์ `MainFrm.cpp:1012` และไม่อยู่ใน RC) |
| 4 | 5499 | รายงานกระดาษทำการ | รายงานกระดาษทำการ.pdf | `working-paper` | `/gl/workingpaper` | W1 |
| 5 | 5505 | รูปแบบงบการเงิน | — | `financial-statement-designer` | `/gl/statement-designer` | W1 (BC เป็นจอออกแบบงบ ไม่ใช่รายงานอ่านอย่างเดียว — R2 ตัดสินวิธีลงทะเบียน) |

### A3. รายงานภาษี (`menuconfig.xml:602-613`)

| ลำดับ | Champ id | ชื่อในจอ Champ | PDF ตัวอย่าง | BC id ปัจจุบัน | route BC | สถานะ wave 1 |
|---|---|---|---|---|---|---|
| 1 | 5536 | รายงานภาษีซื้อ | รายงานภาษีซื้อ.pdf | `vat-buy` | `/report/reportvatbuy` | R7 |
| 2 | 5537 | รายงานภาษีขาย | รายงานภาษีขาย.pdf | `vat-sale` | `/report/reportvatsale` | R7 |
| 3 | 5518 | รายงานภาษีขายต่างประเทศ | รายงานภาษีขายต่างประเทศ.pdf | — | — | R7 |
| 4 | 5517 | รายงานภาษีขายในประเทศ | รายงานภาษีขายในประเทศ.pdf | — | — | R7 |
| 5 | 5539 | รายงานสรุปยอดภาษี | รายงานสรุปยอดภาษี.pdf | `vat-summary-report` | `/report/vatsummary` | R7 |
| 6 | 5522 | รายงานภาษีซื้อที่ยกเลิกข้ามงวด | — | — (แท็บใน `vat-buy`) | — | R7 |
| 7 | 5523 | รายงานภาษีขายที่ยกเลิกข้ามงวด | — | — (แท็บใน `vat-sale`) | — | R7 |
| 8 | 5543 | รายงานภาษีหัก ณ ที่จ่าย (ภ.ง.ด. 53) | รายงานภาษีหัก ณ ที่จ่าย(ภ.ง.ด.53).pdf | `withholding-tax-report` (+ แบบยื่น `vat-pnd53`) | `/report/wht-reports` | R7 (แบบยื่นคงใน TaxFilingWorkbench) |
| 9 | 5541 | รายงานภาษีหัก ณ ที่จ่าย (ภ.ง.ด. 3) | — | `withholding-tax-report` (+ แบบยื่น `vat-pnd3`) | `/report/wht-reports` | R7 (แบบยื่นคงใน TaxFilingWorkbench) |
| 10 | 5540 | รายงานภาษีถูกหัก ณ ที่จ่าย | — | `withholding-tax-received` | `/report/whtreceived` | R7 |

ตัวคั่นอยู่หลังลำดับ 7 (`:610`)

## 2. Table B — รายการ BC ในสามเมนูย่อยที่ Champ ไม่มี (BC-only รอ Q3 (คงที่เดิม))

| เมนูย่อย | BC id | route BC | หมายเหตุ |
|---|---|---|---|
| งบการเงิน | `profit-loss` | `/report/pnl` | BC-only รอ Q3 (คงที่เดิม) |
| งบการเงิน | `balance-sheet` | `/report/balancesheet` | BC-only รอ Q3 (คงที่เดิม) |
| งบการเงิน | `gl-ar-outstanding` | `/report/ar-outstanding` | BC-only รอ Q3 (คงที่เดิม) |
| งบการเงิน | `gl-ap-outstanding` | `/report/ap-outstanding` | BC-only รอ Q3 (คงที่เดิม) |
| งบการเงิน | `gl-bank-unmatched` | `/report/bank-unmatched` | BC-only รอ Q3 (คงที่เดิม) |
| ภาษี | `vat-pp30`, `vat-pp36` | `/report/vatpp30`, `/report/vatpp36` | แบบยื่น (คงใน TaxFilingWorkbench) · BC-only รอ Q3 (คงที่เดิม) |
| ภาษี | `vat-pnd2`, `vat-pnd3`, `vat-pnd53` | `/report/vatpnd2`, `/report/vatpnd3`, `/report/vatpnd53` | แบบยื่น (คงใน TaxFilingWorkbench) · BC-only รอ Q3 (คงที่เดิม); `vat-pnd2` ห้ามลบ (`AGENTS.md`) |
| ภาษี | `tax-forms`, `wht-certificate` | `/report/taxforms`, `/report/whtcertificate` | แบบยื่น (คงใน TaxFilingWorkbench) · BC-only รอ Q3 (คงที่เดิม) |

รายวัน: ไม่มี BC-only

## 3. ผลยืนยัน id↔ชื่อ↔เมนูย่อย และความไม่ตรงที่พบ

ทั้ง 14 id ของ wave 1 (5532, 5533, 5534, 5535, 5506, 5507, 5499, 5530, 5525, 5505, 5520, 5531, 5538, 5521) ตรง id↔ชื่อ↔เมนูย่อยตาม menuconfig และ RC: 5525/5520/5531/5538/5521 อยู่ในกลุ่มซ้อน "รายงานข้อมูลหลักต่างๆ" ภายใต้ "รายงานรายวัน" (ไม่ใช่เมนูย่อยของตัวเอง); 5505 อยู่ใต้ "รายงานงบการเงิน" ความไม่ตรง:

1. **5531** เมนูเขียน "รายงานรายละเอียดสมุดบัญชี" (`menuconfig.xml:576`, `BC5Account.rc:20962`) แต่ชื่อไฟล์ PDF เขียน "…สมุดรายวัน" และ BC เรียกจอว่า "กำหนดสมุดรายวัน" → ใช้ชื่อเมนูตาม Champ เป็นป้ายหลัก
2. **5532** `resourceid="104003"` ใน menuconfig ซ้ำกับชื่อกลุ่ม "รายงานข้อมูลหลักต่างๆ" (`:575`) ส่วน RC ใช้ `104018` (`BC5Account.rc:20990`) → ชื่อจริงคือ "รายงานข้อมูลรายวัน"; PDF ชื่อ "รายงานข้อมูลรายงาน.pdf" น่าจะพิมพ์ผิด (หัว PDF อ่านไม่ได้ ยังไม่ verify)
3. **5530** Champ สะกด "เที่ยบ" แต่ชื่อ PDF และ BC ใช้ "เทียบ" (คงตาม BC)
4. ป๊อปอัปทั้งสามถูก `disable="yes"` ใน menuconfig แต่ภาพแสดงบนจอ (ดู §0)
5. menuconfig มี 5508 แต่ RC ไม่มี; RC มี "รายงานภาษี Tax Machine" (`104024`, `BC5Account.rc:21015`) แต่ menuconfig ไม่มี → ไม่ใส่ทั้งคู่ใน wave 1
6. PDF "รายงานบัญชีแยกประเภท(ตามแหล่งที่มา).pdf" ไม่มีรายการเมนูของตัวเองในสามเมนูย่อย — ยังไม่ verify ที่มา (อาจเป็นมุมมอง/ตัวเลือกของ 5534 หรือเปิดจากจออื่น) ไม่ใส่ใน wave 1
7. BC วาง `budget-comparison-report` (5530) ไว้ในเมนูงบการเงิน แต่ Champ อยู่ใน "รายงานรายวัน" ท้ายสุดหลังตัวคั่น → R6 ย้ายตาม Champ

## 4. ข้อสรุปสำหรับ R6

id ใหม่ทั้งแปดยังไม่มีในโค้ด (grep ใต้ `frontend/src` และ `backend` ไม่พบ) ทุก id ต้องเพิ่มในแคตตาล็อก `menu-data.ts` กลุ่ม `gl-reports` ด้วย (ตาม convention) · ป้ายไทยคัดจาก menuconfig/RC (`:576-598`)

| id ใหม่ | Champ | ป้ายไทย (ตามจอ Champ) | English |
|---|---|---|---|
| `gl-journal-books-report` | 5531 | รายงานรายละเอียดสมุดบัญชี | Journal Book Detail Report |
| `gl-department-report` | 5538 | รายงานรายละเอียดแผนก | Department Detail Report |
| `gl-project-report` | 5521 | รายงานรายละเอียดโครงการ | Project Detail Report |
| `gl-budget-detail` | 5525 | รายงานรายละเอียดงบประมาณ | Budget Detail Report |
| `gl-chart-of-accounts-report` | 5520 | รายงานรายละเอียดผังบัญชี | Chart of Accounts Detail Report |
| `gl-daily-summary-report` | 5533 | รายงานข้อมูลรายวันแบบสรุป | Journal Entries Report (Summary) |
| `general-ledger-summary` | 5535 | รายงานบัญชีแยกประเภทแบบสรุป | General Ledger (Summary) |
| `trial-balance-2` | 5507 | รายงานงบทดลอง 2 | Trial Balance 2 |

ลำดับหลัง wave 1 (id เดิมคงชื่อเดิม; Champ ก่อน, BC-only ท้ายบล็อกตาม convention; ช่อง W2 ยังไม่ใส่ แต่เว้นลำดับไว้):

**`gl-reports-daily`** — ใช้กลุ่มซ้อนแบบ Champ (`MenuBarNode` ซ้อนได้: `menu-bars.ts:25`; มี precedent กลุ่ม "รายงานข้อมูลหลัก" ที่ `:193`; ถ้า R6 ไม่ซ้อนให้เรียงแบนตามลำดับเดียวกัน)
1. submenu `gl-reports-master` ป้าย "รายงานข้อมูลหลักต่างๆ" / "Master Data Reports" (key ใหม่ R6 ตั้ง): `gl-journal-books-report`, `gl-department-report`, `gl-project-report`, (W2: 5524), `gl-budget-detail`, (W2: 5526-5529), `gl-chart-of-accounts-report`
2. divider
3. `gl-daily-report`, `gl-daily-summary-report`, `general-ledger`, `general-ledger-summary`
4. divider
5. `budget-comparison-report`

**`gl-reports-financial`**
1. `trial-balance`, `trial-balance-2`, (W2: 5508), `working-paper`, `financial-statement-designer`
2. BC-only ท้ายบล็อก (ลำดับเดิม ไม่มี divider): `profit-loss`, `balance-sheet`, `gl-ar-outstanding`, `gl-ap-outstanding`, `gl-bank-unmatched`

**`gl-reports-tax`** — wave 1 ไม่เพิ่ม id ใหม่ (คงรายการและลำดับเดิมจนถึง R7):
`vat-buy`, `vat-sale`, `vat-summary-report`, `vat-pp30`, `vat-pp36`, `withholding-tax-report`, `withholding-tax-received`, `vat-pnd2`, `vat-pnd3`, `vat-pnd53`, `tax-forms`, `wht-certificate`
— เมื่อทำ R7 เป้าหมายตาม Champ คือ `vat-buy`, `vat-sale`, (5518, 5517 ใหม่), `vat-summary-report`, divider, `withholding-tax-report`, `withholding-tax-received` แล้วแบบยื่น/BC-only ต่อท้าย
