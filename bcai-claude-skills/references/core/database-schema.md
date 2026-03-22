# Database Schema — PostgreSQL Business Tables

## สถาปัตยกรรม

- **Multi-tenant**: ทุก shop มี PostgreSQL database แยก ชื่อ database = `shopid`
- **Connection**: `mypg.PgSqlFastConnect(shopID)` — ต่อตรงเข้า database ของ shop นั้น
- **ทุก table มี `shopid`** แม้จะอยู่ใน database แยก (เพื่อ safety + query compatibility)

---

## Tables หลัก

### `doc` — เอกสารทุกประเภท

เอกสารการค้าทุกประเภทอยู่ใน table เดียว แยกด้วย `transflag`

| Column | Type | คำอธิบาย |
|--------|------|----------|
| `shopid` | varchar | รหัสร้านค้า (multi-tenant key) |
| `docno` | varchar | เลขที่เอกสาร (PK ร่วมกับ transflag) |
| `transflag` | int | ประเภทเอกสาร (ดูตาราง TransFlag ด้านล่าง) |
| `docdatetime` | timestamptz | วันเวลาเอกสาร |
| `perioddatetime` | timestamptz | วันที่รอบบัญชี |
| `custcode` | varchar | รหัสลูกค้า/ผู้ขาย |
| `totalamount` | numeric | ยอดรวมก่อนภาษี |
| `vatamount` | numeric | ภาษีมูลค่าเพิ่ม |
| `grandtotal` | numeric | ยอดรวมทั้งหมด |
| `roundamount` | numeric | ยอดปัดเศษ |
| `discountamount` | numeric | ส่วนลด |
| `paytype` | int | ประเภทการชำระ (1=เงินสด, 2=เครดิต) |
| `paycashamount` | numeric | จำนวนเงินสดรับ |
| `paycashchange` | numeric | เงินทอน |
| `branchid` | varchar | รหัสสาขา |
| `salechannelcode` | varchar | ช่องทางขาย |
| `iscancel` | boolean | ยกเลิกหรือไม่ |
| `isdelete` | boolean | ลบหรือไม่ (soft delete) |
| `cancelreason` | varchar | เหตุผลการยกเลิก |
| `taxdocno` | varchar | เลขที่ใบกำกับภาษี |
| `guidfixed` | varchar | GUID ของเอกสาร |
| `creator_code` | varchar | รหัสผู้สร้าง |
| `creator_name` | varchar | ชื่อผู้สร้าง |
| `created_at` | timestamptz | วันเวลาที่สร้าง |
| `currency` | varchar | สกุลเงินหลัก (สำหรับลงบัญชี) |
| `doccurrency` | varchar | สกุลเงินของเอกสาร |
| `exchangerate` | numeric | อัตราแลกเปลี่ยน |

**TransFlag — ประเภทเอกสาร:**

| transflag | ประเภท |
|-----------|--------|
| 1 | ใบเสนอราคา (Quotation) |
| 2 | ใบสั่งขาย (Sales Order) |
| 3 | ใบส่งของ/Invoice (Sales Invoice) |
| 8 | ใบลดหนี้ขาย (Sales Credit Note) |
| 12 | ใบรับสินค้า/ซื้อ (Purchase Invoice) |
| 16 | ใบส่งคืน (Purchase Return) |
| 20 | ใบสั่งซื้อ (Purchase Order) |
| 21 | ใบขอซื้อ (Purchase Requisition) |
| 22 | สืบราคา (Request for Quotation) |
| 30 | ปรับสต็อก (Stock Adjustment) |
| 31 | โอนสต็อก (Stock Transfer) |

**Query ตัวอย่าง:**
```sql
SELECT docno, docdatetime, custcode, grandtotal
FROM doc
WHERE shopid = $1
  AND transflag = $2
  AND iscancel = false
  AND (isdelete = false OR isdelete IS NULL)
  AND docdatetime >= $3
ORDER BY docdatetime DESC
LIMIT $4 OFFSET $5
```

---

### `docdetail` — รายการสินค้าในเอกสาร

| Column | Type | คำอธิบาย |
|--------|------|----------|
| `shopid` | varchar | รหัสร้านค้า |
| `docno` | varchar | เลขที่เอกสาร (FK → doc.docno) |
| `transflag` | int | ประเภทเอกสาร (FK → doc.transflag) |
| `linenumber` | int | ลำดับบรรทัด |
| `barcode` | varchar | บาร์โค้ดสินค้า |
| `itemcode` | varchar | รหัสสินค้า |
| `unitcode` | varchar | รหัสหน่วย |
| `totalqty` | numeric | จำนวน |
| `price` | numeric | ราคาต่อหน่วย |
| `sumamount` | numeric | ยอดรวมบรรทัด |
| `discountamount` | numeric | ส่วนลด |
| `calcflag` | numeric | บวก(+1)/ลบ(-1) สำหรับคำนวณ stock |
| `iscalcstock` | smallint | 1=คำนวณ stock, 0=ไม่คำนวณ |
| `unitstand` | numeric | อัตราส่วนหน่วยมาตรฐาน |
| `unitdivide` | numeric | ตัวหาร |
| `whcode` | varchar | รหัสคลัง |
| `locationcode` | varchar | รหัสที่เก็บ |
| `towhcode` | varchar | รหัสคลังปลายทาง (กรณีโอน) |
| `tolocationcode` | varchar | รหัสที่เก็บปลายทาง |
| `description` | text | คำอธิบายรายการ |

