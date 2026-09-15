# ADR 2026-09-15 — Stock & Cost Engine v2: ออกแบบใหม่ให้ "ถูกต้อง" ก่อน "เร็ว"

สถานะ: **ลงมือแล้ว** (ขั้น 0-3 เสร็จ, ขั้น 4-5 บางส่วน, ขั้น 6 ยังไม่ทำ — ดู §7) · แทนที่ [2026-09-15-background-process-engine.md](2026-09-15-background-process-engine.md) ซึ่งเสนอแค่ worker บนโครงเดิม
เกี่ยวข้อง: [20-champ-parity-gap.md](../20-champ-parity-gap.md) §4 · สั่งโดยลุงจืด: "ออกแบบใหม่เลย ให้ทำงานถูกต้อง และเร็ว"

## 1. ทำไมต้องออกแบบใหม่ ไม่ใช่แค่เติม worker

ตรวจโค้ดจริงแล้วพบว่าเส้นทางคิดต้นทุนปัจจุบัน **พิสูจน์ความถูกต้องไม่ได้** ไม่ใช่แค่ช้า:

| # | ปัญหา | หลักฐาน |
|---|---|---|
| C1 | DELETE กับ INSERT **คนละ transaction** → ตายกลางทาง = ต้นทุนหายถาวร; ระหว่างคิด รายงานอ่านเจอ 0 แถว | `process-stock-calc-cost.go:135` (autocommit) vs `:292` (COPY ทีละ batch), `mypg/utils.go:506-514` |
| C2 | จับ lock ไม่ได้ → `return` เงียบ ไม่คืน error → **งานหายไปเฉย ๆ** | `process-stock-calc-cost.go:90-92` |
| C3 | path ที่ไม่ส่ง businesscode สั่ง `DELETE FROM processstockcost WHERE itemcode = $1` → **ลบข้ามบริษัท** | `process-stock-calc-cost.go:128` |
| C4 | ลำดับคิดต้นทุน **ไม่ deterministic** — `ORDER BY itemcode, docdatetime, linenumber, calcflag DESC, docno` เอา `linenumber` มาก่อน `docno` ทำให้เอกสารคนละใบสลับบรรทัดกัน และถ้าเสมอกันหมดไม่มี tiebreak | `process-stock-calc-cost.go:144,147` |
| C5 | `calcseq` (ที่ตั้งใจให้เป็น BehindIndex) **hard-code `1` ทุก producer** | `handlers/kafka/*.go` 12 ไฟล์ เช่น `sale_invoice.go:151` |
| C6 | checksum เรียงคนละแบบกับตัวคิด (ไม่มี `calcflag`) → incremental ตัดสินผิดได้ | `incremental.go:61,73` |
| C7 | `processstockcost` **ไม่มี unique key ทางธุรกิจ** (PK = `id SERIAL`) → UPSERT ไม่ได้ ต้อง DELETE-first ตลอด | `process/build/create-database.go:704-729,764-769` |
| C8 | โหลดทั้งประวัติสินค้าเข้า RAM **3 ชุด** ไม่มี LIMIT/cursor (มีแค่ `runtime.GC()` ทุก 50,000 แถว) | `:332, :341, :468-484` |
| C9 | lock 5 นาทีไม่ต่ออายุ → item ใหญ่โดนแย่ง lock กลางคัน = คิดพร้อมกัน 2 ตัว | `mypostgres/lock.go:151` |
| C10 | `docdetail.iscancel` ถูก insert เป็น `false` เสมอ → รู้การยกเลิกได้จาก `doc.iscancel` เท่านั้น | `mypg/insert_doc.go:403` |
| C11 | `doc`/`docdetail` ไม่มี UNIQUE ใด ๆ — กันซ้ำด้วย DELETE-then-COPY ในโค้ดล้วน | `create-database.go:355-364,511-516` |
| P1 | ไม่มี snapshot ต่องวด → แก้เอกสาร 1 ใบ = คิดใหม่ตั้งแต่ต้นชีวิตสินค้า | ไม่พบตาราง opening/carry-forward ใด ๆ ใน PG |
| P2 | checksum ที่ใช้ตัดสินว่า "เปลี่ยนไหม" ต้องอ่านประวัติทั้งหมด — **แพงพอ ๆ กับคิดใหม่** | `incremental.go:38-80` |

