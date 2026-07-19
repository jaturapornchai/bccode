---
source: backend/internal/product/product/models/product_dimensions.go
tags: [datamodel, general-type]
---

# ProductDimensionPg

โครงสร้าง `ProductDimensionPg` จากโมดูล mainapi `product` มีฟิลด์ตามซอร์ส `product_dimensions.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| ProductGuid | string | - | productguid | ค่าของ ProductGuid ตามฟิลด์ `productguid` ในซอร์ส |
| DimensionGuid | string | - | dimensionguid | ค่าของ DimensionGuid ตามฟิลด์ `dimensionguid` ในซอร์ส |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
