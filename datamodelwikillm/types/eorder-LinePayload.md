---
source: backend/internal/product/eorder/models/line.go
tags: [datamodel, general-type]
---

# LinePayload

โครงสร้าง `LinePayload` จากโมดูล mainapi `eorder` มีฟิลด์ตามซอร์ส `line.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Token | string | - | token | ค่าของ Token ตามฟิลด์ `token` ในซอร์ส |
| Message | string | - | message | ข้อความผลลัพธ์ |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