**แบบอย่างที่ถูกอยู่แล้วในบ้านเรา:** GL v2 (`backend/internal/generalledger/schema.sql`) — PK ทางธุรกิจจริง (`gl_lines` PK `(company,journal_id,line_no)`), CHECK constraint บังคับที่ฐานข้อมูล (`debit>0 AND credit=0` หรือกลับกัน), `gl_events` append-only + trigger กันแก้, `gl_projection_state.sequence`
→ **ดีไซน์นี้คือการยกมาตรฐาน GL v2 มาใช้กับสต็อก**

## 2. หลักการ 5 ข้อ

1. **PK ต้องเป็นคีย์ทางธุรกิจ** ไม่ใช่ `id SERIAL` — เพื่อให้ MERGE/UPSERT ได้ และข้อมูลซ้ำเป็นไปไม่ได้
2. **ความถูกต้องบังคับที่ฐานข้อมูล** ด้วย CHECK / UNIQUE ไม่ใช่ความเชื่อใจโค้ด
3. **ลำดับการคิดต้องไม่มีทางเสมอ** (total order) — input เดิม → ผลลัพธ์เดิมเสมอ
4. **มีจุดตั้งต้นต่องวด** (checkpoint) — คิดใหม่จากงวดที่กระทบ ไม่ใช่จากศูนย์
5. **เขียนผลแบบ atomic** — ไม่มีวินาทีใดที่ผู้ใช้อ่านเจอข้อมูลครึ่ง ๆ

## 3. โครงสร้างใหม่

> ใช้สิทธิ์ [disposable-database-rule] — ข้อมูล DB/Kafka ทิ้งได้ทุก environment ก่อน go-live จึงออกแบบ schema ใหม่ได้ ไม่ต้องทำ migration/backfill

### 3.1 total order key (แก้ C4, C5)

`docdetail` เพิ่ม `behindindex int NOT NULL DEFAULT 0` (ผู้ใช้แก้ได้ เพื่อจัดลำดับเอกสารวันเดียวกัน — ตรงกับ `BehindIndex` ของต้นแบบ) และ **เลิกใช้ `calcseq` ที่ตายแล้ว**

```sql
ALTER TABLE docdetail ADD COLUMN IF NOT EXISTS behindindex int NOT NULL DEFAULT 0;
CREATE UNIQUE INDEX docdetail_bk ON docdetail (businesscode, docno, linenumber);
CREATE INDEX docdetail_calc ON docdetail (businesscode, itemcode, docdatetime, behindindex, docno, linenumber)
  WHERE transflag IN (54,12,310,48,60,58,66,44,16,56,68,72);
```

ลำดับที่ถูกต้อง (เทียบต้นแบบ `stockprocess_class.cs:10783`):

```sql
ORDER BY docdatetime, behindindex, docno, linenumber
```

`(businesscode, docno, linenumber)` unique → **ไม่มีทางเสมอ** = deterministic 100%

### 3.2 `stock_ledger` แทน `processstockcost` (แก้ C7, C3)

```sql
CREATE TABLE stock_ledger (
    businesscode  text NOT NULL,
    itemcode      text NOT NULL,
    whcode        text NOT NULL,
    docdatetime   timestamptz NOT NULL,
    behindindex   int  NOT NULL,
    docno         text NOT NULL,
    linenumber    int  NOT NULL,
    periodkey     char(7) NOT NULL CHECK (periodkey ~ '^[0-9]{4}-[0-9]{2}$'),
    transflag     int  NOT NULL,
    direction     smallint NOT NULL CHECK (direction IN (-1,1)),
    qty           numeric(18,8) NOT NULL CHECK (qty >= 0),
    unitcost      numeric(18,8) NOT NULL CHECK (unitcost >= 0),
    amount        numeric(18,8) NOT NULL,
    balanceqty    numeric(18,8) NOT NULL,
    balanceamount numeric(18,8) NOT NULL,
    avgcost       numeric(18,8) NOT NULL CHECK (avgcost >= 0),
    PRIMARY KEY (businesscode, itemcode, whcode, docdatetime, behindindex, docno, linenumber)
);
CREATE INDEX stock_ledger_period ON stock_ledger (businesscode, periodkey, itemcode, whcode);
```

**PK = ลำดับการคิดเอง** → ข้อมูลซ้ำเป็นไปไม่ได้, MERGE ได้, และ `businesscode` อยู่ในคีย์ตั้งแต่ต้น (C3 หมดไปโดยโครงสร้าง)

