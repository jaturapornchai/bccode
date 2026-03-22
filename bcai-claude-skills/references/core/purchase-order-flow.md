# การจัดซื้อสินค้า — ใบสั่งซื้อ (Purchase Order)

## ใบสั่งซื้อคืออะไร?

ใบสั่งซื้อ (PO) คือเอกสารที่บอกว่า **"เราจะสั่งซื้อสินค้าจากผู้ขายรายนี้ รายการนี้ จำนวนเท่านี้ ราคาเท่านี้"**

**PO ยังไม่มีผลกระทบอะไรกับสต็อกหรือเจ้าหนี้** — เป็นแค่การ "ตกลงจะซื้อ" เท่านั้น ของจะเข้าคลังก็ต่อเมื่อทำใบรับสินค้าแล้ว

---

## Workflow หลัก — จากสั่งซื้อจนถึงจ่ายเงิน

```
 1. สร้างใบสั่งซื้อ (PO)
    "เราจะซื้ออะไร จากใคร เท่าไหร่"
         │
         ▼
 2. ส่งอนุมัติ
    "หัวหน้า/ผู้มีอำนาจ ช่วยอนุมัติที"
         │
    ┌────┴────┐
    ▼         ▼
 อนุมัติ    ปฏิเสธ
    │         │
    │         ▼
    │    แก้ไข → ส่งใหม่
    │
    ▼
 3. รับสินค้า (Purchase Receive)
    "ของมาแล้ว ตรวจนับเรียบร้อย"
    → สต็อกเพิ่มขึ้น ✅
    → ต้นทุนเฉลี่ยคำนวณใหม่
         │
         ▼
 4. ตั้งหนี้ (AP Accrual)
    "บันทึกว่าเราเป็นหนี้ผู้ขาย"
    → ยอดเจ้าหนี้เพิ่มขึ้น ✅
         │
         ▼
 5. จ่ายเงิน (Creditor Payment)
    "จ่ายเงินตามเครดิตเทอม"
    → ยอดเจ้าหนี้ลดลง ✅
```

### สรุปผลกระทบแต่ละขั้นตอน

| ขั้นตอน | สต็อก | เจ้าหนี้ | เงิน |
|---------|-------|---------|------|
| 1. สร้าง PO | ไม่เปลี่ยน | ไม่เปลี่ยน | ไม่เปลี่ยน |
| 2. อนุมัติ PO | ไม่เปลี่ยน | ไม่เปลี่ยน | ไม่เปลี่ยน |
| 3. รับสินค้า | **เพิ่ม** | ไม่เปลี่ยน | ไม่เปลี่ยน |
| 4. ตั้งหนี้ | ไม่เปลี่ยน | **เพิ่ม** | ไม่เปลี่ยน |
| 5. จ่ายเงิน | ไม่เปลี่ยน | **ลด** | **ลด** |

---

## เอกสารที่เกี่ยวข้องทั้งหมด

```
ใบสั่งซื้อ (PO)           TransFlag=6    ← แค่สั่งซื้อ ยังไม่มีผลอะไร
     │
     ├──► ใบรับสินค้า (PP)   TransFlag=310  ← ของเข้าคลัง สต็อกเพิ่ม
     │         │
     │         └──► ตั้งหนี้ (AP)  TransFlag=12  ← บันทึกหนี้ที่ต้องจ่าย
     │
     └──► ใบซื้อทันที (PU)   TransFlag=12  ← ซื้อ+จ่ายเลย (ไม่ต้องรอรับ)

ใบส่งคืน (PT)             TransFlag=16  ← ส่งของคืนผู้ขาย สต็อกลด
```

---

## สถานะใบสั่งซื้อ — เข้าใจง่ายๆ

### สถานะการอนุมัติ

