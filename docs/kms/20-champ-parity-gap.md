# 20 — ช่องว่างเทียบต้นแบบ Champ (D:\project-champ)

ตรวจ 2026-09-15 ตามกฎ "ยึด D:\project-champ เป็นต้นแบบระบบทั้งหมด" (AGENTS.md)
ขอบเขต: ผังเมนู, ชุดรายงาน, ภาษี, เครื่องมือ/workflow, และ **BCProcess / BCWinserviceInstaller** (โฟลเดอร์ใหม่ที่เพิ่งเพิ่มเข้ามาใน Champ)
หลักฐานทั้งหมดอ้าง `file:line` ของซอร์สจริงทั้งสองฝั่ง — ข้อไหนที่ยังไม่ได้ตรวจเขียนกำกับไว้ว่า "ยังไม่ verify"
**อัปเดตฝั่ง BC 2026-09-25:** ตรวจคอลัมน์ BC ใน §1.2, §2–§6 ใหม่หลังถอด MongoDB/Kafka/Redis/ClickHouse ([ADR](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md)) — ฝั่ง Champ ไม่เปลี่ยน

> วิธีถอดรหัสไฟล์ Champ (จำเป็น ไม่งั้นอ่านภาษาไทยไม่ออก): ไฟล์เป็น UTF-16LE หรือ ANSI ที่บรรจุไบต์ TIS-620
> → `buf.toString('utf16le')` (หรือ `'latin1'`) แล้ว `s.replace(/[\u00a1-\u00fb]/g, c => String.fromCharCode(c.charCodeAt(0) - 0xa0 + 0x0e00))`

---

## 1. ระดับเมนู — Champ 490 รายการ vs BC 225 รายการ

ที่มา: `champ/champ/menuconfigxml/menuconfig.xml` (10 โมดูล: PO, Bill, AP, AR, CHQSYS, Bank/Cash, IC, AS, GL, รายงาน) เทียบกับ `flattenMenuItems()` ใน `frontend/src/lib/menu-data.ts` — ตัวเลข BC เป็นของ 2026-09-15; หลังตัดเมนูตาม Champ 2026-09-19 เหลือ 195 รายการ (`frontend/src/lib/menu-data.test.ts:77`, ADR [champ-parity-menu-cut](decisions/2026-09-19-champ-parity-menu-cut.md))

| กลุ่ม | จำนวน | สรุป |
|---|---|---|
| ตรงกัน / มีของใน BC ชื่อไทยต่างกัน | ~347 | เทียบแล้วด้วย grep เส้นทางจริง เช่น `/asset/maintenance`, `/productserialregistry`, `/gl/periodlock`, `/gl/financialclose`, `/tools/ar-bill-balances`, `/banking/*` |
| รายงานตัวเดียวกันแต่แตกมิติ | 60 | เป็นรายงานฐาน 17 ตัว ที่ Champ แตกเป็นเมนูละมิติ (ตามลูกค้า/ตามสินค้า/ตามพนักงานขาย/ตามเขตการขาย ฯลฯ) — BC ควรทำเป็น "มิติที่เลือกได้" ในจอเดียว ไม่ใช่เพิ่ม 60 เมนู |
| **ขาดจริง (ตรวจแล้วว่าไม่มีใน BC)** | 7 | ดูตาราง §1.1 |

### 1.1 รายการที่ขาดจริง

| Champ | สภาพใน BC |
|---|---|
| รายงานผลต่างการตรวจนับสต็อก (count variance) | ไม่มีทั้งจอและ API |
| รายงานวงเงินเครดิตลูกหนี้ (credit limit) | ไม่มี |
| รายงานความเคลื่อนไหว Serial | มีทะเบียน Serial (`/productserialregistry`) แต่ไม่มีรายงานการเคลื่อนไหว |
| รายงานค่าเสื่อมสำหรับ ภ.ง.ด.50 | มี `/report/assetschedule` แต่ไม่ใช่รูปแบบแนบ ภ.ง.ด.50 |
| ประเมินผู้ขาย (Vendor Evaluation) | ไม่มี |
| Statement แบบส่งล่วงหน้า (forward statement) | มี statement ปกติ ไม่มีแบบ forward |
| มิติ "เขตการขาย" (sales territory) | ไม่มีทั้ง master และ field ในเอกสาร |

