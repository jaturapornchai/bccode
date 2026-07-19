---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# ProductRestaurant

โครงสร้าง `ProductRestaurant` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| IsForRestaurant | bool | isforrestaurant | isforrestaurant | ค่าของ IsForRestaurant ตามฟิลด์ `isforrestaurant` ในซอร์ส |
| IsForTakeAway | bool | isfortakeaway | isfortakeaway | ค่าของ IsForTakeAway ตามฟิลด์ `isfortakeaway` ในซอร์ส |
| IsForDelivery | bool | isfordelivery | isfordelivery | ค่าของ IsForDelivery ตามฟิลด์ `isfordelivery` ในซอร์ส |
| IsForCustomer | bool | isforcustomer | isforcustomer | ค่าของ IsForCustomer ตามฟิลด์ `isforcustomer` ในซอร์ส |
| IsForCustomerPreOrder | bool | isforcustomerpreorder | isforcustomerpreorder | ค่าของ IsForCustomerPreOrder ตามฟิลด์ `isforcustomerpreorder` ในซอร์ส |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
