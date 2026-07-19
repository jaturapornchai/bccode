---
source: backend/internal/models/activity.go
tags: [datamodel, general-type, mongo-inline]
---

# ActivityDoc

โครงสร้าง `ActivityDoc` ที่ฝังแบบ inline ในเอกสาร MongoDB ตามซอร์ส `activity.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| CreatedBy | string | createdby | - | ผู้สร้างข้อมูล |
| CreatedAt | time.Time | createdat | - | เวลาสร้างข้อมูล UTC |
| UpdatedBy | string | updatedby,omitempty | - | ผู้แก้ไขข้อมูลล่าสุด |
| UpdatedAt | time.Time | updatedat,omitempty | - | เวลาแก้ไขล่าสุด UTC |
| DeletedBy | string | deletedby,omitempty | - | ผู้ลบข้อมูล |
| DeletedAt | time.Time | deletedat,omitempty | - | เวลาลบข้อมูล UTC |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
