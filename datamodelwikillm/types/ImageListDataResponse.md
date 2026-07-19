---
source: image-model.go
tags: [datamodel, general-type]
---

# ImageListDataResponse

Response สำหรับรายการรูปภาพพร้อม URL/base64 และจำนวนทั้งหมด (total)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Status | string | - | status | สถานะผลลัพธ์ |
| Code | int | - | code | รหัสผลลัพธ์ |
| Count | int | - | count | จำนวนรายการที่คืนมา |
| Total | int64 | - | total | จำนวนทั้งหมด (ไม่รวม limit) |
| Data | [][[ImageListDataItem]] | - | data | รายการรูปภาพพร้อม URL/base64 |

## ความสัมพันธ์

- `Data` — embed [][[ImageListDataItem]]
- คู่กับ request [[ImageListRequest]] (เมื่อ `IncludeData` = true)
