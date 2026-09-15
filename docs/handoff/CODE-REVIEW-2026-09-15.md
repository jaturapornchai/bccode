# Code Review — โค้ดใหม่ (commit 6ec32b12, 84c9ebbb, a79b5279) — 2026-09-15

รีวิวโค้ดใหม่ 3 ชุด: (1) stock costing engine v2, (2) kafka consumer subscriptions, (3) ERP จอจริงแทน mock
สถานะตรวจ: `go test ./internal/goapi/process/stockengine/` **ผ่าน** (unit; integration ข้ามเพราะไม่มี DB)
หมายเหตุ: ข้อที่เป็น "คำถาม business rule" ห้ามแก้เองโดยเดา — ต้องถามลุงจืดตามกฎ AGENTS.md

ความรุนแรง: 🔴 critical | 🟠 high | 🟡 medium | ⚪ low

---

## สถานะการแก้ (อัปเดต 2026-09-15 หลังย้ายรายงานมาอ่าน `stock_ledger`)

| ข้อ | สถานะ | แก้ที่ไหน / หลักฐาน |
|-----|-------|----------------------|
| A1 race ลบ dirty | ✅ แก้แล้ว | `stock_dirty.markedat` + `Persist(..., dirtyBefore)` ลบเฉพาะงานที่เก่ากว่าเวลาที่อ่านข้อมูล — เทส `TestDocumentArrivingDuringCalculationStaysQueued` |
| A2 งวดก่อนหน้าวันที่ 29–31 | ✅ แก้แล้ว | เลิกใช้ `AddDate(0,-1,0)` เปลี่ยนเป็นอ่านยอดปลายงวดล่าสุดที่ `periodkey <` งวดปัจจุบัน (`LoadOpening`) — เทส `TestOpeningBalanceComesFromTheLatestClosedPeriod` |
| A3 โอนคลังหายครึ่งขา | ✅ แก้แล้ว | ตัวรับเอกสารสร้างสองขา (`MapStockTransferToDocDetailStructs`) + `DirectionOfMovement` อ่าน `calcflag` + ขาเข้ารับของด้วยต้นทุนของต้นทาง — เทส `TestWarehouseTransferKeepsTotalStockAndValue`, `TestWarehouseTransferMovesStockBetweenWarehouses` |
| A4 ลบเอกสารแล้วสต็อกไม่คืน | ✅ แก้แล้ว | `LoadMovements` กรอง `isdelete` + `MarkDocumentDirty` ตอนลบ — เทส `TestDeletedDocumentLeavesTheLedger` |
| A5 แก้เอกสารข้ามงวด | ✅ แก้แล้ว | `mypg.DeleteDocPgSql` ทำเครื่องหมายงานคิดใหม่ด้วยวันที่เดิม **ก่อน** ลบ จึงคิดใหม่ตั้งแต่ `min(วันเดิม, วันใหม่)` |
| A6 retry ไม่มี backoff | ✅ แก้แล้ว | `ReleaseDirty` พักงานตามจำนวนครั้ง (30 วิ × attempts) — เทส `TestReleaseReturnsWorkToQueueAfterBackoff` |
| A7 dead letter ผิดคอลัมน์ | ✅ แก้แล้ว | ตารางใหม่ `stock_dead_letter` (business/item/fromdate/attempts/lasterror) แทนการยัดลง `deadletterqueue` ของเอกสาร |
| A8 opening = 0 ทั้งที่มีประวัติ | ✅ แก้แล้ว | ไม่มียอดยกมา → ถอยไปคิดตั้งแต่เอกสารใบแรก (`earliestMovement`) — เทส `TestRecalculationReachesBackWhenNoOpeningExists` |
| A9 starvation | ✅ แก้แล้ว | `enqueuedat` ไม่เลื่อนตามการแตะครั้งหลังแล้ว (ใช้ `markedat` แทน) |
| A10 validate ก่อนเช็ก transflag | ✅ แก้แล้ว | `EnqueueStockRecalculation` + `ProcessDocumentStockCalculation` ข้ามบรรทัดที่ไม่กระทบสต็อกก่อนตรวจ |
| A11 queue status | ✅ แก้แล้ว | `MIN(enqueuedat) FILTER (pending)` + กรอง `businesscode` จริง |
| A12 จุดเล็ก | ⚠️ แก้เกือบครบ | manager ใช้ generation token แล้ว, มีเทสว่ารายชื่อ transflag สองชุดต้องเท่ากัน (`KnownTransFlags`), integration test รันจริงกับ postgres:18 แล้ว — **เหลือ** `utils.go` ยังส่ง `context.Background()` ไม่ใช่ ctx ของ consumer (ทุก consumer สร้าง `context.Background()` เองอยู่แล้ว ผลกระทบต่ำ) |
| B1 topic รับสินค้าที่ตายแล้ว | ✅ แก้แล้ว | ยืนยันว่าไม่มีทั้งผู้ส่งและผู้รับ `when-purchasereceive-*` ในระบบ (เอกสารรับสินค้าเดินผ่าน purchase partial) → ลบ constants + แก้รายการใน status สองจุด |
| B2 เทสว่า handler เรียก hook ครบ | ⚠️ เปลี่ยนวิธี | การตั้งงานคิดใหม่ย้ายไปอยู่ที่ทางผ่านร่วม (`mypg.DeleteDocPgSql`, `DeleteDocumentFromDatabases`) แล้ว จึงไม่ต้องไล่เทสทีละ handler — ยังไม่มีเทสเชิงโครงสร้างยืนยัน |
| C1 tax report กลืนแถว | ✅ แก้แล้ว | คืน 500 + เช็ก `rows.Err()` |
| C2 business rule ภาษี | ❌ ยังไม่แก้ | ต้องถามลุงจืด (0%/ยกเว้น, isdelete, ใบลดหนี้ใน ภ.พ.30, เขตเวลา) |
| C3 i18n backend (gl_poster) | ❌ ยังไม่แก้ | งานย้ายข้อความเป็น key ทั้งชุด ทำแยกคอมมิต |
| C4 gl_poster จุดเสี่ยง | ⚠️ แก้บางส่วน | ตรวจรูปแบบวันที่ก่อนตัด `[:4]` และไม่กลืน error ตอนเปลี่ยนสถานะเป็นจำหน่ายแล้ว — **เหลือ** การนับค่าเสื่อมสะสมที่ยังไม่ผ่านรายการ (เป็นคำถามทางบัญชี ต้องถามก่อน) |
| D1 ป้าย `{th,en}` | ❌ ยังไม่แก้ | งาน i18n frontend ทำแยก |
| D2 timeout ของ proxy | ✅ ตรวจแล้ว | `proxyMainApiJson` มี AbortSignal 20 วินาที — แต่ `processstockcalccost` เป็นงานหนักที่รันแบบรอผล ควรเปลี่ยนจอไปใช้คิว + `queue-status` แทน (งาน frontend ที่ยังค้าง) |

