---
source: process-model.go
tags: [datamodel, general-type]
---

# LocationItemStruct

รายการที่เก็บ (location) หนึ่งรายการใน payload — รหัส ชื่อ และสถานะถูกเลือก (มีเฉพาะ json tag)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Code | string | - | code | รหัสที่เก็บ |
| Title | string | - | title | ชื่อที่เก็บ |
| IsSelected | bool | - | isselected | ถูกเลือกหรือไม่ |

## ความสัมพันธ์
- ถูก embed ใน [[WarehouseListItemStruct]] (field `Locations`)
- `Code` อ้างอิงรหัสที่เก็บในคลังสินค้า
