---
source: image-model.go
tags: [datamodel, general-type]
---

# ImageListRequest

Request body สำหรับดึงรายการรูปภาพ พร้อม pagination (limit/skip) และตัวเลือกให้รวม base64 data

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| Category | string | - | category,omitempty | กรองตามหมวดหมู่ |
| Limit | int64 | - | limit,omitempty | จำนวนสูงสุดต่อหน้า |
| Skip | int64 | - | skip,omitempty | ข้ามกี่รายการ (pagination) |
| IncludeData | bool | - | includedata,omitempty | true = รวม base64 data ด้วย |

## ความสัมพันธ์

- `HoldingCode` — อ้างอิงกลุ่มกิจการ (tenant) ที่ต้องการดูรายการรูป
- ผลลัพธ์คืนเป็น [[ImageListResponse]], [[ImageListDataResponse]] หรือ [[ImageListResponseWithTotal]]