| สถานะ | หน้าตา | ความหมาย |
|--------|--------|----------|
| ยังไม่ส่ง | (ไม่มี badge) | สร้างแล้ว ยังไม่ได้ส่งให้ใครอนุมัติ |
| อนุมัติอัตโนมัติ | `auto_approved` | ไม่มีกฎอนุมัติ หรือคนสร้างมีสิทธิ์พอ ระบบอนุมัติให้เลย |
| รออนุมัติ | `pending` (เหลือง) | ส่งแล้ว รอหัวหน้ากดอนุมัติ |
| อนุมัติแล้ว | `approved` (เขียว) | ผ่านแล้ว สร้างใบรับสินค้าได้ |
| ปฏิเสธ | `rejected` (แดง) | โดนตีกลับ แก้แล้วส่งใหม่ได้ |

### สถานะการรับสินค้า

| สถานะ | Badge | ความหมาย |
|--------|-------|----------|
| ยังไม่รับ | (ไม่มี) | ยังไม่มีใบรับสินค้า |
| รับครบแล้ว | เขียว "รับครบแล้ว" | รับของครบตาม PO |
| รับบางส่วน | เหลือง "รับบางส่วน" | รับแค่บางรายการ/บางจำนวน |
| รับเกิน | เหลือง "รับเกิน" | รับเกินจำนวนที่สั่ง |
| ปิดด้วยมือ | ม่วง "ปิดด้วยมือ" | ไม่รับของที่เหลือแล้ว (มีเหตุผล) |

### ทำอะไรได้ในแต่ละสถานะ?

| สถานะ | แก้ไข | ลบ | พิมพ์ | ส่งอนุมัติ |
|--------|------|------|------|-----------|
| ยังไม่ส่ง | ✅ | ✅ | ❌ | ✅ |
| auto (ยังไม่อ้างอิง) | ✅ | ✅ | ✅ | — |
| auto (อ้างอิงแล้ว) | ❌ | ❌ | ✅ | — |
| รออนุมัติ | ❌ | ❌ | ❌ | ❌ ถอนก่อน |
| อนุมัติแล้ว | ❌ | ❌ | ✅ | — |
| ปฏิเสธ | ✅ | ❌ | ❌ | ✅ ส่งใหม่ |

**กฎเพิ่มเติม:** ถ้ามีคนอนุมัติบางขั้นไปแล้ว (`currentApprovedLevel > 0`) → ห้ามแก้ไขทุกกรณี

---

## ข้อมูลในใบสั่งซื้อ

### หัวเอกสาร

| ข้อมูล | ตัวอย่าง | หมายเหตุ |
|--------|---------|----------|
| เลขที่เอกสาร | PO2026030800001 | สร้างอัตโนมัติ PO + วันที่ + running |
| วันที่ | 2026-03-08 | วันที่สร้างเอกสาร |
| ผู้จำหน่าย | บ.เอบีซี จำกัด | เลือกจากรายชื่อ Creditor |
| ประเภทจัดซื้อ | จัดซื้อทั่วไป | ใช้กำหนด approval rule |
| ภาษีมูลค่าเพิ่ม | รวม VAT 7% | 0=ไม่มี, 1=รวม, 2=แยก, 3=ยกเว้น |
| เครดิตเทอม | 30 วัน | กำหนดวันครบชำระ |
| สกุลเงิน | THB / USD | รองรับหลายสกุลเงิน + อัตราแลกเปลี่ยน |
| ภาษีหัก ณ ที่จ่าย | 3% | อัตราที่รองรับ: 0.5, 1, 2, 3, 5, 10, 15% |

### รายการสินค้า

| ข้อมูล | ตัวอย่าง | หมายเหตุ |
|--------|---------|----------|
| รหัสสินค้า | P001 | ค้นจากบาร์โค้ดหรือรหัส |
| ชื่อสินค้า | กระดาษ A4 | หลายภาษา |
| จำนวน | 100 | ต้อง > 0 |
| หน่วย | รีม | รองรับ multi-unit |
| ราคาต่อหน่วย | 120.00 | ต้อง >= 0 |
| ส่วนลด | 10%+5% | ลดซ้อนได้ |
| ยอดรวม | 10,260.00 | คำนวณอัตโนมัติ |
| คลังสินค้า | WH01 | เลือกคลังที่จะรับเข้า |

---

## Validation — ระบบตรวจอะไรบ้าง?

### ก่อนบันทึก (Frontend)
- ต้องเลือกผู้จำหน่าย
- ต้องมีสินค้าอย่างน้อย 1 รายการ
- จำนวนต้อง > 0, ราคาต้อง >= 0
- ยอดรวมต้องมากกว่า 0