### 1.2 ข้อกล่าวอ้างที่พิสูจน์แล้วว่า **ผิด**
subagent รายงานว่ารายงาน GL ของ BC (`/report/ledger`, `trialbalance`, `balancesheet`, `pnl`, `cashflow`) ตกไปที่จอ placeholder — **ไม่จริง** (ณ 2026-09-15) ทั้งห้าเส้นทางอยู่ใน `GL_MENU_ITEMS` → `isGeneralLedgerRoute()` → จอ GL จริง ที่ยิง `/api/gl/{path}` ผ่าน `authFetch` (`frontend/src/lib/general-ledger-api.ts:118`); ภายหลัง `/report/cashflow` ถูกตัดออกจากเมนูตาม ADR menu cut 2026-09-19 (อีกสี่เส้นทางยังอยู่)

---

## 2. ชุดรายงาน — โครงสร้างต่างกันเชิงระบบ ไม่ใช่แค่จำนวน

1. **สต็อกของ Champ ผูกกับงวดบัญชี** — ยอดคงเหลืออ่านจาก `BCStkPeriodBal` (ยอดต่อ "งวด" ไม่ใช่ยอด ณ ปัจจุบัน) และมีแถวย่อยระดับ lot / serial / shelf
   BC (2026-09-25): รายงานสต็อกอ่านสมุด `stock_ledger` และ worker เก็บยอดต่องวดใน `stock_period_balance` แล้ว (`backend/internal/goapi/process/stockengine/store.go:51-53,244-245`) — ยังไม่มีแถวย่อยระดับ lot/serial/shelf (ยังไม่ verify ว่าจอรายงาน `/report/stockbalanceitem` แสดงยอดต่องวด)
2. **"บัญชีคุมพิเศษสินค้า"** (`ICRepSpecialAccountView.cpp:386-405`) เป็น stock card เต็มรูป: จำนวน/ต้นทุน/มูลค่า ทั้งเข้า-ออก-คงเหลือ ในบรรทัดเดียวกัน — BC ยังไม่มีรายงานรูปนี้
3. **รายงานความเคลื่อนไหว 16 คอลัมน์** (`ICRepMovementAmountView.cpp:168-183`) — BC ไม่มี
4. **อายุหนี้ตั้งค่าเองได้** — Champ อ่านช่วงอายุจาก ini `[AGING_RANGE]` `BeginRange1-4`/`OverRange4` (`AR_Rep_AgeByArDlg.cpp` → `SetAgingRange`/`GetAgingRange`) ค่าเริ่มต้นคือ 1-7 / 8-14 / 15-21 / 22-28 / เกิน 28 วัน **ไม่ใช่ 30/60/90** และมีคอลัมน์เพิ่มสำหรับใบลดหนี้ เงินมัดจำ เช็ครับล่วงหน้า เช็คคืน เช็คยกเลิก
   BC: ช่วงอายุตายตัวในโค้ด
5. **ทุก dialog รายงานของ Champ มีกรอบ filter มาตรฐานชุดเดียวกัน** — from-to ของทุกคีย์ (รวม ยี่ห้อ/แผนก/โครงการ), ช่องเลขงวด, "แสดงยอดยกมา", "แสดง Serial", ปุ่ม Option (F6) สำหรับพิมพ์สกุลเงินบ้าน
   BC: มีแค่ปุ่มช่วงวันที่ 4 แบบ + ช่องค้นหาข้อความ

---

## 3. ภาษี — จุดต่างที่กระทบความถูกต้องของแบบยื่น

