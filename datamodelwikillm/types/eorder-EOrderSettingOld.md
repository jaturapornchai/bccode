---
source: backend/internal/product/eorder/models/eorder_shop.go
tags: [datamodel, general-type]
---

# EOrderSettingOld

โครงสร้าง `EOrderSettingOld` จากโมดูล mainapi `eorder` มีฟิลด์ตามซอร์ส `eorder_shop.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| OrderSetting | [[OrderSetting\|order_models.OrderSetting]] | - | - | โครงสร้างฝัง |
| Branch | [[Branch\|branch_models.Branch]] | - | branch | ค่าของ Branch ตามฟิลด์ `branch` ในซอร์ส |
| Media | [[Media\|models.Media]] | - | media | ค่าของ Media ตามฟิลด์ `media` ในซอร์ส |
| SaleChannels | [][[SaleChannel\|salechannel_models.SaleChannel]] | - | salechannels | ค่าของ SaleChannels ตามฟิลด์ `salechannels` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[OrderSetting]], [[Branch]], [[Media]], [[SaleChannel]]