### หลังกดบันทึก (Backend — 3 ชั้น)
1. **ตรวจ format** — field ที่ required ต้องไม่ว่าง
2. **ตรวจ business logic** — วันที่, ผู้จำหน่าย, VAT type, ตัวเลขห้าม NaN/Infinity
3. **แก้ไขอัตโนมัติ** — exchange rate ≤ 0 → fix เป็น 1.0

### ตรวจสิทธิ์
- `pending` หรือ `approved` → ห้ามแก้ไข (ต้อง reject/withdraw ก่อน)
- มีคนอนุมัติบางขั้นแล้ว → ห้ามแก้ไขเด็ดขาด

---

## ระบบเบื้องหลัง (สำหรับ Developer)

### เลขที่เอกสาร (DocNo)
```
PO + YYYYMMDD + 00001 (running 5 หลัก)
ตัวอย่าง: PO2026030800001
```
- Running number เก็บใน Redis cache
- Fallback: query MongoDB หาเลขล่าสุด

### Data Flow — ข้อมูลไปอยู่ที่ไหนบ้าง?

ระบบใช้ 3 ฐานข้อมูล แต่ละตัวมีหน้าที่ต่างกัน:

```
 ผู้ใช้กดบันทึก PO
      │
      ▼
 ┌─────────────────────────────────────────────────────┐
 │  MongoDB (Primary — ข้อมูลหลัก)                      │
 │  Collection: transactionPurchaseOrder                │
 │  → เก็บ PO ทั้งใบ (header + details + approval)      │
 │  → ใช้สำหรับ CRUD จาก API โดยตรง                     │
 │  → Frontend อ่าน/เขียนผ่าน API ที่นี่                  │
 └────────────────────┬────────────────────────────────┘
                      │ Kafka Message
                      │ (when-purchaseorder-created/updated/deleted)
                      ▼
 ┌─────────────────────────────────────────────────────┐
 │  PostgreSQL (Secondary — สำหรับ query/report)        │
 │  Table: purchase_order_transaction                   │
 │  Table: purchase_order_transaction_detail             │
 │  → Consumer แปลง MongoDB doc → PG row               │
 │  → ใช้สำหรับ JOIN, aggregate, report                 │
 │  → GoAPI ดึงรายการ PO จากที่นี่ (getdoc)              │
 └────────────────────┬────────────────────────────────┘
                      │ Async sync
                      ▼
 ┌─────────────────────────────────────────────────────┐
 │  ClickHouse (Analytics — สำหรับ dashboard/BI)        │
 │  → เก็บข้อมูลสรุป (aggregated)                       │
 │  → ใช้สำหรับ report ยอดสั่งซื้อ, วิเคราะห์ supplier  │
 │  → MCP tools ดึงข้อมูลจากที่นี่ (dashboard KPI)       │
 └─────────────────────────────────────────────────────┘
```

### สรุป: ใครใช้ DB ไหน?

| ทำอะไร | ใช้ DB ไหน | ทำไม |
|--------|-----------|------|
| สร้าง/แก้/ลบ PO | **MongoDB** | Primary storage, flexible schema |
| ดึงรายการ PO (getdoc) | **PostgreSQL** | เร็วกว่า, รองรับ filter/sort/pagination |
| ดู PO ตาม GUID | **MongoDB** | อ่านจาก primary โดยตรง |
| Report ยอดสั่งซื้อ | **ClickHouse** | เร็วมากสำหรับ aggregate |
| MCP tools (dashboard) | **ClickHouse** | Optimized สำหรับ analytics |
| Approval status | **MongoDB** (GoAPI) | เก็บแยก collection |

### Kafka Topics

| Topic | เมื่อไหร่ |
|-------|---------|
| `when-purchaseorder-created` | สร้าง PO ใหม่ |
| `when-purchaseorder-updated` | แก้ไข PO |
| `when-purchaseorder-deleted` | ลบ PO |
| `when-purchaseorder-bulk-created` | สร้างหลายใบ |
| `when-purchaseorder-bulk-updated` | แก้ไขหลายใบ |
| `when-purchaseorder-bulk-deleted` | ลบหลายใบ |

