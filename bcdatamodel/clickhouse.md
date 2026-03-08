# ClickHouse Tables

ตาราง OLAP/Analytics ทั้งหมด 29 ตาราง + 1 Dictionary

**Source:** `internal/goapi/myclickhouse/ensure_tables.go`

## Engine Types

| Engine | ลักษณะ | ตารางที่ใช้ |
|--------|--------|------------|
| **MergeTree** | Append-only, เพิ่มได้อย่างเดียว | docdetail, processstock*, cartorder* |
| **ReplacingMergeTree** | Dedup by ORDER BY key | doc, productbarcode, shop, docpayment |
| **Memory** | In-memory ชั่วคราว | temp_barcode_map |

## Partitioning Strategy

- ส่วนใหญ่ partition by `shopid` (multi-tenant isolation)
- `processstockcost` partition by `(shopid, itemcode)` — heavy query optimization
- `shop` partition by `mongodbname`

---

## Core Transaction Tables

### doc (เอกสาร Header)

**Engine:** ReplacingMergeTree | **Partition:** shopid | **Order:** (guidfixed, shopid, docno)

| Column | Type | Default | หมายเหตุ |
|--------|------|---------|----------|
| shopid | String | — | รหัสร้าน |
| docno | String | — | เลขที่เอกสาร |
| docdatetime | DateTime | — | วันเวลาเอกสาร |
| perioddatetime | DateTime | — | งวดบัญชี |
| taxdocno | String | '' | เลขใบกำกับภาษี |
| totalamount | Float64 | — | ยอดรวม |
| roundamount | Float64 | 0 | ปัดเศษ |
| paytype | Int16 | 0 | ประเภทการชำระ |
| paycashamount | Float64 | 0 | เงินสด |
| paycashchange | Float64 | 0 | เงินทอน |
| paycashbalance | Float64 | 0 | ยอดคงเหลือ |
| deliverycode | String | '' | รหัสจัดส่ง |
| checksum | String | — | checksum |
| branchid | String | '' | รหัสสาขา |
| slipurl | String | '' | URL สลิป |
| salechannelcode | String | — | ช่องทางขาย |
| deliveryamount | Float64 | — | ค่าจัดส่ง |
| iscancel | Bool | false | ยกเลิก |
| cancelreason | String | '' | เหตุผลยกเลิก |
| guidpos | String | '' | GUID POS |
| guidbranch | String | '' | GUID สาขา |
| guidfixed | String | '' | GUID เอกสาร |
| **transflag** | **Int16** | **0** | **ประเภทเอกสาร** (ดู enums.md) |
| isdelete | Bool | 0 | ลบแล้ว |
| currency | String | '' | สกุลเงินหลัก |
| currency_symbol | String | '' | สัญลักษณ์ |
| doc_currency | String | '' | สกุลเงินเอกสาร |
| doc_currency_symbol | String | '' | สัญลักษณ์เอกสาร |
| exchange_rate | Float64 | 1 | อัตราแลกเปลี่ยน |
| totalamount_doc | Float64 | 0 | ยอดรวม (doc currency) |
| approval_status | String | '' | สถานะอนุมัติ |
| isclosedmanual | Bool | false | ปิดเอกสารเอง |
| closedmanual_by_code | String | '' | รหัสผู้ปิด |
| closedmanual_by_name | String | '' | ชื่อผู้ปิด |
| closedmanual_at | DateTime | '1970-01-01' | วันเวลาปิด |
| closedmanual_reason | String | '' | เหตุผลปิด |

**Indexes:** idx_shopid (minmax), idx_branchid (minmax), idx_perioddatetime (minmax), idx_shopid_transflag (minmax)

### docdetail (รายละเอียดสินค้า)

**Engine:** MergeTree | **Partition:** shopid | **Order:** (shopid, docno, line_number)

