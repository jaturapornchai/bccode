---
source: backend/internal/product/product/models/product.go
tags: [datamodel, general-type]
---

# Product

โครงสร้าง `Product` จากโมดูล mainapi `product` มีฟิลด์ตามซอร์ส `product.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| PartitionIdentity | [[PartitionIdentity\|models.PartitionIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| Code | string | code | code | รหัสรายการ |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| GroupCode | string | groupcode | groupcode | ค่าของ GroupCode ตามฟิลด์ `groupcode` ในซอร์ส |
| GroupNames | *[][[NameX\|models.NameX]] | groupnames | groupnames | ค่าของ GroupNames ตามฟิลด์ `groupnames` ในซอร์ส |
| ManufacturerGUID | string | manufacturerguid | manufacturerguid | ค่าของ ManufacturerGUID ตามฟิลด์ `manufacturerguid` ในซอร์ส |
| ManufacturerCode | string | manufacturercode | manufacturercode | ค่าของ ManufacturerCode ตามฟิลด์ `manufacturercode` ในซอร์ส |
| ManufacturerNames | *[][[NameX\|models.NameX]] | manufacturernames | manufacturernames | ค่าของ ManufacturerNames ตามฟิลด์ `manufacturernames` ในซอร์ส |
| Dimensions | [][[product-ProductDimension\|ProductDimension]] | dimensions | dimensions | ค่าของ Dimensions ตามฟิลด์ `dimensions` ในซอร์ส |
| VatType | int8 | vattype | vattype | ค่าของ VatType ตามฟิลด์ `vattype` ในซอร์ส |
| Barcodes | [][[product-Barcodes\|Barcodes]] | - | barcodes,omitempty | ค่าของ Barcodes ตามฟิลด์ `barcodes` ในซอร์ส |
| ItemType | int8 | itemtype | itemtype | ค่าของ ItemType ตามฟิลด์ `itemtype` ในซอร์ส |
| UnitGuid | string | unitguid | unitguid | ค่าของ UnitGuid ตามฟิลด์ `unitguid` ในซอร์ส |
| GroupsuboneGuid | string | groupsuboneguid | groupsuboneguid | ค่าของ GroupsuboneGuid ตามฟิลด์ `groupsuboneguid` ในซอร์ส |
| GroupsuboneCode | string | groupsubonecode | groupsubonecode | ค่าของ GroupsuboneCode ตามฟิลด์ `groupsubonecode` ในซอร์ส |
| GroupsuboneNames | *[][[NameX\|models.NameX]] | groupsubonenames | groupsubonenames | ค่าของ GroupsuboneNames ตามฟิลด์ `groupsubonenames` ในซอร์ส |
| GroupsubtwoGuid | string | groupsubtwoguid | groupsubtwoguid | ค่าของ GroupsubtwoGuid ตามฟิลด์ `groupsubtwoguid` ในซอร์ส |
| GroupsubtwoCode | string | groupsubtwocode | groupsubtwocode | ค่าของ GroupsubtwoCode ตามฟิลด์ `groupsubtwocode` ในซอร์ส |
| GroupsubtwoNames | *[][[NameX\|models.NameX]] | groupsubtwonames | groupsubtwonames | ค่าของ GroupsubtwoNames ตามฟิลด์ `groupsubtwonames` ในซอร์ส |
| BrandGuid | string | brandguid | brandguid | ค่าของ BrandGuid ตามฟิลด์ `brandguid` ในซอร์ส |
| BrandCode | string | brandcode | brandcode | ค่าของ BrandCode ตามฟิลด์ `brandcode` ในซอร์ส |
| BrandNames | *[][[NameX\|models.NameX]] | brandnames | brandnames | ค่าของ BrandNames ตามฟิลด์ `brandnames` ในซอร์ส |
| DesignGuid | string | designguid | designguid | ค่าของ DesignGuid ตามฟิลด์ `designguid` ในซอร์ส |
| DesignCode | string | designcode | designcode | ค่าของ DesignCode ตามฟิลด์ `designcode` ในซอร์ส |
| DesignNames | *[][[NameX\|models.NameX]] | designnames | designnames | ค่าของ DesignNames ตามฟิลด์ `designnames` ในซอร์ส |
| ModelGuid | string | modelguid | modelguid | ค่าของ ModelGuid ตามฟิลด์ `modelguid` ในซอร์ส |
| ModelCode | string | modelcode | modelcode | ค่าของ ModelCode ตามฟิลด์ `modelcode` ในซอร์ส |
| ModelNames | *[][[NameX\|models.NameX]] | modelnames | modelnames | ค่าของ ModelNames ตามฟิลด์ `modelnames` ในซอร์ส |
| PatternGuid | string | patternguid | patternguid | ค่าของ PatternGuid ตามฟิลด์ `patternguid` ในซอร์ส |
| PatternCode | string | patterncode | patterncode | ค่าของ PatternCode ตามฟิลด์ `patterncode` ในซอร์ส |
| PatternNames | *[][[NameX\|models.NameX]] | patternnames | patternnames | ค่าของ PatternNames ตามฟิลด์ `patternnames` ในซอร์ส |
| GradeGuid | string | gradeguid | gradeguid | ค่าของ GradeGuid ตามฟิลด์ `gradeguid` ในซอร์ส |
| GradeCode | string | gradecode | gradecode | ค่าของ GradeCode ตามฟิลด์ `gradecode` ในซอร์ส |
| GradeNames | *[][[NameX\|models.NameX]] | gradenames | gradenames | ค่าของ GradeNames ตามฟิลด์ `gradenames` ในซอร์ส |
| CategoryGuid | string | categoryguid | categoryguid | ค่าของ CategoryGuid ตามฟิลด์ `categoryguid` ในซอร์ส |
| CategoryCode | string | categorycode | categorycode | ค่าของ CategoryCode ตามฟิลด์ `categorycode` ในซอร์ส |
| CategoryNames | *[][[NameX\|models.NameX]] | categorynames | categorynames | ค่าของ CategoryNames ตามฟิลด์ `categorynames` ในซอร์ส |
| ClassGuid | string | classguid | classguid | ค่าของ ClassGuid ตามฟิลด์ `classguid` ในซอร์ส |
| ClassCode | string | classcode | classcode | ค่าของ ClassCode ตามฟิลด์ `classcode` ในซอร์ส |
| ClassNames | *[][[NameX\|models.NameX]] | classnames | classnames | ค่าของ ClassNames ตามฟิลด์ `classnames` ในซอร์ส |
| MaterialType | int8 | materialtype | materialtype | ค่าของ MaterialType ตามฟิลด์ `materialtype` ในซอร์ส |
| TaxType | int8 | taxtype | taxtype | ค่าของ TaxType ตามฟิลด์ `taxtype` ในซอร์ส |
| Manufacturers | *[][[product-ProductManufacturer\|ProductManufacturer]] | manufacturers | manufacturers | ค่าของ Manufacturers ตามฟิลด์ `manufacturers` ในซอร์ส |
| Suppliers | *[][[product-ProductSupplier\|ProductSupplier]] | suppliers | suppliers | ค่าของ Suppliers ตามฟิลด์ `suppliers` ในซอร์ส |
| ImageURI | string | imageuri | imageuri | URI ของรูปภาพ |
| Images | *[][[product-ProductImage\|ProductImage]] | images | images | ค่าของ Images ตามฟิลด์ `images` ในซอร์ส |
| UseImageOrColor | bool | useimageorcolor | useimageorcolor | ค่าของ UseImageOrColor ตามฟิลด์ `useimageorcolor` ในซอร์ส |
| ColorSelect | string | colorselect | colorselect | ค่าของ ColorSelect ตามฟิลด์ `colorselect` ในซอร์ส |
| ColorSelectHex | string | colorselecthex | colorselecthex | ค่าของ ColorSelectHex ตามฟิลด์ `colorselecthex` ในซอร์ส |
| IsSumPoint | bool | issumpoint | issumpoint | ค่าของ IsSumPoint ตามฟิลด์ `issumpoint` ในซอร์ส |
| IsALaCarte | bool | isalacarte | isalacarte | ค่าของ IsALaCarte ตามฟิลด์ `isalacarte` ในซอร์ส |
| IsSplitUnitPrint | bool | issplitunitprint | issplitunitprint | ค่าของ IsSplitUnitPrint ตามฟิลด์ `issplitunitprint` ในซอร์ส |
| IsOnlyStaff | bool | isonlystaff | isonlystaff | ค่าของ IsOnlyStaff ตามฟิลด์ `isonlystaff` ในซอร์ส |
| FoodType | int | foodtype | foodtype | ค่าของ FoodType ตามฟิลด์ `foodtype` ในซอร์ส |
| IsStockForRestaurant | bool | isstockforrestaurant | isstockforrestaurant | ค่าของ IsStockForRestaurant ตามฟิลด์ `isstockforrestaurant` ในซอร์ส |
| Restaurant | [[product-ProductRestaurant\|ProductRestaurant]] | restaurant | restaurant | ค่าของ Restaurant ตามฟิลด์ `restaurant` ในซอร์ส |
| OrderTypes | *[][[product-ProductOrderType\|ProductOrderType]] | ordertypes | ordertypes | ค่าของ OrderTypes ตามฟิลด์ `ordertypes` ในซอร์ส |
| Options | *[][[product-ProductOption\|ProductOption]] | options | options | ค่าของ Options ตามฟิลด์ `options` ในซอร์ส |
| IsAlert | bool | isalert | isalert | ค่าของ IsAlert ตามฟิลด์ `isalert` ในซอร์ส |
| AlertDescription | string | alertdescription | alertdescription | ค่าของ AlertDescription ตามฟิลด์ `alertdescription` ในซอร์ส |
| Description | string | description | description | คำอธิบาย |
| TimeForSales | *[][[product-ProductTimeForSale\|ProductTimeForSale]] | timeforsales | timeforsales | ค่าของ TimeForSales ตามฟิลด์ `timeforsales` ในซอร์ส |
| BusinessTypes | *[][[product-ProductBarcodeBusinessType\|ProductBarcodeBusinessType]] | businesstypes | businesstypes | ค่าของ BusinessTypes ตามฟิลด์ `businesstypes` ในซอร์ส |
| IgnoreBranches | *[][[product-ProductBarcodeBranch\|ProductBarcodeBranch]] | ignorebranches | ignorebranches | ค่าของ IgnoreBranches ตามฟิลด์ `ignorebranches` ในซอร์ส |
| Condition | bool | condition | condition | ค่าของ Condition ตามฟิลด์ `condition` ในซอร์ส |
| DivideValue | float64 | dividevalue | dividevalue | ค่าของ DivideValue ตามฟิลด์ `dividevalue` ในซอร์ส |
| StandValue | float64 | standvalue | standvalue | ค่าของ StandValue ตามฟิลด์ `standvalue` ในซอร์ส |
| IsUseSubBarcodes | bool | isusesubbarcodes | isusesubbarcodes | ค่าของ IsUseSubBarcodes ตามฟิลด์ `isusesubbarcodes` ในซอร์ส |
| RefBarcodes | *[][[product-RefProductBarcode\|RefProductBarcode]] | refbarcodes | refbarcodes | ค่าของ RefBarcodes ตามฟิลด์ `refbarcodes` ในซอร์ส |
| BOM | *[][[product-BOMProductBarcode\|BOMProductBarcode]] | bom | bom | ค่าของ BOM ตามฟิลด์ `bom` ในซอร์ส |
| PackageWeight | float64 | packageweight | packageweight | ค่าของ PackageWeight ตามฟิลด์ `packageweight` ในซอร์ส |
| PackageLength | float64 | packagelength | packagelength | ค่าของ PackageLength ตามฟิลด์ `packagelength` ในซอร์ส |
| PackageWidth | float64 | packagewidth | packagewidth | ค่าของ PackageWidth ตามฟิลด์ `packagewidth` ในซอร์ส |
| PackageHeight | float64 | packageheight | packageheight | ค่าของ PackageHeight ตามฟิลด์ `packageheight` ในซอร์ส |
| MarketplaceProducts | *[][[product-MarketplaceProductMap\|MarketplaceProductMap]] | marketplaceproducts | marketplaceproducts | ค่าของ MarketplaceProducts ตามฟิลด์ `marketplaceproducts` ในซอร์ส |
| OrderPoint | float64 | orderpoint | orderpoint | ค่าของ OrderPoint ตามฟิลด์ `orderpoint` ในซอร์ส |
| MinPoint | float64 | minpoint | minpoint | ค่าของ MinPoint ตามฟิลด์ `minpoint` ในซอร์ส |
| MaxPoint | float64 | maxpoint | maxpoint | ค่าของ MaxPoint ตามฟิลด์ `maxpoint` ในซอร์ส |
| Qty | float64 | qty | qty | จำนวน |
| StockBarcode | string | stockbarcode | stockbarcode | ค่าของ StockBarcode ตามฟิลด์ `stockbarcode` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[PartitionIdentity]], [[NameX]], [[product-ProductDimension|ProductDimension]], [[product-Barcodes|Barcodes]], [[product-ProductManufacturer|ProductManufacturer]], [[product-ProductSupplier|ProductSupplier]], [[product-ProductImage|ProductImage]], [[product-ProductRestaurant|ProductRestaurant]], [[product-ProductOrderType|ProductOrderType]], [[product-ProductOption|ProductOption]], [[product-ProductTimeForSale|ProductTimeForSale]], [[product-ProductBarcodeBusinessType|ProductBarcodeBusinessType]], [[product-ProductBarcodeBranch|ProductBarcodeBranch]], [[product-RefProductBarcode|RefProductBarcode]], [[product-BOMProductBarcode|BOMProductBarcode]], [[product-MarketplaceProductMap|MarketplaceProductMap]]
