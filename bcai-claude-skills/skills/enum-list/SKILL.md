---
name: enum-list
description: ดูรายการ enums ที่ backend กำหนด — ใช้เมื่อต้องการรู้ค่า enum (สถานะเอกสาร, ประเภทสินค้า, payment type, doc status), ตรวจสอบ enum values ก่อนสร้าง dropdown/radio/switch, sync enum ระหว่าง frontend-backend, หรือถามว่า "สถานะ ... มีอะไรบ้าง", "ค่า enum ... คืออะไร"
user-invocable: true
---

# Enum Catalog — ดูรายการ Enums

## วิธีใช้
`/enum-list` — แสดง enum ทั้งหมด
`/enum-list <keyword>` — ค้นหา enum ที่มี keyword

## ขั้นตอนการทำงาน

### 1. ดูรายการ Enums
เรียก MCP tool `list_enums`:
```
Tool: list_enums
Parameters:
  - keyword: "payment" (optional — filter ตาม keyword)
```

### 2. แสดงผล
```
## Enum: DocStatus (สถานะเอกสาร)

| Value | Label (TH) | Label (EN) |
|-------|-----------|-----------|
| 0 | ร่าง | Draft |
| 1 | รออนุมัติ | Pending |
| 2 | อนุมัติ | Approved |
| 3 | ยกเลิก | Cancelled |

## Enum: PaymentType (ประเภทการชำระ)

| Value | Label |
|-------|-------|
| cash | เงินสด |
| transfer | โอนเงิน |
| credit | เครดิต |
| qr | QR Payment |
```

### 3. แนะนำการใช้งานใน Flutter
```dart
// สร้าง Dart enum ตาม backend
enum DocStatus {
  draft(0, 'ร่าง'),
  pending(1, 'รออนุมัติ'),
  approved(2, 'อนุมัติ'),
  cancelled(3, 'ยกเลิก');

  final int value;
  final String label;
  const DocStatus(this.value, this.label);
}
```

## ตัวอย่าง
```
/enum-list                → แสดงทั้งหมด
/enum-list payment        → enum เกี่ยวกับ payment
/enum-list doc             → enum เกี่ยวกับเอกสาร
/enum-list status          → enum เกี่ยวกับสถานะ
/enum-list product         → enum เกี่ยวกับสินค้า
```

## MCP Tools ที่ใช้
| Tool | หน้าที่ |
|------|--------|
| `list_enums` | ดูรายการ enum + values ทั้งหมด |

## หมายเหตุ
- `list_enums` อยู่ใน MCP Dev endpoint เท่านั้น
- Enum values ต้องตรงกันระหว่าง frontend-backend เสมอ
- ถ้า backend เพิ่ม enum value ใหม่ → frontend ต้อง handle ด้วย (default case)