| Column | Type | Default | หมายเหตุ |
|--------|------|---------|----------|
| shopid | String | — | รหัสร้าน |
| docno | String | — | เลขที่เอกสาร |
| docdatetime | DateTime | — | วันเวลา |
| perioddatetime | DateTime | — | งวดบัญชี |
| line_number | UInt32 | — | ลำดับบรรทัด |
| barcode | String | — | barcode |
| unitcode | String | — | รหัสหน่วย |
| qty | Float64 | — | จำนวน |
| price | Float64 | — | ราคา |
| discount | String | — | สูตรส่วนลด |
| whcode | String | — | รหัสคลัง |
| locationcode | String | — | รหัสตำแหน่ง |
| sumofcost | Float64 | — | ต้นทุนรวม |
| discountamount | Float64 | — | จำนวนส่วนลด |
| sumamount | Float64 | — | ยอดรวม |
| branchid | String | — | รหัสสาขา |
| itemnames | String | — | ชื่อสินค้า |
| refguid | String | — | GUID อ้างอิง |
| sumamountchoice | Float64 | — | ยอดตัวเลือก |
| ischoice | Int32 | — | เป็นตัวเลือก |
| guidfixed | String | '' | GUID เอกสาร |
| guidpos | String | '' | GUID POS |
| guidbranch | String | '' | GUID สาขา |
| **transflag** | **Int16** | **0** | **ประเภทเอกสาร** |
| isupdated | Bool | false | อัปเดตแล้ว |
| itemcode | Nullable(String) | NULL | รหัสสินค้า |
| unitstand | Float64 | 1 | ตัวตั้ง |
| unitdivide | Float64 | 1 | ตัวหาร |
| itemname | String | — | ชื่อสินค้า (single) |
| calcflag | Int8 | — | 1=เข้า, -1=ออก |
| calcseq | Int32 | — | ลำดับคำนวณ |
| barcodemain | String | — | barcode หลัก |
| iscalcstock | Int8 | 0 | คำนวณ stock หรือไม่ |
| price_doc | Float64 | 0 | ราคา (doc currency) |
| sumamount_doc | Float64 | 0 | ยอดรวม (doc currency) |
| discountamount_doc | Float64 | 0 | ส่วนลด (doc currency) |
| priceexcludevat_doc | Float64 | 0 | ราคาไม่รวม VAT (doc) |
| sumamountexcludevat_doc | Float64 | 0 | ยอดไม่รวม VAT (doc) |
| totalvaluevat_doc | Float64 | 0 | VAT (doc currency) |

**Indexes:** idx_barcode (bloom_filter 0.01), idx_shopid (minmax), idx_branchid (minmax), idx_perioddatetime (minmax), idx_shopid_transflag (minmax), idx_shopid_isupdated (set 100), idx_shop_item (minmax)

### docpayment (การชำระเงิน)

**Engine:** ReplacingMergeTree | **Partition:** shopid | **Order:** (guidfixed, shopid, docno)

| Column | Type | Default | หมายเหตุ |
|--------|------|---------|----------|
| shopid | String | — | รหัสร้าน |
| branchid | String | — | รหัสสาขา |
| docdatetime | DateTime | — | วันเวลา |
| perioddatetime | DateTime | — | งวดบัญชี |
| amount | Decimal(18,6) | — | จำนวนเงิน |
| description | String | — | คำอธิบาย |
| docno | String | — | เลขที่เอกสาร |
| trans_flag | Int32 | — | ประเภทเอกสาร |
| guidfixed | String | '' | GUID |
| guidbranch | String | '' | GUID สาขา |

### docref (เอกสารอ้างอิง)

**Engine:** MergeTree | **Order:** (shopid, docno, docnotransflag)

| Column | Type | หมายเหตุ |
|--------|------|----------|
| shopid | String | รหัสร้าน |
| docno | String | เลขที่เอกสาร |
| docnotransflag | Int32 | transflag เอกสาร |
| docnoref | String | เลขที่อ้างอิง |
| docnoreftransflag | Int32 | transflag อ้างอิง |

---

## Stock Processing Tables

### processstockcost (ต้นทุนสินค้า — Optimized)

**Engine:** MergeTree | **Partition:** (shopid, itemcode) | **Order:** (shopid, itemcode, docdatetime, whcode, locationcode)

ZSTD compression + LowCardinality — ตารางที่ query หนักที่สุด

