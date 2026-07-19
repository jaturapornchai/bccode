---
source: backend/internal/product/productbarcode/models/product_unit.go
tags: [datamodel, general-type]
---

# ProductUnit

โครงสร้าง `ProductUnit` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_unit.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| Names | *[]models.NameX | names | names | รายการชื่อหลายภาษา |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
