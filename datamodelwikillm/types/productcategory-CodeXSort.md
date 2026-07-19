---
source: backend/internal/product/productcategory/models/productcategory.go
tags: [datamodel, general-type]
---

# CodeXSort

โครงสร้าง `CodeXSort` จากโมดูล mainapi `productcategory` มีฟิลด์ตามซอร์ส `productcategory.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Code | string | code | code | รหัสรายการ |
| XOrder | uint | xorder | xorder | ลำดับการแสดงผล |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[NameX]]