---

## A. Stock Engine v2 (`backend/internal/goapi/process/stockengine/`)

### 🔴 A1. Race: MarkDirty ที่มาระหว่างคำนวณถูกลบทิ้ง → เอกสารไม่เข้า stock_ledger ถาวร
- **ไฟล์:** `store.go:170-174` (Persist ลบ `stock_dirty` แบบไม่มีเงื่อนไขใน tx เดียวกับเขียน ledger)
- **สถานการณ์:** worker claim งานตอน T1 → `LoadMovements` (snapshot) → consumer อีกตัว insert เอกสารใหม่ของสินค้าเดียวกันตอน T2 แล้ว `MarkDirty` (สำเร็จ เพราะ ON CONFLICT อัปเดตแถวเดิม) → worker `Persist` commit ตอน T3 ลบแถว dirty ทิ้ง **ทั้งที่ snapshot ไม่มีเอกสารของ T2** → เอกสารใบนั้นไม่มีใน `stock_ledger` จนกว่าจะมีเอกสารอื่นมาแตะสินค้าตัวเดิมอีก
- advisory lock กันไม่ได้ เพราะ `MarkDirty` ไม่ได้จอง lock เดียวกัน
- **แก้:** จับเวลา snapshot ตอน claim/อ่าน movements แล้วเปลี่ยน DELETE เป็น
  `DELETE FROM stock_dirty WHERE businesscode=$1 AND itemcode=$2 AND enqueuedat <= $snapshotTime`
  (MarkDirty อัปเดต `enqueuedat = now()` ทุกครั้งอยู่แล้ว จึงรอดจากเงื่อนไขนี้เมื่อมาสาย) — หรือหลัง commit ให้ตรวจว่ามี dirty ใหม่กว่า snapshot แล้ว mark ซ้ำ
- **เพิ่มเทส:** integration test ยิง MarkDirty คั่นกลางระหว่าง LoadMovements กับ Persist แล้ว assert ว่างานยังค้าง