| Column | Type | หมายเหตุ |
|--------|------|----------|
| shopid | LowCardinality(String) | รหัสร้าน |
| itemcode | LowCardinality(String) | รหัสสินค้า |
| docdatetime | DateTime (Delta+ZSTD) | วันเวลา |
| docno | String | เลขที่เอกสาร |
| linenumber | UInt16 | ลำดับ |
| transflag | UInt8 | ประเภท |
| barcodemain | String | barcode หลัก |
| barcode | String | barcode |
| unitcode | LowCardinality(String) | หน่วย |
| whcode | LowCardinality(String) | คลัง |
| locationcode | LowCardinality(String) | ตำแหน่ง |
| totalqty | Float64 | จำนวน |
| unitstand | Float64 | ตัวตั้ง |
| unitdivide | Float64 | ตัวหาร |
| price | Float64 | ราคา |
| averagecost | Float64 | ต้นทุนเฉลี่ย |
| calcamount | Float64 | มูลค่าคำนวณ |
| balanceqty | Float64 | ยอดคงเหลือ |
| balanceamount | Float64 | มูลค่าคงเหลือ |
| guid | String | GUID |
| unitcost | Float64 | ต้นทุนต่อหน่วย |
| docref | String | เอกสารอ้างอิง |
| originalqty | Int32 | จำนวนเดิม |
| qty | Int32 | จำนวน |

**Settings:** index_granularity=16384, parts_to_throw_insert=1000, parts_to_delay_insert=500

### processstockdetail (รายละเอียด Stock)

**Engine:** MergeTree | **Partition:** shopid | **Order:** (shopid, barcode)

| Column | Type | หมายเหตุ |
|--------|------|----------|
| shopid | LowCardinality(String) | รหัสร้าน |
| docdatetime | DateTime | วันเวลา |
| docno | String | เลขที่เอกสาร |
| linenumber | UInt16 | ลำดับ |
| transflag | UInt16 | ประเภท |
| calcflag | UInt16 | ทิศทาง stock |
| calcseq | UInt16 | ลำดับคำนวณ |
| barcodemain | String | barcode หลัก |
| barcode | String | barcode |
| unitcode | LowCardinality(String) | หน่วย |
| whcode | LowCardinality(String) | คลัง |
| locationcode | LowCardinality(String) | ตำแหน่ง |
| totalqty | Float64 | จำนวน |
| unitstand | Float64 | ตัวตั้ง |
| unitdivide | Float64 | ตัวหาร |
| price | Float64 | ราคา |
| priceexcludevat | Float64 | ราคาไม่รวม VAT |
| docref | String | อ้างอิง |

### processstocklot (Lot Tracking)

**Engine:** MergeTree | **Partition:** shopid | **Order:** (shopid, barcodemain)

| Column | Type | หมายเหตุ |
|--------|------|----------|
| shopid | LowCardinality(String) | รหัสร้าน |
| docdatetime | DateTime | วันเวลา |
| lotnumber | String | เลข Lot |
| docno | String | เลขที่เอกสาร |
| transflag | UInt16 | ประเภท |
| barcodemain | String | barcode หลัก |
| unitcode | LowCardinality(String) | หน่วย |
| whcode | LowCardinality(String) | คลัง |
| locationcode | LowCardinality(String) | ตำแหน่ง |
| qty | Float64 | จำนวน |
| unitstand | Float64 | ตัวตั้ง |
| unitdivide | Float64 | ตัวหาร |
| price | Float64 | ราคา |
| cost | Float64 | ต้นทุน |
| balanceqty | Float64 | ยอดคงเหลือ |
| balanceamount | Float64 | มูลค่าคงเหลือ |
| guidref | String | GUID อ้างอิง |

### processstock (สรุป Stock — Header)

**Engine:** ReplacingMergeTree | **Partition:** shopid | **Order:** (guidfixed, shopid, docno)

โครงสร้างเหมือน `doc` แต่ใช้ `transflag UInt16` (ไม่มี default)

---

## Master Data Tables

### productbarcode

**Engine:** ReplacingMergeTree | **Partition:** shopid | **Order:** (shopid, barcode)