### Flow เมื่อ PO ถูกอ้างอิง (รับสินค้า)

```
 PO อนุมัติแล้ว
      │
      ▼
 สร้างใบรับสินค้า (Purchase Receive) อ้างอิง PO
      │
      ├──► MongoDB: transactionPurchasePartial (เก็บใบรับ)
      │
      ├──► Kafka: when-purchasereceive-created
      │         │
      │         ▼
      │    PostgreSQL: purchase_receive_transaction (sync)
      │         │
      │         ▼
      │    PostgreSQL: docdetail (สำหรับคำนวณต้นทุน)
      │         │
      │         ▼
      │    PostgreSQL: stockwaitprocess (คิวรอคำนวณ)
      │         │
      │         ▼
      │    PostgreSQL: processstockcost (ผลคำนวณต้นทุน)
      │         │
      │         ▼
      │    PostgreSQL: productbarcode (อัพเดท stock + averagecost)
      │         │
      │         ▼
      │    ClickHouse: stock data (async sync สำหรับ report)
      │
      ├──► MongoDB: transactionPurchasePartial → PO.isref = true
      │    (อัพเดท PO ว่าถูกอ้างอิงแล้ว)
      │
      └──► สร้าง AP Accrual Receive (ตั้งหนี้)
                │
                ├──► MongoDB: transactionAccrualReceive
                ├──► Kafka → PostgreSQL: ap_purchasereceive_transaction
                └──► PostgreSQL: creditor balance เพิ่ม
```

---

## Database Schema — ทุก Field ครบ (สำหรับทำรายงาน/BI)

### MongoDB Collection: `transactionPurchaseOrder`

เก็บ PO ทั้งใบ (header + details) เป็น JSON document เดียว — ใช้ model `Transaction` จาก Go backend

### PostgreSQL: `purchase_order_transaction` (Header)

Embed จาก `TransactionPG` + field เพิ่มเติม

| Column | Type | ความหมาย |
|--------|------|-----------|
| `shopid` | TEXT | รหัสร้านค้า |
| `parid` | TEXT | Partition ID |
| `guidfixed` | TEXT | GUID คงที่ (PK) |
| `transflag` | INT2 | ประเภทเอกสาร (6=PO) |
| `docno` | TEXT | เลขที่เอกสาร |
| `docdate` | TIMESTAMPTZ | วันที่เอกสาร |
| `docreftype` | INT2 | ประเภทเอกสารอ้างอิง |
| `docrefno` | TEXT | เลขที่เอกสารอ้างอิง |
| `docrefdate` | TIMESTAMPTZ | วันที่เอกสารอ้างอิง |
| `branchcode` | TEXT | รหัสสาขา |
| `branchnames` | JSONB | ชื่อสาขา (หลายภาษา) |
| `description` | TEXT | รายละเอียด/หมายเหตุ |
| `taxdocno` | TEXT | เลขที่เอกสารภาษี |
| `taxdocdate` | TIMESTAMPTZ | วันที่เอกสารภาษี |
| `creditorcode` | TEXT | **รหัสเจ้าหนี้/ผู้จำหน่าย** |
| `creditornames` | JSONB | **ชื่อเจ้าหนี้ (หลายภาษา)** |
| `iscancel` | BOOLEAN | ยกเลิกแล้ว |
| `isbom` | BOOLEAN | เป็น BOM |
| `status` | INT2 | สถานะเอกสาร |
| `vattype` | INT2 | ประเภทภาษี (0/1/2/3) |
| `vatrate` | FLOAT8 | อัตราภาษี % |
| `totalvalue` | FLOAT8 | มูลค่ารวม |
| `discountword` | TEXT | ส่วนลด (ข้อความ เช่น "10%+5%") |
| `totaldiscount` | FLOAT8 | ส่วนลดรวม |
| `deliveryamount` | FLOAT8 | ค่าจัดส่ง |
| `totalbeforevat` | FLOAT8 | รวมก่อนภาษี |
| `totalvatvalue` | FLOAT8 | มูลค่าภาษี |
| `totalexceptvat` | FLOAT8 | รวมยกเว้นภาษี |
| `totalaftervat` | FLOAT8 | รวมหลังภาษี |
| `totalamount` | FLOAT8 | **ยอดรวมสุทธิ** |
| `guidref` | TEXT | GUID อ้างอิง |
| `guidpos` | TEXT | GUID ของ POS |
| `devicename` | TEXT | ชื่ออุปกรณ์ |
| `inquirytype` | INT | ประเภทการสอบถาม |
| `ismanualamount` | BOOLEAN | ระบุยอดเงินเอง |
| `pointdiscountamount` | FLOAT8 | ส่วนลดจากแต้ม |
| `paypointamount` | FLOAT8 | จำนวนแต้มที่จ่าย |
| `alcoholamount` | FLOAT8 | มูลค่าแอลกอฮอล์ |
| `otheramount` | FLOAT8 | มูลค่าอื่นๆ |
| `drinkamount` | FLOAT8 | มูลค่าเครื่องดื่ม |
| `foodamount` | FLOAT8 | มูลค่าอาหาร |

