---
source: backend/internal/product/option/models/option.go
tags: [datamodel, general-type]
---

# InventoryOptionPageResponse

โครงสร้าง `InventoryOptionPageResponse` จากโมดูล mainapi `option` มีฟิลด์ตามซอร์ส `option.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Success | bool | - | success | สถานะความสำเร็จ |
| Data | [][[option-InventoryOptionMainInfo\|InventoryOptionMainInfo]] | - | data,omitempty | ค่าของ Data ตามฟิลด์ `data` ในซอร์ส |
| Pagination | [[PaginationDataResponse\|common.PaginationDataResponse]] | - | pagination,omitempty | ค่าของ Pagination ตามฟิลด์ `pagination` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[option-InventoryOptionMainInfo|InventoryOptionMainInfo]], [[PaginationDataResponse]]