> **บทเรียนตอนลงมือ:** ตอนแรกออกแบบให้ `periodkey` เป็นคอลัมน์ที่ฐานข้อมูลสร้างเองจาก `to_char(docdatetime,'YYYY-MM')`
> PostgreSQL ปฏิเสธด้วย `generation expression is not immutable` เพราะผลของ `to_char` กับ `AT TIME ZONE`
> ขึ้นกับฐานข้อมูลเขตเวลาที่แก้ไขได้ (ลองกับ `Asia/Bangkok` ก็ไม่ผ่านเช่นกัน)
> จึงย้ายมาคำนวณในโปรแกรมที่ `stockengine.PeriodKeyOf` ซึ่งตัดงวดตามเขตเวลาธุรกิจอย่างชัดเจน
> และให้ฐานข้อมูลตรวจแค่รูปแบบ — **ผลพลอยได้คือกฎการตัดงวดอยู่ที่เดียวและทดสอบได้**
> (เอกสารเวลา 1 มีนาคม 00:30 น. ตามเวลาไทย ต้องอยู่งวดมีนาคม ไม่ใช่กุมภาพันธ์ตามเวลาสากล)

### 3.3 checkpoint ต่องวด `stock_period_balance` (แก้ P1)

```sql
CREATE TABLE stock_period_balance (
    businesscode text NOT NULL,
    itemcode     text NOT NULL,
    whcode       text NOT NULL,
    periodkey    char(7) NOT NULL,
    closeqty     numeric(18,8) NOT NULL,
    closeamount  numeric(18,8) NOT NULL,
    closeavgcost numeric(18,8) NOT NULL CHECK (closeavgcost >= 0),
    hastrans     boolean NOT NULL,
    PRIMARY KEY (businesscode, itemcode, whcode, periodkey)
);
```

ตรงกับ `bcstkwarehouseperiod` ของต้นแบบ (`ISTRANSnn/QTYnn/AVERAGECOSTnn/AMOUNTnn` × 61 งวด = ~305 คอลัมน์) แต่เป็น **long table**
`hastrans = false` = งวดที่ไม่มีรายการ ให้ยกยอดงวดก่อนมา (ตรรกะเดียวกับ `stockprocess_class.cs:1960-1966`)

### 3.4 สมุดงานค้าง `stock_dirty` (แทน timer ของต้นแบบ, แทน checksum)

```sql
CREATE TABLE stock_dirty (
    businesscode text NOT NULL,
    itemcode     text NOT NULL,
    fromdate     timestamptz NOT NULL,
    reason       text NOT NULL DEFAULT 'doc',
    lasterror    text NOT NULL DEFAULT '',
    enqueuedat   timestamptz NOT NULL DEFAULT now(),
    attempts     int NOT NULL DEFAULT 0,
    leaseowner   text,
    leaseuntil   timestamptz,
    PRIMARY KEY (businesscode, itemcode)
);
```

ตอน projection เขียน `docdetail` เสร็จ:

```sql
INSERT INTO stock_dirty (businesscode, itemcode, fromdate, reason) VALUES ($1,$2,$3,'doc')
ON CONFLICT (businesscode,itemcode)
DO UPDATE SET fromdate = LEAST(stock_dirty.fromdate, EXCLUDED.fromdate), enqueuedat = now();
SELECT pg_notify('stock_dirty', $1||'|'||$2);
```

`LEAST(...)` = เอกสาร 500 ใบที่แตะสินค้าเดียวกันเหลือ **1 แถว** และคิดใหม่เฉพาะตั้งแต่วันเก่าสุดที่กระทบ
→ **`stockcalculationstate` / `stockcalculationstate_company` เลิกใช้ทั้งคู่** (P2 หมดไป — dirty-set บอกตรง ๆ ว่าอะไรเปลี่ยน ไม่ต้องอ่านประวัติมาทำ checksum)

## 4. อัลกอริทึมคิดใหม่ (แก้ C1, C2, C8, C9)

