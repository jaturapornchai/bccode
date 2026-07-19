---
source: backend/internal/models/identity.go
tags: [datamodel, general-type, mongo-inline]
---

# HoldingCodeentity

โครงสร้าง `HoldingCodeentity` ที่ฝังแบบ inline ในโมเดล MongoDB ตามซอร์ส `identity.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | holdingcode | holdingcode | รหัส Holding (tenant) |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