| เรื่อง | Champ | BC |
|---|---|---|
| ทะเบียนภาษี | ตารางเฉพาะ `BCInputTax` / `BCOutputTax` มี `TaxNo`/`TaxDate` **แยกจาก** `DocNo`/`DocDate`, `TaxGroup`, `ExceptTaxAmount`/`ZeroTaxAmount`/`BeforeTaxAmount`, ยอดรวมรายวัน, สรุปใบลดหนี้แยกส่วน | ไม่มีตารางทะเบียนภาษีแยก — รายงานภาษีซื้อ/ขายอ่านรายละเอียด VAT (`details.vats`) ของใบสำคัญ GL ที่ผ่านรายการแล้ว ตามงวดภาษีที่บันทึก ไม่ใช่วันที่ใบสำคัญ ผ่าน `POST /api/report/tax/vat-register` (`backend/internal/goapi/handlers/tax_report.go:66-70`) |
| เอกสารยกเลิก | `CancelOutPeriod` (`myespeciallyformview.cpp:2828-2841`) แยกกรณี **ยกเลิกในงวดเดียวกัน** (ลงบรรทัดกลับรายการติดลบ) กับ **ยกเลิกข้ามงวด** (แยกรายงานต่างหาก) | ไม่มีตัวกรอง `iscancel` แล้ว — รายงานภาษีอ่านใบสำคัญ GL ที่ผ่านรายการ ยกเลิกด้วยการกลับรายการ; WHT ที่ถูกกลับรายการในเดือนหลังยังคงอยู่ในแบบของเดือนที่ยื่นไปแล้ว (`ReversedMonth`, `tax_report.go:318-320`); การแยกยกเลิกในงวด/ข้ามงวดของภาษีซื้อ-ขายแบบ Champ ยังไม่ verify |
| ปรับปรุงภาษีซื้อข้ามงวด | เมนู "ปรับปรุงภาษีซื้อ" แก้ `TaxDate2`/`TaxGroup` เพื่อยกเครดิตภาษีซื้อไปงวดถัดไป | มีเมนู `/transaction/purchasevatadjustment` (`frontend/src/lib/menu-data.ts:642`) แต่เป็น route ERP ที่ backend ถูกถอด → ขึ้น "รอพัฒนา" |
| ภาษีหัก ณ ที่จ่าย | ตาราง `BCAPWTaxList` / `BCARWTaxList` มี BookNo, PayDate, ShortDesc, **PersonType** (0=บุคคล→ภ.ง.ด.3, 1=นิติบุคคล→ภ.ง.ด.53) | ไม่มีทะเบียนภาษีแยกตารางแบบ Champ — คำนวณสดจากรายละเอียด `details.withholdings` ของใบสำคัญ GL ที่ผ่านรายการแล้ว โดยอ่านจาก `gl_lines`/`gl_records` ใน PostgreSQL ผ่าน `POST /api/report/tax/wht` (`backend/internal/goapi/handlers/tax_report.go:361`) จอ ภ.ง.ด.2/3/53 ใช้งานได้แล้ว (ตาม ADR [ภาษีอยู่ในบัญชีแยกประเภท](decisions/2026-09-23-tax-inside-general-ledger.md)) |

**อัปเดต 2026-09-25:** ข้อควรระวังเดิม (ไม่มีคอลัมน์ `businesscode`, `branchno` ว่างเพราะ Kafka consumer, ความหมาย `vattype` ไม่ยืนยัน) อ้างอิงโมดูล Mongo/Kafka ที่ถอดไปแล้วเมื่อ 2026-09-23 (ADR [ถอด MongoDB/Kafka/Redis/ClickHouse](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md)) — ปัจจุบัน `TaxWithholdingRequest.businesscode` กรองตามบริษัทจริง (`tax_report.go:274`) และ `branchno` เติมจาก snapshot รายการ/ทะเบียนคู่ค้าแล้ว (`tax_report.go:301`); คำว่า `vattype` ไม่มีในโค้ดชุดปัจจุบันอีกแล้ว

