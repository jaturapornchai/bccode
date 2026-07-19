---
source: backend/internal/product/bom/models/bom.go
tags: [datamodel, general-type]
---

# ProductBarcodeBOMSaveRequest

โครงสร้าง `ProductBarcodeBOMSaveRequest` จากโมดูล mainapi `bom` มีฟิลด์ตามซอร์ส `bom.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GuidFixed | string | - | guidfixed | GUID ถาวรของรายการ |
| Barcode | string | - | barcode | บาร์โค้ดสินค้า |
| Names | *[][[NameX\|models.NameX]] | - | names | รายการชื่อหลายภาษา |
| ItemUnitCode | string | - | itemunitcode | รหัสหน่วยนับสินค้า |
| ItemUnitNames | *[][[NameX\|models.NameX]] | - | itemunitnames | ค่าของ ItemUnitNames ตามฟิลด์ `itemunitnames` ในซอร์ส |
| Price | float64 | - | price | ราคา |
| OutputQty | float64 | - | outputqty | ค่าของ OutputQty ตามฟิลด์ `outputqty` ในซอร์ส |
| BOM | [][[bom-ProductBarcodeBOMView\|ProductBarcodeBOMView]] | - | bom | ค่าของ BOM ตามฟิลด์ `bom` ในซอร์ส |
| BOMs | [][[bom-ProductBarcodeBOMVersion\|ProductBarcodeBOMVersion]] | - | boms | ค่าของ BOMs ตามฟิลด์ `boms` ในซอร์ส |
| CostMode | string | - | costmode | ค่าของ CostMode ตามฟิลด์ `costmode` ในซอร์ส |
| StandardCost | float64 | - | standardcost | ค่าของ StandardCost ตามฟิลด์ `standardcost` ในซอร์ส |
| LaborCost | float64 | - | laborcost | ค่าของ LaborCost ตามฟิลด์ `laborcost` ในซอร์ส |
| OverheadCost | float64 | - | overheadcost | ค่าของ OverheadCost ตามฟิลด์ `overheadcost` ในซอร์ส |
| ScrapPercent | float64 | - | scrappercent | ค่าของ ScrapPercent ตามฟิลด์ `scrappercent` ในซอร์ส |
| FinishedGoodBarcode | string | - | finishedgoodbarcode | ค่าของ FinishedGoodBarcode ตามฟิลด์ `finishedgoodbarcode` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[NameX]], [[bom-ProductBarcodeBOMView|ProductBarcodeBOMView]], [[bom-ProductBarcodeBOMVersion|ProductBarcodeBOMVersion]]
