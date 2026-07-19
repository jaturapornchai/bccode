---
source: backend/internal/models/name.go
tags: [datamodel, general-type]
---

# NameX

โครงสร้าง `NameX` สำหรับชื่อหลายภาษาที่โมเดล MongoDB อ้างอิง ตามซอร์ส `name.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Code | *string | code | code | รหัสภาษา |
| Name | *string | name | name | ชื่อในภาษานั้น |
| IsAuto | bool | isauto | isauto | ระบุว่าระบบสร้างให้อัตโนมัติ |
| IsDelete | bool | - | isdelete | ระบุว่ารายการถูกลบ |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
