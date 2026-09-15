# ADR 2026-09-15 — เครื่องยนต์ประมวลผลหลังบ้าน: แทน timer ของ BCProcess ด้วย Event + Dirty-Set

สถานะ: **ถูกแทนที่แล้ว (superseded)** โดย [2026-09-15-stock-cost-engine-v2.md](2026-09-15-stock-cost-engine-v2.md) — ฉบับนั้นออกแบบใหม่ทั้งชุดตามคำสั่งลุงจืด "ให้ทำงานถูกต้อง และเร็ว" ฉบับนี้เสนอแค่ worker บนโครงเดิมที่ยังมีบั๊กความถูกต้อง 10 ข้อ · เดิมสถานะ: ข้อเสนอ · ผู้เสนอ: Claude (Fable) · เกี่ยวข้อง: [20-champ-parity-gap.md](../20-champ-parity-gap.md) §4

## บริบท

ต้นแบบ Champ ใช้ Windows Service (`BCProcess\`, `BCWinserviceInstaller\`) วน timer 6 thread ทุก ~1 วินาที
คอยสแกนหาว่า "มีอะไรต้องทำไหม" + watchdog ทุก 10 วินาที — เป็น **polling trick** ที่ทำงานได้เพราะเป็น
โปรแกรมเครื่องเดียว ฐานข้อมูลเดียว

ระบบใหม่มี Kafka + PostgreSQL ต่อ holding อยู่แล้ว จึงไม่ควรลอก timer มาตรง ๆ

## สภาพปัจจุบันที่ตรวจแล้ว (2026-09-15)

| ส่วน | สภาพ | หลักฐาน |
|---|---|---|
| ขนส่งเหตุการณ์ | **มีแล้ว** — mainapi เขียน Mongo → publish topic `when-<entity>-<action>` | `backend/internal/debtaccount/creditor/config/creditor_message_queue_config.go:4-9` |
| Projection ฝั่ง mainapi | **มีแล้ว** 20+ consumer เขียนลง `stocktransaction`/`stocktransactiondetail` | `backend/internal/transaction/transactionconsumer/*`, `backend/internal/transaction/models/stock_transaction_postgres.go:48,88` |
| Projection ฝั่ง goapi | **ต่อไม่ครบ** — เขียน `docdetail` ซึ่งเป็น**ตารางที่เครื่องคิดต้นทุนใช้จริง** แต่ subscribe แค่ 14 จาก 50 handler | `backend/internal/goapi/handlers/kafka/manager.go:10-112` เทียบกับ `handlers/kafka/*.go` |
| คิวงาน PG | **โครงครบแต่ไม่มีคนใช้** — `queues` + `deadletterqueue` + `distributedlocks`, `PopFromQueue` ใช้ `FOR UPDATE SKIP LOCKED` | `backend/internal/goapi/mypostgres/init_schema.go:19-59`, `mypostgres/queue.go:100` |
| ผู้ผลิตงานเข้าคิว | มีจุดเดียวทั้งระบบ | `backend/internal/goapi/handlers/kafka/purchase_order.go:155` |
| ผู้บริโภคคิว | **ไม่มี** | `handlers/process_consumer.go:100-112` คอมเมนต์ทิ้ง, `handlers/health.go:90-93` `return nil` |
| ตัวคิดต้นทุน | มี รับทีละ item | `process/process-stock/process-stock-calc-cost.go:49 ProductCalcCostIncrementalCompany` |
| ตัวจับว่าอะไรเปลี่ยน | มี checksum ต่อ item แต่ต้องอ่านประวัติทั้งชีวิตสินค้าเพื่อคำนวณ | `process/process-stock/incremental.go:38-80`, ตาราง `stockcalculationstate` (`:268`) |

### ปัญหาที่ใหญ่กว่า "ไม่มี worker"

`docdetail` เขียนโดย goapi เท่านั้น (`handlers/kafka/utils.go:344`) และ subscribe ไม่ครบ →
**เอกสารเคลื่อนไหวสต็อกไม่เคยไหลเข้า `docdetail` ผ่าน Kafka เลย**: โอนคลัง, ปรับสต็อก (เพิ่ม/ลด),
รับ/เบิก/คืนสินค้า, **ยอดยกมา (transflag 54)**, ซื้อรับ/ซื้อคืน/ซื้อบางส่วน, ใบขอซื้อ, RFQ
ที่ไหลเข้าคือ ขาย/รับคืนขาย/ใบสั่งขาย/ซื้อ/สินค้า/คลัง เท่านั้น
→ ต่อให้เปิด worker วันนี้ ก็คำนวณจากข้อมูลที่ขาดครึ่ง

## ทางเลือกที่พิจารณา

1. **ลอก timer มาเลย** (goroutine ticker 1 วิ) — ง่ายสุด แต่ยิง query รัวเปล่า ๆ ทุกวินาที × ทุก holding และ checksum ปัจจุบันต้องอ่านประวัติทั้งหมดต่อ item → ยิ่ง scale ยิ่งพัง **ปฏิเสธ**
2. **ใช้ Kafka เป็นคิวงานตรง ๆ** (topic `stock-recalc`, worker เป็น consumer) — Kafka **ยุบงานซ้ำไม่ได้**: เอกสาร 500 ใบที่แตะสินค้าเดียวกัน = 500 message ต้องคิด 500 รอบ (จะ dedupe ในหน่วยความจำก็หายตอน restart) และ **ถามไม่ได้ว่าตอนนี้ค้างกี่ตัว** ซึ่งจอเครื่องมือของผู้ใช้ต้องใช้ **ปฏิเสธเป็นตัวหลัก**
3. **Kafka = ขนส่งเหตุการณ์ + PostgreSQL = สมุดงานค้างแบบยุบซ้ำได้ + worker ตื่นด้วย NOTIFY** ← **เลือกข้อนี้**

## การตัดสินใจ

### ชั้นที่ 1 — เหตุการณ์ (ใช้ของเดิม ไม่สร้างใหม่)
mainapi เขียน Mongo → publish `when-*` → goapi consumer project ลง `docdetail`
**งานที่ต้องทำคือ subscribe ให้ครบ** ใน `manager.go` ไม่ใช่เขียน consumer ใหม่ (handler เขียนไว้ครบแล้ว 50 ตัว)

### ชั้นที่ 2 — สมุดงานค้าง (Dirty-Set) แทน timer
ตารางใหม่ใน PG ของแต่ละ holding:

```sql
CREATE TABLE IF NOT EXISTS stockdirty (
    businesscode varchar(50)  NOT NULL DEFAULT '',
    itemcode     varchar(100) NOT NULL,
    dirtyfrom    timestamptz  NOT NULL,          -- วันเวลาที่เก่าสุดที่กระทบ
    reason       varchar(30)  NOT NULL,          -- doc | manual | period-close | backfill
    enqueuedat   timestamptz  NOT NULL DEFAULT now(),
    attempts     int          NOT NULL DEFAULT 0,
    leaseowner   varchar(64),
    leaseuntil   timestamptz,
    PRIMARY KEY (businesscode, itemcode)
);
CREATE INDEX IF NOT EXISTS idxstockdirtyready ON stockdirty(enqueuedat) WHERE leaseuntil IS NULL;
```

ตอน consumer project แถวลง `docdetail` เสร็จ ให้ยิงต่อ:

```sql
INSERT INTO stockdirty (businesscode, itemcode, dirtyfrom, reason)
VALUES ($1, $2, $3, 'doc')
ON CONFLICT (businesscode, itemcode)
DO UPDATE SET dirtyfrom = LEAST(stockdirty.dirtyfrom, EXCLUDED.dirtyfrom),
              enqueuedat = now();
SELECT pg_notify('stockdirty', $1 || '|' || $2);
```

`LEAST(...)` คือหัวใจ — เอกสาร 500 ใบที่แตะสินค้าเดียวกันเหลือ **1 แถว** และคำนวณใหม่
**เฉพาะตั้งแต่วันเก่าสุดที่กระทบ** ไม่ใช่ทั้งชีวิตสินค้า (สิ่งที่ timer + checksum ทำไม่ได้)

### ชั้นที่ 3 — worker แทน 6 thread
- ตื่นด้วย `pq.NewListener` + `LISTEN stockdirty` (มี `github.com/lib/pq v1.10.9` ใน `backend/go.mod:17` อยู่แล้ว ไม่ต้องเพิ่ม dependency) **ไม่ใช่ ticker**
- poll สำรองทุก 15 วินาที กัน NOTIFY หายตอน reconnect (นี่คือ watchdog ที่ Champ ทำทุก 10 วิ)
- จองงานแบบ lease ไม่ถือ transaction ยาว:

```sql
UPDATE stockdirty SET leaseowner = $1, leaseuntil = now() + interval '5 minutes', attempts = attempts + 1
WHERE (businesscode, itemcode) IN (
    SELECT businesscode, itemcode FROM stockdirty
    WHERE (leaseuntil IS NULL OR leaseuntil < now()) AND attempts < 5
    ORDER BY enqueuedat LIMIT $2 FOR UPDATE SKIP LOCKED)
RETURNING businesscode, itemcode, dirtyfrom, attempts;
```

- เรียก `ProductCalcCostIncrementalCompany(db, holding, business, itemcode, ...)` ต่อ item
- สำเร็จ → `DELETE` แถว · ล้มเหลว → ปล่อย lease หมดอายุเอง (retry อัตโนมัติ) · `attempts >= 5` → ย้ายเข้า `deadletterqueue` (ตารางมีแล้ว `init_schema.go:34`)
- เปิด/ปิดด้วย env `BCAI_WORKER_ENABLED`; รันเป็น goroutine ใน goapi ก่อน (ยังไม่แยก binary)

### Kafka ช่วยตรงไหนที่ timer ทำไม่ได้
1. **ขนส่งข้ามบริการ** — mainapi ไม่ต้องรู้จัก PG ของ goapi (ของเดิม ใช้ต่อ)
2. **เมื่อต้อง scale หลาย worker**: เพิ่ม topic `bc-stock-recalc-v1` โดยใช้ **partition key = `holding|business|itemcode`** → สินค้าตัวเดียวกันตกที่ partition เดิมเสมอ = ไม่มีทางคำนวณชนกัน **โดยไม่ต้องใช้ตาราง `distributedlocks` เลย**
3. **replay** — retention 7 วัน ย้อน offset สร้าง dirty-set ใหม่ได้ถ้า PG เสีย

> **เริ่มด้วย PG-only worker ตัวเดียวต่อ holding ก่อน** แล้วค่อยเพิ่ม Kafka topic เมื่อวัดได้ว่าตัวเดียวไม่พอ — ไม่สร้าง topic ล่วงหน้า (YAGNI)

### แผนที่จาก 6 thread ของ Champ

| Champ | BC |
|---|---|
| ProcessStock (ต้นทุน) | `stockdirty` + worker `stock` |
| ProcessDoc (สถานะเอกสาร) | `docdirty` → `process-doc.ProcessDocumentStatusByDocNo` (มีแล้ว) |
| GL posting (`glstartposting=1`) | `gldirty` → topic `bc-gl-projection-v1` (มีแล้วที่ `generalledger/kafkatransport/transport.go:26`) |
| คิวรายงาน | `reportjobs` (มี status/progress ให้จอ poll) |
| FunctionStock | รวมใน `stockdirty` แยกด้วย `reason` |
| Watchdog 10 วิ | `leaseuntil` หมดอายุ = retry เอง + endpoint `GET /goapi/api/process/queue-status` ให้จอเครื่องมือแสดงคิวค้างจริง |

## ลำดับการทำ

| ขั้น | งาน | ทำไมต้องลำดับนี้ |
|---|---|---|
| 0 | **subscribe kafka handler ให้ครบใน `manager.go`** | ไม่มีข้อมูลก็ไม่มีอะไรให้ worker ทำ — สำคัญกว่าตัว worker |
| 1 | ตาราง `stockdirty` + UPSERT ตอน project เสร็จ | สร้างงานค้างให้ถูกยุบซ้ำตั้งแต่ต้นทาง |
| 2 | worker + LISTEN/NOTIFY + lease + DLQ | ตัว engine |
| 3 | `GET /goapi/api/process/queue-status` + ต่อจอเครื่องมือ | ผู้ใช้เห็นคิวจริง แทน progress bar ปลอม |
| 4 | `behindindex` ใน `docdetail` + ใส่ใน ORDER BY ของตัวคิดต้นทุน | ลำดับเอกสารวันเดียวกันมีผลต่อต้นทุน (ปัจจุบันเรียง `docdatetime, linenumber, docno`) |
| 5 | `stockperiodbal` (snapshot ต่องวด) → `dirtyfrom` เปลี่ยนความหมายเป็น "งวดที่เริ่มคำนวณใหม่" | ทำให้แก้ย้อนหลังไม่ต้องคิดใหม่ทั้งชีวิตสินค้า |
| 6 | นโยบายสต็อกติดลบ 4 แบบ เป็น config ต่อ business | ต้องมีข้อ 5 ก่อนถึงจะตัดแถวที่จุดข้ามศูนย์ได้ถูก |

## ผลที่ตามมา

**ได้:** ไม่ยิง query เปล่าทุกวินาที · งานซ้ำถูกยุบตั้งแต่ต้นทาง · retry/DLQ/สถานะคิวถามได้จริง · ใช้ของเดิมเกือบทั้งหมด (คิว PG, ตัวคิดต้นทุน, ตัว consumer) · ขยายเป็นหลาย worker ได้ทีหลังโดยไม่ต้องรื้อ

**เสีย/เสี่ยง:** เพิ่มตารางและ connection ที่ต้อง LISTEN ค้างไว้ต่อ holding (ต้องคุมจำนวน) · lease หมดอายุกลางคันแปลว่างานอาจถูกคิดซ้ำ — ตัวคิดต้นทุนต้อง idempotent (ยังไม่ verify ว่าเป็นทุกเส้นทาง) · ขั้น 0 จะทำให้ข้อมูลเก่าที่ไม่เคยเข้า `docdetail` ไหลเข้ามาพร้อมกัน ต้องมีแผน backfill · `stockcalculationstate` (checksum) จะกลายเป็นของซ้ำซ้อนกับ `stockdirty` — ต้องตัดสินใจว่าเก็บไว้เป็นตัวตรวจสอบหรือลบทิ้ง
