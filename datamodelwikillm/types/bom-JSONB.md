---
source: backend/internal/product/bom/models/bom_postgres.go
tags: [datamodel, general-type]
---

# JSONB

โครงสร้าง `JSONB` จากโมดูล mainapi `bom` มีฟิลด์ตามซอร์ส `bom_postgres.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Raw | json.RawMessage | - | - | ค่าของ Raw ตามฟิลด์ `Raw` ในซอร์ส |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