---

## 4. BCProcess / BCWinserviceInstaller — เทียบเครื่องยนต์ประมวลผลหลังบ้าน

โฟลเดอร์ใหม่ใน Champ คือ **Windows Service ที่ทำงานประมวลผลหลังบ้านทั้งหมด** (`BCProcess\` ~35MB, `BCWinserviceInstaller\` ~25MB)

**ฝั่ง Champ:**
- `BCSService.cs` — 6 thread วนตรวจคิวทุก ~1 วินาที + watchdog ทุก 10 วินาที
- `stockprocess_class.cs` — คำนวณต้นทุนถัวเฉลี่ยถ่วงน้ำหนัก **แยกตามคลัง** และรองรับ FIFO ตาม lot + ต้นทุนมาตรฐาน
- เก็บ **snapshot ต่องวด** (`bcstkwarehouseperiod` ฯลฯ) → แก้เอกสารย้อนหลังแล้วคำนวณใหม่เฉพาะตั้งแต่งวดที่กระทบ ไม่ใช่ทั้งชีวิตสินค้า
- นโยบายสต็อกติดลบ 4 แบบ พร้อมการ**ตัดแถวตรงจุดที่ยอดข้ามศูนย์**
- เขียนธง `glstartposting=1` เป็นสะพานส่งต่อให้ GL
- คิวแยกสำหรับงานรายงานและ `functionstock` แบบ async

**ฝั่ง BC (ตรวจใหม่ 2026-09-25):**
- `process_consumer.go` (consumer ที่ถูกคอมเมนต์ทิ้ง) และ `performBackgroundTask()` เดิมไม่มีแล้ว (ลบในรอบถอด MongoDB/Kafka 2026-09-23 ตาม [ADR](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md)) — ปัจจุบันมี worker คำนวณต้นทุนต่อฐาน holding อ่านคิว `stock_dirty` ใน PostgreSQL งานที่ล้มเหลวลง `stock_dead_letter` (`backend/internal/goapi/process/stockengine/worker.go`, `manager.go:36`, เปิดที่ `backend/internal/goapi/bootstrap.go:122`)
- มี snapshot ต่องวดใน `stock_period_balance` (`stockengine/store.go:51-53,244-245`)
- คำนวณเฉพาะถัวเฉลี่ยถ่วงน้ำหนักแยกคลัง (`stockengine/calc.go:1,216`) ไม่มี FIFO/lot/ต้นทุนมาตรฐานเป็นวิธีคิดต้นทุน
- มีนโยบายสต็อกติดลบ 4 แบบตาม `x_stock_amount_type` ของ Champ (`stockengine/calc.go:19-31`) — การตัดแถวตรงจุดที่ยอดข้ามศูนย์ยังไม่ verify
- เอกสารขาย/ซื้อ (`/transaction/*`) backend ถูกถอดพร้อม MongoDB จึงยังไม่มีคิว doc-flow ฝั่งขาย

> ผลที่ตามมา: `processstockcalccost` ยังต้องส่ง `itemcodelist` อย่างน้อย 1 รายการ (`backend/internal/goapi/handlers/process_stock_calc_cost.go:84,181`) — จึงยังทำปุ่ม "คำนวณใหม่ทั้งหมด" ผ่าน API นี้ไม่ได้

---

## 5. เครื่องมือ / workflow

1. **สั่งคำนวณใหม่** — Champ ไม่คำนวณเองตรงนั้น แต่ re-queue เข้า `ProcessStock` ให้ service ทำ
2. **`behindindex`** — ลำดับเอกสารที่ลงวันเดียวกัน มีผลต่อการคิดต้นทุน; BC ใช้เรียงสมุดสต็อกแล้ว (`backend/internal/goapi/process/stockengine/query.go:19`)
3. **`IsLockCost` รายบรรทัด** — ล็อกต้นทุนเฉพาะบรรทัดไม่ให้คำนวณทับ; BC ไม่มี
4. **ปิดงวด** — สร้าง JE ปิดบัญชี แล้วตั้ง `BCPeriod.Status=0` ซึ่ง `IsPeriodClose()` ใช้บล็อกการบันทึก **ทุกโมดูล**
   BC: งวดเก็บใน `gl_records` (kind `periods`) มีธง `locked` และบังคับตอนบันทึกสมุดรายวัน (`checkOpenDate`, `backend/internal/generalledger/postgres_guards.go:43-52`) แต่ **ไม่มีโมดูลอื่นเรียกใช้** (grep `kind='periods'` เจอเฉพาะใต้ `internal/generalledger/`) → ปิดงวด GL แล้วเอกสารสต็อก (`/inventory/*`) ยังบันทึกย้อนหลังได้
5. **ปิดปี** — สร้าง JE ยอดยกมา, ล้างรายการเคลื่อนไหว, เก็บภาษีซื้อไว้ 6 งวด / ภาษีขาย 1 งวด; BC: GL มีประมวลผลสิ้นปี (`/gl/year-end`, `backend/internal/generalledger/postgres_processes.go:45`) สร้างร่างยอดยกมาโดยเก็บประวัติไว้ ไม่ล้างรายการ — การยกภาษีซื้อ/ขายข้ามปีแบบ Champ ยังไม่ verify
6. **ยกเลิกเอกสาร** — PO/SO ใช้ "ใบยกเลิก" แยกใบ มีรหัสเหตุผล ยกเลิกเฉพาะจำนวนคงเหลือ และคืน `ReserveQty`; เอกสารบัญชีใช้ soft-cancel คือคงแถวภาษีไว้แต่ตีศูนย์แล้วทำเครื่องหมาย "ยกเลิก"
7. **การอนุมัติ** — Champ มีแค่ธงเดียว ที่ถูกจุดโดยเงื่อนไข "เกินวงเงินเครดิต" เท่านั้น
   → **Champ ไม่มีระบบอนุมัติหลายลำดับ/ตามวงเงิน** จอ workflow อนุมัติของ BC จึงเป็นของใหม่ ไม่ใช่การไล่ให้ทัน parity

---

## 6. ลำดับที่ควรทำ (ข้อเสนอ — รอลุงจืดตัดสิน)

| ลำดับ | งาน | เหตุผล |
|---|---|---|
| ~~P1~~ | worker ประมวลผลหลังบ้าน | **ทำแล้ว** — worker ต้นทุนสต็อกบนคิว PostgreSQL (§4) |
| ~~P1~~ | snapshot ต่องวด + `behindindex` | **ทำแล้ว** ฝั่งสต็อก (§4, §5 ข้อ 2) |
| P1 | ทะเบียนภาษีแยก + `CancelOutPeriod` | กระทบความถูกต้องของแบบยื่นโดยตรง (`businesscode`/`branchno` ในรายงาน WHT ทำแล้ว — ดู §3) |
| ~~P2~~ | WHT บน PostgreSQL + API (ภ.ง.ด.2/3/53, 50 ทวิ) | **ทำแล้ว** — อ่านจาก GL (`/api/report/tax/wht`, `/api/report/tax/wht/certificate`) |
| P2 | period lock ใช้ข้ามโมดูล | ปิดงวดแล้วต้องปิดจริงทุกจอ |
| P2 | กรอบ filter มาตรฐานของรายงาน + อายุหนี้ตั้งค่าได้ | ผู้ใช้เดิมจาก Champ คาดหวังสิ่งนี้ |
| P3 | 7 รายงานที่ขาด + มิติ "เขตการขาย" | เพิ่มทีหลังได้ ไม่บล็อกงานประจำวัน |

**ไม่ทำ:** อย่าเพิ่ม 60 เมนูรายงานแตกมิติ — ทำเป็นตัวเลือกมิติในจอรายงานเดียว
