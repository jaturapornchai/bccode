---
source: backend/internal/product/bom/models/bom.go
tags: [datamodel, general-type]
---

# BOMProductBarcode

โครงสร้าง `BOMProductBarcode` จากโมดูล mainapi `bom` มีฟิลด์ตามซอร์ส `bom.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| BarcodeGuidFixed | string | guidfixed | guidfixed | ค่าของ BarcodeGuidFixed ตามฟิลด์ `guidfixed` ในซอร์ส |
| Level | int | level | level | ค่าของ Level ตามฟิลด์ `level` ในซอร์ส |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| ItemUnitCode | string | itemunitcode | itemunitcode | รหัสหน่วยนับสินค้า |
| ItemUnitNames | *[][[NameX\|models.NameX]] | itemunitnames | itemunitnames | ค่าของ ItemUnitNames ตามฟิลด์ `itemunitnames` ในซอร์ส |
| Barcode | string | barcode | barcode | บาร์โค้ดสินค้า |
| RefType | string | reftype,omitempty | reftype | ค่าของ RefType ตามฟิลด์ `reftype` ในซอร์ส |
| Condition | bool | condition | condition | ค่าของ Condition ตามฟิลด์ `condition` ในซอร์ส |
| DivideValue | float64 | dividevalue | dividevalue | ค่าของ DivideValue ตามฟิลด์ `dividevalue` ในซอร์ส |
| StandValue | float64 | standvalue | standvalue | ค่าของ StandValue ตามฟิลด์ `standvalue` ในซอร์ส |
| Qty | float64 | qty | qty | จำนวน |
| YieldPercent | float64 | yieldpercent,omitempty | yieldpercent | ค่าของ YieldPercent ตามฟิลด์ `yieldpercent` ในซอร์ส |
| AverageCost | float64 | averagecost,omitempty | averagecost | ค่าของ AverageCost ตามฟิลด์ `averagecost` ในซอร์ส |
| Price | float64 | price,omitempty | price | ราคา |
| MaterialType | int8 | materialtype,omitempty | materialtype | ค่าของ MaterialType ตามฟิลด์ `materialtype` ในซอร์ส |
| SubRecipeOutputQty | float64 | subrecipeoutputqty,omitempty | subrecipeoutputqty | ค่าของ SubRecipeOutputQty ตามฟิลด์ `subrecipeoutputqty` ในซอร์ส |
| CostMode | string | costmode | costmode | ค่าของ CostMode ตามฟิลด์ `costmode` ในซอร์ส |
| StandardCost | float64 | standardcost | standardcost | ค่าของ StandardCost ตามฟิลด์ `standardcost` ในซอร์ส |
| LaborCost | float64 | laborcost | laborcost | ค่าของ LaborCost ตามฟิลด์ `laborcost` ในซอร์ส |
| OverheadCost | float64 | overheadcost | overheadcost | ค่าของ OverheadCost ตามฟิลด์ `overheadcost` ในซอร์ส |
| ScrapPercent | float64 | scrappercent | scrappercent | ค่าของ ScrapPercent ตามฟิลด์ `scrappercent` ในซอร์ส |
| FinishedGoodBarcode | string | finishedgoodbarcode | finishedgoodbarcode | ค่าของ FinishedGoodBarcode ตามฟิลด์ `finishedgoodbarcode` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[NameX]]
