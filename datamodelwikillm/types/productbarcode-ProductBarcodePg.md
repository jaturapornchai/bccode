---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# ProductBarcodePg

โครงสร้าง `ProductBarcodePg` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| PartitionIdentity | [[PartitionIdentity\|models.PartitionIdentity]] | - | - | โครงสร้างฝัง |
| Barcode | string | - | barcode | บาร์โค้ดสินค้า |
| Names | JSONB | - | names | รายการชื่อหลายภาษา |
| UnitCode | string | - | itemunitcode | รหัสหน่วยนับ |
| UnitNames | JSONB | - | itemunitnames | ค่าของ UnitNames ตามฟิลด์ `itemunitnames` ในซอร์ส |
| BalanceQty | float64 | - | balanceqty | ค่าของ BalanceQty ตามฟิลด์ `balanceqty` ในซอร์ส |
| MainBarcodeRef | string | - | mainbarcoderef | ค่าของ MainBarcodeRef ตามฟิลด์ `mainbarcoderef` ในซอร์ส |
| StandValue | float64 | - | standvalue | ค่าของ StandValue ตามฟิลด์ `standvalue` ในซอร์ส |
| DivideValue | float64 | - | dividevalue | ค่าของ DivideValue ตามฟิลด์ `dividevalue` ในซอร์ส |
| BalanceAmount | float64 | - | balanceamount | ค่าของ BalanceAmount ตามฟิลด์ `balanceamount` ในซอร์ส |
| AverageCost | float64 | - | averagecost | ค่าของ AverageCost ตามฟิลด์ `averagecost` ในซอร์ส |
| ItemCode | string | - | itemcode | รหัสสินค้า |
| ItemType | int8 | - | itemtype | ค่าของ ItemType ตามฟิลด์ `itemtype` ในซอร์ส |
| MaterialType | int8 | - | materialtype | ค่าของ MaterialType ตามฟิลด์ `materialtype` ในซอร์ส |
| BrandCode | string | - | brandcode | ค่าของ BrandCode ตามฟิลด์ `brandcode` ในซอร์ส |
| BrandName | JSONB | - | brandnames | ค่าของ BrandName ตามฟิลด์ `brandnames` ในซอร์ส |
| CategoryCode | string | - | categorycode | ค่าของ CategoryCode ตามฟิลด์ `categorycode` ในซอร์ส |
| CategoryName | JSONB | - | categorynames | ค่าของ CategoryName ตามฟิลด์ `categorynames` ในซอร์ส |
| ClassCode | string | - | classcode | ค่าของ ClassCode ตามฟิลด์ `classcode` ในซอร์ส |
| ClassNames | JSONB | - | classnames | ค่าของ ClassNames ตามฟิลด์ `classnames` ในซอร์ส |
| DesignCode | string | - | designcode | ค่าของ DesignCode ตามฟิลด์ `designcode` ในซอร์ส |
| DesignNames | JSONB | - | designnames | ค่าของ DesignNames ตามฟิลด์ `designnames` ในซอร์ส |
| GradeCode | string | - | gradecode | ค่าของ GradeCode ตามฟิลด์ `gradecode` ในซอร์ส |
| GradeNames | JSONB | - | gradenames | ค่าของ GradeNames ตามฟิลด์ `gradenames` ในซอร์ส |
| GroupCode | string | - | groupcode | ค่าของ GroupCode ตามฟิลด์ `groupcode` ในซอร์ส |
| GroupNames | JSONB | - | groupnames | ค่าของ GroupNames ตามฟิลด์ `groupnames` ในซอร์ส |
| GroupSubOneCode | string | - | groupsubonecode | ค่าของ GroupSubOneCode ตามฟิลด์ `groupsubonecode` ในซอร์ส |
| GroupSubOneNames | JSONB | - | groupsubonenames | ค่าของ GroupSubOneNames ตามฟิลด์ `groupsubonenames` ในซอร์ส |
| GroupSubTwoCode | string | - | groupsubtwocode | ค่าของ GroupSubTwoCode ตามฟิลด์ `groupsubtwocode` ในซอร์ส |
| GroupSubTwoNames | JSONB | - | groupsubtwonames | ค่าของ GroupSubTwoNames ตามฟิลด์ `groupsubtwonames` ในซอร์ส |
| ModelCode | string | - | modelcode | ค่าของ ModelCode ตามฟิลด์ `modelcode` ในซอร์ส |
| ModelNames | JSONB | - | modelnames | ค่าของ ModelNames ตามฟิลด์ `modelnames` ในซอร์ส |
| PatternCode | string | - | patterncode | ค่าของ PatternCode ตามฟิลด์ `patterncode` ในซอร์ส |
| PatternNames | JSONB | - | patternnames | ค่าของ PatternNames ตามฟิลด์ `patternnames` ในซอร์ส |
| BOM | BOMProductBarcodePg | - | bom | ค่าของ BOM ตามฟิลด์ `bom` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[PartitionIdentity]]
