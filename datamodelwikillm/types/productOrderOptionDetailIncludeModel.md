---
source: backend/internal/goapi/models/global-model.go
tags: [datamodel, general-type]
---

# productOrderOptionDetailIncludeModel

โครงสร้างการอ้างอิงตัวเลือกที่รวมอยู่ในรายละเอียดตัวเลือกสินค้า เป็นโครงสร้างแบบ recursive (Details อ้างอิงชนิดตัวเอง) ใช้แสดงลำดับชั้นของตัวเลือกที่เชื่อมโยงกัน

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Optionguid | string | - | optionguid | รหัส GUID ของตัวเลือกที่อ้างถึง |
| Details | [][[productOrderOptionDetailIncludeModel]] | - | details | รายการลูกแบบ recursive ของตัวเลือกที่รวมอยู่ |

## ความสัมพันธ์
- ฝังตัวเอง (recursive) ผ่านฟิลด์ Details
- ถูกฝังใน [[productOrderOptionDetailModel]] (Includeoptions)
- `Optionguid` — อ้างอิง Guidcode ของ [[productOrderOptionModel]] / [[productOrderOptionDetailModel]]