### 🔴 A2. `previousPeriodKey` ผิดเมื่อ From เป็นวันที่ 29–31 → อ่านงวดปัจจุบันเป็นยอดยกมา → เบิ้ลยอด
- **ไฟล์:** `store.go:38-40` — `PeriodKeyOf(from.AddDate(0, -1, 0))`
- **บั๊ก:** Go `AddDate(0,-1,0)` กับ 2024-03-31 ได้ 2024-02-31 → normalize เป็น **2024-03-02** → PeriodKey = "2024-03" (งวดเดียวกับ From!) ทั้งที่ต้องเป็น "2024-02"
- **ผลกระทบ:** รอบแรกอ่าน opening ไม่เจอ (ถือ 0) — รอด; แต่รอบถัดไป `LoadOpening` อ่าน closing **ของงวดปัจจุบัน** มาเป็นยอดตั้งต้น แล้วเอา movements ตั้งแต่ต้นงวดมาบวกซ้ำ → ยอดเบิ้ลทุกครั้งที่ dirty ตกวันที่ 31 (และ 29–30 ของเดือนสั้น)
- **แก้:** `func previousPeriodKey(from time.Time) string { return PeriodKeyOf(periodStart(from).AddDate(0, 0, -1)) }`
- **เพิ่มเทส:** table-driven วันที่ 31 มี.ค., 31 พ.ค., 30 เม.ย., 29 ก.พ. ฯลฯ

### 🔴 A3. เอกสารโอนคลัง (transflag 72) ตัดออกจากคลังต้นทางอย่างเดียว คลังปลายทางไม่เคยได้รับของ
- **ไฟล์:** `handlers/kafka/stock_transfer.go:211-245` สร้าง docdetail **บรรทัดเดียว** (คลังต้นทาง) — comment ในโค้ดยอมรับเอง ("For now, create a single entry"); `calc.go:47` ให้ 72 = DirectionOut เสมอ; engine ไม่เคยดู `CalcFlag`
- **ผลกระทบ:** สต็อกรวมทั้งบริษัทหายไปทุกครั้งที่โอนคลัง
- **แก้ (ต้องตัดสินใจเชิงออกแบบ ถามลุงจืดก่อน):** โอนคลังต้องเกิด 2 ขา — ออกจากคลังต้นทาง + เข้าคลังปลายทาง ต้องรู้ว่า docdetail เก็บคลังปลายทางไว้ field ใด (towhcode?) แล้วให้ engine ขยาย 1 บรรทัดเป็น 2 ledger rows (Out ต้นทางด้วยต้นทุนเฉลี่ย, In ปลายทางด้วยต้นทุนเดียวกัน)
- **เพิ่มเทส:** โอน A→B แล้วยอดรวม A+B ต้องเท่าเดิม

### 🔴 A4. ลบเอกสารแล้วสต็อกไม่เด้งกลับ — ทั้งไม่กรอง isdelete และไม่ mark dirty
- **ไฟล์:** `store.go:88` กรองเฉพาะ `COALESCE(h.iscancel,FALSE)=FALSE` — ไม่กรอง `h.isdelete`; `handlers/kafka/utils.go:384-415` (`DeleteDocumentFromDatabases`) set `doc.isdelete=true` แล้วจบ ไม่เรียก `MarkDirty`
- **ผลกระทบ:** เอกสารที่ถูกลบยังคงคิดสต็อก/ต้นทุนใน stock_ledger ตลอดไป
- **แก้:** (1) เพิ่ม `AND COALESCE(h.isdelete,FALSE)=FALSE` ใน `LoadMovements` (ตรวจว่า doc มีคอลัมน์ isdelete จริง) (2) consumer ฝั่ง delete ทุกประเภทเอกสารที่ movesStock ต้อง enqueue recalculation ด้วย (อ่าน docdetail ของเอกสารก่อนลบ หรือ mark ตาม docno ที่ลบ)
- **เพิ่มเทส:** สร้างเอกสาร → คิดต้นทุน → ลบเอกสาร → ยอดต้องกลับเท่าเดิม

