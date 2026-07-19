---
source: process-doc-purchase-model.go
tags: [datamodel, general-type]
---

# PurchaseStatusStruct

โมเดลสถานะเอกสารซื้อ ใช้ตอบสถานะการรับสินค้าของเอกสารซื้อทั้งใบ (ปิดหรือยัง, เทียบรับครบ/ไม่ครบ/เกิน) พร้อมรายละเอียดรายบรรทัด

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocNo | string | - | docno | เลขที่เอกสาร |
| IsClosed | bool | - | isclosed | ปิดเอกสารแล้วหรือไม่ |
| IsComparedSuccess | int | - | iscomparedsuccess | ผลเทียบการรับสินค้า: 0=ยังไม่ได้รับสินค้า, 1=รับครบ, 2=รับไม่ครบ, 3=รับเกิน |
| Details | [][[PurchaseStatusDocDetailStruct]] | - | details | รายละเอียดสถานะรายบรรทัดของเอกสาร |

## ความสัมพันธ์

- `Details` → embed [[PurchaseStatusDocDetailStruct]] (รายบรรทัด)
- `DocNo` → อ้างอิงเลขที่เอกสารซื้อ
