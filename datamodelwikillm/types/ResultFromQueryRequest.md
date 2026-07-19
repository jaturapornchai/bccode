---
source: query-result-model.go
tags: [datamodel, general-type]
---

# ResultFromQueryRequest

Request model สำหรับ endpoint `/resultfromquery` — ส่ง SQL SELECT query ไปรันและเก็บผลลัพธ์ไว้ตาม GUID (แปลงมาจาก Python models.py)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัส holding (บังคับ) |
| Query | string | - | query | SQL SELECT query ที่จะรัน (บังคับ) |
| GUID | string | - | guid | GUID ของผลลัพธ์ (ไม่ระบุ = สร้างให้อัตโนมัติ) |

## ความสัมพันธ์

- `HoldingCode` — อ้างอิงกลุ่มกิจการ (holding) เป็นขอบเขต tenant ของ query
