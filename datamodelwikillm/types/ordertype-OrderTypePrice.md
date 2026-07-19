---
source: backend/internal/product/ordertype/models/ordertype.go
tags: [datamodel, general-type]
---

# OrderTypePrice

โครงสร้าง `OrderTypePrice` จากโมดูล mainapi `ordertype` มีฟิลด์ตามซอร์ส `ordertype.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Type | int8 | type | type | ค่าของ Type ตามฟิลด์ `type` ในซอร์ส |
| Price | float64 | price | price | ราคา |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