### PostgreSQL: `purchase_order_transaction_detail` (Detail)

Embed จาก `TransactionDetailPG`

| Column | Type | ความหมาย |
|--------|------|-----------|
| `id` | SERIAL | PK (auto) |
| `shopid` | TEXT | รหัสร้านค้า |
| `guidfixed` | TEXT | GUID (FK → header) |
| `parid` | TEXT | Partition ID |
| `docno` | TEXT | เลขที่เอกสาร |
| `docdate` | TIMESTAMPTZ | วันที่เอกสาร |
| `linenumber` | INT2 | ลำดับรายการ |
| `barcode` | TEXT | บาร์โค้ด |
| `itemcode` | TEXT | รหัสสินค้า (via itemguid) |
| `itemguid` | TEXT | GUID สินค้า |
| `itemnames` | JSONB | ชื่อสินค้า (หลายภาษา) |
| `itemtype` | INT2 | ประเภทสินค้า |
| `unitcode` | TEXT | รหัสหน่วยนับ |
| `unitnames` | JSONB | ชื่อหน่วย (หลายภาษา) |
| `qty` | FLOAT8 | จำนวน |
| `price` | FLOAT8 | ราคาต่อหน่วย |
| `priceexcludevat` | FLOAT8 | ราคาไม่รวม VAT |
| `discount` | TEXT | ส่วนลด (ข้อความ) |
| `discountamount` | FLOAT8 | มูลค่าส่วนลด |
| `sumamount` | FLOAT8 | **ยอดรวมรายการ** |
| `sumamountexcludevat` | FLOAT8 | ยอดรวมไม่รวม VAT |
| `sumamountchoice` | FLOAT8 | ยอดรวมตัวเลือก |
| `totalvaluevat` | FLOAT8 | มูลค่า VAT รายการ |
| `vattype` | INT2 | ประเภทภาษีรายการ |
| `taxtype` | INT2 | ประเภทภาษีอื่น |
| `vatcal` | INT2 | วิธีคำนวณภาษี |
| `whcode` | TEXT | รหัสคลังสินค้า |
| `whnames` | JSONB | ชื่อคลัง (หลายภาษา) |
| `locationcode` | TEXT | รหัสที่เก็บ |
| `locationnames` | JSONB | ชื่อที่เก็บ (หลายภาษา) |
| `standvalue` | FLOAT8 | ค่ามาตรฐานหน่วย (สำหรับ multi-unit) |
| `dividevalue` | FLOAT8 | ค่าหารหน่วย |
| `groupcode` | TEXT | รหัสกลุ่มสินค้า |
| `groupnames` | JSONB | ชื่อกลุ่ม (หลายภาษา) |
| `refguid` | TEXT | GUID อ้างอิง |
| `docref` | TEXT | เลขที่เอกสารอ้างอิง |
| `docrefdatetime` | TIMESTAMPTZ | วันที่เอกสารอ้างอิง |
| `remark` | TEXT | หมายเหตุรายการ |
| `ischoice` | INT2 | เป็น choice item |
| `foodtype` | INT2 | ประเภทอาหาร |
| `whcodedestination` | TEXT | รหัสคลังปลายทาง |
| `whcodedestinationnames` | JSONB | ชื่อคลังปลายทาง |
| `locationcodedestination` | TEXT | รหัสที่เก็บปลายทาง |
| `locationdestination` | JSONB | ชื่อที่เก็บปลายทาง |