### 🟠 A5. แก้เอกสารข้ามงวด ทิ้ง ledger เก่าค้างในงวดเดิม
- **ไฟล์:** flow update ของทุก consumer (เช่น `sale_invoice.go:83` ลบเอกสารเดิม → insert ใหม่ → `ProcessDocumentStockCalculation` ด้วยวันที่**ใหม่**)
- **สถานการณ์:** เอกสารย้ายจาก 5 ม.ค. เป็น 20 ก.พ. → `MarkDirty(from=20 ก.พ.)` → recalc ตั้งแต่ 1 ก.พ. → ledger แถวเดิมในเดือน ม.ค. (และ closing งวด ม.ค.) ไม่ถูกแตะ → ยอดเดิมค้าง + ยอดใหม่เข้ามา = เบิ้ล
- **แก้:** ก่อน `DeleteDocPgSql` ให้อ่าน `docdatetime` เดิมไว้ แล้ว `MarkDirty` ด้วย `min(วันเดิม, วันใหม่)`

### 🟠 A6. Retry ไม่มี backoff — error ชั่วคราวไหลเข้า dead letter ภายในไม่กี่มิลลิวินาที
- **ไฟล์:** `worker.go:75-89` (inner loop drain ต่อเนื่อง) + `dirty.go:99-112` (`ReleaseDirty` เคลียร์ lease เป็น NULL ทันที)
- **สถานการณ์:** DB สะดุด 1 วิ → claim→fail→release→claim ใหม่ทันที ครบ 5 attempts ใน loop เดียว → `MoveToDeadLetter` ทั้งที่เป็น error ชั่วคราว
- **แก้:** `ReleaseDirty` ตั้ง `leaseuntil = now() + (attempts * 30 วินาที)` แทน NULL (lease หมดอายุแล้วจะถูก claim ใหม่เองตามกลไกเดิม)

### 🟠 A7. Dead letter เขียนข้อมูลผิดคอลัมน์
- **ไฟล์:** `dirty.go:127-130` — ใส่ `businesscode` ลงคอลัมน์ `holdingcode`, `itemcode` ลง `docno`, `transflag='0'` (string)
- **ผลกระทบ:** จอ/รายงานที่อ่าน `deadletterqueue` ตีความผิดหมด; holding จริงไม่ถูกบันทึก
- **แก้:** ส่ง holdingCode เข้ามาด้วย (Worker มีอยู่แล้ว) แมปคอลัมน์ให้ถูก หรือสร้างตาราง dead letter เฉพาะของ stock engine

### 🟠 A8. Opening = 0 ทั้งที่มีประวัติก่อน From (ครั้งแรกที่ deploy หรือสินค้าที่เพิ่งถูกแตะ)
- **ไฟล์:** `store.go:50-71` (`LoadOpening` คืน map ว่างถ้าไม่มีแถวใน `stock_period_balance`)
- **สถานการณ์:** ระบบเพิ่งเปิดใช้ — ไม่มี period balance เลย; เอกสารใหม่วันนี้ mark dirty from=วันนี้ → recalc เฉพาะงวดนี้ โดย opening=0 ทั้งที่สินค้ามียอดสะสมหลายปีใน docdetail → ยอดคงเหลือผิดตั้งแต่แถวแรก
- **แก้:** ถ้าไม่มี period balance ของงวดก่อน **และ** มี docdetail ก่อน `periodStart(From)` → ขยาย From ไปที่เอกสารแรกสุด (หรือมี migration seed ยอดยกมาก่อนเปิดระบบ — ต้องมีแผน backfill ประกอบ ADR)

### 🟡 A9. Starvation: สินค้าที่ถูกแตะบ่อยไม่เคยได้คำนวณ
- **ไฟล์:** `dirty.go:44-46` (MarkDirty อัปเดต `enqueuedat = now()` ทุกครั้ง) + `dirty.go:75` (`ORDER BY enqueuedat`)
- **แก้:** อย่าอัปเดต `enqueuedat` ตอน conflict (เก็บเวลา enqueue แรก) หรือเรียงตาม `fromdate`

### 🟡 A10. `EnqueueStockRecalculation` ตรวจ business/item ก่อนเช็ก transflag → เอกสารที่ไม่กระทบสต็อกทำให้ทั้ง batch พัง
- **ไฟล์:** `stock_engine_hook.go:30-35` — return error ถ้า business/item ว่าง **ก่อน** `DirectionOf` skip; ใบสั่งซื้อ/ใบสั่งขายที่ไม่มี itemcode บางบรรทัดจะทำให้ทั้ง enqueue ล้ม
- **แก้:** ย้าย validation ไปหลังผ่าน `DirectionOf` แล้ว (เช็กเฉพาะเอกสารที่กระทบสต็อกจริง)

