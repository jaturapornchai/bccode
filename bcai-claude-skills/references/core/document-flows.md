# Document Flows — State Machine ของทุก Module

## Document Types

| Code | ชื่อ | ฝั่งที่ใช้ |
|------|------|----------|
| `PO` | Purchase Order | จัดซื้อ |
| `PI` | Purchase Invoice | จัดซื้อ |
| `SO` | Sales Order | ขาย |
| `SI` | Sales Invoice | ขาย |
| `SA` | Stock Adjustment | คลังสินค้า |
| `TR` | Transfer (โอนย้าย) | คลังสินค้า |

## State Machine หลัก

```
                    [แก้ไขแล้วส่งใหม่]
                         ↑
draft ──submit──► pending ──approve──► approved ──complete──► completed
                    │
                    └──reject──► rejected

[ยกเลิกได้ทุก state ยกเว้น completed]
draft/pending/approved ──cancel──► cancelled
```

## กฎต่อ State

| State | แก้ไขได้? | ลบได้? | ส่ง approve ได้? | พิมพ์ได้? |
|-------|----------|--------|-----------------|----------|
| `draft` | yes | yes | yes | yes |
| `pending` | no | no | no | yes |
| `rejected` | yes | no | yes (ส่งใหม่) | yes |
| `approved` | no | no | no | yes |
| `completed` | no | no | no | yes |
| `cancelled` | no | no | no | yes |

## Approval Workflow

### เงื่อนไขก่อน approve
- เอกสารต้องอยู่ใน state `pending`
- user ที่ approve ต้องมี role `approver` สำหรับ document type นั้น
- ยอดรวมต้องมากกว่า 0
- ต้องมีรายการสินค้าอย่างน้อย 1 รายการ

### เมื่อ approve สำเร็จ
- เปลี่ยน `docstatus` → `approved`
- ส่ง LINE OA notification แจ้งผู้สร้างเอกสาร
- บันทึก audit log (ใครอนุมัติ, เวลา)

### เมื่อ reject
- เปลี่ยน `docstatus` → `rejected`
- บันทึก `reject_reason`
- ส่ง LINE OA แจ้งผู้สร้าง

## Flow แยกตาม Module

### Purchase Flow
```
PO (draft) → PO (approved) → PI (สร้างใหม่อ้างอิง PO) → PI (approved)
```
- PI ต้องอ้างอิง PO ที่ approved แล้วเท่านั้น
- Stock เพิ่มเมื่อ PI `completed`

### Sales Flow
```
SO (draft) → SO (approved) → SI (สร้างจาก SO) → SI (approved) → SI (completed)
```
- Stock ลดเมื่อ SI `completed`
- ถ้า SI ยกเลิก → stock คืน

### Stock Adjustment
```
SA (draft) → SA (approved) → SA (completed)
```
- Stock เปลี่ยนทันทีเมื่อ `completed`
- ต้องระบุเหตุผลการปรับ

### Transfer
```
TR (draft) → TR (approved) → TR (completed)
```
- Stock ออกจาก warehouse ต้นทาง + เข้า warehouse ปลายทาง เมื่อ `completed`
- ทั้งสอง warehouse ต้องอยู่ใน shop เดียวกัน

## Document Number Format

- docno สร้างโดย backend อัตโนมัติ
- Format: `{PREFIX}{YY}{MM}{RUNNING_NUMBER}` เช่น `PO2501001`
- ห้าม frontend สร้าง docno เอง

## Database Fields สำคัญ (table: doc)

```sql
docno        VARCHAR  -- เลขที่เอกสาร (unique per shop per doctype)
doctype      VARCHAR  -- PO/PI/SO/SI/SA/TR
docstatus    VARCHAR  -- draft/pending/approved/completed/cancelled/rejected
docdatetime  TIMESTAMPTZ -- วันที่เอกสาร
shopid       VARCHAR  -- multi-tenant key
createdby    VARCHAR  -- user ที่สร้าง
approvedby   VARCHAR  -- user ที่อนุมัติ
approvedat   TIMESTAMPTZ
totalamount  NUMERIC  -- ยอดรวมก่อน VAT
vatamount    NUMERIC  -- ภาษีมูลค่าเพิ่ม
grandtotal   NUMERIC  -- ยอดรวมหลัง VAT
```
