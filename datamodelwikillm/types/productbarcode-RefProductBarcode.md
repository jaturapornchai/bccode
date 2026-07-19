---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# RefProductBarcode

โครงสร้าง `RefProductBarcode` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GuidFixed | string | guidfixed | guidfixed | GUID ถาวรของรายการ |
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| ItemUnitCode | string | itemunitcode | itemunitcode | รหัสหน่วยนับสินค้า |
| ItemUnitNames | *[][[NameX\|models.NameX]] | itemunitnames | itemunitnames | ค่าของ ItemUnitNames ตามฟิลด์ `itemunitnames` ในซอร์ส |
| Barcode | string | barcode | barcode | บาร์โค้ดสินค้า |
| Condition | bool | condition | condition | ค่าของ Condition ตามฟิลด์ `condition` ในซอร์ส |
| DivideValue | float64 | dividevalue | dividevalue | ค่าของ DivideValue ตามฟิลด์ `dividevalue` ในซอร์ส |
| StandValue | float64 | standvalue | standvalue | ค่าของ StandValue ตามฟิลด์ `standvalue` ในซอร์ส |
| Qty | float64 | qty | qty | จำนวน |
| SellerSKU | string | sellersku | sellersku | ค่าของ SellerSKU ตามฟิลด์ `sellersku` ในซอร์ส |
| SkuPackageWeight | float64 | skupackageweight | skupackageweight | ค่าของ SkuPackageWeight ตามฟิลด์ `skupackageweight` ในซอร์ส |
| SkuPackageLength | float64 | skupackagelength | skupackagelength | ค่าของ SkuPackageLength ตามฟิลด์ `skupackagelength` ในซอร์ส |
| SkuPackageWidth | float64 | skupackagewidth | skupackagewidth | ค่าของ SkuPackageWidth ตามฟิลด์ `skupackagewidth` ในซอร์ส |
| SkuPackageHeight | float64 | skupackageheight | skupackageheight | ค่าของ SkuPackageHeight ตามฟิลด์ `skupackageheight` ในซอร์ส |
| MarketplaceSKUMappings | *[][[productbarcode-MarketplaceSKUMap\|MarketplaceSKUMap]] | marketplaceskumappings | marketplaceskumappings | ค่าของ MarketplaceSKUMappings ตามฟิลด์ `marketplaceskumappings` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[NameX]], [[productbarcode-MarketplaceSKUMap|MarketplaceSKUMap]]
