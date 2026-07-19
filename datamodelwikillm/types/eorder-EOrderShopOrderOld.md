---
source: backend/internal/product/eorder/models/eorder_shop.go
tags: [datamodel, general-type]
---

# EOrderShopOrderOld

โครงสร้าง `EOrderShopOrderOld` จากโมดูล mainapi `eorder` มีฟิลด์ตามซอร์ส `eorder_shop.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| EOrderSettingOld | [[eorder-EOrderSettingOld\|EOrderSettingOld]] | - | - | โครงสร้างฝัง |
| DeviceInfo | [[OrderDevice\|device_models.OrderDevice]] | deviceinfo | deviceinfo | ค่าของ DeviceInfo ตามฟิลด์ `deviceinfo` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[eorder-EOrderSettingOld|EOrderSettingOld]], [[OrderDevice]]
