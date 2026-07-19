---
source: backend/internal/product/productbarcode/models/product_choice.go
tags: [datamodel, general-type]
---

# ProductChoice

โครงสร้าง `ProductChoice` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_choice.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GUID | string | guid | guid | GUID ของรายการ |
| RefBarcode | string | refbarcode | refbarcode | ค่าของ RefBarcode ตามฟิลด์ `refbarcode` ในซอร์ส |
| RefProductCode | string | refproductcode | refproductcode | ค่าของ RefProductCode ตามฟิลด์ `refproductcode` ในซอร์ส |
| RefBarcodeNames | *[]models.NameX | refbarcodenames | refbarcodenames | ค่าของ RefBarcodeNames ตามฟิลด์ `refbarcodenames` ในซอร์ส |
| Names | *[]models.NameX | names | names | รายการชื่อหลายภาษา |
| RefUnitCode | string | refunitcode | refunitcode | ค่าของ RefUnitCode ตามฟิลด์ `refunitcode` ในซอร์ส |
| RefUnitNames | *[]models.NameX | refunitnames | refunitnames | ค่าของ RefUnitNames ตามฟิลด์ `refunitnames` ในซอร์ส |
| Price | *string | price | price | ราคา |
| Qty | float64 | qty | qty | จำนวน |
| ImageURI | string | imageuri | imageuri | URI ของรูปภาพ |
| IsStock | bool | isstock | isstock | ค่าของ IsStock ตามฟิลด์ `isstock` ในซอร์ส |
| IsDefault | bool | isdefault | isdefault | ค่าของ IsDefault ตามฟิลด์ `isdefault` ในซอร์ส |
| VatCal | int8 | vatcal | vatcal | ค่าของ VatCal ตามฟิลด์ `vatcal` ในซอร์ส |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