### PostgreSQL: ตาราง Purchase Receive + AP (โครงสร้างเหมือนกัน)

| ตาราง | หน้าที่ | Field เพิ่ม |
|-------|--------|-----------|
| `purchasereceive_transaction` | Header รับสินค้า | `creditorcode`, `creditornames` |
| `purchasereceive_transaction_detail` | Detail รับสินค้า | (เหมือน PO detail) |
| `ap_purchasereceive_transaction` | Header ตั้งหนี้ | `creditorcode`, `creditornames` |
| `ap_purchasereceive_transaction_detail` | Detail ตั้งหนี้ | (เหมือน PO detail) |

### PostgreSQL (GoAPI): `docdetail` — Input คำนวณสต็อก/ต้นทุน

| Column | Type | ความหมาย |
|--------|------|-----------|
| `id` | SERIAL | PK |
| `docdatetime` | TIMESTAMPTZ | วันเวลาเอกสาร (ใช้เรียงลำดับ) |
| `docno` | TEXT | เลขที่เอกสาร |
| `docref` | TEXT | เลขที่อ้างอิง |
| `description` | TEXT | รายละเอียด |
| `linenumber` | INT | ลำดับรายการ |
| `transflag` | INT | ประเภทเอกสาร |
| `calcflag` | INT | 1=เพิ่มสต็อก, 2=ลดสต็อก |
| `calcseq` | INT | ลำดับการคำนวณ |
| `iscancel` | BOOLEAN | ยกเลิก (default false) |
| `itemcode` | TEXT | รหัสสินค้าหลัก |
| `barcodemain` | TEXT | บาร์โค้ดหลัก |
| `barcode` | TEXT | บาร์โค้ดย่อย |
| `unitcode` | TEXT | หน่วยนับ |
| `whcode` | TEXT | คลังสินค้า |
| `locationcode` | TEXT | ที่เก็บ |
| `totalqty` | NUMERIC(18,8) | จำนวน |
| `unitstand` | NUMERIC(18,8) | ค่ามาตรฐานหน่วย |
| `unitdivide` | NUMERIC(18,8) | ค่าหารหน่วย |
| `price` | NUMERIC(18,2) | ราคา |
| `priceexcludevat` | NUMERIC(18,2) | ราคาไม่รวม VAT |
| `sumamount` | NUMERIC(18,2) | **มูลค่ารวม (ใช้คำนวณต้นทุน)** |
| `isupdated` | BOOLEAN | อัพเดทแล้ว (default false) |
| `iscalcstock` | INT | คำนวณสต็อกแล้ว (default 0) |
| `price_doc` | NUMERIC(18,2) | ราคา (สกุลเงินเอกสาร) |
| `sumamount_doc` | NUMERIC(18,2) | มูลค่า (สกุลเงินเอกสาร) |
| `discountamount_doc` | NUMERIC(18,2) | ส่วนลด (สกุลเงินเอกสาร) |
| `priceexcludevat_doc` | NUMERIC(18,2) | ราคาไม่รวม VAT (สกุลเงินเอกสาร) |
| `sumamountexcludevat_doc` | NUMERIC(18,2) | มูลค่าไม่รวม VAT (สกุลเงินเอกสาร) |
| `totalvaluevat_doc` | NUMERIC(18,2) | มูลค่า VAT (สกุลเงินเอกสาร) |

### PostgreSQL (GoAPI): `processstockcost` — ผลคำนวณต้นทุน

