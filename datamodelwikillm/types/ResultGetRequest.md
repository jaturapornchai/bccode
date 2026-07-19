---
source: query-result-model.go
tags: [datamodel, general-type]
---

# ResultGetRequest

Request model สำหรับ endpoint `/resultget` — ดึงผลลัพธ์ query ที่เก็บไว้ตาม GUID แบบแบ่งหน้า (limit/offset)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัส holding (บังคับ) |
| GUID | string | - | guid | GUID ของผลลัพธ์ query ที่ต้องการดึง (บังคับ) |
| Limit | int | - | limit | จำนวนแถวที่ต้องการ (สูงสุด 1000) |
| Offset | int | - | offset | จำนวนแถวที่ข้าม |

## ความสัมพันธ์

- `HoldingCode` — อ้างอิงกลุ่มกิจการ (holding) เป็นขอบเขต tenant
- `GUID` — อ้างอิงผลลัพธ์ query ที่สร้างจาก [[ResultFromQueryRequest]]
