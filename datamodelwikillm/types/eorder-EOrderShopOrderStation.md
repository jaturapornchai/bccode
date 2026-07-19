---
source: backend/internal/product/eorder/models/eorder_shop.go
tags: [datamodel, general-type]
---

# EOrderShopOrderStation

โครงสร้าง `EOrderShopOrderStation` จากโมดูล mainapi `eorder` มีฟิลด์ตามซอร์ส `eorder_shop.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| OrderDevice | [[OrderDevice\|device_models.OrderDevice]] | - | - | โครงสร้างฝัง |
| Setting | [[eorder-EOrderSetting\|EOrderSetting]] | - | ordersetting | ค่าของ Setting ตามฟิลด์ `ordersetting` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[OrderDevice]], [[eorder-EOrderSetting|EOrderSetting]]