| Column | Type | ความหมาย |
|--------|------|-----------|
| `id` | SERIAL | PK |
| `docdatetime` | TIMESTAMPTZ | วันเวลาเอกสาร |
| `docno` | TEXT | เลขที่เอกสาร |
| `docref` | TEXT | เลขที่อ้างอิง |
| `linenumber` | INT | ลำดับรายการ |
| `transflag` | INT | ประเภทเอกสาร |
| `itemcode` | TEXT | รหัสสินค้า |
| `barcode` | TEXT | บาร์โค้ด |
| `unitcode` | TEXT | หน่วยนับมาตรฐาน |
| `whcode` | TEXT | คลังสินค้า |
| `locationcode` | TEXT | ที่เก็บ |
| `totalqty` | NUMERIC(18,8) | จำนวนที่เปลี่ยน |
| `unitstand` | NUMERIC(18,8) | ค่ามาตรฐานหน่วย |
| `unitdivide` | NUMERIC(18,8) | ค่าหารหน่วย |
| `price` | NUMERIC(18,2) | ราคา |
| `averagecost` | NUMERIC(18,2) | **ต้นทุนเฉลี่ย ณ เวลานั้น** |
| `calcamount` | NUMERIC(18,2) | **มูลค่าที่คำนวณ** |
| `balanceqty` | NUMERIC(18,8) | **ยอดคงเหลือ** |
| `balanceamount` | NUMERIC(18,2) | **มูลค่าคงเหลือ** |
| `unitcost` | NUMERIC(18,2) | ต้นทุนต่อหน่วย |
| `guid` | TEXT | GUID อ้างอิง |

### ClickHouse: `docdetail` — Analytics รายการสินค้า

Engine: `MergeTree` PARTITION BY `shopid` ORDER BY `(shopid, docno, line_number)`

| Column | Type | ความหมาย |
|--------|------|-----------|
| `shopid` | String | ร้านค้า |
| `docno` | String | เลขที่เอกสาร |
| `docdatetime` | DateTime | วันเวลา |
| `perioddatetime` | DateTime | วันที่งวด |
| `line_number` | UInt32 | ลำดับรายการ |
| `transflag` | Int16 | ประเภทเอกสาร |
| `calcflag` | Int8 | flag คำนวณ |
| `calcseq` | Int32 | ลำดับคำนวณ |
| `guidfixed` | String | GUID เอกสาร |
| `guidpos` | String | GUID POS |
| `guidbranch` | String | GUID สาขา |
| `branchid` | String | รหัสสาขา |
| `itemcode` | Nullable(String) | รหัสสินค้า |
| `barcode` | String | บาร์โค้ด |
| `barcodemain` | String | บาร์โค้ดหลัก |
| `itemname` | String | ชื่อสินค้า |
| `itemnames` | String | ชื่อสินค้า (JSON) |
| `unitcode` | String | หน่วยนับ |
| `unitstand` | Float64 | ค่ามาตรฐาน (default 1.0) |
| `unitdivide` | Float64 | ค่าหาร (default 1.0) |
| `qty` | Float64 | จำนวน |
| `price` | Float64 | ราคา |
| `discount` | String | ส่วนลด |
| `discountamount` | Float64 | มูลค่าส่วนลด |
| `sumamount` | Float64 | ยอดรวม |
| `sumofcost` | Float64 | **ต้นทุนรวม** |
| `whcode` | String | คลัง |
| `locationcode` | String | ที่เก็บ |
| `refguid` | String | GUID อ้างอิง |
| `ischoice` | Int32 | choice item |
| `sumamountchoice` | Float64 | ยอด choice |
| `isupdated` | Bool | อัพเดทแล้ว |
| `iscalcstock` | Int8 | คำนวณสต็อกแล้ว |
| `price_doc` | Float64 | ราคา (สกุลเงินเอกสาร) |
| `sumamount_doc` | Float64 | ยอดรวม (สกุลเงินเอกสาร) |
| `discountamount_doc` | Float64 | ส่วนลด (สกุลเงินเอกสาร) |
| `priceexcludevat_doc` | Float64 | ราคาไม่รวม VAT (สกุลเงินเอกสาร) |
| `sumamountexcludevat_doc` | Float64 | ยอดรวมไม่รวม VAT (สกุลเงินเอกสาร) |
| `totalvaluevat_doc` | Float64 | มูลค่า VAT (สกุลเงินเอกสาร) |

### ClickHouse: `processstockcost` — Analytics ต้นทุน