### 🟡 A11. Queue status คลาดเคลื่อนเล็กน้อย
- **ไฟล์:** `dirty.go:154-160` — `MIN(enqueuedat)` รวมแถวที่กำลัง process ด้วย (OldestWait ควรจำกัดเฉพาะ pending); `handlers/stock_queue_status.go:49` รับ `businesscode` แต่ไม่ใช้กรอง → คิวรวมทั้ง holding
- **แก้:** เติมเงื่อนไข pending ใน MIN; ตัดสินใจว่าจะกรอง business หรือเอาพารามิเตอร์ออก

### ⚪ A12. จุดเล็กอื่น ๆ
- `manager.go:50,91-96` — race ตอน Stop/Start: goroutine worker เก่า defer `delete(running, holdingCode)` ลง map **ใหม่** ได้ → spawn ซ้ำ; แก้ด้วย generation token
- `worker.go:26` — `AfterRecalculate` เป็น global var; ต้อง set ก่อน `StartWorkers` เท่านั้น (เขียน comment กำกับหรือทำ setter ที่ panic ถ้าเรียกหลัง start)
- `utils.go:182` — `EnqueueStockRecalculation(context.Background(), ...)` ไม่ส่งต่อ ctx ของ consumer (timeout/cancel หลุด)
- `calc.go:208-210` + `store.go:155-160` — recalc ลบ ledger ตั้งแต่ต้นงวดทั้งหมดของสินค้านั้น รวม transflag ที่ไม่อยู่ในตัวกรอง; วันนี้ list ตรงกัน (`myglobal.TransFlagsToProcess` = DirectionOf keys) แต่ถ้าอนาคตถอด flag ออกจาก list ประวัติจะหายเงียบ — เขียนเทสกันไว้ว่า 2 list ต้องเท่ากันเสมอ
- `store_integration_test.go` — integration tests ถูก skip เมื่อไม่มี DB; ควรมี CI job ที่ยิง Postgres จริง (testcontainers) เพราะบั๊ก A1/A2 จับได้เฉพาะ integration

---

## B. Kafka consumers (`backend/internal/goapi/handlers/kafka/`)

### 🟡 B1. `TOPIC_PURCHASE_RECEIVE_*` เป็น dead constants + status endpoint โฆษณาเกินจริง
- **ไฟล์:** `constants.go:43-46` ประกาศ topic แต่ไม่มี handler ชื่อ PurchaseReceive ใน package นี้เลย; `manager.go:30-118` ไม่มี consumer group; แต่ `api.go:301-303` กับ `handlers/kafka.go:1227-1229` ยัง list ใน "purchase"
- **แก้:** ยืนยันว่ารับสินค้า (purchase receive) ไหลผ่าน service อื่น (`internal/transaction/transactionconsumer/purchasereceive`) จริงหรือไม่ — ถ้าใช่ ลบ constants/รายการ status ออก; ถ้าไม่ใช่ = เอกสารยังถูกลืมอีกประเภท (ประเด็นเดียวกับที่ commit นี้ตั้งใจแก้)

### 🟡 B2. ยังไม่มีเทสยืนยันว่า "ทุก handler ที่ movesStock เรียก ProcessDocumentStockCalculation จริง"
- commit 84c9ebbb เพิ่ม `manager_test.go` ตรวจ subscription แต่ไม่ได้ตรวจว่า handler แต่ละตัวเรียก hook คิดต้นทุน (เช่น delete flow ขาดตาม A4)
- **แก้:** เพิ่มเทสเชิงโครงสร้าง (grep-based หรือ interface) ว่า handler ทุกตัวในกลุ่ม movesStock เรียก `ProcessDocumentStockCalculation` ทั้ง create/update **และ** delete

---

## C. Backend รายงานภาษี + สินทรัพย์ถาวร (commit a79b5279)

### 🟡 C1. `tax_report.go` กลืนแถวที่ scan ไม่สำเร็จเงียบ ๆ และไม่เช็ก `rows.Err()`
- **ไฟล์:** `handlers/tax_report.go:113-126` — `rows.Scan` error → `continue`; รายงานภาษีขาดแถวโดยไม่มีใครรู้
- **แก้:** return error 500 แทน continue; เช็ก `rows.Err()` หลัง loop