```
งานหนึ่งชิ้น = (businesscode, itemcode, fromdate)

1. pg_advisory_xact_lock(hashtext('stock:'||businesscode||':'||itemcode))
      ← ปล่อยอัตโนมัติตอนจบ transaction ไม่มีวันหมดอายุกลางคัน ไม่มี leak   [C9]
      ← จับไม่ได้ = คืน error ให้ผู้เรียก งานยังค้างใน stock_dirty ไม่หายเงียบ [C2]
2. prevPeriod = งวดก่อนหน้าของ fromdate
   opening   = SELECT closeqty/closeamount/closeavgcost FROM stock_period_balance
               WHERE periodkey = prevPeriod        ← ไม่มีแถว = เริ่มจากศูนย์   [P1]
3. streaming cursor บน docdetail ตั้งแต่ fromdate (rows.Next ทีละแถว ไม่โหลดเข้า RAM) [C8]
      JOIN doc ON (businesscode, docno) เพื่อกรอง doc.iscancel = false          [C10]
      ORDER BY docdatetime, behindindex, docno, linenumber                      [C4]
4. วนคิด moving weighted average ต่อคลัง + ใช้นโยบายสต็อกติดลบตาม config
5. ทั้งหมดใน transaction เดียว:                                                 [C1]
      DELETE FROM stock_ledger         WHERE businesscode=$1 AND itemcode=$2 AND docdatetime >= $3;
      INSERT INTO stock_ledger ...     (COPY เข้า temp แล้ว INSERT..SELECT)
      MERGE  INTO stock_period_balance ... (อัปเดตทุกงวดตั้งแต่ fromdate)
      DELETE FROM stock_dirty          WHERE businesscode=$1 AND itemcode=$2;
   COMMIT;
```

**ความเร็วมาจากโครงสร้าง ไม่ใช่การจูน:** ขอบเขตงานเปลี่ยนจาก *ทั้งประวัติสินค้า* → *เฉพาะตั้งแต่งวดที่แก้*
(ยังไม่ได้ measure ตัวเลขจริง — ต้องวัดหลัง implement ตามกฎ VERIFY BEFORE DONE)

### นโยบายสต็อกติดลบ (ตามต้นแบบ `stockprocess_class.cs:405`)

`productcostingconfig` ต่อ business: `0` มูลค่าติดลบได้ · `1` มูลค่าติดลบให้เป็น 0 · `3` ปรับต้นทุนและมูลค่าเป็น 0 เมื่อติดลบ · `4` ติดลบให้ใช้ต้นทุนมาตรฐาน/ล่าสุด
พร้อม **แตกบรรทัดตอนยอดข้ามศูนย์** (`:3017,3038`)

## 5. worker

- goroutine ใน goapi (ยังไม่แยก binary), เปิดด้วย env `BCAI_STOCK_WORKER=1`
- ตื่นด้วย `pq.NewListener` + `LISTEN stock_dirty` (`lib/pq v1.10.9` มีอยู่แล้ว `backend/go.mod:17`) — **ไม่ใช่ ticker 1 วินาทีแบบต้นแบบ**
- poll สำรองทุก 15 วินาที = watchdog (ต้นแบบทำทุก 10 วินาที)
- จองงานด้วย lease (`leaseuntil`) + `FOR UPDATE SKIP LOCKED`; ล้มเหลวปล่อยหมดอายุเอง; `attempts >= 5` → `deadletterqueue`
- ขนานได้ N ตัวโดยปลอดภัย เพราะ advisory lock คีย์ = item

**Kafka อยู่ตรงไหน:** คงบทบาทเดิม = ขนส่งเหตุการณ์ข้ามบริการ (mainapi → goapi projection) **ไม่ใช้เป็นคิวงาน** เพราะ Kafka ยุบงานซ้ำไม่ได้และถามสถานะคิวไม่ได้
เมื่อใดต้องกระจาย worker ข้ามเครื่อง ค่อยเพิ่ม topic `bc-stock-recalc-v1` โดย partition key = `holding|business|itemcode` — **ยังไม่ทำตอนนี้**

## 6. ของเดิมที่ต้องซ่อมควบคู่ (ไม่งั้น engine ใหม่ไม่มีข้อมูลป้อน)

`docdetail` เขียนโดย goapi เท่านั้น (`handlers/kafka/utils.go:344`) แต่ `manager.go:10-112` subscribe แค่ 14 จาก 50 handler
→ โอนคลัง, ปรับสต็อกเพิ่ม/ลด, รับ-เบิก-คืนสินค้า, **ยอดยกมา (transflag 54)**, ซื้อรับ/ซื้อคืน/ซื้อบางส่วน, ใบขอซื้อ, RFQ **ไม่เคยไหลเข้า `docdetail` ผ่าน Kafka**

