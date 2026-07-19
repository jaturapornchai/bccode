---
source: mongo-model.go
tags: [datamodel, general-type]
---

# ResultModel

โครงสร้างผลลัพธ์มาตรฐานสำหรับตอบกลับการทำงาน (response/result) บอกสถานะสำเร็จ, guid อ้างอิง และข้อมูล pagination (จำนวน record/หน้าทั้งหมด)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Success | bool | success | success | สถานะว่าการทำงานสำเร็จหรือไม่ |
| Guid | string | guid | guid | รหัสอ้างอิง (GUID) ของผลลัพธ์/รายการ |
| TotalRecord | int | totalrecord | totalrecord | จำนวน record ทั้งหมด |
| TotalPage | int | totalpage | totalpage | จำนวนหน้าทั้งหมด (pagination) |

## ความสัมพันธ์

- ไม่มี struct ซ้อน (nested) และไม่มี field อ้างอิงด้วยรหัส (code reference) ในโมเดลนี้
