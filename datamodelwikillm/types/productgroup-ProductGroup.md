---
source: backend/internal/product/productgroup/models/productgroup.go
tags: [datamodel, general-type]
---

# ProductGroup

โครงสร้าง `ProductGroup` จากโมดูล mainapi `productgroup` มีฟิลด์ตามซอร์ส `productgroup.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| PartitionIdentity | [[PartitionIdentity\|models.PartitionIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| Code | string | code | code | รหัสรายการ |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| ParentGUID | string | parentguid | parentguid | GUID ของรายการแม่ |
| ParentGUIDAll | string | parentguidall | parentguidall | ลำดับ GUID ของรายการแม่ทั้งหมด |
| XSorts | *[][[XSort\|models.XSort]] | xsorts | xsorts | ค่าของ XSorts ตามฟิลด์ `xsorts` ในซอร์ส |
| ChildCount | int | childcount | childcount | จำนวนรายการลูก |
| IsDisabled | bool | isdisabled | isdisabled | ค่าของ IsDisabled ตามฟิลด์ `isdisabled` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[PartitionIdentity]], [[NameX]], [[XSort]]