ดูรายละเอียดที่ [products.md](products.md#productbarcode-clickhouse-table)

### productbarcodeprocess (Processing optimization)

**Engine:** MergeTree | **Order:** (shopid, barcode)

| Column | Type | หมายเหตุ |
|--------|------|----------|
| shopid | LowCardinality(String) | รหัสร้าน |
| barcode | String | barcode |
| name0 | String | ชื่อ |
| unitname | LowCardinality(String) | หน่วย |
| unitcode | LowCardinality(String) | รหัสหน่วย |
| itemtype | UInt8 | ประเภท |
| barcoderef | String | barcode อ้างอิง |
| barcoderefunitstand | Float64 | ตัวตั้ง ref |
| barcoderefunitdivide | Float64 | ตัวหาร ref |
| isstock | UInt8 | เป็น stock |
| itemcode | String | รหัสสินค้า |

### productbarcoderef (Barcode mapping)

**Engine:** MergeTree | **Partition:** shopid | **Order:** (shopid, barcode, id)

| Column | Type | หมายเหตุ |
|--------|------|----------|
| id | UInt64 | ID |
| shopid | String | รหัสร้าน |
| barcode | String | barcode |
| barcoderef | String | barcode อ้างอิง |
| itemcode | String | รหัสสินค้า |
| unitcode | String | รหัสหน่วย |
| standvalue | Float64 | ตัวตั้ง (default 1) |
| dividevalue | Float64 | ตัวหาร (default 1) |

### productbarcodeimport (Staging)

**Engine:** MergeTree | **Order:** (taskid, rownumber)

ตาราง staging สำหรับ import สินค้า — มี field ครบทุกมิติ (group, brand, category, etc.)

---

## Other Tables

### cartorder / cartorderdetail
ตะกร้าสินค้า (Cart order) — สำหรับ e-commerce

### token
เก็บ authentication token

| Column | Type | หมายเหตุ |
|--------|------|----------|
| tokenid | String | Token ID |
| shopid | String | รหัสร้าน |
| userid | String | รหัสผู้ใช้ |
| active | UInt8 | Active (default 1) |

### userlogin
เก็บ user login info + shop list

### shop
Master shop data (ดู [organization.md](organization.md#shop-ร้านค้า))

### task_status
Background task tracking

| Column | Type | หมายเหตุ |
|--------|------|----------|
| task_id | String | Task ID |
| shop_id | String | รหัสร้าน |
| status | String | สถานะ (pending/running/completed/failed) |
| error_message | String | ข้อผิดพลาด |
| progress | Int32 | เปอร์เซ็นต์ |
| created_at | DateTime('UTC') | สร้างเมื่อ |
| updated_at | DateTime('UTC') | อัปเดตเมื่อ |
| completed_at | DateTime('UTC') | เสร็จเมื่อ |

### result / resultfordashboard
เก็บผลลัพธ์การคำนวณ (JSON format)

### stockwaitprocess
คิว stock ที่รอประมวลผล

### temp_barcode_map (Memory engine)
ตาราง temporary ใน RAM — ใช้ map barcode → itemcode ตอนคำนวณ

---

## Dictionary

### productbarcode_dict

```sql
CREATE DICTIONARY productbarcode_dict (
  barcode String,
  itemcode String,
  unitstand Float64,
  unitdivide Float64
) PRIMARY KEY barcode
SOURCE(CLICKHOUSE(QUERY '...'))
LAYOUT(COMPLEX_KEY_HASHED())
```

ใช้สำหรับ lookup barcode → itemcode ในการ query — เร็วกว่า JOIN

---

## Query Patterns ที่ใช้บ่อย

### Stock Balance
```sql
SELECT itemcode,
  SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance
FROM docdetail
WHERE shopid = ? AND transflag IN (1,3,5,7,9,11,13,16,18,20,30,31,32,33,34,35,36)
GROUP BY itemcode
```

### Sales Summary
```sql
SELECT SUM(totalamount) as total_sales, COUNT(*) as transaction_count
FROM doc
WHERE shopid = ? AND transflag = 44 AND iscancel = false
  AND docdatetime BETWEEN ? AND ?
```

### Inventory Value
```sql
SELECT pb.itemcode, pb.name0 as name,
  SUM((dd.qty * dd.calcflag) * dd.unitstand / NULLIF(dd.unitdivide,0)) as stock_qty
FROM docdetail dd
JOIN productbarcode pb ON dd.shopid = pb.shopid AND dd.barcode = pb.barcode
WHERE dd.shopid = ? AND dd.transflag IN (...)
GROUP BY pb.itemcode, pb.name0
```