## 7. ลำดับการทำ และสถานะจริง (อัปเดต 2026-09-15 หลังลงมือแล้ว)

| ขั้น | งาน | สถานะ |
|---|---|---|
| 0 | subscribe handler ให้ครบใน `manager.go` | **เสร็จ** — เปิดครบ 17 group (เดิม 6) เขียนเป็นตารางเดียว + เทสต์กันลืม (`manager_test.go`) |
| 1 | `behindindex` + ดัชนีลำดับการคิด + แก้ ORDER BY | **เสร็จ** — `create-database.go` เพิ่มคอลัมน์ + `idx_docdetail_calcorder`; ลำดับใหม่คือ `docdatetime, behindindex, docno, linenumber` |
| 2 | `stock_ledger` + `stock_period_balance` + เขียนใน transaction เดียว | **เสร็จ** — `create-stock-engine.go`, `stockengine/store.go` (`Persist`) |
| 3 | `stock_dirty` + worker + advisory lock + streaming | **เสร็จ** — `stockengine/dirty.go`, `worker.go`, `manager.go`; ใช้ `pg_try_advisory_xact_lock` |
| 4 | นโยบายสต็อกติดลบ 4 แบบ | **เสร็จเฉพาะการตีมูลค่า** (`calc.go` `applyPolicy`) — **ยังไม่ทำการแตกบรรทัดตรงจุดที่ยอดข้ามศูนย์** ตามต้นแบบ |
| 5 | `POST /goapi/api/process/queue-status` + ต่อจอเครื่องมือ | **เสร็จครึ่งเดียว** — API พร้อมแล้ว (`handlers/stock_queue_status.go`) แต่ **จอเครื่องมือยังไม่เรียกใช้** |
| 6 | เลิกใช้ `processstockcost`/`processstocklot`/`distributedlocks`/`stockcalculationstate*` + ย้ายรายงานมาอ่าน `stock_ledger` | **เสร็จ** — ทุกรายงานอ่าน `stock_ledger` กรองด้วย `businesscode`; ตัวคิดต้นทุนเดิม ตารางเดิม ตัวล็อกเดิม สำเนาบน ClickHouse และทางถอย `BCAI_STOCK_ENGINE=v1` ถูกลบทิ้งทั้งหมด (กฎ "ไม่ต้องสนใจข้อมูลเก่า เดินหน้าอย่างเดียว") |
| 7 | แก้ผลการรีวิวโค้ด `docs/handoff/CODE-REVIEW-2026-09-15.md` | **เสร็จ 15 จาก 19 ข้อ** — ดูตารางสถานะในไฟล์รีวิว; ที่เหลือเป็นคำถามทางบัญชีและงาน i18n |

### สิ่งที่ยังไม่ได้ทำ และเหตุผล

1. **`behindindex` ยังไม่มีจอให้ผู้ใช้แก้** — ทุกแถวเป็น 0 การเรียงจึงตกไปที่ `docno` ซึ่งยังคงที่และคิดซ้ำได้ผลเดิม
   แต่ผู้ใช้ยังกำหนดลำดับเอกสารในวันเดียวกันเองไม่ได้แบบต้นแบบ
2. **การแตกบรรทัดตอนยอดข้ามศูนย์** — ต้นแบบแตกรายการเดียวเป็นสองบรรทัดตรงจุดที่ยอดเปลี่ยนจากบวกเป็นลบ
   ที่นี่ยังคิดเป็นบรรทัดเดียว ตัวเลขยอดคงเหลือถูกต้อง แต่รายงานที่ต้องการเห็นจุดข้ามศูนย์จะยังไม่ตรงต้นแบบ
3. **ยังไม่ทดสอบกับข้อมูลจริงขนาดใหญ่** — ตัวเลขความเร็วที่ดีขึ้นเป็นผลเชิงโครงสร้าง (ขอบเขตงานเล็กลง) ยังไม่ได้วัดจริง
4. **จอเครื่องมือยังสั่งคิดต้นทุนแบบรอผล** — `processstockcalccost` ผ่าน proxy ที่มี timeout 20 วินาที
   ควรเปลี่ยนเป็นฝากงานเข้าคิวแล้วดูความคืบหน้าจาก `queue-status` (ต่อกับขั้น 5 ที่ยังค้าง)

### หลักฐานการทดสอบ

