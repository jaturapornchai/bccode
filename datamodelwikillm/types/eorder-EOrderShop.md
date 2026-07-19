---
source: backend/internal/product/eorder/models/eorder_shop.go
tags: [datamodel, general-type]
---

# EOrderShop

โครงสร้าง `EOrderShop` จากโมดูล mainapi `eorder` มีฟิลด์ตามซอร์ส `eorder_shop.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| Name1 | string | - | name1 | ค่าของ Name1 ตามฟิลด์ `name1` ในซอร์ส |
| ProfilePicture | string | - | profilepicture | ค่าของ ProfilePicture ตามฟิลด์ `profilepicture` ในซอร์ส |
| TotalTable | int | - | totaltable | ค่าของ TotalTable ตามฟิลด์ `totaltable` ในซอร์ส |
| OrderStation | [[eorder-EOrderShopOrderStation\|EOrderShopOrderStation]] | - | orderstation,omitempty | ค่าของ OrderStation ตามฟิลด์ `orderstation` ในซอร์ส |
| Kitchens | [][[Kitchen\|kitchen_models.Kitchen]] | kitchens | kitchens | ค่าของ Kitchens ตามฟิลด์ `kitchens` ในซอร์ส |
| Notify | [][[NotifyInfo\|notify_models.NotifyInfo]] | - | notify | ค่าของ Notify ตามฟิลด์ `notify` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[eorder-EOrderShopOrderStation|EOrderShopOrderStation]], [[Kitchen]], [[NotifyInfo]]