### 🟡 C2. คำถาม business rule ของรายงานภาษี (ห้ามแก้เอง — ถามลุงจืด)
- `buildVatRegisterQuery` กรอง `totalvatvalue <> 0` → เอกสาร 0%/ยกเว้นหายจากทะเบียนภาษี ทั้งที่ทะเบียนภาษีขาย/ซื้อปกติต้องแสดงครบ
- ไม่กรอง `isdelete` (กรองแค่ `iscancel`)
- ใบลดหนี้ (รับคืน/ส่งคืน) ไม่ได้รวมใน ภ.พ.30 summary — ยืนยันว่า return ถูกเก็บในตารางเดียวกันเป็นยอดติดลบหรือเป็นตารางแยก
- `INTERVAL '7 hour'` hard-code — ตรงกับ TimeZone=UTC rule แต่ควรอ้างอิงค่ากลาง (มี `businessLocation` ใน stockengine แล้ว)

### 🟡 C3. `gl_poster.go` เพิ่ม `fmt.Errorf` ภาษาไทยใหม่หลายสิบจุด ขัดกฎ i18n ของ backend
- **ไฟล์:** `internal/fixedasset/gl_poster.go` (เช่น :90, :109, :136, :223, :301, :318, :367, :518 ฯลฯ) — กฎ AGENTS.md: backend ต้องคืน key + `language.Text(key, lang)` ไม่ใช่ข้อความไทยตรง ๆ; โค้ดใหม่ควรไม่เพิ่มจำนวน 294 จุดเดิม
- **แก้:** เปลี่ยนเป็น error key แล้วเพิ่มแถวใน `backend/assets/language/languages.tsv` ครบ 13 คอลัมน์

### 🟡 C4. `gl_poster.go` จุดเสี่ยงอื่น
- `:539, :559` — `disposal.DisposalDate[:4]` panic ถ้าวันที่สั้นกว่า 4 ตัวอักษร → validate รูปแบบ `YYYY-MM-DD` ก่อน
- `:590` — อัปเดตสถานะสินทรัพย์เป็น disposed แล้ว **กลืน error** (`_, _ =`) → ขายแล้วแต่สถานะไม่เปลี่ยน
- `:299-302` — GL post สำเร็จแต่ `UpdateMany` mark isposted ล้ม → ข้อมูลไม่สอดคล้อง (journal ออกแล้ว แต่รายการดูเหมือนยังไม่ผ่านรายการ); idempotent replay ช่วยบางส่วน แต่ควร log/แจ้งชัดและมีคำสั่ง reconcile
- `:395-405` — คำนวณต้นทุนสะสมจาก schedule item ล่าสุดโดยไม่สน `isposted` → งวดที่ยังไม่ผ่านรายการถูกนับรวมใน disposal ได้

---

## D. Frontend (commit a79b5279)

### 🟡 D1. ยังใช้ป้าย `{ th, en }` สองภาษาในไฟล์ที่แก้ (รองรับ 2 จาก 12 ภาษา)
- **ไฟล์:** `frontend/src/lib/erp-reports.ts:8-444` (ทุก config มี `label/title/description: { th, en }`) — ไม่ใช่ของที่เพิ่มใน commit นี้ (diff ไม่มีไทยเพิ่มเลย ตรวจแล้ว) แต่ตามกฎ AGENTS.md ห้ามใช้แพทเทิร์น `t(th, en)` และพบจอเดิมที่ยังใช้ต้องอัปเกรด
- **แก้:** ย้ายเป็นคีย์ใน `languages.tsv` + `backendText()` (มี reference implementation ที่ `system-settings/utils.ts` `fieldLabel()`)

### ⚪ D2. `route.ts` proxy — ตรวจ timeout ของ upstream
- **ไฟล์:** `frontend/src/app/api/goapi/[...goPath]/route.ts` — allowlist/sanitize ดีแล้ว; แต่ไม่มี timeout ชัดเจนรอบ `proxyMainApiJson` (ตรวจว่า `workspace-api.ts` ใส่ AbortSignal ไว้หรือไม่); POST allowlist มี `processstockcalccost` (งานหนัก) — ยืนยันว่าตั้งใจเปิด

---

## สรุปลำดับแก้ (แนะนำ)

1. **A2** (บั๊กวันที่ 29–31) — 1 บรรทัด + เทส กระทบยอดเงินทุกงวด
2. **A1 + A4** (race ลบ dirty / ลบเอกสารไม่คืนสต็อก) — เอกสารหายจาก ledger ถาวร
3. **A3** (โอนคลังหายครึ่งขา) — ต้องถามลุงจืดเรื่อง field คลังปลายทางก่อนเขียนโค้ด
4. **A5, A6, A7, A8** — ความถูกต้องของยอดย้อนหลัง + ความทนทานของคิว
5. **C3** (i18n backend) + **C4** — กฎโปรเจ็กต์ + panic risk
6. เหลือเป็น medium/low ตามตาราง