Engine: `MergeTree` PARTITION BY `(shopid, itemcode)` ORDER BY `(shopid, itemcode, docdatetime, whcode, locationcode)`
CODEC: ZSTD(1) ทุก column — bloom_filter บน `docno`+`barcode`, minmax บน `docdatetime`

| Column | Type | ความหมาย |
|--------|------|-----------|
| `shopid` | LowCardinality(String) | ร้านค้า |
| `itemcode` | LowCardinality(String) | รหัสสินค้า |
| `docdatetime` | DateTime | วันเวลา |
| `docno` | String | เลขที่เอกสาร |
| `docref` | String | เลขที่อ้างอิง |
| `linenumber` | UInt16 | ลำดับรายการ |
| `transflag` | UInt8 | ประเภทเอกสาร |
| `barcodemain` | String | บาร์โค้ดหลัก |
| `barcode` | String | บาร์โค้ด |
| `unitcode` | LowCardinality(String) | หน่วยนับ |
| `whcode` | LowCardinality(String) | คลัง |
| `locationcode` | LowCardinality(String) | ที่เก็บ |
| `totalqty` | Float64 | จำนวน |
| `unitstand` | Float64 | ค่ามาตรฐาน |
| `unitdivide` | Float64 | ค่าหาร |
| `price` | Float64 | ราคา |
| `averagecost` | Float64 | **ต้นทุนเฉลี่ย** |
| `calcamount` | Float64 | **มูลค่าคำนวณ** |
| `balanceqty` | Float64 | **ยอดคงเหลือ** |
| `balanceamount` | Float64 | **มูลค่าคงเหลือ** |
| `unitcost` | Float64 | ต้นทุนต่อหน่วย |
| `guid` | String | GUID |
| `originalqty` | Int32 | จำนวนเดิม |
| `qty` | Int32 | จำนวน |

---

### API Endpoints

| ทำอะไร | Method | Path |
|--------|--------|------|
| สร้าง PO | POST | `/transaction/purchase-order` |
| แก้ไข PO | PUT | `/transaction/purchase-order/:id` |
| ลบ PO | DELETE | `/transaction/purchase-order/:id` |
| ดึงรายการ | POST | `/goapi/getdoc` (system=purchase-order) |
| ส่งอนุมัติ | POST | `/goapi/api/approval/po-status/submit` |
| ถอนอนุมัติ | POST | `/goapi/api/approval/po-status/withdraw` |
| ปิดด้วยมือ | POST | `/goapi/api/purchase-order/manual-close` |

### ไฟล์ Code สำคัญ

**Backend (Go):**
| ไฟล์ | ทำอะไร |
|------|--------|
| `internal/transaction/purchaseorder/purchaseorder_http.go` | API routes ทั้งหมด |
| `internal/transaction/purchaseorder/services/purchaseorder_http_service.go` | Business logic + สร้างเลขเอกสาร |
| `internal/transaction/purchaseorder/validators/po_validator.go` | Validation 3 ชั้น |
| `internal/transaction/models/transaction.go` | โครงสร้างข้อมูลหลัก |
| `internal/transaction/purchasepartial/` | ใบรับสินค้า |
| `internal/transaction/accrualreceive/` | ตั้งหนี้เจ้าหนี้ |

**Frontend (Flutter):**
| ไฟล์ | ทำอะไร |
|------|--------|
| `lib/screens/purchaseorder/purchaseorder_edit_screen.dart` | หน้าสร้าง/แก้ไข |
| `lib/screens/purchaseorder/purchaseorder_list_screen.dart` | หน้ารายการ |
| `lib/screens/purchaseorder/utils/po_workflow_manager.dart` | จัดการ flow ทั้งหมด |
| `lib/screens/purchaseorder/utils/po_approval_helper.dart` | กฎสิทธิ์ตามสถานะ |
| `lib/screens/purchaseorder/utils/po_save_helper.dart` | เตรียมข้อมูลก่อนบันทึก |
| `lib/screens/purchaseorder/components/po_list_card.dart` | Card + status badges |
| `lib/bloc/trans/trans_bloc.dart` | State management (ใช้ร่วมกับ transaction อื่น) |
