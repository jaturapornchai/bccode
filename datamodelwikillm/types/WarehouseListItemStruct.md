---
source: process-model.go
tags: [datamodel, general-type]
---

# WarehouseListItemStruct

รายการคลังสินค้าหนึ่งรายการใน payload (รหัส ชื่อ สถานะถูกเลือก) พร้อมรายการที่เก็บย่อย (มีเฉพาะ json tag)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Code | string | - | code | รหัสคลังสินค้า |
| Title | string | - | title | ชื่อคลังสินค้า |
| IsSelected | bool | - | isselected | ถูกเลือกหรือไม่ |
| Locations | [][[LocationItemStruct]] | - | locations,omitempty | รายการที่เก็บในคลัง |

## ความสัมพันธ์
- ถูก embed ใน [[PayLoadCommandStruct]] (field `WarehouseList`)
- Embedded: [[LocationItemStruct]] (slice ใน `Locations`)
- `Code` อ้างอิงรหัสคลังสินค้า