**Query ตัวอย่าง (doc JOIN docdetail):**
```sql
SELECT d.docno, d.docdatetime, dd.barcode, dd.itemcode, dd.totalqty, dd.price
FROM docdetail dd
INNER JOIN doc d ON dd.docno = d.docno AND dd.transflag = d.transflag
WHERE dd.barcode = $1
  AND d.transflag = $2
  AND d.iscancel = false
  AND (d.isdelete = false OR d.isdelete IS NULL)
ORDER BY d.docdatetime DESC
```

---

### `product` — สินค้า

| Column | Type | คำอธิบาย |
|--------|------|----------|
| `shopid` | varchar | รหัสร้านค้า |
| `itemcode` | varchar | รหัสสินค้า (PK ร่วมกับ shopid) |
| `name0` | varchar | ชื่อสินค้า (ภาษาหลัก) |
| `name1` | varchar | ชื่อสินค้า (ภาษาอังกฤษ) |
| `groupcode` | varchar | รหัสกลุ่มสินค้า |
| `unitcode` | varchar | หน่วยหลัก |
| `isdelete` | boolean | ลบหรือไม่ (soft delete) |
| `isinactive` | boolean | ปิดใช้งานหรือไม่ |

---

### `productbarcode` — Barcodes

| Column | Type | คำอธิบาย |
|--------|------|----------|
| `shopid` | varchar | รหัสร้านค้า |
| `barcode` | varchar | บาร์โค้ด (PK ร่วมกับ shopid) |
| `itemcode` | varchar | รหัสสินค้า (FK → product.itemcode) |
| `unitcode` | varchar | หน่วย |
| `unitstand` | numeric | อัตราส่วนหน่วยมาตรฐาน |
| `unitdivide` | numeric | ตัวหาร |
| `price1` | numeric | ราคาขาย 1 |
| `price2` | numeric | ราคาขาย 2 |
| `price3` | numeric | ราคาขาย 3 |
| `barcodemain` | varchar | barcode หลักของสินค้า |

---

### `stockbalance` — ยอดคงเหลือ

| Column | Type | คำอธิบาย |
|--------|------|----------|
| `shopid` | varchar | รหัสร้านค้า |
| `itemcode` | varchar | รหัสสินค้า |
| `whcode` | varchar | รหัสคลัง |
| `locationcode` | varchar | รหัสที่เก็บ |
| `balance` | numeric | ยอดคงเหลือ (หน่วยมาตรฐาน) |

> บางระบบคำนวณ balance จาก `docdetail` โดยตรง แทนที่จะใช้ `stockbalance`

---

### `stockcard` — ประวัติ Stock Movement

| Column | Type | คำอธิบาย |
|--------|------|----------|
| `shopid` | varchar | รหัสร้านค้า |
| `docno` | varchar | เลขที่เอกสาร |
| `transflag` | int | ประเภทเอกสาร |
| `docdatetime` | timestamptz | วันเวลา |
| `itemcode` | varchar | รหัสสินค้า |
| `barcode` | varchar | บาร์โค้ด |
| `whcode` | varchar | รหัสคลัง |
| `locationcode` | varchar | รหัสที่เก็บ |
| `qty` | numeric | จำนวน (บวก=รับ, ลบ=จ่าย) |
| `balance` | numeric | ยอดคงเหลือหลังรายการ |
| `price` | numeric | ราคาต้นทุน |

---

### `customer` — ลูกค้า

| Column | Type | คำอธิบาย |
|--------|------|----------|
| `shopid` | varchar | รหัสร้านค้า |
| `custcode` | varchar | รหัสลูกค้า (PK ร่วมกับ shopid) |
| `name0` | varchar | ชื่อลูกค้า |
| `name1` | varchar | ชื่อลูกค้า (EN) |
| `address` | text | ที่อยู่ |
| `tel` | varchar | เบอร์โทร |
| `taxid` | varchar | เลขผู้เสียภาษี |
| `creditlimit` | numeric | วงเงินเครดิต |
| `creditterm` | int | เครดิต (วัน) |
| `isdelete` | boolean | soft delete |

---

### `supplier` — ซัพพลายเออร์/เจ้าหนี้

โครงสร้างคล้าย `customer` แต่ชื่อ table คือ `supplier`

---

## MongoDB Collections

MongoDB เก็บข้อมูล denormalized สำหรับ read-heavy operations

| Collection | ข้อมูล |
|-----------|--------|
| `productbarcode` | สินค้า + barcode + ราคา + หน่วยทุกหน่วย |
| `customer` | ข้อมูลลูกค้าพร้อม balance |
| `supplier` | ข้อมูล supplier/creditor |
| `employee` | พนักงาน |
| `warehouse` | คลังสินค้า + locations |
| `trans` | ข้อมูล transaction แบบ denormalized |

ทุก MongoDB document มี field `shopid` และ `guidfixed`

---

## ClickHouse Tables (Analytics)

| Table | ข้อมูล |
|-------|--------|
| `sales_logs` | บันทึก transactions ทุกรายการสำหรับ analytics |
| `stock_logs` | บันทึก stock movements |

ใช้สำหรับ **read-only analytics** เท่านั้น — ห้าม write จาก handler ปกติ

---

## Tips

- ใช้ `AND (isdelete = false OR isdelete IS NULL)` เสมอ — ไม่ใช่แค่ `AND isdelete = false`
- `docno + transflag` เป็น composite key ของ doc และ docdetail
- ยอดคงเหลือ stock = `SUM((totalqty * calcflag) * unitstand / unitdivide)` จาก docdetail ที่ `iscalcstock = 1`