- 20 เทสต์ตรรกะการคำนวณ (`calc_test.go`) — ครอบคลุมถัวเฉลี่ยถ่วงน้ำหนัก ทิศทางเอกสารทั้ง 12 ประเภท
  การเรียงที่ต้องได้ผลเดิมแม้สลับลำดับข้อมูลเข้า การเริ่มจากยอดยกมา นโยบายติดลบทั้ง 4 แบบ การตัดงวดตามเขตเวลาไทย
  ราคาต่อหน่วยที่ต้องคิดตามหน่วยนับในเอกสาร และการโอนคลังสองขาที่ต้องไม่ทำให้ยอดรวมเปลี่ยน
- 19 เทสต์ที่รันจริงกับ PostgreSQL 18 (`store_integration_test.go`, build tag `integration`) — คิดซ้ำได้ผลเดิม
  ไม่ลบข้ามบริษัท เอกสารยกเลิกไม่ถูกคิด คิดเฉพาะช่วงท้ายได้ผลเท่าคิดใหม่ทั้งหมด งานซ้ำถูกยุบ การจองงานไม่ซ้อนกัน
  งานล้มเหลวกลับเข้าคิวหลังพักตามจำนวนครั้ง คิดพร้อมกันถูกกันด้วยล็อก สินค้าคนละตัวไม่บล็อกกัน
  ฐานข้อมูลปฏิเสธค่าที่เป็นไปไม่ได้ worker ถูกปลุกด้วยสัญญาณจริง เอกสารที่เข้ามาระหว่างคำนวณไม่หลุดคิว
  เอกสารที่ถูกลบหลุดออกจากสมุดสต็อก ยอดยกมามาจากงวดล่าสุดที่มีจริง และการโอนคลังย้ายของโดยยอดรวมไม่เปลี่ยน
- 4 เทสต์กันการลืมเปิดตัวรับข้อมูล + กันรายชื่อ transflag สองชุดไม่ตรงกัน (`manager_test.go`)
- 2 เทสต์ว่าใบโอนคลังถูกแตกเป็นสองขาและข้อมูลคลังปลายทางไม่ตกหล่น (`stock_transfer_test.go`)

วิธีรันเทสต์ที่ต้องใช้ฐานข้อมูลจริง:

```
docker run -d --name bc-stockengine-test -e POSTGRES_HOST_AUTH_METHOD=trust -p 55433:5432 postgres:18-alpine
BC_STOCK_TEST_POSTGRES_DSN="postgres://postgres@localhost:55433/postgres?sslmode=disable" \
  go test -tags integration ./internal/goapi/process/stockengine/
```

## 8. ผลที่ตามมา

**ได้:** ต้นทุนพิสูจน์ได้ว่า deterministic · ไม่มีหน้าต่างที่รายงานอ่านเจอ 0 แถว · ลบข้ามบริษัทเป็นไปไม่ได้เชิงโครงสร้าง · แก้ย้อนหลังคิดเฉพาะงวดที่กระทบ · ไม่มี OOM จาก item ใหญ่ · เลิกพึ่ง checksum ที่แพง · สอดคล้องกับ GL v2 ที่ทำถูกอยู่แล้ว

**เสีย/เสี่ยง:** เป็นการรื้อ schema สต็อกทั้งชุด (ทำได้เพราะ DB disposable ก่อน go-live — หลัง go-live จะทำแบบนี้ไม่ได้) · ต้องแก้ทุกรายงานที่อ่าน `processstockcost` (`process-stock-balance-*.go` 3 ไฟล์ + ClickHouse replace path) · `behindindex` ต้องมีจอให้ผู้ใช้แก้ลำดับ ไม่งั้นได้ค่า 0 เท่ากันหมดแล้วกลับไปพึ่ง docno เหมือนเดิม · advisory lock กันได้เฉพาะภายใน PostgreSQL instance เดียวกัน (พอสำหรับสถาปัตยกรรม per-holding DB ปัจจุบัน) · ขั้น 0 จะทำให้ข้อมูลไหลเข้าเพิ่มหลายเท่า ต้องทดสอบบน dev ก่อน

**ยังไม่ verify:** ตัวเลขความเร็วจริง (ต้องวัดหลัง implement) · จำนวนจุดในโค้ด/รายงานที่อ้าง `processstockcost` ทั้งหมด · ผลกระทบต่อ ClickHouse path (`clickhouse-replace-process-stock-cost.go`)
